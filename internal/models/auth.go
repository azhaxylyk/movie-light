package models

import (
	"database/sql"
	"errors"
	"log"

	"github.com/gofrs/uuid"
	"golang.org/x/crypto/bcrypt"
)

type OAuthUserInfo struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

func AuthenticateOrRegisterOAuthUser(email, username, provider string, moderatorRequest bool) (string, error) {
	var userID string

	err := db.QueryRow("SELECT id FROM users WHERE email = ?", email).Scan(&userID)
	if err != nil {
		if err == sql.ErrNoRows {
			newUUID, err := uuid.NewV4()
			if err != nil {
				return "", err
			}
			userID = newUUID.String()
			sessionUUID, err := uuid.NewV4()
			if err != nil {
				return "", err
			}
			sessionToken := sessionUUID.String()

			_, err = db.Exec("INSERT INTO users (id, email, username, provider, session_token) VALUES (?, ?, ?, ?, ?)",
				userID, email, username, provider, sessionToken)
			if err != nil {
				return "", err
			}
			return sessionToken, nil
		}
		return "", err
	}

	sessionUUID, err := uuid.NewV4()
	if err != nil {
		return "", err
	}
	sessionToken := sessionUUID.String()

	_, err = db.Exec("UPDATE users SET session_token = ? WHERE id = ?", sessionToken, userID)
	if err != nil {
		return "", err
	}

	return sessionToken, nil
}

func CheckEmailExists(email string) (bool, error) {
	var exists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE email = ?)", email).Scan(&exists)
	if err != nil {
		log.Printf("Error checking if email exists: %v", err)
	}
	return exists, err
}

func CheckUsernameExists(username string) (bool, error) {
	var exists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE username = ?)", username).Scan(&exists)
	if err != nil {
		log.Printf("Error checking if username exists: %v", err)
	}
	return exists, err
}

func RegisterUser(email, username, password string, isModeratorRequest bool) (string, error) {

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Error hashing password: %v", err)
		return "", err
	}

	newUUID, err := uuid.NewV4()
	if err != nil {
		log.Printf("Error generating user UUID: %v", err)
		return "", err
	}
	userID := newUUID.String()

	sessionUUID, err := uuid.NewV4()
	if err != nil {
		log.Printf("Error generating session UUID: %v", err)
		return "", err
	}
	sessionToken := sessionUUID.String()

	provider := "email"

	role := "user"

	_, err = db.Exec("INSERT INTO users (id, email, username, password, session_token, role, provider) VALUES (?, ?, ?, ?, ?, ?, ?)",
		userID, email, username, hashedPassword, sessionToken, role, provider)
	if err != nil {
		log.Printf("Error inserting user into database: %v", err)
		return "", err
	}

	return sessionToken, nil
}

func AuthenticateUser(email, password string) (string, error) {
	var userID, hashedPassword string

	err := db.QueryRow("SELECT id, password FROM users WHERE email = ?", email).Scan(&userID, &hashedPassword)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", errors.New("invalid credentials")
		}
		return "", err
	}

	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		return "", errors.New("invalid credentials")
	}

	sessionUUID, _ := uuid.NewV4()
	sessionToken := sessionUUID.String()

	_, err = db.Exec("UPDATE users SET session_token = ? WHERE id = ?", sessionToken, userID)
	if err != nil {
		return "", err
	}

	return sessionToken, nil
}

func GetIDBySessionToken(sessionToken string) (string, string, error) {
	var username string
	var userID string
	err := db.QueryRow("SELECT id, username FROM users WHERE session_token = ?", sessionToken).Scan(&userID, &username)
	if err != nil {
		return "", "", err
	}
	return userID, username, nil
}
