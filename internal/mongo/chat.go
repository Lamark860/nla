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

type ChatRepo struct {
	sessions *mongo.Collection
	messages *mongo.Collection
}

func NewChatRepo(db *mongo.Database) *ChatRepo {
	repo := &ChatRepo{
		sessions: db.Collection("chat_sessions"),
		messages: db.Collection("chat_messages"),
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	repo.messages.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "session_id", Value: 1}, {Key: "created_at", Value: 1}},
	})
	return repo
}

// CreateSession inserts a new chat session.
func (r *ChatRepo) CreateSession(ctx context.Context, s *model.ChatSession) error {
	_, err := r.sessions.InsertOne(ctx, s)
	if err != nil {
		return fmt.Errorf("insert session: %w", err)
	}
	return nil
}

// GetSession returns a single session by ID.
func (r *ChatRepo) GetSession(ctx context.Context, sessionID string) (*model.ChatSession, error) {
	var s model.ChatSession
	err := r.sessions.FindOne(ctx, bson.M{"session_id": sessionID}).Decode(&s)
	if err != nil {
		return nil, fmt.Errorf("find session: %w", err)
	}
	return &s, nil
}

// ListSessions returns all sessions, newest first.
func (r *ChatRepo) ListSessions(ctx context.Context) ([]model.ChatSession, error) {
	opts := options.Find().SetSort(bson.D{{Key: "updated_at", Value: -1}})
	cursor, err := r.sessions.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, fmt.Errorf("find sessions: %w", err)
	}
	defer cursor.Close(ctx)

	var results []model.ChatSession
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}
	return results, nil
}

// DeleteSession removes a session and all its messages.
func (r *ChatRepo) DeleteSession(ctx context.Context, sessionID string) error {
	_, err := r.sessions.DeleteOne(ctx, bson.M{"session_id": sessionID})
	if err != nil {
		return fmt.Errorf("delete session: %w", err)
	}
	_, err = r.messages.DeleteMany(ctx, bson.M{"session_id": sessionID})
	if err != nil {
		return fmt.Errorf("delete messages: %w", err)
	}
	return nil
}

// UpdateSessionTimestamp updates the updatedAt timestamp.
func (r *ChatRepo) UpdateSessionTimestamp(ctx context.Context, sessionID string) error {
	_, err := r.sessions.UpdateOne(ctx,
		bson.M{"session_id": sessionID},
		bson.M{"$set": bson.M{"updated_at": time.Now()}},
	)
	return err
}

// AddMessage inserts a chat message.
func (r *ChatRepo) AddMessage(ctx context.Context, msg *model.ChatMessage) error {
	_, err := r.messages.InsertOne(ctx, msg)
	if err != nil {
		return fmt.Errorf("insert message: %w", err)
	}
	return nil
}

// GetMessages returns all messages for a session, ordered by time.
func (r *ChatRepo) GetMessages(ctx context.Context, sessionID string) ([]model.ChatMessage, error) {
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: 1}})
	cursor, err := r.messages.Find(ctx, bson.M{"session_id": sessionID}, opts)
	if err != nil {
		return nil, fmt.Errorf("find messages: %w", err)
	}
	defer cursor.Close(ctx)

	var results []model.ChatMessage
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}
	return results, nil
}
