package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"movie-light/internal/models"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
	"unicode"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

var googleOAuthConfig *oauth2.Config
var githubOAuthConfig *oauth2.Config

func InitOAuthConfigs() {
	googleOAuthConfig = &oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		RedirectURL:  "https://localhost:8080/auth/google/callback",
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
		Endpoint:     google.Endpoint,
	}

	githubOAuthConfig = &oauth2.Config{
		ClientID:     os.Getenv("GITHUB_CLIENT_ID"),
		ClientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),
		RedirectURL:  "https://localhost:8080/auth/github/callback",
		Scopes:       []string{"user:email"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://github.com/login/oauth/authorize",
			TokenURL: "https://github.com/login/oauth/access_token",
		},
	}
}

func GitHubAuthHandler(w http.ResponseWriter, r *http.Request) {
	moderatorRequest := r.FormValue("moderator_request") == "on"

	http.SetCookie(w, &http.Cookie{
		Name:     "moderator_request",
		Value:    fmt.Sprintf("%v", moderatorRequest),
		Expires:  time.Now().Add(10 * time.Minute),
		HttpOnly: true,
		Path:     "/",
	})

	url := githubOAuthConfig.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func GitHubCallbackHandler(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	cookie, err := r.Cookie("moderator_request")
	if err != nil {
		log.Println("GitHubCallbackHandler: Error reading moderator_request cookie:", err)
	}

	moderatorRequest := cookie != nil && cookie.Value == "true"

	token, err := githubOAuthConfig.Exchange(context.Background(), code)
	if err != nil {
		log.Println("GitHubCallbackHandler: GitHub OAuth2 exchange failed:", err)
		http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
		return
	}

	client := githubOAuthConfig.Client(context.Background(), token)
	userInfo, err := getGitHubUserInfo(client)
	if err != nil {
		log.Println("GitHubCallbackHandler: Failed to get GitHub user info:", err)
		http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
		return
	}

	sessionToken, err := models.AuthenticateOrRegisterOAuthUser(userInfo.Email, userInfo.Name, "github", moderatorRequest)
	if err != nil {
		log.Println("GitHubCallbackHandler: Failed to authenticate/register user:", err)
		http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
		return
	}

	cookie = &http.Cookie{
		Name:    "session_token",
		Value:   sessionToken,
		Expires: time.Now().Add(24 * time.Hour),
		Path:    "/",
	}
	http.SetCookie(w, cookie)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func GoogleAuthHandler(w http.ResponseWriter, r *http.Request) {
	moderatorRequest := r.FormValue("moderator_request") == "on"

	http.SetCookie(w, &http.Cookie{
		Name:     "moderator_request",
		Value:    fmt.Sprintf("%v", moderatorRequest),
		Expires:  time.Now().Add(10 * time.Minute),
		HttpOnly: true,
		Path:     "/",
	})
	url := googleOAuthConfig.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func GoogleCallbackHandler(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	cookie, err := r.Cookie("moderator_request")
	if err != nil {
		log.Println("GoogleCallbackHandler: Error reading moderator_request cookie:", err)
	}

	moderatorRequest := cookie != nil && cookie.Value == "true"

	token, err := googleOAuthConfig.Exchange(context.Background(), code)
	if err != nil {
		log.Println("GoogleCallbackHandler: Google OAuth2 exchange failed:", err)
		http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
		return
	}
	client := googleOAuthConfig.Client(context.Background(), token)
	userInfo, err := getGoogleUserInfo(client)
	if err != nil {
		log.Println("GoogleCallbackHandler: Failed to get Google user info:", err)
		http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
		return
	}

	sessionToken, err := models.AuthenticateOrRegisterOAuthUser(userInfo.Email, userInfo.Name, "google", moderatorRequest)
	if err != nil {
		log.Println("GoogleCallbackHandler: Failed to authenticate/register user:", err)
		http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
		return
	}

	cookie = &http.Cookie{
		Name:     "session_token",
		Value:    sessionToken,
		Expires:  time.Now().Add(24 * time.Hour),
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, cookie)

	http.Redirect(w, r, "/", http.StatusSeeOther)
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

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		email := sanitizeInput(r.FormValue("email"))
		username := sanitizeInput(r.FormValue("username"))
		password := r.FormValue("password")
		moderatorRequest := r.FormValue("moderator_request") == "on"
		if email == "" || username == "" || password == "" {
			ErrorHandler(w, r, http.StatusBadRequest, "All fields are required")
			log.Println("Error: Missing required fields")
			return
		}

		if !isValidEmail(email) {
			ErrorHandler(w, r, http.StatusBadRequest, "Invalid email format")
			log.Println("Error: Invalid email format")
			return
		}

		emailExists, err := models.CheckEmailExists(email)
		if err != nil {
			log.Printf("Error checking if email exists: %v", err)
			ErrorHandler(w, r, http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
			return
		}
		if emailExists {
			renderTemplateWithError(w, "register", "Email is already registered")
			log.Println("Error: Email already exists")
			return
		}

		usernameExists, err := models.CheckUsernameExists(username)
		if err != nil {
			log.Printf("Error checking if username exists: %v", err)
			ErrorHandler(w, r, http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
			return
		}
		if usernameExists {
			renderTemplateWithError(w, "register", "Username is already taken")
			log.Println("Error: Username already exists")
			return
		}

		sessionToken, err := models.RegisterUser(email, username, password, moderatorRequest)
		if err != nil {
			log.Printf("Error during user registration: %v", err)
			ErrorHandler(w, r, http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
			return
		}

		cookie := http.Cookie{
			Name:     "session_token",
			Value:    sessionToken,
			Expires:  time.Now().Add(24 * time.Hour),
			Path:     "/",
			HttpOnly: true,
			Secure:   false,
			SameSite: http.SameSiteLaxMode,
		}

		http.SetCookie(w, &cookie)

		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	tmpl, _ := template.ParseFiles("web/templates/register.html")
	tmpl.Execute(w, nil)
}

func sanitizeInput(input string) string {
	trimmed := strings.TrimSpace(input)
	return strings.Map(func(r rune) rune {
		if unicode.IsPrint(r) && !unicode.IsSpace(r) {
			return r
		}
		return -1
	}, trimmed)
}

func renderTemplateWithError(w http.ResponseWriter, templateName, errorMessage string) {
	tmpl, _ := template.ParseFiles("web/templates/" + templateName + ".html")
	tmpl.Execute(w, struct{ Error string }{Error: errorMessage})
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		email := r.FormValue("email")
		password := r.FormValue("password")

		sessionToken, err := models.AuthenticateUser(email, password)
		if err != nil {
			tmpl, _ := template.ParseFiles("web/templates/login.html")
			tmpl.Execute(w, struct{ Error string }{Error: "Invalid email or password"})
			return
		}

		cookie := http.Cookie{
			Name:     "session_token",
			Value:    sessionToken,
			Expires:  time.Now().Add(24 * time.Hour),
			Path:     "/",
			HttpOnly: true,
			Secure:   false,
			SameSite: http.SameSiteLaxMode,
		}

		http.SetCookie(w, &cookie)

		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	tmpl, _ := template.ParseFiles("web/templates/login.html")
	tmpl.Execute(w, nil)
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	cookie := http.Cookie{
		Name:    "session_token",
		Value:   "",
		Expires: time.Now().Add(-1 * time.Hour),
	}
	http.SetCookie(w, &cookie)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func isValidEmail(email string) bool {
	return emailRegex.MatchString(email)
}

func getGitHubUserInfo(client *http.Client) (*models.OAuthUserInfo, error) {
	resp, err := client.Get("https://api.github.com/user")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	type GitHubUser struct {
		Email string `json:"email"`
		Name  string `json:"name"`
	}

	var user GitHubUser
	err = json.NewDecoder(resp.Body).Decode(&user)
	if err != nil {
		return nil, err
	}

	if user.Email == "" {
		emailResp, err := client.Get("https://api.github.com/user/emails")
		if err != nil {
			return nil, err
		}
		defer emailResp.Body.Close()

		var emails []struct {
			Email   string `json:"email"`
			Primary bool   `json:"primary"`
		}
		err = json.NewDecoder(emailResp.Body).Decode(&emails)
		if err != nil {
			return nil, err
		}

		for _, e := range emails {
			if e.Primary {
				user.Email = e.Email
				break
			}
		}
	}

	return &models.OAuthUserInfo{
		Email: user.Email,
		Name:  user.Name,
	}, nil
}
