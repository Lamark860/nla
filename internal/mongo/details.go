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

const dohodCacheTTLDays = 30

type DetailsRepo struct {
	col *mongo.Collection
}

func NewDetailsRepo(db *mongo.Database) *DetailsRepo {
	return &DetailsRepo{col: db.Collection("dohod_details")}
}

// Get returns cached dohod data for a bond. Returns nil if not found or stale (>30 days).
func (r *DetailsRepo) Get(ctx context.Context, isin string) (*model.DohodBondData, error) {
	var d model.DohodBondData
	err := r.col.FindOne(ctx, bson.M{"isin": isin}).Decode(&d)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, fmt.Errorf("find dohod details: %w", err)
	}

	// Check TTL
	if time.Since(d.UpdatedAt) > time.Duration(dohodCacheTTLDays)*24*time.Hour {
		return nil, nil // stale
	}

	return &d, nil
}

// Upsert inserts or updates dohod data for a bond
func (r *DetailsRepo) Upsert(ctx context.Context, d *model.DohodBondData) error {
	d.UpdatedAt = time.Now()
	opts := options.Update().SetUpsert(true)
	_, err := r.col.UpdateOne(ctx, bson.M{"isin": d.ISIN}, bson.M{"$set": d}, opts)
	if err != nil {
		return fmt.Errorf("upsert dohod details: %w", err)
	}
	return nil
}

// GetBySecid returns cached dohod data by secid
func (r *DetailsRepo) GetBySecid(ctx context.Context, secid string) (*model.DohodBondData, error) {
	var d model.DohodBondData
	err := r.col.FindOne(ctx, bson.M{"secid": secid}).Decode(&d)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, fmt.Errorf("find dohod details by secid: %w", err)
	}

	if time.Since(d.UpdatedAt) > time.Duration(dohodCacheTTLDays)*24*time.Hour {
		return nil, nil
	}

	return &d, nil
}
