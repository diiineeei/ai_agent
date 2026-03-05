package handler

import (
	"encoding/json"
	"net/http"

	"ai_agent/internal/model"
	"ai_agent/internal/repository"
)

type AgentConfigHandler struct {
	repo repository.AgentConfigRepository
}

func NewAgentConfigHandler(repo repository.AgentConfigRepository) *AgentConfigHandler {
	return &AgentConfigHandler{repo: repo}
}

// List handles GET /agent-configs
func (h *AgentConfigHandler) List(w http.ResponseWriter, r *http.Request) {
	configs, err := h.repo.List(r.Context())
	if err != nil {
		jsonError(w, "erro ao listar agentes: "+err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, http.StatusOK, configs)
}

// GetByID handles GET /agent-configs/{id}
func (h *AgentConfigHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	cfg, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		jsonError(w, "erro ao buscar agente: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if cfg == nil {
		jsonError(w, "agente não encontrado", http.StatusNotFound)
		return
	}
	jsonResponse(w, http.StatusOK, cfg)
}

// Create handles POST /agent-configs
func (h *AgentConfigHandler) Create(w http.ResponseWriter, r *http.Request) {
	var cfg model.AgentConfig
	if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
		jsonError(w, "body inválido: "+err.Error(), http.StatusBadRequest)
		return
	}
	if cfg.Name == "" || cfg.Model == "" {
		jsonError(w, "campos 'name' e 'model' são obrigatórios", http.StatusBadRequest)
		return
	}
	created, err := h.repo.Create(r.Context(), cfg)
	if err != nil {
		jsonError(w, "erro ao criar agente: "+err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, http.StatusCreated, created)
}

// Update handles PUT /agent-configs/{id}
func (h *AgentConfigHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var cfg model.AgentConfig
	if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
		jsonError(w, "body inválido: "+err.Error(), http.StatusBadRequest)
		return
	}
	if cfg.Name == "" || cfg.Model == "" {
		jsonError(w, "campos 'name' e 'model' são obrigatórios", http.StatusBadRequest)
		return
	}
	updated, err := h.repo.Update(r.Context(), id, cfg)
	if err != nil {
		jsonError(w, "erro ao atualizar agente: "+err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, http.StatusOK, updated)
}

// Delete handles DELETE /agent-configs/{id}
func (h *AgentConfigHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := h.repo.Delete(r.Context(), id); err != nil {
		jsonError(w, "erro ao excluir agente: "+err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, http.StatusOK, map[string]string{"message": "agente removido com sucesso"})
}
