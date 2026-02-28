package handlers

import (
	"net/http"
	"time"

	"desent-api/internal/models"
	"desent-api/internal/utils"
)

type AuthHandler struct {
	jwtSecret string
	jwtTTL    time.Duration
}

func NewAuthHandler(jwtSecret string, jwtTTL time.Duration) *AuthHandler {
	return &AuthHandler{jwtSecret: jwtSecret, jwtTTL: jwtTTL}
}

func (h *AuthHandler) CreateToken(w http.ResponseWriter, r *http.Request) {
	var req models.TokenRequest
	if err := decodeJSON(r.Body, &req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON_BODY", "invalid JSON body")
		return
	}

	if req.Username != "admin" || req.Password != "password" {
		writeError(w, http.StatusUnauthorized, "INVALID_CREDENTIALS", "invalid credentials")
		return
	}

	token, err := utils.GenerateToken(req.Username, h.jwtSecret, h.jwtTTL)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		return
	}

	writeJSON(w, http.StatusOK, models.TokenResponse{Token: token})
}
