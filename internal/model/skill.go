package model

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type SkillDocument struct {
	ID          bson.ObjectID `bson:"_id,omitempty" json:"id"`
	Name        string             `bson:"name"          json:"name"`
	Description string             `bson:"description"   json:"description"`
	Enabled     bool               `bson:"enabled"       json:"enabled"`
	CreatedAt   time.Time          `bson:"created_at"    json:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at"    json:"updated_at"`
}
