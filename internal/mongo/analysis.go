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

type AnalysisRepo struct {
	col *mongo.Collection
}

func NewAnalysisRepo(db *mongo.Database) *AnalysisRepo {
	return &AnalysisRepo{col: db.Collection("bond_analyses")}
}

func (r *AnalysisRepo) Save(ctx context.Context, a *model.BondAnalysis) error {
	a.SavedAt = time.Now()
	if a.Timestamp.IsZero() {
		a.Timestamp = time.Now()
	}
	_, err := r.col.InsertOne(ctx, a)
	if err != nil {
		return fmt.Errorf("save analysis: %w", err)
	}
	return nil
}

func (r *AnalysisRepo) GetBySecid(ctx context.Context, secid string) ([]model.BondAnalysis, error) {
	opts := options.Find().SetSort(bson.D{{Key: "timestamp", Value: -1}})
	cursor, err := r.col.Find(ctx, bson.M{"secid": secid}, opts)
	if err != nil {
		return nil, fmt.Errorf("find analyses: %w", err)
	}
	defer cursor.Close(ctx)

	var results []model.BondAnalysis
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}
	return results, nil
}

func (r *AnalysisRepo) GetByID(ctx context.Context, id string) (*model.BondAnalysis, error) {
	var a model.BondAnalysis
	err := r.col.FindOne(ctx, bson.M{"_id": id}).Decode(&a)
	if err != nil {
		return nil, fmt.Errorf("analysis not found: %w", err)
	}
	return &a, nil
}

func (r *AnalysisRepo) GetStats(ctx context.Context, secid string) (*model.AnalysisStats, error) {
	analyses, err := r.GetBySecid(ctx, secid)
	if err != nil {
		return nil, err
	}

	stats := &model.AnalysisStats{Total: len(analyses)}
	if len(analyses) > 0 {
		stats.LastAnalysis = &analyses[0].Timestamp
		sum := 0.0
		count := 0
		for _, a := range analyses {
			if a.Rating != nil {
				sum += float64(*a.Rating)
				count++
			}
		}
		if count > 0 {
			stats.AvgRating = sum / float64(count)
		}
	}
	return stats, nil
}

// GetLatestRatings returns the most recent rating per SECID (batch)
func (r *AnalysisRepo) GetLatestRatings(ctx context.Context, secids []string) (map[string]int, error) {
	if len(secids) == 0 {
		return map[string]int{}, nil
	}

	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{"secid": bson.M{"$in": secids}, "rating": bson.M{"$ne": nil}}}},
		{{Key: "$sort", Value: bson.D{{Key: "timestamp", Value: -1}}}},
		{{Key: "$group", Value: bson.M{
			"_id":    "$secid",
			"rating": bson.M{"$first": "$rating"},
		}}},
	}

	cursor, err := r.col.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	result := make(map[string]int)
	for cursor.Next(ctx) {
		var doc struct {
			SECID  string `bson:"_id"`
			Rating int    `bson:"rating"`
		}
		if cursor.Decode(&doc) == nil {
			result[doc.SECID] = doc.Rating
		}
	}
	return result, nil
}
