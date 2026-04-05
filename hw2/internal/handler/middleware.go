package handler

import (
	"net/http"
	"strings"

	"github.com/KazikovAP/backend_course_start/hw2/internal/service"
)

func AuthMiddleware(auth *service.AuthService, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		if header == "" {
			writeJSON(w, http.StatusUnauthorized, errorResponse("missing token"))
			return
		}

		tokenStr := strings.TrimPrefix(header, "Bearer ")
		if !auth.ValidateToken(tokenStr) {
			writeJSON(w, http.StatusUnauthorized, errorResponse("invalid token"))
			return
		}

		next(w, r)
	}
}
