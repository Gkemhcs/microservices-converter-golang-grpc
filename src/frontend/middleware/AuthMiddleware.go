package middleware

import (
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	
)

// AuthMiddleware validates the session for protected routes
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)

		// Check if the user is logged in
		if loggedIn := session.Get("logged_in"); loggedIn == nil || !loggedIn.(bool) {
			c.Redirect(http.StatusFound,"/user/login")
			return
		}

		// Allow the request to proceed
		c.Next()
	}
}
