package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"movie-light/internal/models"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var googleOAuthConfig *oauth2.Config

func InitOAuthConfigs() {
	googleOAuthConfig = &oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		RedirectURL:  "https://localhost:8080/auth/google/callback",
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
		Endpoint:     google.Endpoint,
	}
}

func GoogleAuthHandler(c *gin.Context) {
	moderatorRequest := c.Query("moderator_request") == "on"

	c.SetCookie("moderator_request", fmt.Sprintf("%v", moderatorRequest), 600, "/", "", false, true)

	url := googleOAuthConfig.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	c.Redirect(http.StatusTemporaryRedirect, url)
}

func GoogleCallbackHandler(c *gin.Context) {
	code := c.Query("code")
	cookie, err := c.Cookie("moderator_request")
	if err != nil {
		log.Println("GoogleCallbackHandler: Error reading moderator_request cookie:", err)
	}

	moderatorRequest := cookie == "true"

	token, err := googleOAuthConfig.Exchange(c.Request.Context(), code)
	if err != nil {
		log.Println("GoogleCallbackHandler: Google OAuth2 exchange failed:", err)
		c.Redirect(http.StatusTemporaryRedirect, "/login")
		return
	}

	client := googleOAuthConfig.Client(c.Request.Context(), token)
	userInfo, err := getGoogleUserInfo(client)
	if err != nil {
		log.Println("GoogleCallbackHandler: Failed to get Google user info:", err)
		c.Redirect(http.StatusTemporaryRedirect, "/login")
		return
	}

	sessionToken, err := models.AuthenticateOrRegisterOAuthUser(userInfo.Email, userInfo.Name, "google", moderatorRequest)
	if err != nil {
		log.Println("GoogleCallbackHandler: Failed to authenticate/register user:", err)
		c.Redirect(http.StatusTemporaryRedirect, "/login")
		return
	}

	c.SetCookie("session_token", sessionToken, 86400, "/", "", false, true)
	c.Redirect(http.StatusSeeOther, "/")
}

func getGoogleUserInfo(client *http.Client) (*models.OAuthUserInfo, error) {
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		log.Println("getGoogleUserInfo: Failed to call Google API:", err)
		return nil, err
	}
	defer resp.Body.Close()

	var userInfo models.OAuthUserInfo
	err = json.NewDecoder(resp.Body).Decode(&userInfo)
	if err != nil {
		log.Println("getGoogleUserInfo: Failed to decode user info:", err)
	}
	return &userInfo, err
}
