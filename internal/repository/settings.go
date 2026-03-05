package repository

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type SettingsRepository interface {
	GetSystemInstruction(ctx context.Context) (string, error)
	SetSystemInstruction(ctx context.Context, value string) error
}

type MongoSettingsRepository struct {
	col *mongo.Collection
}

func NewMongoSettingsRepository(col *mongo.Collection) *MongoSettingsRepository {
	return &MongoSettingsRepository{col: col}
}

func (r *MongoSettingsRepository) GetSystemInstruction(ctx context.Context) (string, error) {
	var doc struct {
		Value string `bson:"value"`
	}
	err := r.col.FindOne(ctx, bson.M{"_id": "system_instruction"}).Decode(&doc)
	if err == mongo.ErrNoDocuments {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return doc.Value, nil
}

func (r *MongoSettingsRepository) SetSystemInstruction(ctx context.Context, value string) error {
	filter := bson.M{"_id": "system_instruction"}
	update := bson.M{"$set": bson.M{"value": value}}
	_, err := r.col.UpdateOne(ctx, filter, update, options.UpdateOne().SetUpsert(true))
	return err
}
