package auth

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var googleOAuthConfig = &oauth2.Config{
	RedirectURL:  "http://localhost:8080/auth/google/callback",
	ClientID:     "527452364696-388fencg8tha8tp3so6ad8puqblvhfe7.apps.googleusercontent.com",
	ClientSecret: "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",
	Scopes: []string{"https://www.googleapis.com/auth/userinfo.profile",
		"https://www.googleapis.com/auth/userinfo.email",
	},
	Endpoint: google.Endpoint,
}

var database *sql.DB

func ConnectDB() error {
	var err error
	database, err = sql.Open("sqlite3", "database/forum.db")
	if err != nil {
		return err
	}
	return nil
}

func DownloadImage(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	imageBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return imageBytes, nil
}

func InsertOrUpdateUser(name, surname, email string, image []byte) error {
	err := ConnectDB()
	if err != nil {
		return err
	}

	exists, _, err := getUserID(email)
	if err != nil {
		return err
	}

	if exists {
		stmt, err := database.Prepare("UPDATE users SET name = ?, surname = ?, image = ? WHERE email = ?")
		if err != nil {
			return err
		}
		defer stmt.Close()

		_, err = stmt.Exec(name, surname, image, email)
		if err != nil {
			return err
		}
	} else {
		stmt, err := database.Prepare("INSERT INTO users(username, name, surname, password, email, image) VALUES(?, ?, ?, ?, ?, ?)")
		if err != nil {
			return err
		}
		defer stmt.Close()

		_, err = stmt.Exec("", name, surname, "", email, image)
		if err != nil {
			return err
		}
	}

	return nil
}

func HandleGoogleLogin(w http.ResponseWriter, r *http.Request) {
	url := googleOAuthConfig.AuthCodeURL("random")
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func HandleGoogleCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	token, err := googleOAuthConfig.Exchange(context.Background(), code)
	if err != nil {
		log.Fatalf("Code exchange failed: %s", err.Error())
		return
	}

	client := googleOAuthConfig.Client(context.Background(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		log.Fatalf("Failed to get user info: %s", err.Error())
		return
	}
	defer resp.Body.Close()

	var userInfo struct {
		ID            string `json:"id"`
		Email         string `json:"email"`
		VerifiedEmail bool   `json:"verified_email"`
		Name          string `json:"given_name"`
		Surname       string `json:"family_name"`
		Picture       string `json:"picture"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		log.Fatalf("Failed to decode user info: %s", err.Error())
		return
	}

	imageBytes, err := DownloadImage(userInfo.Picture)
	if err != nil {
		log.Fatalf("Failed to download user profile image: %s", err.Error())
		return
	}

	err = InsertOrUpdateUser(userInfo.Name, userInfo.Surname, userInfo.Email, imageBytes)
	if err != nil {
		log.Fatalf("Failed to insert or update user: %s", err.Error())
		return
	}

	_, userID, err := getUserID(userInfo.Email)
	if err != nil {
		fmt.Println(err)
	}
	cookie := http.Cookie{
		Name:     "session_token",
		Value:    strconv.Itoa(userID),
		Expires:  time.Now().Add(24 * time.Hour),
		MaxAge:   86400,
		Path:     "/",
		HttpOnly: true,
	}
	http.SetCookie(w, &cookie)

	http.Redirect(w, r, "/profile", http.StatusSeeOther)
}

func getUserID(email string) (bool, int, error) {
	err := ConnectDB()
	if err != nil {
		return false, 0, err
	}
	var userID int
	query := "SELECT id FROM users WHERE email = ?"
	err = database.QueryRow(query, email).Scan(&userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, 0, nil
		}
		return false, 0, err
	}
	return true, userID, nil
}
