package skills

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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
// Usa o provider do agente configurado: Gemini ou Ollama.
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

	if agentConfig.Provider == "ollama" {
		return generateWithOllama(ctx, agentConfig, prompt)
	}
	return generateWithGemini(ctx, geminiClient, agentConfig.Model, prompt)
}

func generateWithGemini(ctx context.Context, client *genai.Client, modelName, prompt string) ([]string, error) {
	if modelName == "" {
		modelName = "gemini-2.5-flash"
	}
	chat, err := client.Chats.Create(ctx, modelName, &genai.GenerateContentConfig{}, nil)
	if err != nil {
		return nil, fmt.Errorf("criando chat para sugestões: %w", err)
	}
	resp, err := chat.SendMessage(ctx, genai.Part{Text: prompt})
	if err != nil {
		return nil, fmt.Errorf("gerando sugestões: %w", err)
	}
	return parseQuestions(resp.Text()), nil
}

func generateWithOllama(ctx context.Context, agentConfig model.AgentConfig, prompt string) ([]string, error) {
	baseURL := strings.TrimRight(agentConfig.BaseURL, "/")
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}

	body, _ := json.Marshal(map[string]any{
		"model": agentConfig.Model,
		"messages": []map[string]any{
			{"role": "user", "content": prompt},
		},
		"stream": false,
	})

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/api/chat", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("criando request ollama: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("chamando ollama: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		raw, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ollama retornou %d: %s", resp.StatusCode, strings.TrimSpace(string(raw)))
	}

	var result struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decodificando resposta ollama: %w", err)
	}

	return parseQuestions(result.Message.Content), nil
}

func parseQuestions(text string) []string {
	lines := strings.Split(strings.TrimSpace(text), "\n")
	questions := make([]string, 0, 3)
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			questions = append(questions, line)
		}
	}
	return questions
}

var _ Skill = (*SuggestQuestionsSkill)(nil)
