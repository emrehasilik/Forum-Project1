package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

func authenticateUser(email, password string) (bool, int, error) {
	var storedPassword string
	var userID int
	query := "SELECT id, password FROM users WHERE email = ?"
	err := database.QueryRow(query, email).Scan(&userID, &storedPassword)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, 0, nil
		}
		return false, 0, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(password))
	if err != nil {
		return false, 0, nil
	}
	return true, userID, nil
}

// Kullanıcıyı giriş yapmış olarak ayarlar
func setSession(w http.ResponseWriter, userID int, email string, rememberMe bool) error {
	var expiration time.Time
	if rememberMe {
		expiration = time.Now().Add(30 * 24 * time.Hour)
	} else {
		expiration = time.Now().Add(24 * time.Hour)
	}

	cookie := http.Cookie{
		Name:     "session_token",
		Value:    strconv.Itoa(userID),
		Expires:  expiration,
		HttpOnly: true,
	}
	http.SetCookie(w, &cookie)

	// Kullanıcı ID ve email'i bellekte sakla
	sessionStore[strconv.Itoa(userID)] = email

	return nil
}

// Bellekte oturumları sakla (demonstrasyon amacıyla)
var sessionStore = map[string]string{}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		cookie, _ := r.Cookie("session_token")
		if cookie != nil {
			http.Redirect(w, r, "/home", http.StatusSeeOther)
			return
		}
		tmpl, _ := template.ParseFiles("./static/html/login.html")
		tmpl.Execute(w, nil)
		return
	}

	email := r.FormValue("email")
	password := r.FormValue("password")
	rememberMe := r.FormValue("remember_me") == "on"

	connectDatabase()

	authenticated, userID, err := authenticateUser(email, password)
	if err != nil {
		http.Error(w, "Authentication error", http.StatusInternalServerError)
		fmt.Println("Authentication error")
		return
	}

	if !authenticated {
		tmpl, _ := template.ParseFiles("./static/html/login.html")
		data := struct {
			ErrorMessage string
		}{
			ErrorMessage: "Geçersiz e-posta veya şifre",
		}
		tmpl.Execute(w, data)

		return
	}

	err = setSession(w, userID, email, rememberMe)
	if err != nil {
		http.Error(w, "Session error", http.StatusInternalServerError)
		fmt.Println("Session error")
		return
	}

	// Ana sayfaya yönlendir
	http.Redirect(w, r, "/homepage", http.StatusSeeOther)
}
