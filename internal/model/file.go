package model

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type FileDocument struct {
	ID          bson.ObjectID `bson:"_id,omitempty" json:"id"`
	Name        string        `bson:"name"          json:"name"`
	ContentType string        `bson:"content_type"  json:"content_type"`
	Size        int64         `bson:"size"          json:"size"`
	Content     string        `bson:"content"       json:"content,omitempty"`
	Embeddings  []float64     `bson:"embeddings"    json:"embeddings,omitempty"`
	UploadedAt  time.Time     `bson:"uploaded_at"   json:"uploaded_at"`
}

// FileListItem is used in list responses to omit large fields.
type FileListItem struct {
	ID          bson.ObjectID `json:"id"`
	Name        string             `json:"name"`
	ContentType string             `json:"content_type"`
	Size        int64              `json:"size"`
	UploadedAt  time.Time          `json:"uploaded_at"`
}
