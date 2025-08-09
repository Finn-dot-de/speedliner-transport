package middleware

import (
	"context"
	"net/http"
	"speedliner-server/src/utils/users"
)

func RoleMiddleware(allowedRoles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("char")
			if err != nil || cookie.Value == "" {
				http.Error(w, "Not authenticated", http.StatusUnauthorized)
				return
			}

			charID := cookie.Value
			ok, err := users.HasRole(charID, allowedRoles...)
			if err != nil {
				http.Error(w, "Error checking user role", http.StatusInternalServerError)
				return
			}
			if !ok {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			ctx := context.WithValue(r.Context(), contextKey("char"), charID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
