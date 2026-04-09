package mongo

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"nla/internal/model"
)

type IssuerRepo struct {
	col *mongo.Collection
}

func NewIssuerRepo(db *mongo.Database) *IssuerRepo {
	return &IssuerRepo{col: db.Collection("bond_issuers")}
}

// GetAll returns all issuer mappings
func (r *IssuerRepo) GetAll(ctx context.Context) ([]model.BondIssuer, error) {
	cursor, err := r.col.Find(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("find issuers: %w", err)
	}
	defer cursor.Close(ctx)

	var results []model.BondIssuer
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}
	return results, nil
}

// GetBySecid returns issuer mapping for a single bond
func (r *IssuerRepo) GetBySecid(ctx context.Context, secid string) (*model.BondIssuer, error) {
	var issuer model.BondIssuer
	err := r.col.FindOne(ctx, bson.M{"secid": secid}).Decode(&issuer)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, fmt.Errorf("find issuer by secid: %w", err)
	}
	return &issuer, nil
}

// GetBySecids returns issuer mappings for multiple bonds
func (r *IssuerRepo) GetBySecids(ctx context.Context, secids []string) (map[string]model.BondIssuer, error) {
	cursor, err := r.col.Find(ctx, bson.M{"secid": bson.M{"$in": secids}})
	if err != nil {
		return nil, fmt.Errorf("find issuers by secids: %w", err)
	}
	defer cursor.Close(ctx)

	result := make(map[string]model.BondIssuer, len(secids))
	for cursor.Next(ctx) {
		var issuer model.BondIssuer
		if err := cursor.Decode(&issuer); err != nil {
			continue
		}
		result[issuer.SECID] = issuer
	}
	return result, nil
}
