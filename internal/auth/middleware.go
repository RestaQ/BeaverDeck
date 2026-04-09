package auth

import (
	"context"
	"database/sql"
	"net/http"
	"strings"

	"beaverdeck/internal/users"
)

type ctxKey string

const userKey ctxKey = "beaverdeck-user"

func Middleware(userStore *users.Store) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if userStore == nil {
				http.Error(w, "user store is not configured", http.StatusServiceUnavailable)
				return
			}

			token := tokenFromRequest(r)
			u, err := userStore.Authenticate(r.Context(), token)
			if err != nil {
				if err == sql.ErrNoRows {
					http.Error(w, "unauthorized", http.StatusUnauthorized)
					return
				}
				http.Error(w, "authorization failed", http.StatusInternalServerError)
				return
			}
			requestedUser := strings.TrimSpace(r.Header.Get("X-BeaverDeck-Username"))
			if requestedUser == "" {
				requestedUser = strings.TrimSpace(r.URL.Query().Get("username"))
			}
			if requestedUser == "" || requestedUser != u.Username {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), userKey, *u)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func UserFromContext(ctx context.Context) (users.UserWithToken, bool) {
	v := ctx.Value(userKey)
	u, ok := v.(users.UserWithToken)
	return u, ok
}

func IsAdmin(ctx context.Context) bool {
	u, ok := UserFromContext(ctx)
	if !ok {
		return false
	}
	return u.RoleMode == string(users.RoleAdmin)
}

func tokenFromRequest(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	if strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
		return strings.TrimSpace(authHeader[len("Bearer "):])
	}
	return r.URL.Query().Get("token")
}
