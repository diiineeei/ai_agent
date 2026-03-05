package repository

import (
	"context"
	"fmt"
	"time"

	"ai_agent/internal/model"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type SessionRepository interface {
	Save(ctx context.Context, sessionID string, history []model.Content) error
	Load(ctx context.Context, sessionID string) ([]model.Content, error)
	Delete(ctx context.Context, sessionID string) error
	GetAgentConfigID(ctx context.Context, sessionID string) (string, error)
	SetAgentConfigID(ctx context.Context, sessionID, agentConfigID string) error
	ListAll(ctx context.Context) ([]model.SessionSummary, error)
}

type sessionDoc struct {
	ID            string          `bson:"_id"`
	History       []model.Content `bson:"history"`
	AgentConfigID string          `bson:"agent_config_id,omitempty"`
	UpdatedAt     time.Time       `bson:"updated_at,omitempty"`
}

type MongoSessionRepository struct {
	coll *mongo.Collection
}

func NewMongoSessionRepository(coll *mongo.Collection) *MongoSessionRepository {
	return &MongoSessionRepository{coll: coll}
}

func (r *MongoSessionRepository) Save(ctx context.Context, sessionID string, history []model.Content) error {
	filter := bson.M{"_id": sessionID}
	update := bson.M{"$set": bson.M{
		"history":    history,
		"updated_at": time.Now(),
	}}
	_, err := r.coll.UpdateOne(ctx, filter, update, options.UpdateOne().SetUpsert(true))
	if err != nil {
		return fmt.Errorf("saving session %q: %w", sessionID, err)
	}
	return nil
}

func (r *MongoSessionRepository) Load(ctx context.Context, sessionID string) ([]model.Content, error) {
	var doc sessionDoc
	err := r.coll.FindOne(ctx, bson.M{"_id": sessionID}).Decode(&doc)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("loading session %q: %w", sessionID, err)
	}
	return doc.History, nil
}

func (r *MongoSessionRepository) Delete(ctx context.Context, sessionID string) error {
	_, err := r.coll.DeleteOne(ctx, bson.M{"_id": sessionID})
	if err != nil {
		return fmt.Errorf("deleting session %q: %w", sessionID, err)
	}
	return nil
}

func (r *MongoSessionRepository) GetAgentConfigID(ctx context.Context, sessionID string) (string, error) {
	var doc sessionDoc
	err := r.coll.FindOne(ctx, bson.M{"_id": sessionID}).Decode(&doc)
	if err == mongo.ErrNoDocuments {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("getting agent config id for session %q: %w", sessionID, err)
	}
	return doc.AgentConfigID, nil
}

func (r *MongoSessionRepository) SetAgentConfigID(ctx context.Context, sessionID, agentConfigID string) error {
	filter := bson.M{"_id": sessionID}
	update := bson.M{"$set": bson.M{"agent_config_id": agentConfigID}}
	_, err := r.coll.UpdateOne(ctx, filter, update, options.UpdateOne().SetUpsert(true))
	if err != nil {
		return fmt.Errorf("setting agent config id for session %q: %w", sessionID, err)
	}
	return nil
}

func (r *MongoSessionRepository) ListAll(ctx context.Context) ([]model.SessionSummary, error) {
	pipeline := mongo.Pipeline{
		bson.D{{Key: "$project", Value: bson.D{
			{Key: "agent_config_id", Value: 1},
			{Key: "updated_at", Value: 1},
			{Key: "message_count", Value: bson.D{
				{Key: "$size", Value: bson.D{
					{Key: "$ifNull", Value: bson.A{"$history", bson.A{}}},
				}},
			}},
		}}},
		bson.D{{Key: "$sort", Value: bson.D{
			{Key: "updated_at", Value: -1},
		}}},
	}
	cursor, err := r.coll.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("listing sessions: %w", err)
	}
	var summaries []model.SessionSummary
	if err := cursor.All(ctx, &summaries); err != nil {
		return nil, fmt.Errorf("decoding sessions: %w", err)
	}
	return summaries, nil
}
