package handler

import (
	"encoding/json"
	"net/http"

	"nla/internal/model"
	"nla/internal/service"
)

type RatingHandler struct {
	ratingService *service.RatingService
}

func NewRatingHandler(ratingService *service.RatingService) *RatingHandler {
	return &RatingHandler{ratingService: ratingService}
}

// GetAllRatings GET /api/v1/ratings
func (h *RatingHandler) GetAllRatings(w http.ResponseWriter, r *http.Request) {
	ratings, err := h.ratingService.GetAll(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, ratings)
}

// GetIssuerRating GET /api/v1/ratings/{issuer}
func (h *RatingHandler) GetIssuerRating(w http.ResponseWriter, r *http.Request) {
	issuer := r.URL.Query().Get("issuer")
	if issuer == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "issuer query parameter is required"})
		return
	}

	rating, err := h.ratingService.GetByIssuer(r.Context(), issuer)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, rating)
}

// UpsertRating POST /api/v1/ratings
func (h *RatingHandler) UpsertRating(w http.ResponseWriter, r *http.Request) {
	var rating model.IssuerRating
	if err := json.NewDecoder(r.Body).Decode(&rating); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if rating.Issuer == "" || rating.Agency == "" || rating.Rating == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "issuer, agency, and rating are required"})
		return
	}

	if rating.Score < 1 || rating.Score > 10 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "score must be between 1 and 10"})
		return
	}

	if err := h.ratingService.Upsert(r.Context(), &rating); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, rating)
}

// BulkUpsertRatings POST /api/v1/ratings/bulk
func (h *RatingHandler) BulkUpsertRatings(w http.ResponseWriter, r *http.Request) {
	var ratings []model.IssuerRating
	if err := json.NewDecoder(r.Body).Decode(&ratings); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	for _, rating := range ratings {
		if rating.Issuer == "" || rating.Agency == "" || rating.Rating == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "all ratings must have issuer, agency, and rating"})
			return
		}
		if rating.Score < 1 || rating.Score > 10 {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "all scores must be between 1 and 10"})
			return
		}
	}

	if err := h.ratingService.BulkUpsert(r.Context(), ratings); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]int{"upserted": len(ratings)})
}
