package mongo

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"nla/internal/model"
)

type RatingRepo struct {
	col *mongo.Collection
}

func NewRatingRepo(db *mongo.Database) *RatingRepo {
	repo := &RatingRepo{col: db.Collection("issuer_ratings")}
	// Create index on issuer for fast lookup
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	repo.col.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "issuer", Value: 1}},
		Options: options.Index().SetBackground(true),
	})
	return repo
}

func (r *RatingRepo) Upsert(ctx context.Context, rating *model.IssuerRating) error {
	rating.UpdatedAt = time.Now()
	filter := bson.M{"issuer": rating.Issuer, "agency": rating.Agency}
	update := bson.M{"$set": rating}
	opts := options.Update().SetUpsert(true)
	_, err := r.col.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return fmt.Errorf("upsert rating: %w", err)
	}
	return nil
}

func (r *RatingRepo) GetByIssuer(ctx context.Context, issuer string) ([]model.IssuerRating, error) {
	cursor, err := r.col.Find(ctx, bson.M{"issuer": issuer})
	if err != nil {
		return nil, fmt.Errorf("find ratings: %w", err)
	}
	defer cursor.Close(ctx)

	var results []model.IssuerRating
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}
	return results, nil
}

func (r *RatingRepo) GetAll(ctx context.Context) ([]model.IssuerRating, error) {
	opts := options.Find().SetSort(bson.D{{Key: "issuer", Value: 1}})
	cursor, err := r.col.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, fmt.Errorf("find all ratings: %w", err)
	}
	defer cursor.Close(ctx)

	var results []model.IssuerRating
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}
	return results, nil
}

func (r *RatingRepo) BulkUpsert(ctx context.Context, ratings []model.IssuerRating) error {
	if len(ratings) == 0 {
		return nil
	}

	models := make([]mongo.WriteModel, 0, len(ratings))
	for i := range ratings {
		ratings[i].UpdatedAt = time.Now()
		filter := bson.M{"issuer": ratings[i].Issuer, "agency": ratings[i].Agency}
		update := bson.M{"$set": ratings[i]}
		models = append(models, mongo.NewUpdateOneModel().SetFilter(filter).SetUpdate(update).SetUpsert(true))
	}

	_, err := r.col.BulkWrite(ctx, models)
	if err != nil {
		return fmt.Errorf("bulk upsert ratings: %w", err)
	}
	return nil
}

func (r *RatingRepo) Delete(ctx context.Context, issuer, agency string) error {
	_, err := r.col.DeleteOne(ctx, bson.M{"issuer": issuer, "agency": agency})
	if err != nil {
		return fmt.Errorf("delete rating: %w", err)
	}
	return nil
}
