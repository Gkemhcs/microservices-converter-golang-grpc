package main

import (
	"converter/frontend/handlers"
	"converter/frontend/middleware"
	"converter/frontend/routes"
	"converter/frontend/utils"
	"fmt"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/redis"
	"github.com/gin-gonic/gin"
	ginprometheus "github.com/zsais/go-gin-prometheus"
	"go.opentelemetry.io/otel"
)

var (
	ClientID     = utils.GetEnv("GOOGLE_CLIENT_ID", "27377431828-slhtq2am6nagu69kfmb4vn5pl8g8j4ma.apps.googleusercontent.com")
	ClientSecret = utils.GetEnv("GOOGLE_CLIENT_SECRET", "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")

	RedirectURL = utils.GetEnv("REDIRECT_CALLBACK_URL", "http://localhost:8080/google/auth/callback")
	REDIS_HOST  = utils.GetEnv("REDIS_HOST", "localhost")

	REDIS_PORT   = utils.GetEnv("REDIS_PORT", "6379")
	REDIS_SECRET = utils.GetEnv("REDIS_SECRET", "idwoeoejinkwnks")

	HOST = utils.GetEnv("SERVER_HOST", "0.0.0.0")
	PORT = utils.GetEnv("SERVER_PORT", "8080")
)

func main() {
	store, err := redis.NewStore(10, "tcp", fmt.Sprintf("%s:%s", REDIS_HOST, REDIS_PORT), "", []byte("secret-key"))
	if err != nil {
		fmt.Print(http.StatusInternalServerError, "Error connecting to redis")
	}

	router := gin.Default()

	// Initialize the Prometheus middleware
	p := ginprometheus.NewPrometheus("gin")
	p.Use(router)

	logger := utils.NewLogger()
	router.Use(middleware.LoggingMiddleware(logger))

	router.Use(sessions.Sessions("converter-frontend", store))

	router.Static("/static", "./static")
	router.LoadHTMLGlob("templates/*")

	cleanup := utils.InitTracer("frontend", logger)
	defer cleanup()

	tracer := otel.Tracer("frontend")

	dbConn, err := utils.ConnectToDB(logger)
	if err != nil {
		panic(err)
	}
	dbClient := handlers.DBClient{
		DB: dbConn,
	}
	routes.SetupUserRoutes(router, logger, &dbClient)
	routes.SetupServiceRoutes(router, logger, &tracer, &dbClient)
	router.GET("/", func(c *gin.Context) {
		_, span := tracer.Start(c.Request.Context(), "home-route")
		defer span.End()
		session := sessions.Default(c)

		email := session.Get("email")
		if email == nil {
			c.HTML(http.StatusOK, "home.html", gin.H{})
			return
		}
		c.HTML(http.StatusOK, "home.html", gin.H{"Email": email.(string)})
	})
	router.GET("/health", healthCheck)

	router.Run(fmt.Sprintf("%s:%s", HOST, PORT)) // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}

func healthCheck(c *gin.Context) {

	c.JSON(http.StatusOK, gin.H{
		"status":  "succeeded",
		"message": "server receiving requests",
	})
}
