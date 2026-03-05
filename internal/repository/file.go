package repository

import (
	"context"
	"fmt"

	"ai_agent/internal/model"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type FileRepository interface {
	Save(ctx context.Context, file model.FileDocument) (string, error)
	ListAll(ctx context.Context) ([]model.FileDocument, error)
	FindByID(ctx context.Context, id string) (*model.FileDocument, error)
	Delete(ctx context.Context, id string) error
	ListAllWithEmbeddings(ctx context.Context) ([]model.FileDocument, error)
}

type MongoFileRepository struct {
	coll *mongo.Collection
}

func NewMongoFileRepository(coll *mongo.Collection) *MongoFileRepository {
	return &MongoFileRepository{coll: coll}
}

func (r *MongoFileRepository) Save(ctx context.Context, file model.FileDocument) (string, error) {
	result, err := r.coll.InsertOne(ctx, file)
	if err != nil {
		return "", fmt.Errorf("saving file: %w", err)
	}
	id, ok := result.InsertedID.(interface{ Hex() string })
	if !ok {
		return fmt.Sprintf("%v", result.InsertedID), nil
	}
	return id.Hex(), nil
}

func (r *MongoFileRepository) ListAll(ctx context.Context) ([]model.FileDocument, error) {
	projection := bson.M{"content": 0, "embeddings": 0}
	cursor, err := r.coll.Find(ctx, bson.M{}, options.Find().SetProjection(projection))
	if err != nil {
		return nil, fmt.Errorf("listing files: %w", err)
	}
	var docs []model.FileDocument
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, fmt.Errorf("decoding files: %w", err)
	}
	return docs, nil
}

func (r *MongoFileRepository) FindByID(ctx context.Context, id string) (*model.FileDocument, error) {
	objID, err := parseObjectID(id)
	if err != nil {
		return nil, err
	}
	var doc model.FileDocument
	if err := r.coll.FindOne(ctx, bson.M{"_id": objID}).Decode(&doc); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, fmt.Errorf("finding file %q: %w", id, err)
	}
	return &doc, nil
}

func (r *MongoFileRepository) Delete(ctx context.Context, id string) error {
	objID, err := parseObjectID(id)
	if err != nil {
		return err
	}
	_, err = r.coll.DeleteOne(ctx, bson.M{"_id": objID})
	if err != nil {
		return fmt.Errorf("deleting file %q: %w", id, err)
	}
	return nil
}

func (r *MongoFileRepository) ListAllWithEmbeddings(ctx context.Context) ([]model.FileDocument, error) {
	cursor, err := r.coll.Find(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("listing files with embeddings: %w", err)
	}
	var docs []model.FileDocument
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, fmt.Errorf("decoding files with embeddings: %w", err)
	}
	return docs, nil
}

func parseObjectID(id string) (bson.ObjectID, error) {
	objID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return bson.ObjectID{}, fmt.Errorf("invalid file id %q: %w", id, err)
	}
	return objID, nil
}
