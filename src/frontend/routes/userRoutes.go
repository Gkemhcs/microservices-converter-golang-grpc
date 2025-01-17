package routes

import (
	"converter/frontend/handlers"
	"converter/frontend/middleware"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func SetupUserRoutes(r *gin.Engine, logger *logrus.Logger, dbClient *handlers.DBClient) {

	userHandler := handlers.UserHandler{
		Logger:   logger,
		DBClient: dbClient}
	userRoutes := r.Group("/user")
	{

		userRoutes.GET("/login", userHandler.LoginUser)
		userRoutes.POST("/google/auth/callback", userHandler.GoogleAuthCallback)
		userRoutes.Use(middleware.AuthMiddleware())
		{
			userRoutes.GET("/logout", userHandler.LogoutUser)
			userRoutes.GET("/profile", userHandler.GetUserProfile)
			userRoutes.GET("/download", userHandler.GetDownloadUrl)
		}

	}
}
