package middleware

import (
	"database/sql"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/tahsin005/affpilot-auth/internal/utils"
)


// self only middleware
func SelfOnlyMiddleware() mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := r.Context().Value("user").(*Claims)
			if !ok {
				utils.WriteError(w, http.StatusUnauthorized, "Unauthorized")
				return
			}

			vars := mux.Vars(r)
			targetUserID := vars["user_id"]

			if targetUserID != claims.UserID {
				utils.WriteError(w, http.StatusForbidden, "Forbidden: You can only perform this action on your own account")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// roles based middleware
func RoleMiddleware(allowedRoles ...string) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := r.Context().Value("user").(*Claims)
			if !ok {
				utils.WriteError(w, http.StatusUnauthorized, "Unauthorized")
				return
			}

			for _, role := range allowedRoles {
				if claims.Role == role {
					next.ServeHTTP(w, r)
					return
				}
			}

			utils.WriteError(w, http.StatusForbidden, "Forbidden: Insufficient role")
		})
	}
}

func PermissionMiddleware(db *sql.DB, requiredPermission string) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := r.Context().Value("user").(*Claims)
			if !ok {
				utils.WriteError(w, http.StatusUnauthorized, "Unauthorized")
				return
			}

			var hasPermission bool
			query := `
				SELECT EXISTS (
					SELECT 1
					FROM user_roles ur
					JOIN role_permissions rp ON ur.role_id = rp.role_id
					JOIN permissions p ON rp.permission_id = p.id
					WHERE ur.user_id = $1
					AND p.name = $2
				)
			`
			err := db.QueryRow(query, claims.UserID, requiredPermission).Scan(&hasPermission)
			if err != nil {
				utils.WriteError(w, http.StatusInternalServerError, "Internal server error")
				return
			}
			
			if !hasPermission {
				utils.WriteError(w, http.StatusForbidden, "Forbidden")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// system admin only middleware
func SystemAdminOnlyMiddleware() mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := r.Context().Value("user").(*Claims)
			if !ok {
				utils.WriteError(w, http.StatusUnauthorized, "Unauthorized")
				return
			}

			if claims.Role != "system_admin" {
				utils.WriteError(w, http.StatusForbidden, "Forbidden: System Admin access required")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// self or authroized middleware (user / admin+ or user / moderator+)
func SelfOrAuthorizedMiddleware(db *sql.DB, requiredPermission string) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := r.Context().Value("user").(*Claims)
			if !ok {
				utils.WriteError(w, http.StatusUnauthorized, "Unauthorized")
				return
			}

			vars := mux.Vars(r)
			targetUserID := vars["user_id"]

			if targetUserID == claims.UserID {
				next.ServeHTTP(w, r)
				return
			}

			var hasPermission bool
			query := `
				SELECT EXISTS (
					SELECT 1
					FROM user_roles ur
					JOIN role_permissions rp ON ur.role_id = rp.role_id
					JOIN permissions p ON rp.permission_id = p.id
					WHERE ur.user_id = $1
					AND p.name = $2
				)
			`
			err := db.QueryRow(query, claims.UserID, requiredPermission).Scan(&hasPermission)
			if err != nil {
				utils.WriteError(w, http.StatusInternalServerError, "Internal server error")
				return
			}

			if !hasPermission {
				utils.WriteError(w, http.StatusForbidden, "Forbidden")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}