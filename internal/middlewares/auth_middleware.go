package middlewares

import (
	"encoding/json"
	"net/http"
	"strings"

	"desent-api/internal/utils"
)

type unauthorizedResponse struct {
	ErrorCode string `json:"error_code"`
	Message   string `json:"message"`
}

func RequireBearerAuth(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := strings.TrimSpace(r.Header.Get("Authorization"))
			if authHeader == "" {
				writeUnauthorized(w)
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || strings.TrimSpace(parts[1]) == "" {
				writeUnauthorized(w)
				return
			}

			if err := utils.ValidateToken(strings.TrimSpace(parts[1]), jwtSecret); err != nil {
				writeUnauthorized(w)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func writeUnauthorized(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	_ = json.NewEncoder(w).Encode(unauthorizedResponse{
		ErrorCode: "UNAUTHORIZED",
		Message:   "unauthorized",
	})
}
