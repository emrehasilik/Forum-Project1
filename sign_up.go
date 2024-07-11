package main

import (
	"database/sql"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

var tmpl = template.Must(template.ParseFiles("static/html/sign_up.html"))

func SignupHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		cookie, _ := r.Cookie("session_token")
		if cookie != nil {
			http.Redirect(w, r, "/home", http.StatusSeeOther)
			return
		}
		tmpl.Execute(w, nil)
		return
	}

	username := r.FormValue("username")
	name := r.FormValue("name")
	surname := r.FormValue("surname")
	email := r.FormValue("email")
	password := r.FormValue("password")
	// tckimlik :=r.FormValue("Tc-Kimlik")

	message := ""

	if username == "" || name == "" || surname == "" || email == "" || password == "" {
		message = "Lütfen tüm alanları doldurun."
		tmpl.Execute(w, struct{ Message string }{Message: message})
		return
	}

	connectDatabase()

	var existingUsername, existingEmail string
	err := database.QueryRow("SELECT username FROM users WHERE username = ?", username).Scan(&existingUsername)
	if err != nil && err != sql.ErrNoRows {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = database.QueryRow("SELECT email FROM users WHERE email = ?", email).Scan(&existingEmail)
	if err != nil && err != sql.ErrNoRows {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if existingUsername != "" {
		message += "Kullanıcı adı mevcut."
	}

	if existingEmail != "" {
		message += " E-posta adresi mevcut."
	}

	if message != "" {
		tmpl.Execute(w, struct{ Message string }{Message: message})
		return
	}

	// Şifreyi hashleyelim
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Kullanıcı adı ve e-posta adresi yoksa, yeni kayıt oluşturulur.
	stmt, err := database.Prepare("INSERT INTO users(username, name, surname, email, password, image, banner) VALUES(?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(username, name, surname, email, hashedPassword, getDefaultImage(), getDefaultImage())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Kayıt başarılı mesajı
	tmpl.Execute(w, nil)
}

func getDefaultImage() []byte {
	file, err := os.Open("static/images/newuser.png")
	if err != nil {
		log.Println("Error retrieving the file")
		return nil
	}
	defer file.Close()

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		log.Println("Error reading file")
		return nil
	}
	return fileBytes
}
