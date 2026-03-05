package handler

import (
	"encoding/json"
	"net/http"

	"ai_agent/internal/model"
	"ai_agent/internal/repository"
)

type FeedbackHandler struct {
	repo repository.FeedbackRepository
}

func NewFeedbackHandler(repo repository.FeedbackRepository) *FeedbackHandler {
	return &FeedbackHandler{repo: repo}
}

// Submit handles POST /feedback — upserts a thumbs-up or thumbs-down rating for a message.
func (h *FeedbackHandler) Submit(w http.ResponseWriter, r *http.Request) {
	var req struct {
		SessionID     string `json:"session_id"`
		MessageIndex  int    `json:"message_index"`
		AgentConfigID string `json:"agent_config_id"`
		Rating        string `json:"rating"` // "up" | "down"
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "body inválido: "+err.Error(), http.StatusBadRequest)
		return
	}
	if req.SessionID == "" || (req.Rating != "up" && req.Rating != "down") {
		jsonError(w, "campos 'session_id' e 'rating' (up|down) são obrigatórios", http.StatusBadRequest)
		return
	}
	f := model.Feedback{
		SessionID:     req.SessionID,
		MessageIndex:  req.MessageIndex,
		AgentConfigID: req.AgentConfigID,
		Rating:        req.Rating,
	}
	if err := h.repo.Upsert(r.Context(), f); err != nil {
		jsonError(w, "erro ao salvar avaliação: "+err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, http.StatusOK, map[string]string{"message": "avaliação registrada"})
}

// ForSession handles GET /feedback?session_id=xxx — returns all ratings for a session.
func (h *FeedbackHandler) ForSession(w http.ResponseWriter, r *http.Request) {
	sessionID := r.URL.Query().Get("session_id")
	if sessionID == "" {
		jsonError(w, "parâmetro 'session_id' é obrigatório", http.StatusBadRequest)
		return
	}
	feedback, err := h.repo.GetForSession(r.Context(), sessionID)
	if err != nil {
		jsonError(w, "erro ao buscar avaliações: "+err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, http.StatusOK, feedback)
}

// Stats handles GET /feedback/stats — returns thumbs up/down counts grouped by agent.
func (h *FeedbackHandler) Stats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.repo.StatsByAgent(r.Context())
	if err != nil {
		jsonError(w, "erro ao buscar estatísticas: "+err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, http.StatusOK, stats)
}
