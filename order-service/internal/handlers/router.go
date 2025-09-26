package handlers

import (
	"database/sql"
	"fmt"
	_ "savannah-store/order-service/docs"
	"savannah-store/order-service/internal/logger"
	auth "savannah-store/order-service/internal/middleware"
	"savannah-store/order-service/internal/repository"

	"github.com/go-redis/redis"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"net/http"
	"os"

	echoLogger "github.com/mudphilo/echo-logger"

	amqp "github.com/rabbitmq/amqp091-go"
	echoSwagger "github.com/swaggo/echo-swagger"
)

// router and DB instance
type App struct {
	DB              *sql.DB
	ArriveTime      int64
	E               *echo.Echo
	RedisConnection *redis.Client
	RabbitMQConn    *amqp.Connection
}

// Initialize initializes the app with predefined configuration
func (a *App) Initialize() {

	a.RedisConnection = repository.RedisClient()
	a.RabbitMQConn = repository.GetRabbitMQConnection()

	dbName := os.Getenv("ORDER_DB_NAME")

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

	// Cart routes
	a.E.POST("/cart", a.AddToCart, auth.RoleMiddleware(a.DB, "customer", "admin"))
	a.E.GET("/cart", a.ViewCart, auth.RoleMiddleware(a.DB, "customer", "admin"))
	a.E.PUT("/cart", a.UpdateCart, auth.RoleMiddleware(a.DB, "customer", "admin"))
	a.E.DELETE("/cart", a.DeleteCart, auth.RoleMiddleware(a.DB, "customer", "admin"))

	// Order routes
	a.E.POST("/orders", a.PlaceOrder, auth.RoleMiddleware(a.DB, "customer", "admin"))
	a.E.GET("/orders", a.ViewOrders, auth.RoleMiddleware(a.DB, "customer", "admin"))
	a.E.DELETE("/orders", a.DeleteOrder, auth.RoleMiddleware(a.DB, "admin"))

	//status
	a.E.POST("/", a.GetStatus)
	a.E.GET("/", a.GetStatus)
	a.E.GET("/docs/*", echoSwagger.WrapHandler)
}

func (a *App) GetStatus(c echo.Context) error {

	return c.JSON(repository.CheckConnectionStatus())

}

// Run the app on it's router
func (a *App) Run() {

	host := os.Getenv("SYSTEM_HOST")
	port := os.Getenv("ORDER_SYSTEM_PORT")
	server := fmt.Sprintf("%s:%s", host, port)
	logger.Info("Auth service started %v", server)
	a.E.Logger.Fatal(a.E.Start(server))
}
