package repository

import (
	"context"
	"time"

	"ai_agent/internal/model"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type ChessRepository interface {
	Save(ctx context.Context, game model.ChessGame) error
	Load(ctx context.Context, sessionID string) (*model.ChessGame, error)
	Delete(ctx context.Context, sessionID string) error
}

type mongoChessRepository struct {
	coll *mongo.Collection
}

func NewMongoChessRepository(coll *mongo.Collection) ChessRepository {
	return &mongoChessRepository{coll: coll}
}

func (r *mongoChessRepository) Save(ctx context.Context, game model.ChessGame) error {
	game.UpdatedAt = time.Now()
	_, err := r.coll.UpdateOne(
		ctx,
		bson.M{"session_id": game.SessionID},
		bson.M{"$set": game},
		options.UpdateOne().SetUpsert(true),
	)
	return err
}

func (r *mongoChessRepository) Load(ctx context.Context, sessionID string) (*model.ChessGame, error) {
	var game model.ChessGame
	err := r.coll.FindOne(ctx, bson.M{"session_id": sessionID}).Decode(&game)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	return &game, err
}

func (r *mongoChessRepository) Delete(ctx context.Context, sessionID string) error {
	_, err := r.coll.DeleteOne(ctx, bson.M{"session_id": sessionID})
	return err
}
