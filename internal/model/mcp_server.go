package model

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// McpServer armazena a configuração de um servidor MCP externo.
type McpServer struct {
	ID          bson.ObjectID     `bson:"_id,omitempty" json:"id"`
	Name        string            `bson:"name"          json:"name"`
	Description string            `bson:"description"   json:"description"`
	Transport   string            `bson:"transport"     json:"transport"` // "stdio" | "http"
	Command     string            `bson:"command"       json:"command"`   // para stdio
	Args        []string          `bson:"args"          json:"args"`      // para stdio
	URL         string            `bson:"url"           json:"url"`       // para http
	Env         map[string]string `bson:"env"           json:"env"`
	Enabled     bool              `bson:"enabled"       json:"enabled"`
	CreatedAt   time.Time         `bson:"created_at"    json:"created_at"`
	UpdatedAt   time.Time         `bson:"updated_at"    json:"updated_at"`
}
