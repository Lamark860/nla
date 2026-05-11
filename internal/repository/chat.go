package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"nla/internal/model"
)

// ChatRepo persists chat sessions and their messages. CASCADE DELETE on
// chat_messages → chat_sessions handles message cleanup automatically.
type ChatRepo struct {
	pool *pgxpool.Pool
}

func NewChatRepo(pool *pgxpool.Pool) *ChatRepo {
	return &ChatRepo{pool: pool}
}

func (r *ChatRepo) CreateSession(ctx context.Context, s *model.ChatSession) error {
	err := r.pool.QueryRow(ctx, `
		INSERT INTO chat_sessions (session_id, title, agent_type)
		VALUES ($1, $2, $3)
		RETURNING created_at, updated_at`,
		s.SessionID, s.Title, s.AgentType).
		Scan(&s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return fmt.Errorf("insert session: %w", err)
	}
	return nil
}

func (r *ChatRepo) GetSession(ctx context.Context, sessionID string) (*model.ChatSession, error) {
	var s model.ChatSession
	err := r.pool.QueryRow(ctx, `
		SELECT session_id, title, agent_type, created_at, updated_at
		FROM chat_sessions WHERE session_id = $1`, sessionID).
		Scan(&s.SessionID, &s.Title, &s.AgentType, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("find session: %w", err)
	}
	return &s, nil
}

func (r *ChatRepo) ListSessions(ctx context.Context) ([]model.ChatSession, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT session_id, title, agent_type, created_at, updated_at
		FROM chat_sessions
		ORDER BY updated_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("find sessions: %w", err)
	}
	defer rows.Close()

	var results []model.ChatSession
	for rows.Next() {
		var s model.ChatSession
		if err := rows.Scan(&s.SessionID, &s.Title, &s.AgentType, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, err
		}
		results = append(results, s)
	}
	return results, rows.Err()
}

func (r *ChatRepo) DeleteSession(ctx context.Context, sessionID string) error {
	// CASCADE handles chat_messages cleanup.
	_, err := r.pool.Exec(ctx, `DELETE FROM chat_sessions WHERE session_id = $1`, sessionID)
	if err != nil {
		return fmt.Errorf("delete session: %w", err)
	}
	return nil
}

func (r *ChatRepo) UpdateSessionTimestamp(ctx context.Context, sessionID string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE chat_sessions SET updated_at = NOW() WHERE session_id = $1`, sessionID)
	return err
}

func (r *ChatRepo) AddMessage(ctx context.Context, msg *model.ChatMessage) error {
	err := r.pool.QueryRow(ctx, `
		INSERT INTO chat_messages (session_id, role, content)
		VALUES ($1, $2, $3)
		RETURNING created_at`,
		msg.SessionID, msg.Role, msg.Content).
		Scan(&msg.CreatedAt)
	if err != nil {
		return fmt.Errorf("insert message: %w", err)
	}
	return nil
}

func (r *ChatRepo) GetMessages(ctx context.Context, sessionID string) ([]model.ChatMessage, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT session_id, role, content, created_at
		FROM chat_messages WHERE session_id = $1
		ORDER BY created_at`, sessionID)
	if err != nil {
		return nil, fmt.Errorf("find messages: %w", err)
	}
	defer rows.Close()

	var results []model.ChatMessage
	for rows.Next() {
		var m model.ChatMessage
		if err := rows.Scan(&m.SessionID, &m.Role, &m.Content, &m.CreatedAt); err != nil {
			return nil, err
		}
		results = append(results, m)
	}
	return results, rows.Err()
}
