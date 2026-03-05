package skills

import (
	"context"
	"fmt"

	"ai_agent/internal/agent"
	"ai_agent/internal/repository"
)

// SearchDocumentsSkill performs semantic search over files stored in MongoDB.
type SearchDocumentsSkill struct {
	fileRepo repository.FileRepository
	embedder *agent.Embedder
}

func NewSearchDocumentsSkill(fileRepo repository.FileRepository, embedder *agent.Embedder) *SearchDocumentsSkill {
	return &SearchDocumentsSkill{fileRepo: fileRepo, embedder: embedder}
}

func (s *SearchDocumentsSkill) Name() string { return "search_documents" }

func (s *SearchDocumentsSkill) Declaration() *agent.FunctionDeclaration {
	return &agent.FunctionDeclaration{
		Name:        "search_documents",
		Description: "Realiza busca semântica nos documentos enviados pelo usuário.",
		ParametersSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"query": map[string]any{
					"type":        "string",
					"description": "Texto descrevendo a informação que deseja encontrar",
				},
			},
			"required": []string{"query"},
		},
		ResponseSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"results": map[string]any{
					"type": "array",
					"items": map[string]any{
						"type": "object",
						"properties": map[string]any{
							"filename": map[string]any{"type": "string"},
							"content":  map[string]any{"type": "string"},
							"score":    map[string]any{"type": "number"},
						},
					},
				},
			},
		},
		FunctionCall: s.execute,
	}
}

func (s *SearchDocumentsSkill) execute(ctx context.Context, args map[string]any) (map[string]any, error) {
	query, ok := args["query"].(string)
	if !ok || query == "" {
		return nil, fmt.Errorf("argumento query é obrigatório")
	}

	// Load all documents with embeddings from MongoDB on each call.
	// This ensures newly uploaded files are immediately searchable.
	docs, err := s.fileRepo.ListAllWithEmbeddings(ctx)
	if err != nil {
		return nil, fmt.Errorf("carregando documentos para busca: %w", err)
	}

	s.embedder.IndexFromDocuments(docs)
	searchResults := s.embedder.Search(query, 3)

	out := make([]map[string]any, 0, len(searchResults))
	for _, r := range searchResults {
		out = append(out, map[string]any{
			"filename": r.Filename,
			"content":  r.Text,
			"score":    r.Score,
		})
	}
	return map[string]any{"results": out}, nil
}

var _ Skill = (*SearchDocumentsSkill)(nil)
