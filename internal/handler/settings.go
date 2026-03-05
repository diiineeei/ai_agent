package handler

import (
	"encoding/json"
	"net/http"

	"ai_agent/internal/repository"
)

type SettingsHandler struct {
	settingsRepo repository.SettingsRepository
}

func NewSettingsHandler(settingsRepo repository.SettingsRepository) *SettingsHandler {
	return &SettingsHandler{settingsRepo: settingsRepo}
}

// GetSystemInstruction handles GET /settings/system-instruction
func (h *SettingsHandler) GetSystemInstruction(w http.ResponseWriter, r *http.Request) {
	value, err := h.settingsRepo.GetSystemInstruction(r.Context())
	if err != nil {
		jsonError(w, "erro ao carregar instrução: "+err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, http.StatusOK, map[string]string{"value": value})
}

// SetSystemInstruction handles PUT /settings/system-instruction
func (h *SettingsHandler) SetSystemInstruction(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Value string `json:"value"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonError(w, "body inválido: "+err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.settingsRepo.SetSystemInstruction(r.Context(), body.Value); err != nil {
		jsonError(w, "erro ao salvar instrução: "+err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, http.StatusOK, map[string]string{"value": body.Value})
}
