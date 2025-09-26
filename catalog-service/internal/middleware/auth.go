package middleware

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"database/sql"
	

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

type JwtCustomClaims struct {
	UserID int64  `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}
// RoleMiddleware validates API Key (JWT) and checks user role + expiry
func RoleMiddleware(db *sql.DB, allowedRoles ...string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Get API-Key header
			apiKey := c.Request().Header.Get("api-key")
			if apiKey == "" {
				return c.JSON(http.StatusUnauthorized, echo.Map{"error": "missing api-key header"})
			}

			// Parse JWT
			claims := &JwtCustomClaims{}
			token, err := jwt.ParseWithClaims(apiKey, claims, func(t *jwt.Token) (interface{}, error) {
				return []byte(os.Getenv("JWT_SECRET")), nil
			})
			if err != nil || !token.Valid {
				return c.JSON(http.StatusUnauthorized, echo.Map{"error": "invalid or expired token"})
			}

			// Check token expiry
			if claims.ExpiresAt != nil && claims.ExpiresAt.Time.Before(time.Now()) {
				return c.JSON(http.StatusUnauthorized, echo.Map{"error": "token expired"})
			}

			// Get role from DB
			var roleName string
			err = db.QueryRow(`
				SELECT r.name 
				FROM authdb.users u 
				JOIN authdb.roles r ON u.role_id = r.id 
				WHERE u.id = ?`,
				claims.UserID,
			).Scan(&roleName)
			if err != nil {
				if err == sql.ErrNoRows {
					return c.JSON(http.StatusForbidden, echo.Map{"error": "user not found"})
				}
				return c.JSON(http.StatusInternalServerError, echo.Map{"error": "failed to fetch role"})
			}

			// Check if role is allowed
			if !isRoleAllowed(roleName, allowedRoles) {
				return c.JSON(http.StatusForbidden, echo.Map{"error": fmt.Sprintf("role '%s' not authorized", roleName)})
			}

			// Store user info in context
			c.Set("user_id", claims.UserID)
			c.Set("email", claims.Email)
			c.Set("role", roleName)

			return next(c)
		}
	}
}

// Helper to check allowed roles
func isRoleAllowed(userRole string, allowedRoles []string) bool {
	for _, r := range allowedRoles {
		if strings.EqualFold(r, "all") || strings.EqualFold(userRole, r) {
			return true
		}
	}
	return false
}
