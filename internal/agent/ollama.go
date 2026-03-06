package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"ai_agent/internal/model"
	"ai_agent/internal/repository"
)

// OllamaAgent implements Agent using Ollama's native /api/chat endpoint.
type OllamaAgent struct {
	baseURL     string
	model       string
	sysInstruct string
	funcsMap    map[string]*FunctionDeclaration
	sessionRepo repository.SessionRepository
	httpClient  *http.Client
}

func NewOllama(baseURL, modelName, sysInstruct string, repo repository.SessionRepository) *OllamaAgent {
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}
	baseURL = strings.TrimRight(baseURL, "/")
	return &OllamaAgent{
		baseURL:     baseURL,
		model:       modelName,
		sysInstruct: sysInstruct,
		funcsMap:    make(map[string]*FunctionDeclaration),
		sessionRepo: repo,
		httpClient:  &http.Client{},
	}
}

func (a *OllamaAgent) AddFunctionCall(fn *FunctionDeclaration) error {
	if fn.Name == "" {
		return fmt.Errorf("function declaration must have a name")
	}
	if fn.FunctionCall == nil {
		return fmt.Errorf("function declaration %q must have a FunctionCall callback", fn.Name)
	}
	a.funcsMap[fn.Name] = fn
	return nil
}

// ── Ollama API types ─────────────────────────────────────

type ollamaMessage struct {
	Role      string       `json:"role"`
	Content   string       `json:"content"`
	ToolCalls []ollamaCall `json:"tool_calls,omitempty"`
}

type ollamaCall struct {
	Function ollamaFunc `json:"function"`
}

type ollamaFunc struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"` // some models return object, others a JSON string
}

// args parses Arguments tolerating both object {"key":"val"} and JSON-encoded string formats.
func (f ollamaFunc) args() map[string]any {
	var m map[string]any
	if json.Unmarshal(f.Arguments, &m) == nil {
		return m
	}
	// fallback: some models encode arguments as a JSON string
	var s string
	if json.Unmarshal(f.Arguments, &s) == nil {
		json.Unmarshal([]byte(s), &m) //nolint:errcheck
	}
	return m
}

type ollamaTool struct {
	Type     string        `json:"type"`
	Function ollamaToolDef `json:"function"`
}

type ollamaToolDef struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Parameters  any    `json:"parameters"`
}

type ollamaChatRequest struct {
	Model    string          `json:"model"`
	Messages []ollamaMessage `json:"messages"`
	Tools    []ollamaTool    `json:"tools,omitempty"`
	Stream   bool            `json:"stream"`
}

type ollamaChatResponse struct {
	Message         ollamaMessage `json:"message"`
	Done            bool          `json:"done"`
	EvalCount       int32         `json:"eval_count"`
	PromptEvalCount int32         `json:"prompt_eval_count"`
}

// ── Send ────────────────────────────────────────────────

func (a *OllamaAgent) Send(ctx context.Context, sessionID, prompt string) (string, TokenUsage, error) {
	// Load history and build message list
	messages := []ollamaMessage{}
	if a.sysInstruct != "" {
		messages = append(messages, ollamaMessage{Role: "system", Content: a.sysInstruct})
	}
	if a.sessionRepo != nil && sessionID != "" {
		history, err := a.sessionRepo.Load(ctx, sessionID)
		if err != nil {
			return "", TokenUsage{}, fmt.Errorf("loading history: %w", err)
		}
		messages = append(messages, modelContentsToOllama(history)...)
	}
	messages = append(messages, ollamaMessage{Role: "user", Content: prompt})

	tools := a.buildTools()
	text, allMessages, usage, err := a.chat(ctx, messages, tools)
	if err != nil {
		return "", TokenUsage{}, err
	}

	if a.sessionRepo != nil && sessionID != "" {
		// Save only text turns (skip system message at index 0)
		start := 0
		if a.sysInstruct != "" {
			start = 1
		}
		if err := a.sessionRepo.Save(ctx, sessionID, ollamaToModelContents(allMessages[start:])); err != nil {
			return "", TokenUsage{}, fmt.Errorf("saving history: %w", err)
		}
	}

	return text, usage, nil
}

// chat runs the function-call loop against Ollama, accumulating token usage.
func (a *OllamaAgent) chat(ctx context.Context, messages []ollamaMessage, tools []ollamaTool) (string, []ollamaMessage, TokenUsage, error) {
	resp, err := a.callOllama(ctx, messages, tools)
	if err != nil {
		return "", nil, TokenUsage{}, err
	}
	usage := TokenUsage{
		PromptTokens:   resp.PromptEvalCount,
		ResponseTokens: resp.EvalCount,
		TotalTokens:    resp.PromptEvalCount + resp.EvalCount,
	}

	msg := resp.Message
	messages = append(messages, msg)

	if len(msg.ToolCalls) == 0 {
		return msg.Content, messages, usage, nil
	}

	// Execute tool calls and collect results
	for _, tc := range msg.ToolCalls {
		fn, ok := a.funcsMap[tc.Function.Name]
		if !ok {
			return "", nil, TokenUsage{}, fmt.Errorf("model requested unknown function %q", tc.Function.Name)
		}
		result, err := fn.FunctionCall(ctx, tc.Function.args())
		if err != nil {
			return "", nil, TokenUsage{}, fmt.Errorf("executing function %q: %w", tc.Function.Name, err)
		}
		resultJSON, _ := json.Marshal(result)
		messages = append(messages, ollamaMessage{Role: "tool", Content: string(resultJSON)})
	}

	text, messages, nextUsage, err := a.chat(ctx, messages, tools)
	return text, messages, usage.add(nextUsage), err
}

func (a *OllamaAgent) callOllama(ctx context.Context, messages []ollamaMessage, tools []ollamaTool) (*ollamaChatResponse, error) {
	body, err := json.Marshal(ollamaChatRequest{
		Model:    a.model,
		Messages: messages,
		Tools:    tools,
		Stream:   false,
	})
	if err != nil {
		return nil, fmt.Errorf("marshaling request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.baseURL+"/api/chat", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	httpResp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("calling ollama: %w", err)
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		raw, _ := io.ReadAll(httpResp.Body)
		return nil, fmt.Errorf("ollama returned %d: %s", httpResp.StatusCode, strings.TrimSpace(string(raw)))
	}

	var resp ollamaChatResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}
	return &resp, nil
}

func (a *OllamaAgent) buildTools() []ollamaTool {
	if len(a.funcsMap) == 0 {
		return nil
	}
	tools := make([]ollamaTool, 0, len(a.funcsMap))
	for _, fn := range a.funcsMap {
		tools = append(tools, ollamaTool{
			Type: "function",
			Function: ollamaToolDef{
				Name:        fn.Name,
				Description: fn.Description,
				Parameters:  fn.ParametersSchema,
			},
		})
	}
	return tools
}

// ── History conversion ───────────────────────────────────

func modelContentsToOllama(contents []model.Content) []ollamaMessage {
	var msgs []ollamaMessage
	for _, c := range contents {
		for _, p := range c.Parts {
			switch {
			case p.FunctionResponse != nil:
				// Tool result stored as FunctionResponse — convert back to "tool" role
				resultJSON, _ := json.Marshal(p.FunctionResponse.Response)
				msgs = append(msgs, ollamaMessage{Role: "tool", Content: string(resultJSON)})
			case p.FunctionCall != nil:
				argsRaw, _ := json.Marshal(p.FunctionCall.Args)
				msgs = append(msgs, ollamaMessage{
					Role: "assistant",
					ToolCalls: []ollamaCall{{
						Function: ollamaFunc{Name: p.FunctionCall.Name, Arguments: json.RawMessage(argsRaw)},
					}},
				})
			case p.Text != "":
				role := "user"
				if c.Role == "model" {
					role = "assistant"
				}
				msgs = append(msgs, ollamaMessage{Role: role, Content: p.Text})
			}
		}
	}
	return msgs
}

func ollamaToModelContents(msgs []ollamaMessage) []model.Content {
	var contents []model.Content
	for _, m := range msgs {
		switch m.Role {
		case "user":
			contents = append(contents, model.Content{
				Role:  "user",
				Parts: []model.Part{{Text: m.Content}},
			})
		case "assistant":
			if len(m.ToolCalls) > 0 {
				var parts []model.Part
				for _, tc := range m.ToolCalls {
					parts = append(parts, model.Part{
						FunctionCall: &struct {
							Name string         `bson:"name"  json:"name"`
							Args map[string]any `bson:"args"  json:"args"`
						}{Name: tc.Function.Name, Args: tc.Function.args()},
					})
				}
				contents = append(contents, model.Content{Role: "model", Parts: parts})
			} else if m.Content != "" {
				contents = append(contents, model.Content{
					Role:  "model",
					Parts: []model.Part{{Text: m.Content}},
				})
			}
		case "tool":
			// Store as FunctionResponse so modelContentsToOllama can reconstruct "tool" role correctly
			var resp any
			json.Unmarshal([]byte(m.Content), &resp) //nolint:errcheck
			contents = append(contents, model.Content{
				Role: "user",
				Parts: []model.Part{{
					FunctionResponse: &struct {
						Name     string `bson:"name"     json:"name"`
						Response any    `bson:"response" json:"response"`
					}{Name: "tool", Response: resp},
				}},
			})
		}
	}
	return contents
}

var _ Agent = (*OllamaAgent)(nil)
