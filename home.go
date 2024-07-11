package main

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

func homePageHandler(w http.ResponseWriter, r *http.Request) {
	cookie, _ := r.Cookie("session_token")
	if r.Method == http.MethodPost {
		title := r.FormValue("title")
		content := r.FormValue("content")
		categories := findHashtaggedWords(content)
		err := r.ParseMultipartForm(10 << 20) // 10 MB
		if err != nil {
			fmt.Println(err)
			http.Error(w, "Error parsing form", http.StatusInternalServerError)
			return
		}

		postImage, _, err := r.FormFile("postPicture")
		var fileBytes []byte

		if err != nil {
			if err != http.ErrMissingFile {
				fmt.Println(err)
				log.Println("Error retrieving the profile picture")
			} else {
				fileBytes = []byte{} // No file uploaded, set to empty byte slice
			}
		} else {
			fileBytes, err = io.ReadAll(postImage)
			if err != nil {
				log.Println("Error reading file")
				return
			}
		}

		if len(title) >= 5 && len(content) >= 10 {
			userID, _ := strconv.Atoi(cookie.Value)
			insertPost(userID, title, content,categories, fileBytes)
		}
		http.Redirect(w, r, "/home", http.StatusSeeOther)
	} else {
		topPosts, err := getTopPosts()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if cookie != nil {
			currentUser, _ := getUser(cookie)
			homeData := struct {
				User       User
				Posts      []PostData
				IsLoggedIn bool
				TopPosts   []PostData
			}{
				User:       currentUser,
				Posts:      getAllPosts(),
				IsLoggedIn: true,
				TopPosts:   topPosts,
			}

			tmpl := template.Must(template.ParseFiles("static/html/homepage.html"))
			err := tmpl.Execute(w, homeData)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		} else {
			homeData2 := struct {
				Posts      []PostData
				IsLoggedIn bool
				TopPosts   []PostData
			}{
				Posts:      getAllPosts(),
				IsLoggedIn: false,
				TopPosts:   topPosts,
			}

			tmpl := template.Must(template.ParseFiles("static/html/homepage.html"))
			err := tmpl.Execute(w, homeData2)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		}
	}
}

func findHashtaggedWords(text string) string {
	// Split the text into words
	words := strings.Fields(text)
	// Create a slice to store the results
	var results []string

	// Loop through each word
	for _, word := range words {
		// Check if the word starts with '#'
		if strings.HasPrefix(word, "#") {
			results = append(results, word)
		}
	}

	// Join the results slice with ',' and return the string
	return strings.Join(results, ",")
}
