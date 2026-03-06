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

type AgentConfigRepository interface {
	List(ctx context.Context) ([]model.AgentConfig, error)
	GetByID(ctx context.Context, id string) (*model.AgentConfig, error)
	Create(ctx context.Context, cfg model.AgentConfig) (*model.AgentConfig, error)
	Update(ctx context.Context, id string, cfg model.AgentConfig) (*model.AgentConfig, error)
	Delete(ctx context.Context, id string) error
	CountAll(ctx context.Context) (int64, error)
}

type MongoAgentConfigRepository struct {
	coll *mongo.Collection
}

func NewMongoAgentConfigRepository(coll *mongo.Collection) *MongoAgentConfigRepository {
	return &MongoAgentConfigRepository{coll: coll}
}

func (r *MongoAgentConfigRepository) List(ctx context.Context) ([]model.AgentConfig, error) {
	cursor, err := r.coll.Find(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("listing agent configs: %w", err)
	}
	var docs []model.AgentConfig
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, fmt.Errorf("decoding agent configs: %w", err)
	}
	return docs, nil
}

func (r *MongoAgentConfigRepository) GetByID(ctx context.Context, id string) (*model.AgentConfig, error) {
	objID, err := parseObjectID(id)
	if err != nil {
		return nil, err
	}
	var doc model.AgentConfig
	if err := r.coll.FindOne(ctx, bson.M{"_id": objID}).Decode(&doc); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, fmt.Errorf("finding agent config %q: %w", id, err)
	}
	return &doc, nil
}

func (r *MongoAgentConfigRepository) Create(ctx context.Context, cfg model.AgentConfig) (*model.AgentConfig, error) {
	now := time.Now()
	cfg.CreatedAt = now
	cfg.UpdatedAt = now
	result, err := r.coll.InsertOne(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("creating agent config: %w", err)
	}
	id, ok := result.InsertedID.(bson.ObjectID)
	if !ok {
		return nil, fmt.Errorf("unexpected inserted ID type")
	}
	return r.GetByID(ctx, id.Hex())
}

func (r *MongoAgentConfigRepository) Update(ctx context.Context, id string, cfg model.AgentConfig) (*model.AgentConfig, error) {
	objID, err := parseObjectID(id)
	if err != nil {
		return nil, err
	}
	update := bson.M{"$set": bson.M{
		"name":               cfg.Name,
		"avatar":             cfg.Avatar,
		"system_instruction": cfg.SystemInstruction,
		"model":              cfg.Model,
		"provider":           cfg.Provider,
		"base_url":           cfg.BaseURL,
		"enabled_skills":     cfg.EnabledSkills,
		"updated_at":         time.Now(),
	}}
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	var updated model.AgentConfig
	err = r.coll.FindOneAndUpdate(ctx, bson.M{"_id": objID}, update, opts).Decode(&updated)
	if err != nil {
		return nil, fmt.Errorf("updating agent config %q: %w", id, err)
	}
	return &updated, nil
}

func (r *MongoAgentConfigRepository) Delete(ctx context.Context, id string) error {
	objID, err := parseObjectID(id)
	if err != nil {
		return err
	}
	_, err = r.coll.DeleteOne(ctx, bson.M{"_id": objID})
	if err != nil {
		return fmt.Errorf("deleting agent config %q: %w", id, err)
	}
	return nil
}

func (r *MongoAgentConfigRepository) CountAll(ctx context.Context) (int64, error) {
	count, err := r.coll.CountDocuments(ctx, bson.M{})
	if err != nil {
		return 0, fmt.Errorf("counting agent configs: %w", err)
	}
	return count, nil
}
