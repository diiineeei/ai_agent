package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"ai_agent/internal/model"
	"ai_agent/internal/repository"

	"google.golang.org/genai"
)

type AgentConfigHandler struct {
	repo         repository.AgentConfigRepository
	geminiClient *genai.Client
}

func NewAgentConfigHandler(repo repository.AgentConfigRepository, geminiClient *genai.Client) *AgentConfigHandler {
	return &AgentConfigHandler{repo: repo, geminiClient: geminiClient}
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

// ImproveInstruction handles POST /agent-configs/improve-instruction
// Uses the specified model to rewrite a system instruction via a single prompt call.
func (h *AgentConfigHandler) ImproveInstruction(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Model       string `json:"model"`
		Instruction string `json:"instruction"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonError(w, "body inválido: "+err.Error(), http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(body.Instruction) == "" {
		jsonError(w, "campo 'instruction' é obrigatório", http.StatusBadRequest)
		return
	}
	if body.Model == "" {
		jsonError(w, "campo 'model' é obrigatório", http.StatusBadRequest)
		return
	}

	prompt := fmt.Sprintf(`Você é um especialista em engenharia de prompts para agentes de IA.
Melhore a instrução de sistema abaixo, tornando-a mais clara, específica e eficaz.
Retorne APENAS a instrução melhorada, sem nenhuma explicação adicional.

Instrução:
%s`, strings.TrimSpace(body.Instruction))

	contents := []*genai.Content{
		{Role: "user", Parts: []*genai.Part{{Text: prompt}}},
	}

	resp, err := h.geminiClient.Models.GenerateContent(r.Context(), body.Model, contents, nil)
	if err != nil {
		jsonError(w, "erro ao chamar o modelo: "+err.Error(), http.StatusInternalServerError)
		return
	}

	jsonResponse(w, http.StatusOK, map[string]string{"instruction": strings.TrimSpace(resp.Text())})
}
