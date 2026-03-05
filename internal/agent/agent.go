package agent

import (
	"context"
	"fmt"

	"ai_agent/internal/model"
	"ai_agent/internal/repository"

	"google.golang.org/genai"
)

// FunctionCallFn is the callback executed when Gemini invokes a function.
type FunctionCallFn func(ctx context.Context, args map[string]any) (map[string]any, error)

// FunctionDeclaration describes a skill the agent can invoke.
type FunctionDeclaration struct {
	Name             string
	Description      string
	ParametersSchema any
	ResponseSchema   any
	FunctionCall     FunctionCallFn
}

// Agent wraps the Gemini client and manages function calling and session persistence.
type Agent struct {
	client            *genai.Client
	model             string
	systemInstruction string
	functionsMap      map[string]*FunctionDeclaration
	sessionRepo       repository.SessionRepository
}

func New(client *genai.Client, model, systemInstruction string) *Agent {
	return &Agent{
		client:            client,
		model:             model,
		systemInstruction: systemInstruction,
		functionsMap:      make(map[string]*FunctionDeclaration),
	}
}

func NewWithRepo(client *genai.Client, model, systemInstruction string, repo repository.SessionRepository) *Agent {
	a := New(client, model, systemInstruction)
	a.sessionRepo = repo
	return a
}

// AddFunctionCall registers a skill with the agent.
func (a *Agent) AddFunctionCall(fn *FunctionDeclaration) error {
	if fn.Name == "" {
		return fmt.Errorf("function declaration must have a name")
	}
	if fn.FunctionCall == nil {
		return fmt.Errorf("function declaration %q must have a FunctionCall callback", fn.Name)
	}
	a.functionsMap[fn.Name] = fn
	return nil
}

func (a *Agent) getTools() []*genai.Tool {
	if len(a.functionsMap) == 0 {
		return nil
	}
	declarations := make([]*genai.FunctionDeclaration, 0, len(a.functionsMap))
	for _, fn := range a.functionsMap {
		declarations = append(declarations, &genai.FunctionDeclaration{
			Name:                 fn.Name,
			Description:          fn.Description,
			ParametersJsonSchema: fn.ParametersSchema,
			ResponseJsonSchema:   fn.ResponseSchema,
		})
	}
	return []*genai.Tool{{FunctionDeclarations: declarations}}
}

// TokenUsage holds the token consumption data for a single agent interaction.
type TokenUsage struct {
	PromptTokens   int32 `json:"prompt_tokens"`
	ResponseTokens int32 `json:"response_tokens"`
	TotalTokens    int32 `json:"total_tokens"`
}

func (u TokenUsage) add(other TokenUsage) TokenUsage {
	return TokenUsage{
		PromptTokens:   u.PromptTokens + other.PromptTokens,
		ResponseTokens: u.ResponseTokens + other.ResponseTokens,
		TotalTokens:    u.TotalTokens + other.TotalTokens,
	}
}

func usageFromResp(resp *genai.GenerateContentResponse) TokenUsage {
	if resp.UsageMetadata == nil {
		return TokenUsage{}
	}
	return TokenUsage{
		PromptTokens:   resp.UsageMetadata.PromptTokenCount,
		ResponseTokens: resp.UsageMetadata.CandidatesTokenCount,
		TotalTokens:    resp.UsageMetadata.TotalTokenCount,
	}
}

// Send processes the user prompt, handles function calling loop, and persists the session.
func (a *Agent) Send(ctx context.Context, sessionID, prompt string) (string, TokenUsage, error) {
	history, err := a.loadHistory(ctx, sessionID)
	if err != nil {
		return "", TokenUsage{}, err
	}

	config := &genai.GenerateContentConfig{
		SystemInstruction: &genai.Content{
			Parts: []*genai.Part{{Text: a.systemInstruction}},
		},
		Tools: a.getTools(),
	}

	chat, err := a.client.Chats.Create(ctx, a.model, config, history)
	if err != nil {
		return "", TokenUsage{}, fmt.Errorf("creating chat session: %w", err)
	}

	resp, err := chat.SendMessage(ctx, genai.Part{Text: prompt})
	if err != nil {
		return "", TokenUsage{}, fmt.Errorf("sending message: %w", err)
	}

	text, usage, err := a.processResponse(ctx, chat, resp)
	if err != nil {
		return "", TokenUsage{}, err
	}

	if err := a.saveHistory(ctx, sessionID, chat.History(true)); err != nil {
		return "", TokenUsage{}, err
	}

	return text, usage, nil
}

// processResponse handles function calls recursively until the model returns text.
func (a *Agent) processResponse(ctx context.Context, chat *genai.Chat, resp *genai.GenerateContentResponse) (string, TokenUsage, error) {
	usage := usageFromResp(resp)

	calls := resp.FunctionCalls()
	if len(calls) == 0 {
		return resp.Text(), usage, nil
	}

	parts := make([]*genai.Part, 0, len(calls))
	for _, call := range calls {
		fn, ok := a.functionsMap[call.Name]
		if !ok {
			return "", TokenUsage{}, fmt.Errorf("model requested unknown function %q", call.Name)
		}

		result, err := fn.FunctionCall(ctx, call.Args)
		if err != nil {
			return "", TokenUsage{}, fmt.Errorf("executing function %q: %w", call.Name, err)
		}

		parts = append(parts, &genai.Part{
			FunctionResponse: &genai.FunctionResponse{
				Name:     call.Name,
				Response: result,
			},
		})
	}

	nextResp, err := chat.Send(ctx, parts...)
	if err != nil {
		return "", TokenUsage{}, fmt.Errorf("sending function responses: %w", err)
	}

	text, nextUsage, err := a.processResponse(ctx, chat, nextResp)
	return text, usage.add(nextUsage), err
}

// GetSession returns the conversation history for a session.
func (a *Agent) GetSession(ctx context.Context, sessionID string) ([]model.Content, error) {
	if a.sessionRepo == nil {
		return nil, nil
	}
	return a.sessionRepo.Load(ctx, sessionID)
}

// ClearSession deletes the session history.
func (a *Agent) ClearSession(ctx context.Context, sessionID string) error {
	if a.sessionRepo == nil {
		return nil
	}
	return a.sessionRepo.Delete(ctx, sessionID)
}

func (a *Agent) loadHistory(ctx context.Context, sessionID string) ([]*genai.Content, error) {
	if a.sessionRepo == nil || sessionID == "" {
		return nil, nil
	}
	stored, err := a.sessionRepo.Load(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	return toGenAIContents(stored), nil
}

func (a *Agent) saveHistory(ctx context.Context, sessionID string, history []*genai.Content) error {
	if a.sessionRepo == nil || sessionID == "" {
		return nil
	}
	return a.sessionRepo.Save(ctx, sessionID, toModelContents(history))
}

// toModelContents converts genai contents to our internal model for persistence.
// genai.Content.Role is string; genai.FunctionResponse.Response is map[string]any.
func toModelContents(contents []*genai.Content) []model.Content {
	result := make([]model.Content, 0, len(contents))
	for _, c := range contents {
		mc := model.Content{Role: c.Role}
		for _, p := range c.Parts {
			mp := model.Part{}
			if p.Text != "" {
				mp.Text = p.Text
			}
			if p.FunctionCall != nil {
				mp.FunctionCall = &struct {
					Name string         `bson:"name"  json:"name"`
					Args map[string]any `bson:"args"  json:"args"`
				}{
					Name: p.FunctionCall.Name,
					Args: p.FunctionCall.Args,
				}
			}
			if p.FunctionResponse != nil {
				// FunctionResponse.Response is map[string]any in the genai SDK.
				mp.FunctionResponse = &struct {
					Name     string `bson:"name"     json:"name"`
					Response any    `bson:"response" json:"response"`
				}{
					Name:     p.FunctionResponse.Name,
					Response: p.FunctionResponse.Response,
				}
			}
			mc.Parts = append(mc.Parts, mp)
		}
		result = append(result, mc)
	}
	return result
}

// toGenAIContents converts persisted model contents back to genai format.
// model.Part.FunctionResponse.Response is any (from MongoDB), so we type-assert to map[string]any.
func toGenAIContents(contents []model.Content) []*genai.Content {
	result := make([]*genai.Content, 0, len(contents))
	for _, c := range contents {
		gc := &genai.Content{Role: c.Role}
		for _, p := range c.Parts {
			gp := &genai.Part{}
			if p.Text != "" {
				gp.Text = p.Text
			}
			if p.FunctionCall != nil {
				gp.FunctionCall = &genai.FunctionCall{
					Name: p.FunctionCall.Name,
					Args: p.FunctionCall.Args,
				}
			}
			if p.FunctionResponse != nil {
				resp, _ := p.FunctionResponse.Response.(map[string]any)
				gp.FunctionResponse = &genai.FunctionResponse{
					Name:     p.FunctionResponse.Name,
					Response: resp,
				}
			}
			gc.Parts = append(gc.Parts, gp)
		}
		result = append(result, gc)
	}
	return result
}
