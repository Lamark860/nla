package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"nla/internal/middleware"
	"nla/internal/repository"
)

type FavoriteHandler struct {
	repo *repository.FavoriteRepository
}

func NewFavoriteHandler(repo *repository.FavoriteRepository) *FavoriteHandler {
	return &FavoriteHandler{repo: repo}
}

// ListFavorites returns all favorite secids for the authenticated user
func (h *FavoriteHandler) ListFavorites(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	secids, err := h.repo.GetSecIDs(r.Context(), userID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list favorites"})
		return
	}
	if secids == nil {
		secids = []string{}
	}

	writeJSON(w, http.StatusOK, map[string]any{"secids": secids, "count": len(secids)})
}

// AddFavorite adds a bond to user's favorites
func (h *FavoriteHandler) AddFavorite(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	secid := chi.URLParam(r, "secid")
	if secid == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "secid required"})
		return
	}

	if err := h.repo.Add(r.Context(), userID, secid); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to add favorite"})
		return
	}

	writeJSON(w, http.StatusCreated, map[string]string{"secid": secid, "status": "added"})
}

// RemoveFavorite removes a bond from user's favorites
func (h *FavoriteHandler) RemoveFavorite(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	secid := chi.URLParam(r, "secid")
	if secid == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "secid required"})
		return
	}

	if err := h.repo.Remove(r.Context(), userID, secid); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to remove favorite"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"secid": secid, "status": "removed"})
}

// CheckFavorites checks which of the given secids are in user's favorites
func (h *FavoriteHandler) CheckFavorites(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	secidsParam := r.URL.Query().Get("secids")
	if secidsParam == "" {
		writeJSON(w, http.StatusOK, map[string]any{"favorites": map[string]bool{}})
		return
	}

	secids := strings.Split(secidsParam, ",")
	// Limit to prevent abuse
	if len(secids) > 200 {
		secids = secids[:200]
	}

	result, err := h.repo.CheckMultiple(r.Context(), userID, secids)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to check favorites"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"favorites": result})
}

// ToggleFavorite toggles a bond in user's favorites (add if not present, remove if present)
func (h *FavoriteHandler) ToggleFavorite(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	var req struct {
		SecID string `json:"secid"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.SecID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "secid required"})
		return
	}

	isFav, err := h.repo.IsFavorite(r.Context(), userID, req.SecID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to check favorite"})
		return
	}

	if isFav {
		if err := h.repo.Remove(r.Context(), userID, req.SecID); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to remove favorite"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"secid": req.SecID, "is_favorite": false})
	} else {
		if err := h.repo.Add(r.Context(), userID, req.SecID); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to add favorite"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"secid": req.SecID, "is_favorite": true})
	}
}
