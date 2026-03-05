package agent

import (
	"hash/fnv"
	"math"
	"strings"

	"ai_agent/internal/model"
)

const embeddingSize = 512

// EmbeddedDocument holds a document and its embedding vector.
type EmbeddedDocument struct {
	Filename  string
	Text      string
	Embedding []float32
}

// SearchResult is returned by the Search method.
type SearchResult struct {
	Filename string
	Text     string
	Score    float32
}

// Embedder computes and searches document embeddings using FNV-hash feature hashing.
// This is a lightweight, dependency-free approach suited for keyword-adjacent matching.
type Embedder struct {
	documents []EmbeddedDocument
}

func NewEmbedder() *Embedder {
	return &Embedder{}
}

// Embed computes a normalized embedding vector for the given text.
func (e *Embedder) Embed(text string) []float64 {
	vec := e.embed(text)
	result := make([]float64, len(vec))
	for i, v := range vec {
		result[i] = float64(v)
	}
	return result
}

// IndexFromDocuments replaces the current index with documents from the database.
// Pre-computed embeddings are used directly, avoiding redundant computation.
func (e *Embedder) IndexFromDocuments(docs []model.FileDocument) {
	e.documents = make([]EmbeddedDocument, 0, len(docs))
	for _, d := range docs {
		vec := make([]float32, len(d.Embeddings))
		for i, v := range d.Embeddings {
			vec[i] = float32(v)
		}
		e.documents = append(e.documents, EmbeddedDocument{
			Filename:  d.Name,
			Text:      d.Content,
			Embedding: vec,
		})
	}
}

// Search returns the topK most relevant documents for the query.
func (e *Embedder) Search(query string, topK int) []SearchResult {
	queryVec := e.embed(query)
	type scored struct {
		doc   EmbeddedDocument
		score float32
	}
	scores := make([]scored, 0, len(e.documents))
	for _, doc := range e.documents {
		score := cosineSimilarity(queryVec, doc.Embedding)
		scores = append(scores, scored{doc, score})
	}
	// Simple selection sort for top-k (small N expected)
	for i := 0; i < len(scores) && i < topK; i++ {
		maxIdx := i
		for j := i + 1; j < len(scores); j++ {
			if scores[j].score > scores[maxIdx].score {
				maxIdx = j
			}
		}
		scores[i], scores[maxIdx] = scores[maxIdx], scores[i]
	}
	if topK > len(scores) {
		topK = len(scores)
	}
	results := make([]SearchResult, 0, topK)
	for i := 0; i < topK; i++ {
		results = append(results, SearchResult{
			Filename: scores[i].doc.Filename,
			Text:     scores[i].doc.Text,
			Score:    scores[i].score,
		})
	}
	return results
}

// embed computes a normalized FNV-hash feature vector for the text.
func (e *Embedder) embed(text string) []float32 {
	vec := make([]float32, embeddingSize)
	words := strings.Fields(strings.ToLower(text))
	for _, word := range words {
		h := fnv.New32a()
		h.Write([]byte(word))
		idx := int(h.Sum32()) % embeddingSize
		if idx < 0 {
			idx = -idx
		}
		vec[idx]++
	}
	return normalize(vec)
}

func normalize(vec []float32) []float32 {
	var sum float64
	for _, v := range vec {
		sum += float64(v * v)
	}
	if sum == 0 {
		return vec
	}
	norm := float32(math.Sqrt(sum))
	for i := range vec {
		vec[i] /= norm
	}
	return vec
}

func cosineSimilarity(a, b []float32) float32 {
	if len(a) != len(b) {
		return 0
	}
	var dot float64
	for i := range a {
		dot += float64(a[i]) * float64(b[i])
	}
	return float32(dot)
}
