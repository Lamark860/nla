package model

import "time"

// ChatSession represents a chat room with an AI agent
type ChatSession struct {
	SessionID string    `json:"session_id" bson:"session_id"`
	Title     string    `json:"title" bson:"title"`
	AgentType string    `json:"agent_type" bson:"agent_type"`
	CreatedAt time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time `json:"updated_at" bson:"updated_at"`
}

// ChatMessage represents a single message in a chat session
type ChatMessage struct {
	SessionID string    `json:"session_id" bson:"session_id"`
	Role      string    `json:"role" bson:"role"` // user, assistant, system
	Content   string    `json:"content" bson:"content"`
	CreatedAt time.Time `json:"created_at" bson:"created_at"`
}

// Agent represents a chat agent configuration
type Agent struct {
	Type        string `json:"type"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
	PromptFile  string `json:"-"`
}

// CreateSessionRequest for POST /chat/sessions
type CreateSessionRequest struct {
	AgentType string `json:"agent_type"`
	Title     string `json:"title"`
}

// SendMessageRequest for POST /chat/sessions/:id/messages
type SendMessageRequest struct {
	Content string `json:"content"`
}

// SendMessageResponse returned after sending a message
type SendMessageResponse struct {
	UserMessage      ChatMessage `json:"user_message"`
	AssistantMessage ChatMessage `json:"assistant_message"`
}
