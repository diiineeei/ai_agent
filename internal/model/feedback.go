package model

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type Feedback struct {
	ID            bson.ObjectID `bson:"_id,omitempty"   json:"id"`
	SessionID     string        `bson:"session_id"      json:"session_id"`
	MessageIndex  int           `bson:"message_index"   json:"message_index"`
	AgentConfigID string        `bson:"agent_config_id" json:"agent_config_id"`
	Rating        string        `bson:"rating"          json:"rating"` // "up" | "down"
	CreatedAt     time.Time     `bson:"created_at"      json:"created_at"`
}

type AgentFeedbackStats struct {
	AgentConfigID string `bson:"_id"         json:"agent_config_id"`
	ThumbsUp      int    `bson:"thumbs_up"   json:"thumbs_up"`
	ThumbsDown    int    `bson:"thumbs_down" json:"thumbs_down"`
}
