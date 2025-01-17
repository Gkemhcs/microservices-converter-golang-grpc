package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"google.golang.org/api/idtoken"
)

const (
	ClientID = "27377431828-slhtq2am6nagu69kfmb4vn5pl8g8j4ma.apps.googleusercontent.com"
)

type UserHandler struct {
	*logrus.Logger
	*DBClient
}

func (h *UserHandler) GetUserProfile(c *gin.Context) {
	session := sessions.Default(c)
	userPicture := session.Get("picture")
	email := session.Get("email").(string)
	userName := session.Get("name")
	downloads,err:=h.DBClient.FetchRecentDownloads(email)
	if err!=nil{
		h.Logger.Errorf("error while fetching downloads %v",err)
		c.HTML(http.StatusOK, "profile.html", gin.H{"ProfileImage": userPicture.(string), "UserName": userName.(string), "Email": email})
		return 
	}

	c.HTML(http.StatusOK, "profile.html", gin.H{"ProfileImage": userPicture.(string), "UserName": userName.(string), "Email": email,"Downloads":downloads})
}
func (h *UserHandler) GoogleAuthCallback(c *gin.Context) {

	idToken := c.PostForm("credential")
	if idToken == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing ID token"})
		return
	}

	payload, err := idtoken.Validate(c.Request.Context(), idToken, ClientID)

	if err != nil {
		h.Logger.Errorf("failed to validate id token: %v", err)

		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid ID token"})
		return
	}

	userID := payload.Claims["sub"].(string)
	userName := payload.Claims["name"].(string)
	userEmail := payload.Claims["email"].(string)
	userPicture := payload.Claims["picture"].(string)

	// Save user data in session
	session := sessions.Default(c)
	session.Set("userID", userID)
	session.Set("name", userName)
	session.Set("email", userEmail)
	session.Set("picture", userPicture)
	session.Set("logged_in", true)

	if err := session.Save(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save session"})
		return
	}
	h.Logger.Info(fmt.Sprintf("%s logged in", userEmail))
	c.Redirect(http.StatusFound, "/")

}

func (h *UserHandler) LogoutUser(c *gin.Context) {
	session := sessions.Default(c)
	user := session.Get("email")
	// Clear session data
	session.Clear()

	// Save the session (effectively logs the user out)
	if err := session.Save(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to clear session"})
		return
	}
	h.Logger.Info(fmt.Sprintf("%s logged out", user))
	c.Redirect(http.StatusFound, "/")

}
func (h *UserHandler) LoginUser(c *gin.Context) {
	c.HTML(http.StatusOK, "login.html", gin.H{})
}

func (h *UserHandler) GetDownloadUrl(c *gin.Context) {
	downloadUrl := c.Query("url")
	c.HTML(http.StatusOK, "download.html", gin.H{"Email": GetEmail(c), "Url": downloadUrl})
}
