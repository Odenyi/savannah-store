package controllers

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"savannah-store/auth-service/internal/library"

	"github.com/golang-jwt/jwt/v5"
	"time"

	"github.com/go-redis/redis"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

// User represents the user model
type User struct {
	ID     int64  `json:"id"`
	Email  string `json:"email"`
	Phone  string `json:"phone"`
	RoleID int64  `json:"role_id"`
}

// UserCreateRequest represents the expected request body
type UserCreateRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Phone    string `json:"phone"`
	RoleName string `json:"role_name"`
}


type UserLoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UserLoginResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}

func UserLogin(c echo.Context, db *sql.DB, rdb *redis.Client, jwtSecret string) error {
	var req UserLoginRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}

	// Fetch user from DB
	var (
		userID       int64
		hashedPass   string
		roleID       int64
		email        string
		phone        string
	)
	err := db.QueryRowContext(c.Request().Context(),
		`SELECT id, password_hash, role_id, email, phone FROM user WHERE email = ? OR phone = ?`,
		req.Email, req.Email,
	).Scan(&userID, &hashedPass, &roleID, &email, &phone)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid credentials"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "server error"})
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(hashedPass), []byte(req.Password)); err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid credentials"})
	}

	// Fetch role permissions
	rows, err := db.QueryContext(c.Request().Context(),
		`SELECT p.name FROM permissions p
		 JOIN role_permissions rp ON rp.permission_id = p.id
		 WHERE rp.role_id = ?`, roleID,
	)
	if err != nil {
		log.Println("failed to fetch permissions:", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "server error"})
	}
	defer rows.Close()

	perms := []string{}
	for rows.Next() {
		var perm string
		if err := rows.Scan(&perm); err != nil {
			log.Println("permission scan error:", err)
			continue
		}
		perms = append(perms, perm)
	}

	// Generate JWT token
	expiry := time.Now().Add(24 * time.Hour)
	claims := jwt.MapClaims{
		"sub":   userID,
		"email": email,
		"phone": phone,
		"role":  roleID,
		"perms": perms,
		"exp":   expiry.Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to generate token"})
	}

	// Store token in Redis
	
	if err := library.SetRedisKeyWithExpiry(rdb, tokenStr, fmt.Sprintf("%d",userID), 24*3600); err != nil {
		log.Println("redis set error:", err)
	}

	// Return token
	resp := UserLoginResponse{
		Token:     tokenStr,
		ExpiresAt: expiry,
	}
	return c.JSON(http.StatusOK, resp)
}

// UserCreate creates a new user with role and logs its permissions
func UserCreate(c echo.Context, db *sql.DB) (*User, error) {
	ctx := c.Request().Context()

	// Parse request body
	var req UserCreateRequest
	if err := c.Bind(&req); err != nil {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusInternalServerError, "failed to hash password")
	}

	// Get role_id
	var roleID int64
	err = db.QueryRowContext(ctx, "SELECT id FROM roles WHERE name = ?", req.RoleName).Scan(&roleID)
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "invalid role name")
	}

	// Insert user
	res, err := db.ExecContext(ctx, `
        INSERT INTO user (email, password_hash, phone, role_id) VALUES (?, ?, ?, ?)`,
		req.Email, string(hashedPassword), req.Phone, roleID,
	)
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	userID, err := res.LastInsertId()
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	// Fetch permissions
	rows, err := db.QueryContext(ctx, `
        SELECT p.name 
        FROM permissions p
        JOIN role_permissions rp ON rp.permission_id = p.id
        WHERE rp.role_id = ?`, roleID)
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	defer rows.Close()

	var perms []string
	for rows.Next() {
		var perm string
		if err := rows.Scan(&perm); err != nil {
			return nil, echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		perms = append(perms, perm)
	}

	log.Printf("Created user %s with role %s and permissions: %v", req.Email, req.RoleName, perms)

	return &User{ID: userID, Email: req.Email, Phone: req.Phone, RoleID: roleID}, nil
}
