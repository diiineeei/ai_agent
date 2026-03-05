package skills

import (
	"context"
	"fmt"
	"strings"

	"ai_agent/internal/agent"
	"ai_agent/internal/model"
	"ai_agent/internal/repository"

	"google.golang.org/genai"
)

// SuggestQuestionsSkill gera perguntas relevantes que o usuário pode fazer ao agente,
// com base no histórico da conversa e nas características do agente.
type SuggestQuestionsSkill struct {
	geminiClient *genai.Client
	sessionRepo  repository.SessionRepository
	agentConfig  model.AgentConfig
	sessionID    string
}

func NewSuggestQuestionsSkill(
	geminiClient *genai.Client,
	sessionRepo repository.SessionRepository,
	agentConfig model.AgentConfig,
	sessionID string,
) *SuggestQuestionsSkill {
	return &SuggestQuestionsSkill{
		geminiClient: geminiClient,
		sessionRepo:  sessionRepo,
		agentConfig:  agentConfig,
		sessionID:    sessionID,
	}
}

func (s *SuggestQuestionsSkill) Name() string { return "suggest_questions" }

func (s *SuggestQuestionsSkill) Declaration() *agent.FunctionDeclaration {
	return &agent.FunctionDeclaration{
		Name:        "suggest_questions",
		Description: "Gera uma lista de perguntas que o usuário pode fazer ao assistente, com base no histórico da conversa e no papel do agente.",
		ParametersSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"focus": map[string]any{
					"type":        "string",
					"description": "Tema ou área específica para direcionar as sugestões (opcional)",
				},
			},
		},
		ResponseSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"questions": map[string]any{
					"type":  "array",
					"items": map[string]any{"type": "string"},
				},
			},
		},
		FunctionCall: s.execute,
	}
}

func (s *SuggestQuestionsSkill) execute(ctx context.Context, args map[string]any) (map[string]any, error) {
	focus, _ := args["focus"].(string)

	questions, err := GenerateSuggestions(ctx, s.geminiClient, s.sessionRepo, s.agentConfig, s.sessionID, focus)
	if err != nil {
		return nil, err
	}
	return map[string]any{"questions": questions}, nil
}

// GenerateSuggestions é a lógica compartilhada entre a skill e o handler HTTP.
func GenerateSuggestions(
	ctx context.Context,
	geminiClient *genai.Client,
	sessionRepo repository.SessionRepository,
	agentConfig model.AgentConfig,
	sessionID string,
	focus string,
) ([]string, error) {
	history, err := sessionRepo.Load(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("carregando histórico: %w", err)
	}

	// Resumo das últimas mensagens (máx 10 turnos)
	var sb strings.Builder
	count := 0
	for _, c := range history {
		if count >= 10 {
			break
		}
		if c.Role != "user" && c.Role != "model" {
			continue
		}
		for _, p := range c.Parts {
			if p.Text != "" {
				role := "Usuário"
				if c.Role == "model" {
					role = "Assistente"
				}
				fmt.Fprintf(&sb, "%s: %s\n", role, p.Text)
				count++
				break
			}
		}
	}

	focusLine := ""
	if focus != "" {
		focusLine = "Foco especial em: " + focus + "\n\n"
	}

	prompt := fmt.Sprintf(
		"Você é o assistente chamado '%s'.\nSua função: %s\n\nHistórico recente da conversa:\n%s\n%sCom base no papel do assistente e no histórico acima, sugira exatamente 3 perguntas relevantes e úteis que o usuário pode fazer a seguir. Retorne apenas as perguntas, uma por linha, sem numeração, sem marcadores.",
		agentConfig.Name,
		agentConfig.SystemInstruction,
		sb.String(),
		focusLine,
	)

	modelName := agentConfig.Model
	if modelName == "" {
		modelName = "gemini-2.5-flash"
	}

	chat, err := geminiClient.Chats.Create(ctx, modelName, &genai.GenerateContentConfig{}, nil)
	if err != nil {
		return nil, fmt.Errorf("criando chat para sugestões: %w", err)
	}

	resp, err := chat.SendMessage(ctx, genai.Part{Text: prompt})
	if err != nil {
		return nil, fmt.Errorf("gerando sugestões: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(resp.Text()), "\n")
	questions := make([]string, 0, 3)
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			questions = append(questions, line)
		}
	}
	return questions, nil
}

var _ Skill = (*SuggestQuestionsSkill)(nil)
