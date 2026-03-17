package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"ai_agent/internal/model"
	"ai_agent/internal/repository"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

// ClaudeAgent implements Agent using Anthropic's Claude API.
type ClaudeAgent struct {
	client      *anthropic.Client
	model       string
	sysInstruct string
	funcsMap    map[string]*FunctionDeclaration
	sessionRepo repository.SessionRepository
}

func NewClaude(apiKey, modelName, sysInstruct string, repo repository.SessionRepository) *ClaudeAgent {
	if modelName == "" {
		modelName = string(anthropic.ModelClaudeSonnet4_6)
	}
	client := anthropic.NewClient(option.WithAPIKey(apiKey))
	return &ClaudeAgent{
		client:      &client,
		model:       modelName,
		sysInstruct: sysInstruct,
		funcsMap:    make(map[string]*FunctionDeclaration),
		sessionRepo: repo,
	}
}

func (a *ClaudeAgent) AddFunctionCall(fn *FunctionDeclaration) error {
	if fn.Name == "" {
		return fmt.Errorf("function declaration must have a name")
	}
	if fn.FunctionCall == nil {
		return fmt.Errorf("function declaration %q must have a FunctionCall callback", fn.Name)
	}
	a.funcsMap[fn.Name] = fn
	return nil
}

func (a *ClaudeAgent) Send(ctx context.Context, sessionID, prompt string) (string, TokenUsage, error) {
	messages, err := a.loadHistory(ctx, sessionID)
	if err != nil {
		return "", TokenUsage{}, err
	}
	messages = append(messages, anthropic.NewUserMessage(anthropic.NewTextBlock(prompt)))

	text, newMessages, usage, err := a.chat(ctx, messages)
	if err != nil {
		return "", TokenUsage{}, err
	}

	if err := a.saveHistory(ctx, sessionID, newMessages); err != nil {
		return "", TokenUsage{}, err
	}

	return text, usage, nil
}

func (a *ClaudeAgent) chat(ctx context.Context, messages []anthropic.MessageParam) (string, []anthropic.MessageParam, TokenUsage, error) {
	params := anthropic.MessageNewParams{
		Model:     anthropic.Model(a.model),
		MaxTokens: 8192,
		Messages:  messages,
	}
	if a.sysInstruct != "" {
		params.System = []anthropic.TextBlockParam{
			{Text: a.sysInstruct},
		}
	}
	if tools := a.buildTools(); len(tools) > 0 {
		params.Tools = tools
	}

	resp, err := a.client.Messages.New(ctx, params)
	if err != nil {
		return "", nil, TokenUsage{}, fmt.Errorf("calling claude: %w", err)
	}

	usage := TokenUsage{
		PromptTokens:   int32(resp.Usage.InputTokens),
		ResponseTokens: int32(resp.Usage.OutputTokens),
		TotalTokens:    int32(resp.Usage.InputTokens + resp.Usage.OutputTokens),
	}

	// Convert response content blocks to param union blocks (for history)
	assistantBlocks := make([]anthropic.ContentBlockParamUnion, 0, len(resp.Content))
	var toolUseBlocks []anthropic.ToolUseBlock
	var textContent string

	for _, block := range resp.Content {
		switch b := block.AsAny().(type) {
		case anthropic.TextBlock:
			textContent = b.Text
			assistantBlocks = append(assistantBlocks, anthropic.NewTextBlock(b.Text))
		case anthropic.ToolUseBlock:
			toolUseBlocks = append(toolUseBlocks, b)
			assistantBlocks = append(assistantBlocks, anthropic.NewToolUseBlock(b.ID, json.RawMessage(b.Input), b.Name))
		}
	}

	// Append assistant turn to history
	messages = append(messages, anthropic.NewAssistantMessage(assistantBlocks...))

	if len(toolUseBlocks) == 0 {
		return textContent, messages, usage, nil
	}

	// Execute tools and build user tool_result message
	toolResultBlocks := make([]anthropic.ContentBlockParamUnion, 0, len(toolUseBlocks))
	for _, tb := range toolUseBlocks {
		fn, ok := a.funcsMap[tb.Name]
		if !ok {
			log.Printf("[skill] model=%q skill=%q (não existe — alucinação)", a.model, tb.Name)
			toolResultBlocks = append(toolResultBlocks, anthropic.NewToolResultBlock(tb.ID, fmt.Sprintf(`{"error":"função %q não existe"}`, tb.Name), true))
			continue
		}

		var args map[string]any
		json.Unmarshal(tb.Input, &args) //nolint:errcheck
		log.Printf("[skill] model=%q skill=%q args=%v", a.model, tb.Name, args)

		result, err := fn.FunctionCall(ctx, args)
		if err != nil {
			return "", nil, TokenUsage{}, fmt.Errorf("executing function %q: %w", tb.Name, err)
		}
		resultJSON, _ := json.Marshal(result)
		toolResultBlocks = append(toolResultBlocks, anthropic.NewToolResultBlock(tb.ID, string(resultJSON), false))
	}

	messages = append(messages, anthropic.NewUserMessage(toolResultBlocks...))

	text, messages, nextUsage, err := a.chat(ctx, messages)
	return text, messages, usage.add(nextUsage), err
}

func (a *ClaudeAgent) buildTools() []anthropic.ToolUnionParam {
	if len(a.funcsMap) == 0 {
		return nil
	}
	tools := make([]anthropic.ToolUnionParam, 0, len(a.funcsMap))
	for _, fn := range a.funcsMap {
		schemaRaw, _ := json.Marshal(fn.ParametersSchema)
		var inputSchema anthropic.ToolInputSchemaParam
		json.Unmarshal(schemaRaw, &inputSchema) //nolint:errcheck

		tp := anthropic.ToolParam{
			Name:        fn.Name,
			Description: anthropic.String(fn.Description),
			InputSchema: inputSchema,
		}
		tools = append(tools, anthropic.ToolUnionParam{OfTool: &tp})
	}
	return tools
}

// ── History persistence ──────────────────────────────────

func (a *ClaudeAgent) loadHistory(ctx context.Context, sessionID string) ([]anthropic.MessageParam, error) {
	if a.sessionRepo == nil || sessionID == "" {
		return nil, nil
	}
	stored, err := a.sessionRepo.Load(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	return modelContentsToClaudeMessages(stored, len(a.funcsMap) == 0), nil
}

func (a *ClaudeAgent) saveHistory(ctx context.Context, sessionID string, messages []anthropic.MessageParam) error {
	if a.sessionRepo == nil || sessionID == "" {
		return nil
	}
	return a.sessionRepo.Save(ctx, sessionID, claudeMessagesToModelContents(messages))
}

func modelContentsToClaudeMessages(contents []model.Content, stripTools bool) []anthropic.MessageParam {
	var msgs []anthropic.MessageParam
	for _, c := range contents {
		role := anthropic.MessageParamRoleUser
		if c.Role == "model" {
			role = anthropic.MessageParamRoleAssistant
		}
		var blocks []anthropic.ContentBlockParamUnion
		for _, p := range c.Parts {
			switch {
			case p.Text != "":
				blocks = append(blocks, anthropic.NewTextBlock(p.Text))
			case p.FunctionCall != nil && !stripTools:
				argsJSON, _ := json.Marshal(p.FunctionCall.Args)
				blocks = append(blocks, anthropic.NewToolUseBlock(p.FunctionCall.Name, json.RawMessage(argsJSON), p.FunctionCall.Name))
			case p.FunctionResponse != nil && !stripTools:
				resultJSON, _ := json.Marshal(p.FunctionResponse.Response)
				blocks = append(blocks, anthropic.NewToolResultBlock(p.FunctionResponse.Name, string(resultJSON), false))
			}
		}
		if len(blocks) > 0 {
			msgs = append(msgs, anthropic.MessageParam{Role: role, Content: blocks})
		}
	}
	return msgs
}

func claudeMessagesToModelContents(messages []anthropic.MessageParam) []model.Content {
	var contents []model.Content
	for _, m := range messages {
		role := "user"
		if m.Role == anthropic.MessageParamRoleAssistant {
			role = "model"
		}
		var parts []model.Part
		for _, block := range m.Content {
			switch {
			case block.OfText != nil:
				if block.OfText.Text != "" {
					parts = append(parts, model.Part{Text: block.OfText.Text})
				}
			case block.OfToolUse != nil:
				var args map[string]any
				if raw, ok := block.OfToolUse.Input.(json.RawMessage); ok {
					json.Unmarshal(raw, &args) //nolint:errcheck
				}
				parts = append(parts, model.Part{
					FunctionCall: &struct {
						Name string         `bson:"name"  json:"name"`
						Args map[string]any `bson:"args"  json:"args"`
					}{Name: block.OfToolUse.Name, Args: args},
				})
			case block.OfToolResult != nil:
				var resp any
				if len(block.OfToolResult.Content) > 0 && block.OfToolResult.Content[0].OfText != nil {
					json.Unmarshal([]byte(block.OfToolResult.Content[0].OfText.Text), &resp) //nolint:errcheck
				}
				parts = append(parts, model.Part{
					FunctionResponse: &struct {
						Name     string `bson:"name"     json:"name"`
						Response any    `bson:"response" json:"response"`
					}{Name: block.OfToolResult.ToolUseID, Response: resp},
				})
			}
		}
		if len(parts) > 0 {
			contents = append(contents, model.Content{Role: role, Parts: parts})
		}
	}
	return contents
}

var _ Agent = (*ClaudeAgent)(nil)
