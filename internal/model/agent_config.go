package model

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type AgentConfig struct {
	ID                bson.ObjectID `bson:"_id,omitempty"      json:"id"`
	Name              string        `bson:"name"               json:"name"`
	SystemInstruction string        `bson:"system_instruction" json:"system_instruction"`
	Model             string        `bson:"model"              json:"model"`
	EnabledSkills     []string      `bson:"enabled_skills"     json:"enabled_skills"`
	CreatedAt         time.Time     `bson:"created_at"         json:"created_at"`
	UpdatedAt         time.Time     `bson:"updated_at"         json:"updated_at"`
}
