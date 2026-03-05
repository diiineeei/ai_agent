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

type SkillRepository interface {
	ListAll(ctx context.Context) ([]model.SkillDocument, error)
	ListEnabled(ctx context.Context) ([]model.SkillDocument, error)
	Toggle(ctx context.Context, name string) (*model.SkillDocument, error)
	Seed(ctx context.Context, skill model.SkillDocument) error
}

type MongoSkillRepository struct {
	coll *mongo.Collection
}

func NewMongoSkillRepository(coll *mongo.Collection) *MongoSkillRepository {
	return &MongoSkillRepository{coll: coll}
}

func (r *MongoSkillRepository) ListAll(ctx context.Context) ([]model.SkillDocument, error) {
	cursor, err := r.coll.Find(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("listing skills: %w", err)
	}
	var docs []model.SkillDocument
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, fmt.Errorf("decoding skills: %w", err)
	}
	return docs, nil
}

func (r *MongoSkillRepository) ListEnabled(ctx context.Context) ([]model.SkillDocument, error) {
	cursor, err := r.coll.Find(ctx, bson.M{"enabled": true})
	if err != nil {
		return nil, fmt.Errorf("listing enabled skills: %w", err)
	}
	var docs []model.SkillDocument
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, fmt.Errorf("decoding enabled skills: %w", err)
	}
	return docs, nil
}

func (r *MongoSkillRepository) Toggle(ctx context.Context, name string) (*model.SkillDocument, error) {
	var current model.SkillDocument
	if err := r.coll.FindOne(ctx, bson.M{"name": name}).Decode(&current); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("skill %q not found", name)
		}
		return nil, fmt.Errorf("finding skill %q: %w", name, err)
	}

	update := bson.M{"$set": bson.M{
		"enabled":    !current.Enabled,
		"updated_at": time.Now(),
	}}
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	var updated model.SkillDocument
	err := r.coll.FindOneAndUpdate(ctx, bson.M{"name": name}, update, opts).Decode(&updated)
	if err != nil {
		return nil, fmt.Errorf("toggling skill %q: %w", name, err)
	}
	return &updated, nil
}

func (r *MongoSkillRepository) Seed(ctx context.Context, skill model.SkillDocument) error {
	now := time.Now()
	filter := bson.M{"name": skill.Name}
	update := bson.M{
		"$setOnInsert": bson.M{
			"name":        skill.Name,
			"description": skill.Description,
			"enabled":     skill.Enabled,
			"created_at":  now,
			"updated_at":  now,
		},
	}
	_, err := r.coll.UpdateOne(ctx, filter, update, options.UpdateOne().SetUpsert(true))
	if err != nil {
		return fmt.Errorf("seeding skill %q: %w", skill.Name, err)
	}
	return nil
}
