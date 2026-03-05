package handler

import (
	"encoding/json"
	"net/http"

	"ai_agent/internal/agent"
	"ai_agent/internal/repository"
	"ai_agent/internal/skills"

	"google.golang.org/genai"
)

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

	a := agent.NewWithRepo(h.geminiClient, cfg.Model, cfg.SystemInstruction, h.sessionRepo)

	if err := h.registry.LoadByNames(ctx, a, cfg.EnabledSkills); err != nil {
		jsonError(w, "erro ao carregar skills: "+err.Error(), http.StatusInternalServerError)
		return
	}

	response, err := a.Send(ctx, req.SessionID, req.Prompt)
	if err != nil {
		jsonError(w, "erro ao processar prompt: "+err.Error(), http.StatusInternalServerError)
		return
	}

	jsonResponse(w, http.StatusOK, map[string]string{
		"session_id": req.SessionID,
		"response":   response,
		"agent_name": cfg.Name,
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
