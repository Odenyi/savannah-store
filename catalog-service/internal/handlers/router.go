package handlers

import (
	"database/sql"
	"fmt"
	_ "savannah-store/catalog-service/docs"
	"savannah-store/catalog-service/internal/logger"
	"savannah-store/catalog-service/internal/repository"
	auth "savannah-store/catalog-service/internal/middleware"

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

	dbName := os.Getenv("CATALOG_DB_NAME")

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

	
	// Category routes
	a.E.POST("/catalog/categories",a.CreateCategory,auth.RoleMiddleware(a.DB, "admin"))          
	a.E.GET("/catalog/categories", a.ViewCategories)           
	a.E.PUT("/catalog/categories/:id", a.UpdateCategory,auth.RoleMiddleware(a.DB, "admin")) 
	a.E.GET("/categories/:id/average-price",a.GetAveragePrice,auth.RoleMiddleware(a.DB, "admin"))  

	// Product routes
	a.E.POST("/catalog/products", a.CreateProduct,auth.RoleMiddleware(a.DB, "admin"))             
	a.E.GET("/catalog/products", a.ViewProducts)               
	a.E.PUT("/catalog/products/:id", a.UpdateProduct,auth.RoleMiddleware(a.DB, "admin"))          



	
	
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
	port := os.Getenv("IDENTITY_SYSTEM_PORT")
	server := fmt.Sprintf("%s:%s", host, port)
	logger.Info("Auth service started %v",server)
	a.E.Logger.Fatal(a.E.Start(server))
}
