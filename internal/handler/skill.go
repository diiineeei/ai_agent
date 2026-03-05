package handler

import (
	"encoding/json"
	"net/http"

	"ai_agent/internal/repository"
)

type SkillHandler struct {
	skillRepo repository.SkillRepository
}

func NewSkillHandler(skillRepo repository.SkillRepository) *SkillHandler {
	return &SkillHandler{skillRepo: skillRepo}
}

// List handles GET /skills — returns all skills with their enabled status.
func (h *SkillHandler) List(w http.ResponseWriter, r *http.Request) {
	skills, err := h.skillRepo.ListAll(r.Context())
	if err != nil {
		jsonError(w, "erro ao listar skills", http.StatusInternalServerError)
		return
	}
	jsonResponse(w, http.StatusOK, skills)
}

// Toggle handles PUT /skills/{name}/toggle — flips the enabled flag for a skill.
func (h *SkillHandler) Toggle(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if name == "" {
		jsonError(w, "nome da skill é obrigatório", http.StatusBadRequest)
		return
	}

	updated, err := h.skillRepo.Toggle(r.Context(), name)
	if err != nil {
		jsonError(w, err.Error(), http.StatusNotFound)
		return
	}
	jsonResponse(w, http.StatusOK, updated)
}

func jsonResponse(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func jsonError(w http.ResponseWriter, msg string, status int) {
	jsonResponse(w, status, map[string]string{"error": msg})
}
