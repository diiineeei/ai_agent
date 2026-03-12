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

type McpServerRepository interface {
	ListAll(ctx context.Context) ([]model.McpServer, error)
	GetByID(ctx context.Context, id bson.ObjectID) (*model.McpServer, error)
	Create(ctx context.Context, s model.McpServer) (*model.McpServer, error)
	Update(ctx context.Context, id bson.ObjectID, s model.McpServer) (*model.McpServer, error)
	Delete(ctx context.Context, id bson.ObjectID) error
	Toggle(ctx context.Context, id bson.ObjectID) (*model.McpServer, error)
}

type MongoMcpServerRepository struct {
	coll *mongo.Collection
}

func NewMongoMcpServerRepository(coll *mongo.Collection) *MongoMcpServerRepository {
	return &MongoMcpServerRepository{coll: coll}
}

func (r *MongoMcpServerRepository) ListAll(ctx context.Context) ([]model.McpServer, error) {
	cursor, err := r.coll.Find(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("listing mcp servers: %w", err)
	}
	var docs []model.McpServer
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, fmt.Errorf("decoding mcp servers: %w", err)
	}
	return docs, nil
}

func (r *MongoMcpServerRepository) GetByID(ctx context.Context, id bson.ObjectID) (*model.McpServer, error) {
	var doc model.McpServer
	if err := r.coll.FindOne(ctx, bson.M{"_id": id}).Decode(&doc); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("mcp server %q not found", id.Hex())
		}
		return nil, fmt.Errorf("finding mcp server: %w", err)
	}
	return &doc, nil
}

func (r *MongoMcpServerRepository) Create(ctx context.Context, s model.McpServer) (*model.McpServer, error) {
	now := time.Now()
	s.ID = bson.NewObjectID()
	s.CreatedAt = now
	s.UpdatedAt = now
	if s.Args == nil {
		s.Args = []string{}
	}
	if s.Env == nil {
		s.Env = map[string]string{}
	}

	if _, err := r.coll.InsertOne(ctx, s); err != nil {
		return nil, fmt.Errorf("inserting mcp server: %w", err)
	}
	return &s, nil
}

func (r *MongoMcpServerRepository) Update(ctx context.Context, id bson.ObjectID, s model.McpServer) (*model.McpServer, error) {
	if s.Args == nil {
		s.Args = []string{}
	}
	if s.Env == nil {
		s.Env = map[string]string{}
	}

	update := bson.M{"$set": bson.M{
		"name":        s.Name,
		"description": s.Description,
		"transport":   s.Transport,
		"command":     s.Command,
		"args":        s.Args,
		"url":         s.URL,
		"env":         s.Env,
		"enabled":     s.Enabled,
		"updated_at":  time.Now(),
	}}
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	var updated model.McpServer
	if err := r.coll.FindOneAndUpdate(ctx, bson.M{"_id": id}, update, opts).Decode(&updated); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("mcp server %q not found", id.Hex())
		}
		return nil, fmt.Errorf("updating mcp server: %w", err)
	}
	return &updated, nil
}

func (r *MongoMcpServerRepository) Delete(ctx context.Context, id bson.ObjectID) error {
	res, err := r.coll.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return fmt.Errorf("deleting mcp server: %w", err)
	}
	if res.DeletedCount == 0 {
		return fmt.Errorf("mcp server %q not found", id.Hex())
	}
	return nil
}

func (r *MongoMcpServerRepository) Toggle(ctx context.Context, id bson.ObjectID) (*model.McpServer, error) {
	var current model.McpServer
	if err := r.coll.FindOne(ctx, bson.M{"_id": id}).Decode(&current); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("mcp server %q not found", id.Hex())
		}
		return nil, fmt.Errorf("finding mcp server: %w", err)
	}
	update := bson.M{"$set": bson.M{
		"enabled":    !current.Enabled,
		"updated_at": time.Now(),
	}}
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	var updated model.McpServer
	if err := r.coll.FindOneAndUpdate(ctx, bson.M{"_id": id}, update, opts).Decode(&updated); err != nil {
		return nil, fmt.Errorf("toggling mcp server: %w", err)
	}
	return &updated, nil
}
