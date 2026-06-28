package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/yadavsushil07/GolangTemplate/internal/model"
	"github.com/yadavsushil07/GolangTemplate/internal/service"
)

type contextKey string

const (
	ContextUserID   contextKey = "userID"
	ContextUserRole contextKey = "userRole"
)

func Auth(authSvc *service.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := extractToken(r)
			if token == "" {
				http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
				return
			}
			userID, role, ok := authSvc.ValidateJWT(token)
			if !ok {
				http.Error(w, `{"error":"invalid or expired token"}`, http.StatusUnauthorized)
				return
			}
			ctx := context.WithValue(r.Context(), ContextUserID, userID)
			ctx = context.WithValue(ctx, ContextUserRole, role)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// VendorOrAdmin allows both vendor and admin roles.
func VendorOrAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		role, _ := r.Context().Value(ContextUserRole).(string)
		if role != model.RoleVendor && role != model.RoleAdmin {
			http.Error(w, `{"error":"forbidden"}`, http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// VendorOnly kept as alias for VendorOrAdmin for backward compatibility.
var VendorOnly = VendorOrAdmin

// AdminOnly allows only admin role.
func AdminOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		role, _ := r.Context().Value(ContextUserRole).(string)
		if role != model.RoleAdmin {
			http.Error(w, `{"error":"forbidden"}`, http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func UserIDFromContext(ctx context.Context) (int64, bool) {
	id, ok := ctx.Value(ContextUserID).(int64)
	return id, ok
}

func UserRoleFromContext(ctx context.Context) string {
	role, _ := ctx.Value(ContextUserRole).(string)
	return role
}

func extractToken(r *http.Request) string {
	bearer := r.Header.Get("Authorization")
	if strings.HasPrefix(bearer, "Bearer ") {
		return strings.TrimPrefix(bearer, "Bearer ")
	}
	if c, err := r.Cookie("auth_token"); err == nil {
		return c.Value
	}
	return ""
}
