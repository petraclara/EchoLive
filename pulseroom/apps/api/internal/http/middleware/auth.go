package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/pulseroom/api/internal/auth"
	"github.com/pulseroom/api/internal/httpx"
)

type ctxKey string

const OrganizerIDKey ctxKey = "organizer_id"

func OrganizerAuth(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := BearerFromRequest(r)
			if token == "" {
				httpx.Error(w, http.StatusUnauthorized, "missing token")
				return
			}
			claims, err := auth.ParseToken(jwtSecret, token)
			if err != nil {
				httpx.Error(w, http.StatusUnauthorized, "invalid token")
				return
			}
			ctx := context.WithValue(r.Context(), OrganizerIDKey, claims.OrganizerID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func OrganizerID(r *http.Request) uuid.UUID {
	return r.Context().Value(OrganizerIDKey).(uuid.UUID)
}
