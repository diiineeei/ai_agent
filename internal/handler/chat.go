package handler

import (
	"encoding/json"
	"net/http"

	"ai_agent/internal/agent"
	"ai_agent/internal/repository"
	"ai_agent/internal/skills"

	"google.golang.org/genai"
)

func containsSkill(list []string, name string) bool {
	for _, s := range list {
		if s == name {
			return true
		}
	}
	return false
}

type ChatHandler struct {
	geminiClient    *genai.Client
	sessionRepo     repository.SessionRepository
	agentConfigRepo repository.AgentConfigRepository
	registry        *skills.SkillRegistry
}

func NewChatHandler(
	geminiClient *genai.Client,
	sessionRepo repository.SessionRepository,
	agentConfigRepo repository.AgentConfigRepository,
	registry *skills.SkillRegistry,
) *ChatHandler {
	return &ChatHandler{
		geminiClient:    geminiClient,
		sessionRepo:     sessionRepo,
		agentConfigRepo: agentConfigRepo,
		registry:        registry,
	}
}

type promptRequest struct {
	SessionID     string `json:"session_id"`
	Prompt        string `json:"prompt"`
	AgentConfigID string `json:"agent_config_id,omitempty"`
}

// SendPrompt handles POST /prompt — sends a user message to the agent.
// A fresh agent is created per request so only currently-enabled skills are used.
func (h *ChatHandler) SendPrompt(w http.ResponseWriter, r *http.Request) {
	var req promptRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "body inválido: "+err.Error(), http.StatusBadRequest)
		return
	}
	if req.Prompt == "" {
		jsonError(w, "campo 'prompt' é obrigatório", http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	storedID, err := h.sessionRepo.GetAgentConfigID(ctx, req.SessionID)
	if err != nil {
		jsonError(w, "erro ao verificar sessão: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if storedID == "" {
		if req.AgentConfigID == "" {
			jsonError(w, "campo 'agent_config_id' é obrigatório para novas sessões", http.StatusBadRequest)
			return
		}
		if err := h.sessionRepo.SetAgentConfigID(ctx, req.SessionID, req.AgentConfigID); err != nil {
			jsonError(w, "erro ao vincular agente à sessão: "+err.Error(), http.StatusInternalServerError)
			return
		}
		storedID = req.AgentConfigID
	}

	cfg, err := h.agentConfigRepo.GetByID(ctx, storedID)
	if err != nil || cfg == nil {
		jsonError(w, "agente não encontrado: "+storedID, http.StatusBadRequest)
		return
	}

	var a agent.Agent
	if cfg.Provider == "ollama" {
		a = agent.NewOllama(cfg.BaseURL, cfg.Model, cfg.SystemInstruction, h.sessionRepo)
	} else {
		a = agent.NewWithRepo(h.geminiClient, cfg.Model, cfg.SystemInstruction, h.sessionRepo)
	}

	if err := h.registry.LoadByNames(ctx, a, cfg.EnabledSkills); err != nil {
		jsonError(w, "erro ao carregar skills: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Inject contextual skills that require session context at instantiation time.
	// suggest_questions requires the Gemini client, so skip for Ollama agents.
	if cfg.Provider != "ollama" &&
		containsSkill(cfg.EnabledSkills, "suggest_questions") && h.registry.IsEnabled(ctx, "suggest_questions") {
		sq := skills.NewSuggestQuestionsSkill(h.geminiClient, h.sessionRepo, *cfg, req.SessionID)
		if err := a.AddFunctionCall(sq.Declaration()); err != nil {
			jsonError(w, "erro ao registrar skill suggest_questions: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	response, usage, err := a.Send(ctx, req.SessionID, req.Prompt)
	if err != nil {
		jsonError(w, "erro ao processar prompt: "+err.Error(), http.StatusInternalServerError)
		return
	}

	jsonResponse(w, http.StatusOK, map[string]any{
		"session_id":  req.SessionID,
		"response":    response,
		"agent_name":  cfg.Name,
		"token_usage": usage,
	})
}

// GetHistory handles GET /history — returns the conversation history for a session.
func (h *ChatHandler) GetHistory(w http.ResponseWriter, r *http.Request) {
	sessionID := r.URL.Query().Get("session_id")
	if sessionID == "" {
		jsonError(w, "parâmetro 'session_id' é obrigatório", http.StatusBadRequest)
		return
	}

	history, err := h.sessionRepo.Load(r.Context(), sessionID)
	if err != nil {
		jsonError(w, "erro ao carregar histórico: "+err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, http.StatusOK, history)
}

// ListSessions handles GET /sessions — returns all session summaries sorted by last activity.
func (h *ChatHandler) ListSessions(w http.ResponseWriter, r *http.Request) {
	sessions, err := h.sessionRepo.ListAll(r.Context())
	if err != nil {
		jsonError(w, "erro ao listar sessões: "+err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, http.StatusOK, sessions)
}

// RenameSession handles PUT /sessions/{id}/name — sets an optional display name for a session.
func (h *ChatHandler) RenameSession(w http.ResponseWriter, r *http.Request) {
	sessionID := r.PathValue("id")
	var body struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonError(w, "body inválido: "+err.Error(), http.StatusBadRequest)
		return
	}
	if err := h.sessionRepo.SetName(r.Context(), sessionID, body.Name); err != nil {
		jsonError(w, "erro ao renomear sessão: "+err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, http.StatusOK, map[string]string{"message": "sessão renomeada com sucesso"})
}

// DeleteHistory handles DELETE /history — clears the session history.
func (h *ChatHandler) DeleteHistory(w http.ResponseWriter, r *http.Request) {
	sessionID := r.URL.Query().Get("session_id")
	if sessionID == "" {
		jsonError(w, "parâmetro 'session_id' é obrigatório", http.StatusBadRequest)
		return
	}

	if err := h.sessionRepo.Delete(r.Context(), sessionID); err != nil {
		jsonError(w, "erro ao apagar histórico: "+err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, http.StatusOK, map[string]string{"message": "histórico removido com sucesso"})
}
