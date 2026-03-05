package model

import "time"

type SessionSummary struct {
	ID            string    `bson:"_id"             json:"session_id"`
	Name          string    `bson:"name"            json:"name"`
	AgentConfigID string    `bson:"agent_config_id" json:"agent_config_id"`
	MessageCount  int       `bson:"message_count"   json:"message_count"`
	UpdatedAt     time.Time `bson:"updated_at"      json:"updated_at"`
}
