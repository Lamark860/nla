package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"nla/internal/model"
	"nla/internal/service"
)

type ChatHandler struct {
	chatService *service.ChatService
}

func NewChatHandler(chatService *service.ChatService) *ChatHandler {
	return &ChatHandler{chatService: chatService}
}

// GetAgents GET /api/v1/chat/agents
func (h *ChatHandler) GetAgents(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, h.chatService.GetAgents())
}

// ListSessions GET /api/v1/chat/sessions
func (h *ChatHandler) ListSessions(w http.ResponseWriter, r *http.Request) {
	sessions, err := h.chatService.ListSessions(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	if sessions == nil {
		sessions = []model.ChatSession{}
	}
	writeJSON(w, http.StatusOK, sessions)
}

// CreateSession POST /api/v1/chat/sessions
func (h *ChatHandler) CreateSession(w http.ResponseWriter, r *http.Request) {
	var req model.CreateSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if req.AgentType == "" {
		req.AgentType = "analyst"
	}

	session, err := h.chatService.CreateSession(r.Context(), &req)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusCreated, session)
}

// GetSession GET /api/v1/chat/sessions/{id}
func (h *ChatHandler) GetSession(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "id")

	session, err := h.chatService.GetSession(r.Context(), sessionID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "session not found"})
		return
	}

	messages, err := h.chatService.GetMessages(r.Context(), sessionID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	if messages == nil {
		messages = []model.ChatMessage{}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"session":  session,
		"messages": messages,
	})
}

// DeleteSession DELETE /api/v1/chat/sessions/{id}
func (h *ChatHandler) DeleteSession(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "id")

	if err := h.chatService.DeleteSession(r.Context(), sessionID); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// SendMessage POST /api/v1/chat/sessions/{id}/messages
func (h *ChatHandler) SendMessage(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "id")

	var req model.SendMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if req.Content == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "content is required"})
		return
	}

	resp, err := h.chatService.SendMessage(r.Context(), sessionID, req.Content)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, resp)
}
