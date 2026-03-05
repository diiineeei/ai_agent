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

type FeedbackRepository interface {
	Upsert(ctx context.Context, f model.Feedback) error
	GetForSession(ctx context.Context, sessionID string) ([]model.Feedback, error)
	StatsByAgent(ctx context.Context) ([]model.AgentFeedbackStats, error)
}

type MongoFeedbackRepository struct {
	coll *mongo.Collection
}

func NewMongoFeedbackRepository(coll *mongo.Collection) *MongoFeedbackRepository {
	return &MongoFeedbackRepository{coll: coll}
}

func (r *MongoFeedbackRepository) Upsert(ctx context.Context, f model.Feedback) error {
	filter := bson.M{"session_id": f.SessionID, "message_index": f.MessageIndex}
	update := bson.M{
		"$set": bson.M{
			"rating":          f.Rating,
			"agent_config_id": f.AgentConfigID,
		},
		"$setOnInsert": bson.M{
			"created_at": time.Now(),
		},
	}
	_, err := r.coll.UpdateOne(ctx, filter, update, options.UpdateOne().SetUpsert(true))
	if err != nil {
		return fmt.Errorf("upserting feedback: %w", err)
	}
	return nil
}

func (r *MongoFeedbackRepository) GetForSession(ctx context.Context, sessionID string) ([]model.Feedback, error) {
	cursor, err := r.coll.Find(ctx, bson.M{"session_id": sessionID})
	if err != nil {
		return nil, fmt.Errorf("getting feedback for session %q: %w", sessionID, err)
	}
	var feedback []model.Feedback
	if err := cursor.All(ctx, &feedback); err != nil {
		return nil, fmt.Errorf("decoding feedback: %w", err)
	}
	return feedback, nil
}

func (r *MongoFeedbackRepository) StatsByAgent(ctx context.Context) ([]model.AgentFeedbackStats, error) {
	pipeline := mongo.Pipeline{
		bson.D{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$agent_config_id"},
			{Key: "thumbs_up", Value: bson.D{{Key: "$sum", Value: bson.D{
				{Key: "$cond", Value: bson.A{
					bson.D{{Key: "$eq", Value: bson.A{"$rating", "up"}}},
					1, 0,
				}},
			}}}},
			{Key: "thumbs_down", Value: bson.D{{Key: "$sum", Value: bson.D{
				{Key: "$cond", Value: bson.A{
					bson.D{{Key: "$eq", Value: bson.A{"$rating", "down"}}},
					1, 0,
				}},
			}}}},
		}}},
		bson.D{{Key: "$sort", Value: bson.D{{Key: "_id", Value: 1}}}},
	}
	cursor, err := r.coll.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("aggregating feedback stats: %w", err)
	}
	var stats []model.AgentFeedbackStats
	if err := cursor.All(ctx, &stats); err != nil {
		return nil, fmt.Errorf("decoding feedback stats: %w", err)
	}
	return stats, nil
}
