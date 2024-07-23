package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/facebook"
)

var facebookOAuthConfig *oauth2.Config

func init() {
	err := godotenv.Load("auth/.env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	facebookOAuthConfig = &oauth2.Config{
		RedirectURL:  "http://localhost:8080/auth/facebook/callback",
		ClientID:     os.Getenv("FACEBOOK_CLIENT_ID"),
		ClientSecret: os.Getenv("FACEBOOK_CLIENT_SECRET"),
		Scopes:       []string{"public_profile", "email"},
		Endpoint:     facebook.Endpoint,
	}
}

func HandleFacebookLogin(w http.ResponseWriter, r *http.Request) {
	url := facebookOAuthConfig.AuthCodeURL("random")
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func HandleFacebookCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	token, err := facebookOAuthConfig.Exchange(context.Background(), code)
	if err != nil {
		log.Fatalf("Code exchange failed: %s", err.Error())
		return
	}

	client := facebookOAuthConfig.Client(context.Background(), token)
	resp, err := client.Get("https://graph.facebook.com/me?fields=id,name,email,picture")
	if err != nil {
		log.Fatalf("Failed to get user info: %s", err.Error())
		return
	}
	defer resp.Body.Close()

	var userInfo struct {
		ID      string `json:"id"`
		Name    string `json:"name"`
		Email   string `json:"email"`
		Picture struct {
			Data struct {
				URL string `json:"url"`
			} `json:"data"`
		} `json:"picture"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		log.Fatalf("Failed to decode user info: %s", err.Error())
		return
	}

	imageBytes, err := DownloadImage(userInfo.Picture.Data.URL)
	if err != nil {
		log.Fatalf("Failed to download user profile image: %s", err.Error())
		return
	}

	err = InsertOrUpdateUser(userInfo.Name, "", userInfo.Email, imageBytes)
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
