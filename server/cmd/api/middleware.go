package main

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
)

type authContextKey string

const (
	authUsernameKey authContextKey = "username"
	authRoleKey     authContextKey = "role"
)

func (a *app) requireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		h := r.Header.Get("Authorization")
		if h == "" {
			writeJSON(w, http.StatusUnauthorized, apiError{Error: "missing_authorization"})
			return
		}

		const prefix = "Bearer "
		if !strings.HasPrefix(h, prefix) {
			writeJSON(w, http.StatusUnauthorized, apiError{Error: "invalid_authorization"})
			return
		}

		tokenString := strings.TrimSpace(strings.TrimPrefix(h, prefix))
		if tokenString == "" {
			writeJSON(w, http.StatusUnauthorized, apiError{Error: "invalid_authorization"})
			return
		}

		claims := &tokenClaims{}
		parsed, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (any, error) {
			return []byte(a.cfg.jwtSecret), nil
		})
		if err != nil || !parsed.Valid {
			writeJSON(w, http.StatusUnauthorized, apiError{Error: "invalid_token"})
			return
		}

		username := claims.Subject
		if username == "" {
			writeJSON(w, http.StatusUnauthorized, apiError{Error: "invalid_token"})
			return
		}

		got, err := a.st.GetUserRoleApprovedByUsername(r.Context(), username)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				writeJSON(w, http.StatusUnauthorized, apiError{Error: "invalid_token"})
				return
			}
			writeJSON(w, http.StatusInternalServerError, apiError{Error: "db_error"})
			return
		}
		role := got.Role
		approved := got.Approved

		if !approved && role != "admin" {
			writeJSON(w, http.StatusForbidden, apiError{Error: "account_pending_approval"})
			return
		}

		ctx := context.WithValue(r.Context(), authUsernameKey, username)
		ctx = context.WithValue(ctx, authRoleKey, role)
		next(w, r.WithContext(ctx))
	}
}

func authUsername(ctx context.Context) string {
	v, _ := ctx.Value(authUsernameKey).(string)
	return v
}

func authRole(ctx context.Context) string {
	v, _ := ctx.Value(authRoleKey).(string)
	return v
}
