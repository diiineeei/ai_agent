package handler

import (
	"net/http"

	"ai_agent/internal/repository"
	"ai_agent/internal/skills"

	"google.golang.org/genai"
)

type SuggestHandler struct {
	geminiClient    *genai.Client
	claudeAPIKey    string
	sessionRepo     repository.SessionRepository
	agentConfigRepo repository.AgentConfigRepository
}

func NewSuggestHandler(
	geminiClient *genai.Client,
	claudeAPIKey string,
	sessionRepo repository.SessionRepository,
	agentConfigRepo repository.AgentConfigRepository,
) *SuggestHandler {
	return &SuggestHandler{
		geminiClient:    geminiClient,
		claudeAPIKey:    claudeAPIKey,
		sessionRepo:     sessionRepo,
		agentConfigRepo: agentConfigRepo,
	}
}

// Suggest handles GET /suggest-questions?session_id=xxx
// Retorna uma lista de perguntas que o usuário pode fazer ao agente.
func (h *SuggestHandler) Suggest(w http.ResponseWriter, r *http.Request) {
	sessionID := r.URL.Query().Get("session_id")
	if sessionID == "" {
		jsonError(w, "parâmetro 'session_id' é obrigatório", http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	agentConfigID, err := h.sessionRepo.GetAgentConfigID(ctx, sessionID)
	if err != nil {
		jsonError(w, "erro ao buscar sessão: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if agentConfigID == "" {
		jsonResponse(w, http.StatusOK, map[string]any{"questions": []string{}})
		return
	}

	cfg, err := h.agentConfigRepo.GetByID(ctx, agentConfigID)
	if err != nil || cfg == nil {
		jsonResponse(w, http.StatusOK, map[string]any{"questions": []string{}})
		return
	}

	// Só gera sugestões se a skill estiver habilitada no agente
	skillEnabled := false
	for _, s := range cfg.EnabledSkills {
		if s == "suggest_questions" {
			skillEnabled = true
			break
		}
	}
	if !skillEnabled {
		jsonResponse(w, http.StatusOK, map[string]any{"questions": []string{}})
		return
	}

	questions, err := skills.GenerateSuggestions(ctx, h.geminiClient, h.claudeAPIKey, h.sessionRepo, *cfg, sessionID, "")
	if err != nil {
		jsonError(w, "erro ao gerar sugestões: "+err.Error(), http.StatusInternalServerError)
		return
	}

	jsonResponse(w, http.StatusOK, map[string]any{"questions": questions})
}
