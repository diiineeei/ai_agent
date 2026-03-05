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
	geminiClient       *genai.Client
	model              string
	defaultInstruction string
	sessionRepo        repository.SessionRepository
	registry           *skills.SkillRegistry
	settingsRepo       repository.SettingsRepository
}

func NewChatHandler(
	geminiClient *genai.Client,
	model, defaultInstruction string,
	sessionRepo repository.SessionRepository,
	registry *skills.SkillRegistry,
	settingsRepo repository.SettingsRepository,
) *ChatHandler {
	return &ChatHandler{
		geminiClient:       geminiClient,
		model:              model,
		defaultInstruction: defaultInstruction,
		sessionRepo:        sessionRepo,
		registry:           registry,
		settingsRepo:       settingsRepo,
	}
}

type promptRequest struct {
	SessionID string `json:"session_id"`
	Prompt    string `json:"prompt"`
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

	instruction, err := h.settingsRepo.GetSystemInstruction(r.Context())
	if err != nil || instruction == "" {
		instruction = h.defaultInstruction
	}

	a := agent.NewWithRepo(h.geminiClient, h.model, instruction, h.sessionRepo)

	if err := h.registry.LoadEnabled(r.Context(), a); err != nil {
		jsonError(w, "erro ao carregar skills: "+err.Error(), http.StatusInternalServerError)
		return
	}

	response, err := a.Send(r.Context(), req.SessionID, req.Prompt)
	if err != nil {
		jsonError(w, "erro ao processar prompt: "+err.Error(), http.StatusInternalServerError)
		return
	}

	jsonResponse(w, http.StatusOK, map[string]string{
		"session_id": req.SessionID,
		"response":   response,
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
