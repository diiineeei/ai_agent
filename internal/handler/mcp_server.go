package handler

import (
	"encoding/json"
	"net/http"

	"ai_agent/internal/model"
	"ai_agent/internal/repository"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type McpServerHandler struct {
	repo repository.McpServerRepository
}

func NewMcpServerHandler(repo repository.McpServerRepository) *McpServerHandler {
	return &McpServerHandler{repo: repo}
}

// List handles GET /mcp-servers
func (h *McpServerHandler) List(w http.ResponseWriter, r *http.Request) {
	servers, err := h.repo.ListAll(r.Context())
	if err != nil {
		jsonError(w, "erro ao listar servidores MCP", http.StatusInternalServerError)
		return
	}
	jsonResponse(w, http.StatusOK, servers)
}

// Create handles POST /mcp-servers
func (h *McpServerHandler) Create(w http.ResponseWriter, r *http.Request) {
	var s model.McpServer
	if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
		jsonError(w, "JSON inválido", http.StatusBadRequest)
		return
	}
	if s.Name == "" || s.Transport == "" {
		jsonError(w, "name e transport são obrigatórios", http.StatusBadRequest)
		return
	}

	created, err := h.repo.Create(r.Context(), s)
	if err != nil {
		jsonError(w, "erro ao criar servidor MCP", http.StatusInternalServerError)
		return
	}
	jsonResponse(w, http.StatusCreated, created)
}

// Update handles PUT /mcp-servers/{id}
func (h *McpServerHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := bson.ObjectIDFromHex(r.PathValue("id"))
	if err != nil {
		jsonError(w, "ID inválido", http.StatusBadRequest)
		return
	}

	var s model.McpServer
	if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
		jsonError(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	updated, err := h.repo.Update(r.Context(), id, s)
	if err != nil {
		jsonError(w, err.Error(), http.StatusNotFound)
		return
	}
	jsonResponse(w, http.StatusOK, updated)
}

// Delete handles DELETE /mcp-servers/{id}
func (h *McpServerHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := bson.ObjectIDFromHex(r.PathValue("id"))
	if err != nil {
		jsonError(w, "ID inválido", http.StatusBadRequest)
		return
	}

	if err := h.repo.Delete(r.Context(), id); err != nil {
		jsonError(w, err.Error(), http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Toggle handles PUT /mcp-servers/{id}/toggle
func (h *McpServerHandler) Toggle(w http.ResponseWriter, r *http.Request) {
	id, err := bson.ObjectIDFromHex(r.PathValue("id"))
	if err != nil {
		jsonError(w, "ID inválido", http.StatusBadRequest)
		return
	}

	updated, err := h.repo.Toggle(r.Context(), id)
	if err != nil {
		jsonError(w, err.Error(), http.StatusNotFound)
		return
	}
	jsonResponse(w, http.StatusOK, updated)
}
