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

type QueueRepo struct {
	col *mongo.Collection
}

func NewQueueRepo(db *mongo.Database) *QueueRepo {
	return &QueueRepo{col: db.Collection("queue_jobs")}
}

func (r *QueueRepo) Create(ctx context.Context, job *model.QueueJob) error {
	job.CreatedAt = time.Now()
	job.UpdatedAt = time.Now()
	_, err := r.col.InsertOne(ctx, job)
	if err != nil {
		return fmt.Errorf("create job: %w", err)
	}
	return nil
}

func (r *QueueRepo) GetByID(ctx context.Context, id string) (*model.QueueJob, error) {
	var job model.QueueJob
	err := r.col.FindOne(ctx, bson.M{"_id": id}).Decode(&job)
	if err != nil {
		return nil, fmt.Errorf("job not found: %w", err)
	}
	return &job, nil
}

func (r *QueueRepo) MarkRunning(ctx context.Context, id string) error {
	return r.updateStatus(ctx, id, model.JobStatusRunning, nil, "")
}

func (r *QueueRepo) MarkDone(ctx context.Context, id string, result interface{}) error {
	return r.updateStatus(ctx, id, model.JobStatusDone, result, "")
}

func (r *QueueRepo) MarkError(ctx context.Context, id string, errMsg string) error {
	return r.updateStatus(ctx, id, model.JobStatusError, nil, errMsg)
}

func (r *QueueRepo) updateStatus(ctx context.Context, id string, status string, result interface{}, errMsg string) error {
	update := bson.M{
		"$set": bson.M{
			"status":     status,
			"updated_at": time.Now(),
		},
	}
	if result != nil {
		update["$set"].(bson.M)["result"] = result
	}
	if errMsg != "" {
		update["$set"].(bson.M)["error"] = errMsg
	}

	_, err := r.col.UpdateOne(ctx, bson.M{"_id": id}, update)
	return err
}

// FindPending finds a pending job with the same type and secid (for dedup)
func (r *QueueRepo) FindPending(ctx context.Context, jobType string, secid string) (*model.QueueJob, error) {
	filter := bson.M{
		"type":   jobType,
		"secid":  secid,
		"status": bson.M{"$in": []string{model.JobStatusPending, model.JobStatusRunning}},
	}
	var job model.QueueJob
	err := r.col.FindOne(ctx, filter).Decode(&job)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &job, nil
}

// FetchPending atomically claims the next pending job
func (r *QueueRepo) FetchPending(ctx context.Context) (*model.QueueJob, error) {
	opts := options.FindOneAndUpdate().
		SetSort(bson.D{{Key: "created_at", Value: 1}}).
		SetReturnDocument(options.After)

	filter := bson.M{"status": model.JobStatusPending}
	update := bson.M{
		"$set": bson.M{
			"status":     model.JobStatusRunning,
			"updated_at": time.Now(),
		},
	}

	var job model.QueueJob
	err := r.col.FindOneAndUpdate(ctx, filter, update, opts).Decode(&job)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &job, nil
}

// GetStats returns counts by status
func (r *QueueRepo) GetStats(ctx context.Context) (map[string]int, error) {
	pipeline := mongo.Pipeline{
		{{Key: "$group", Value: bson.M{
			"_id":   "$status",
			"count": bson.M{"$sum": 1},
		}}},
	}

	cursor, err := r.col.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	stats := make(map[string]int)
	for cursor.Next(ctx) {
		var doc struct {
			Status string `bson:"_id"`
			Count  int    `bson:"count"`
		}
		if cursor.Decode(&doc) == nil {
			stats[doc.Status] = doc.Count
		}
	}
	return stats, nil
}
