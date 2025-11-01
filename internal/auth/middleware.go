package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"github.com/roblesvargas97/estimago/internal/utils"
)

// context keys for authentication metadata.
type ctxKey string

const (
	ctxUserIDKey ctxKey = "auth_user_id"
	ctxEmailKey  ctxKey = "auth_email"
	ctxPlanKey   ctxKey = "auth_plan"
)

// JWTMiddleware validates JWT tokens and injects identity into the request context.
func JWTMiddleware(cfg Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := strings.TrimSpace(r.Header.Get("Authorization"))
			if authHeader == "" || !strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
				utils.WriteErr(w, http.StatusUnauthorized, "unauthorized", "invalid token")
				return
			}

			tokenString := strings.TrimSpace(authHeader[len("Bearer "):])
			if tokenString == "" {
				utils.WriteErr(w, http.StatusUnauthorized, "unauthorized", "invalid token")
				return
			}

			claims, err := ParseJWT(tokenString, cfg.JWTSecret)
			if err != nil {
				utils.WriteErr(w, http.StatusUnauthorized, "unauthorized", "invalid token")
				return
			}

			userID, err := uuid.Parse(claims.Subject)
			if err != nil {
				utils.WriteErr(w, http.StatusUnauthorized, "unauthorized", "invalid token")
				return
			}

			ctx := context.WithValue(r.Context(), ctxUserIDKey, userID)
			ctx = context.WithValue(ctx, ctxEmailKey, claims.Email)
			ctx = context.WithValue(ctx, ctxPlanKey, claims.Plan)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// UserIDFromCtx extracts the authenticated user ID from the context.
func UserIDFromCtx(r *http.Request) (uuid.UUID, bool) {
	v := r.Context().Value(ctxUserIDKey)
	id, ok := v.(uuid.UUID)
	return id, ok
}

// EmailFromCtx extracts the authenticated email from the context.
func EmailFromCtx(r *http.Request) (string, bool) {
	v := r.Context().Value(ctxEmailKey)
	email, ok := v.(string)
	return email, ok
}

// PlanFromCtx extracts the authenticated plan from the context.
func PlanFromCtx(r *http.Request) (string, bool) {
	v := r.Context().Value(ctxPlanKey)
	plan, ok := v.(string)
	return plan, ok
}
