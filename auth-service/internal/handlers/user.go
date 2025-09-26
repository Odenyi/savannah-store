package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"savannah-store/auth-service/internal/controllers"
	"savannah-store/auth-service/internal/library"
	"savannah-store/auth-service/internal/models"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"golang.org/x/oauth2"
)

// StartGoogleAuth godoc
// @Summary Start Google OAuth flow
// @Description Initiates Google OAuth by generating a URL and optional state for signup.
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body models.AuthRequest false "Optional phone and usertype"
// @Success 200 {object} map[string]interface{} "google_url and state"
// @Failure 500 {object} map[string]string "failed to store usertype or phone"
// @Router /auth/google/start [post]
func (a *App) StartGoogleAuth(c echo.Context) error {
	var req struct {
		Phone    string `json:"phone"`
		Usertype string `json:"usertype"`
	}

	_ = c.Bind(&req) // ignore error if empty

	// Generate a state token to link this OAuth request
	state := fmt.Sprintf("state-%d", time.Now().UnixNano())
	if req.Usertype != "" {
		usertypekey := fmt.Sprintf("signup:usertype:%s", state)
		err := library.SetRedisKeyWithExpiry(a.RedisConnection, usertypekey, req.Usertype, 600)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, echo.Map{"error": "failed to store usertype"})
		}
	}

	if req.Phone != "" {
		// Store phone in Redis temporarily
		redisKey := fmt.Sprintf("signup:phone:%s", state)
		err := library.SetRedisKeyWithExpiry(a.RedisConnection, redisKey, req.Phone, 600)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, echo.Map{"error": "failed to store phone"})
		}
	}

	// Generate Google OAuth URL
	url := a.GoogleOauthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)
	return c.JSON(http.StatusOK, echo.Map{
		"google_url": url,
		"state":      state,
	})
}

// UserSignup godoc
// @Summary Manual user signup
// @Description Creates a new user directly without Google OAuth (for alternative signup flows).
// @Tags Auth
// @Accept json
// @Produce json
// @Param user body models.UserSignupRequest true "User signup payload"
// @Success 201 {object} map[string]interface{} "Created user"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 500 {object} map[string]string "Failed to create user"
// @Router /auth/signup [post]
func (a *App) GoogleCallback(c echo.Context) error {
	code := c.QueryParam("code")
	state := c.QueryParam("state")
	if code == "" || state == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "missing code or state"})
	}

	// Exchange code for access token
	token, err := a.GoogleOauthConfig.Exchange(context.Background(), code)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "failed to exchange token"})
	}

	client := a.GoogleOauthConfig.Client(context.Background(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "failed to get user info"})
	}
	defer resp.Body.Close()

	var userInfo struct {
		Email string `json:"email"`
		Name  string `json:"name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "failed to parse user info"})
	}

	// Retrieve phone & usertype from Redis
	redisKey := fmt.Sprintf("signup:phone:%s", state)
	usertypeKey := fmt.Sprintf("signup:usertype:%s", state)
	phone, _ := library.GetRedisKey(a.RedisConnection, redisKey)
	usertype, _ := library.GetRedisKey(a.RedisConnection, usertypeKey)
	if usertype == "" {
		usertype = "customer"
	}

	// Map usertype → role_id
	var roleID int
	err = a.DB.QueryRow("SELECT id FROM roles WHERE name = ?", usertype).Scan(&roleID)
	if err != nil {
		// fallback: assign customer role
		_ = a.DB.QueryRow("SELECT id FROM roles WHERE name = 'customer'").Scan(&roleID)
	}

	// Check if user exists
	var userID int64
	err = a.DB.QueryRow("SELECT id FROM users WHERE email = ?", userInfo.Email).Scan(&userID)
	if err != nil {
		if err == sql.ErrNoRows {
			// User does not exist → create new one
			res, insertErr := a.DB.Exec(`
				INSERT INTO users (email, email_verified, full_name, phone, role_id)
				VALUES (?, 1, ?, ?, ?)`,
				userInfo.Email, userInfo.Name, phone, roleID,
			)
			if insertErr != nil {
				return c.JSON(http.StatusInternalServerError, echo.Map{"error": "failed to create user"})
			}
			userID, _ = res.LastInsertId()
		} else {
			return c.JSON(http.StatusInternalServerError, echo.Map{"error": "failed to check user"})
		}
	}
	//querry the correct role id
	a.DB.QueryRow("SELECT role_id,name FROM users WHERE id = ?", userID).Scan(&roleID, &usertype)
	// Generate JWT
	claims := &models.JwtCustomClaims{
		UserID: userID,
		Email:  userInfo.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		},
	}
	var jwtSecret = []byte(os.Getenv("JWT_SECRET"))

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, signErr := jwtToken.SignedString(jwtSecret)
	if signErr != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "failed to sign token"})
	}

	// Clean up Redis keys
	a.RedisConnection.Del(redisKey)
	a.RedisConnection.Del(usertypeKey)

	return c.JSON(http.StatusOK, echo.Map{
		"message": "authenticated",
		"token":   tokenString,
		"user": echo.Map{
			"id":       userID,
			"email":    userInfo.Email,
			"name":     userInfo.Name,
			"phone":    phone,
			"role_id":  roleID,
			"usertype": usertype,
		},
	})
}

func (a *App) UserSignup(c echo.Context) error {
	user, err := controllers.UserCreate(c, a.DB)
	if err != nil {
		// err might already be an echo.HTTPError
		if httpErr, ok := err.(*echo.HTTPError); ok {
			return c.JSON(httpErr.Code, map[string]string{"error": httpErr.Message.(string)})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"id":    user.ID,
		"email": user.Email,
		"phone": user.Phone,
		"role":  user.RoleID,
	})
}
