package handler

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"time"

	"ai_agent/internal/agent"
	"ai_agent/internal/model"
	"ai_agent/internal/repository"

	"github.com/ledongthuc/pdf"
)

type FileHandler struct {
	fileRepo repository.FileRepository
	embedder *agent.Embedder
}

func NewFileHandler(fileRepo repository.FileRepository, embedder *agent.Embedder) *FileHandler {
	return &FileHandler{fileRepo: fileRepo, embedder: embedder}
}

// Upload handles POST /files — accepts multipart file, extracts text, and stores in MongoDB.
func (h *FileHandler) Upload(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		jsonError(w, "erro ao processar form: "+err.Error(), http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		jsonError(w, "campo 'file' é obrigatório", http.StatusBadRequest)
		return
	}
	defer file.Close()

	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	var text string
	switch {
	case strings.HasPrefix(contentType, "text/"):
		data, err := io.ReadAll(file)
		if err != nil {
			jsonError(w, "erro ao ler arquivo: "+err.Error(), http.StatusInternalServerError)
			return
		}
		text = string(data)
	case contentType == "application/pdf":
		extracted, err := extractPDFText(file)
		if err != nil {
			jsonError(w, "erro ao extrair texto do PDF: "+err.Error(), http.StatusInternalServerError)
			return
		}
		text = extracted
	default:
		jsonError(w, "tipo de arquivo não suportado: "+contentType, http.StatusUnsupportedMediaType)
		return
	}

	embeddings := h.embedder.Embed(text)

	doc := model.FileDocument{
		Name:        header.Filename,
		ContentType: contentType,
		Size:        header.Size,
		Content:     text,
		Embeddings:  embeddings,
		UploadedAt:  time.Now(),
	}

	id, err := h.fileRepo.Save(r.Context(), doc)
	if err != nil {
		jsonError(w, "erro ao salvar arquivo: "+err.Error(), http.StatusInternalServerError)
		return
	}

	jsonResponse(w, http.StatusCreated, map[string]any{
		"id":          id,
		"name":        doc.Name,
		"size":        doc.Size,
		"uploaded_at": doc.UploadedAt,
	})
}

// List handles GET /files — lists all files without content or embeddings.
func (h *FileHandler) List(w http.ResponseWriter, r *http.Request) {
	files, err := h.fileRepo.ListAll(r.Context())
	if err != nil {
		jsonError(w, "erro ao listar arquivos: "+err.Error(), http.StatusInternalServerError)
		return
	}

	items := make([]model.FileListItem, 0, len(files))
	for _, f := range files {
		items = append(items, model.FileListItem{
			ID:          f.ID,
			Name:        f.Name,
			ContentType: f.ContentType,
			Size:        f.Size,
			UploadedAt:  f.UploadedAt,
		})
	}
	jsonResponse(w, http.StatusOK, items)
}

// Delete handles DELETE /files/{id} — removes a file by its MongoDB ObjectID.
func (h *FileHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		jsonError(w, "id é obrigatório", http.StatusBadRequest)
		return
	}

	if err := h.fileRepo.Delete(r.Context(), id); err != nil {
		jsonError(w, "erro ao deletar arquivo: "+err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, http.StatusOK, map[string]string{"message": "arquivo removido com sucesso"})
}

// extractPDFText extracts plain text from a PDF file using ledongthuc/pdf.
func extractPDFText(r io.Reader) (string, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return "", err
	}

	reader, err := pdf.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	for i := 1; i <= reader.NumPage(); i++ {
		page := reader.Page(i)
		if page.V.IsNull() {
			continue
		}
		content, err := page.GetPlainText(nil)
		if err != nil {
			continue
		}
		sb.WriteString(content)
		sb.WriteString("\n")
	}
	return sb.String(), nil
}
