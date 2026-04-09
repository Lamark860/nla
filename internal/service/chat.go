package service

import (
	"context"
	"crypto/rand"
	"fmt"
	"time"

	"nla/internal/client/openai"
	"nla/internal/model"
	mongoRepo "nla/internal/mongo"
)

// Predefined agents
var defaultAgents = []model.Agent{
	{
		Type:        "analyst",
		Name:        "Финансовый аналитик",
		Description: "Анализ облигаций, кредитные рейтинги, стратегии",
		Icon:        "chart-bar",
		PromptFile:  "",
	},
	{
		Type:        "educator",
		Name:        "Обучение",
		Description: "Объяснение терминов, принципы инвестирования",
		Icon:        "academic-cap",
		PromptFile:  "",
	},
}

var agentSystemPrompts = map[string]string{
	"analyst": `Ты — профессиональный финансовый аналитик по облигациям РФ. 
Специализация: корпоративные, субфедеральные, муниципальные облигации, ОФЗ.
Отвечай кратко и конкретно. Используй данные MOEX когда уместно.
Формат ответа: обычный текст с абзацами, без markdown-заголовков.
Язык: русский.`,

	"educator": `Ты — финансовый преподаватель, объясняешь сложные термины простым языком.
Специализация: облигации, фиксированный доход, кредитные рейтинги, дюрация, доходность.
Используй аналогии и примеры из реальной жизни. Формат: обычный текст.
Язык: русский.`,
}

type ChatService struct {
	repo   *mongoRepo.ChatRepo
	openai *openai.Client
}

func NewChatService(repo *mongoRepo.ChatRepo, openaiClient *openai.Client) *ChatService {
	return &ChatService{
		repo:   repo,
		openai: openaiClient,
	}
}

func (s *ChatService) GetAgents() []model.Agent {
	return defaultAgents
}

func (s *ChatService) CreateSession(ctx context.Context, req *model.CreateSessionRequest) (*model.ChatSession, error) {
	// Validate agent type
	if !s.isValidAgent(req.AgentType) {
		return nil, fmt.Errorf("unknown agent type: %s", req.AgentType)
	}

	title := req.Title
	if title == "" {
		title = "Новый чат"
	}

	session := &model.ChatSession{
		SessionID: newUID(),
		Title:     title,
		AgentType: req.AgentType,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.repo.CreateSession(ctx, session); err != nil {
		return nil, err
	}
	return session, nil
}

func (s *ChatService) GetSession(ctx context.Context, sessionID string) (*model.ChatSession, error) {
	return s.repo.GetSession(ctx, sessionID)
}

func (s *ChatService) ListSessions(ctx context.Context) ([]model.ChatSession, error) {
	return s.repo.ListSessions(ctx)
}

func (s *ChatService) DeleteSession(ctx context.Context, sessionID string) error {
	return s.repo.DeleteSession(ctx, sessionID)
}

func (s *ChatService) GetMessages(ctx context.Context, sessionID string) ([]model.ChatMessage, error) {
	return s.repo.GetMessages(ctx, sessionID)
}

func (s *ChatService) SendMessage(ctx context.Context, sessionID string, content string) (*model.SendMessageResponse, error) {
	// Get session to determine agent type
	session, err := s.repo.GetSession(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("session not found: %w", err)
	}

	// Save user message
	userMsg := &model.ChatMessage{
		SessionID: sessionID,
		Role:      "user",
		Content:   content,
		CreatedAt: time.Now(),
	}
	if err := s.repo.AddMessage(ctx, userMsg); err != nil {
		return nil, err
	}

	// Build conversation for OpenAI
	history, err := s.repo.GetMessages(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	messages := s.buildOpenAIMessages(session.AgentType, history)

	// Call OpenAI
	response, err := s.openai.ChatCompletion(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("openai: %w", err)
	}

	// Save assistant message
	assistantMsg := &model.ChatMessage{
		SessionID: sessionID,
		Role:      "assistant",
		Content:   response,
		CreatedAt: time.Now(),
	}
	if err := s.repo.AddMessage(ctx, assistantMsg); err != nil {
		return nil, err
	}

	// Update session timestamp
	_ = s.repo.UpdateSessionTimestamp(ctx, sessionID)

	return &model.SendMessageResponse{
		UserMessage:      *userMsg,
		AssistantMessage: *assistantMsg,
	}, nil
}

func (s *ChatService) buildOpenAIMessages(agentType string, history []model.ChatMessage) []openai.ChatMessage {
	var messages []openai.ChatMessage

	// System prompt
	systemPrompt, ok := agentSystemPrompts[agentType]
	if !ok {
		systemPrompt = agentSystemPrompts["analyst"]
	}
	messages = append(messages, openai.ChatMessage{
		Role:    "system",
		Content: systemPrompt,
	})

	// Keep last 20 messages for context window management
	start := 0
	if len(history) > 20 {
		start = len(history) - 20
	}

	for _, msg := range history[start:] {
		messages = append(messages, openai.ChatMessage{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	return messages
}

func (s *ChatService) isValidAgent(agentType string) bool {
	for _, a := range defaultAgents {
		if a.Type == agentType {
			return true
		}
	}
	return false
}

func newUID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}
