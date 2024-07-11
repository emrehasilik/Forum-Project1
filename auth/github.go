package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

var githubOAuthConfig = &oauth2.Config{
	RedirectURL:  "http://localhost:8080/auth/github/callback",
	ClientID:     "Ov23liJ1gCxwPv6pT79K",
	ClientSecret: "596b06de3634313475b92eacba984dabb87d28fd",
	Scopes:       []string{"user:email"},
	Endpoint:     github.Endpoint,
}

func HandleGithubLogin(w http.ResponseWriter, r *http.Request) {
	url := githubOAuthConfig.AuthCodeURL("random")
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func HandleGithubCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	token, err := githubOAuthConfig.Exchange(context.Background(), code)
	if err != nil {
		log.Fatalf("Code exchange failed: %s", err.Error())
		return
	}

	client := githubOAuthConfig.Client(context.Background(), token)
	resp, err := client.Get("https://api.github.com/user")
	if err != nil {
		log.Fatalf("Failed to get user info: %s", err.Error())
		return
	}
	defer resp.Body.Close()

	var userInfo struct {
		ID    int    `json:"id"`
		Email string `json:"email"`
		Name  string `json:"name"`
		Login string `json:"login"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		log.Fatalf("Failed to decode user info: %s", err.Error())
		return
	}

	// Fetch email if not provided
	if userInfo.Email == "" {
		emailResp, err := client.Get("https://api.github.com/user/emails")
		if err != nil {
			log.Fatalf("Failed to get user emails: %s", err.Error())
			return
		}
		defer emailResp.Body.Close()

		var emails []struct {
			Email    string `json:"email"`
			Primary  bool   `json:"primary"`
			Verified bool   `json:"verified"`
		}
		if err := json.NewDecoder(emailResp.Body).Decode(&emails); err != nil {
			log.Fatalf("Failed to decode user emails: %s", err.Error())
			return
		}

		for _, email := range emails {
			if email.Primary && email.Verified {
				userInfo.Email = email.Email
				break
			}
		}
	}

	err = InsertOrUpdateUser(userInfo.Name, userInfo.Login, userInfo.Email, nil)
	if err != nil {
		log.Fatalf("Failed to insert or update user: %s", err.Error())
		return
	}

	_, userID, err := getUserID(userInfo.Email)
	if err != nil {

		fmt.Println(err)
	}
	fmt.Println(userID)
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
