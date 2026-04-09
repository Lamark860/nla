package handler

import (
	"encoding/json"
	"net/http"

	"nla/internal/middleware"
	"nla/internal/model"
	"nla/internal/service"
)

type Handler struct {
	auth      *service.AuthService
	Bond      *BondHandler
	Analysis  *AnalysisHandler
	Rating    *RatingHandler
	Chat      *ChatHandler
	Favorite  *FavoriteHandler
	JWTSecret string
}

func New(auth *service.AuthService, jwtSecret string) *Handler {
	return &Handler{
		auth:      auth,
		JWTSecret: jwtSecret,
	}
}

func (h *Handler) SetBondHandler(bh *BondHandler) {
	h.Bond = bh
}

func (h *Handler) SetAnalysisHandler(ah *AnalysisHandler) {
	h.Analysis = ah
}

func (h *Handler) SetRatingHandler(rh *RatingHandler) {
	h.Rating = rh
}

func (h *Handler) SetChatHandler(ch *ChatHandler) {
	h.Chat = ch
}

func (h *Handler) SetFavoriteHandler(fh *FavoriteHandler) {
	h.Favorite = fh
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req model.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	resp, err := h.auth.Register(r.Context(), req)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusCreated, resp)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req model.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	resp, err := h.auth.Login(r.Context(), req)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	user, err := h.auth.GetUser(r.Context(), userID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "user not found"})
		return
	}

	writeJSON(w, http.StatusOK, user)
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
