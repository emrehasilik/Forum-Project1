package main

import (
	"database/sql"
	"encoding/base64"
	"fmt"
	"html/template"
	"log"
	"net/http"

	_ "github.com/mattn/go-sqlite3"
)

type sPostData struct {
	PostID           int
	PostTitle        string
	PostContent      string
	PostLikeCount    int
	PostDislikeCount int
	PostImage        string
}

type TemplateData struct {
	Posts      []sPostData
	IsLoggedIn bool
}

// convertImageToBase64 converts image data to a base64-encoded string
func convertImageToBase64(image []byte) string {
	if image == nil {
		return ""
	}
	return base64.StdEncoding.EncodeToString(image)
}

// handleFilter handles filtering requests based on categories, likes, and all posts
func handleFilter(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		fmt.Println("test")
		// Render the response using the template
		tmpl := template.Must(template.ParseFiles("static/html/discovered.html"))
		err := tmpl.Execute(w, TemplateData{})
		if err != nil {
			log.Println("Error executing template:", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	} else {
		// Open database connection
		db, err := sql.Open("sqlite3", "database/forum.db")
		if err != nil {
			log.Println("Error connecting to database:", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		defer db.Close()

		// Parse form data
		err = r.ParseForm()
		if err != nil {
			log.Println("Error parsing form data:", err)
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		// Get filter type from form data
		filter := r.FormValue("filter")
		keyword := r.Form.Get("keyword")
		keyword = "%" + keyword + "%"
		var query string

		switch filter {
		case "allPosts":
			query = `
			SELECT id, title, content, like_count, dislike_count, image 
			FROM posts 
			WHERE title LIKE ? OR content LIKE ? OR category LIKE ?
			ORDER BY created_at DESC
		`
		case "categories":
			query = `
			SELECT id, title, content, like_count, dislike_count, image 
			FROM posts 
			WHERE category LIKE ? AND (title LIKE ? OR content LIKE ? OR category LIKE ?)
			ORDER BY created_at DESC
		`
		default:
			log.Println("Unknown filter type:", filter)
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		// Execute the query
		rows, err := db.Query(query, keyword, keyword, keyword)
		if err != nil {
			log.Println("Error querying database:", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		// Process query results
		var posts []sPostData
		for rows.Next() {
			var post sPostData
			var image []byte
			err = rows.Scan(&post.PostID, &post.PostTitle, &post.PostContent, &post.PostLikeCount, &post.PostDislikeCount, &image)
			if err != nil {
				log.Println("Error scanning row:", err)
				continue
			}
			post.PostImage = convertImageToBase64(image)
			posts = append(posts, post)
		}
		if err := rows.Err(); err != nil {
			log.Println("Error iterating over rows:", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		fmt.Println(posts)
		tmpl := template.Must(template.ParseFiles("static/html/discovered.html"))
		err = tmpl.Execute(w, posts)
		if err != nil {
			log.Println("Error executing template:", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	}
}
