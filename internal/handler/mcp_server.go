package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"time"

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

// Ping handles POST /mcp-servers/{id}/ping
func (h *McpServerHandler) Ping(w http.ResponseWriter, r *http.Request) {
	id, err := bson.ObjectIDFromHex(r.PathValue("id"))
	if err != nil {
		jsonError(w, "ID inválido", http.StatusBadRequest)
		return
	}

	srv, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		jsonError(w, "servidor não encontrado", http.StatusNotFound)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	online, latencyMs, msg := pingMcpServer(ctx, srv)
	jsonResponse(w, http.StatusOK, map[string]any{
		"online":     online,
		"latency_ms": latencyMs,
		"message":    msg,
	})
}

func pingMcpServer(ctx context.Context, srv *model.McpServer) (online bool, latencyMs int64, message string) {
	initReq := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"ping","version":"1.0"}}}`

	start := time.Now()

	switch srv.Transport {
	case "http":
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, srv.URL, bytes.NewBufferString(initReq))
		if err != nil {
			return false, 0, fmt.Sprintf("erro ao criar requisição: %v", err)
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return false, 0, fmt.Sprintf("sem resposta: %v", err)
		}
		defer resp.Body.Close()
		io.Copy(io.Discard, resp.Body)

		ms := time.Since(start).Milliseconds()
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return true, ms, fmt.Sprintf("online (%d ms)", ms)
		}
		return false, ms, fmt.Sprintf("HTTP %d", resp.StatusCode)

	case "stdio":
		cmd := exec.CommandContext(ctx, srv.Command, srv.Args...)
		for k, v := range srv.Env {
			cmd.Env = append(cmd.Env, k+"="+v)
		}

		stdin, err := cmd.StdinPipe()
		if err != nil {
			return false, 0, fmt.Sprintf("erro ao criar pipe: %v", err)
		}
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			return false, 0, fmt.Sprintf("erro ao criar pipe: %v", err)
		}

		if err := cmd.Start(); err != nil {
			return false, 0, fmt.Sprintf("falha ao iniciar processo: %v", err)
		}
		defer cmd.Process.Kill() //nolint:errcheck

		if _, err := fmt.Fprintln(stdin, initReq); err != nil {
			return false, 0, fmt.Sprintf("erro ao escrever: %v", err)
		}
		stdin.Close()

		buf := make([]byte, 4096)
		n, _ := stdout.Read(buf)
		ms := time.Since(start).Milliseconds()

		if n > 0 {
			return true, ms, fmt.Sprintf("online (%d ms)", ms)
		}
		return false, ms, "processo iniciou mas não respondeu"

	default:
		return false, 0, fmt.Sprintf("transport %q não suportado para ping", srv.Transport)
	}
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
