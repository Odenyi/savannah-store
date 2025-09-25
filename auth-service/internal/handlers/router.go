package handlers

import (
	"database/sql"
	"fmt"
	"savannah-store/auth-service/internal/logger"
	_ "savannah-store/auth-service/docs"
	"savannah-store/auth-service/internal/repository"

	"github.com/go-redis/redis"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	echoSwagger "github.com/swaggo/echo-swagger"

	"net/http"
	"os"

	echoLogger "github.com/mudphilo/echo-logger"

	amqp "github.com/rabbitmq/amqp091-go"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// router and DB instance
type App struct {
	DB              *sql.DB
	ArriveTime      int64
	E               *echo.Echo
	RedisConnection *redis.Client
	RabbitMQConn    *amqp.Connection
	GoogleOauthConfig *oauth2.Config
}

// Initialize initializes the app with predefined configuration
func (a *App) Initialize() {

	a.RedisConnection = repository.RedisClient()
	a.RabbitMQConn = repository.GetRabbitMQConnection()
	

	dbName := os.Getenv("AUTH_DB_NAME")
	var googleOauthConfig *oauth2.Config
	// Google OAuth2 setup
	googleOauthConfig = &oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_KEY"),
		RedirectURL:  "https://auth.vaslinkcomm.com/auth/google/callback",
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
		Endpoint:     google.Endpoint,
	}
	a.GoogleOauthConfig = googleOauthConfig

	dbO := repository.DbInstance(dbName)
	a.DB = dbO

	a.setRouters()

}

// setRouters sets the all required router
func (a *App) setRouters() {

	// init webserver
	a.E = echo.New()

	a.E.Use(middleware.Gzip())
	a.E.IPExtractor = echo.ExtractIPFromXFFHeader()
	// add recovery middleware to make the system null safe
	a.E.Use(middleware.Recover()) // change due to swagger
	a.E.Use(session.Middleware(sessions.NewCookieStore([]byte(os.Getenv("SESSION_SECRET")))))

	// setup log format and parameters to log for every request

	a.E.Use(echoLogger.Logger())

	allowedMethods := []string{http.MethodGet, http.MethodHead, http.MethodPut, http.MethodPatch, http.MethodPost, http.MethodDelete}
	AllowOrigins := []string{"*"}

	//setup CORS
	corsConfig := middleware.CORSConfig{
		AllowOrigins: AllowOrigins, // in production limit this to only known hosts
		AllowHeaders: AllowOrigins,
		AllowMethods: allowedMethods,
	}

	a.E.Use(middleware.CORSWithConfig(corsConfig))
	
	

	// Routes
	a.E.POST("/auth/google/start", a.StartGoogleAuth)
	a.E.GET("/auth/google/callback", a.GoogleCallback)
	

	//status
	a.E.GET("/docs/*", echoSwagger.WrapHandler)
	a.E.POST("/", a.GetStatus)
	a.E.GET("/", a.GetStatus)
}

func (a *App) GetStatus(c echo.Context) error {

	return c.JSON(repository.CheckConnectionStatus())

}

// Run the app on it's router
func (a *App) Run() {

	host := os.Getenv("SYSTEM_HOST")
	port := os.Getenv("AUTH_SYSTEM_PORT")
	server := fmt.Sprintf("%s:%s", host, port)
	logger.Info("Auth service started %v",server)
	a.E.Logger.Fatal(a.E.Start(server))
}
