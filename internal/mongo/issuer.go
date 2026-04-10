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

// ToggleHidden sets is_hidden for all bonds of an emitter
func (r *IssuerRepo) ToggleHidden(ctx context.Context, emitterID int64, hidden bool) (int64, error) {
	res, err := r.col.UpdateMany(ctx,
		bson.M{"emitter_id": emitterID},
		bson.M{"$set": bson.M{"is_hidden": hidden}},
	)
	if err != nil {
		return 0, fmt.Errorf("toggle issuer %d: %w", emitterID, err)
	}
	return res.ModifiedCount, nil
}

// GetHiddenEmitterIDs returns all emitter IDs that are hidden
func (r *IssuerRepo) GetHiddenEmitterIDs(ctx context.Context) ([]int64, error) {
	cursor, err := r.col.Distinct(ctx, "emitter_id", bson.M{"is_hidden": true})
	if err != nil {
		return nil, err
	}
	ids := make([]int64, 0)
	for _, v := range cursor {
		switch n := v.(type) {
		case int32:
			ids = append(ids, int64(n))
		case int64:
			ids = append(ids, n)
		case float64:
			ids = append(ids, int64(n))
		}
	}
	return ids, nil
}

// UpdateEmitterName sets emitter_name for all bonds of a given emitter_id
func (r *IssuerRepo) UpdateEmitterName(ctx context.Context, emitterID int64, name string) error {
	_, err := r.col.UpdateMany(ctx,
		bson.M{"emitter_id": emitterID, "emitter_name": ""},
		bson.M{"$set": bson.M{"emitter_name": name, "updated_at": time.Now()}},
	)
	if err != nil {
		return fmt.Errorf("update emitter name %d: %w", emitterID, err)
	}
	return nil
}

// Upsert creates or updates bond_issuer by secid
func (r *IssuerRepo) Upsert(ctx context.Context, issuer *model.BondIssuer) error {
	now := time.Now()
	opts := options.Update().SetUpsert(true)
	_, err := r.col.UpdateOne(ctx,
		bson.M{"secid": issuer.SECID},
		bson.M{
			"$set": bson.M{
				"emitter_id":   issuer.EmitterID,
				"emitter_name": issuer.EmitterName,
				"updated_at":   now,
			},
			"$setOnInsert": bson.M{
				"secid":      issuer.SECID,
				"is_hidden":  false,
				"needs_sync": true,
				"created_at": now,
			},
		},
		opts,
	)
	if err != nil {
		return fmt.Errorf("upsert issuer %s: %w", issuer.SECID, err)
	}
	return nil
}

// GetAllSecids returns all secids present in the collection
func (r *IssuerRepo) GetAllSecids(ctx context.Context) (map[string]bool, error) {
	cursor, err := r.col.Find(ctx, bson.M{}, options.Find().SetProjection(bson.M{"secid": 1}))
	if err != nil {
		return nil, fmt.Errorf("find all secids: %w", err)
	}
	defer cursor.Close(ctx)

	result := make(map[string]bool)
	for cursor.Next(ctx) {
		var doc struct {
			SECID string `bson:"secid"`
		}
		if err := cursor.Decode(&doc); err != nil {
			continue
		}
		result[doc.SECID] = true
	}
	return result, nil
}

// GetOneSecidPerEmitter returns one sample secid for each emitter_id.
// Used for bulk rating sync — we only need one bond per emitter to fetch dohod.ru ratings.
func (r *IssuerRepo) GetOneSecidPerEmitter(ctx context.Context) (map[int64]string, error) {
	pipeline := bson.A{
		bson.M{"$match": bson.M{"emitter_id": bson.M{"$gt": 0}, "is_hidden": bson.M{"$ne": true}}},
		bson.M{"$group": bson.M{
			"_id":   "$emitter_id",
			"secid": bson.M{"$first": "$secid"},
		}},
	}
	cursor, err := r.col.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("aggregate emitters: %w", err)
	}
	defer cursor.Close(ctx)

	result := make(map[int64]string)
	for cursor.Next(ctx) {
		var doc struct {
			EmitterID int64  `bson:"_id"`
			SECID     string `bson:"secid"`
		}
		if err := cursor.Decode(&doc); err != nil {
			continue
		}
		if doc.EmitterID > 0 && doc.SECID != "" {
			result[doc.EmitterID] = doc.SECID
		}
	}
	return result, nil
}
