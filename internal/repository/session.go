package repository

import (
	"context"
	"fmt"

	"ai_agent/internal/model"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type SessionRepository interface {
	Save(ctx context.Context, sessionID string, history []model.Content) error
	Load(ctx context.Context, sessionID string) ([]model.Content, error)
	Delete(ctx context.Context, sessionID string) error
}

type sessionDoc struct {
	ID      string          `bson:"_id"`
	History []model.Content `bson:"history"`
}

type MongoSessionRepository struct {
	coll *mongo.Collection
}

func NewMongoSessionRepository(coll *mongo.Collection) *MongoSessionRepository {
	return &MongoSessionRepository{coll: coll}
}

func (r *MongoSessionRepository) Save(ctx context.Context, sessionID string, history []model.Content) error {
	filter := bson.M{"_id": sessionID}
	update := bson.M{"$set": bson.M{"history": history}}
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
