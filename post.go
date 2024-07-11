package main

import (
	"fmt"
	"html/template"
	"net/http"
)

func PostPageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		cookie, _ := r.Cookie("session_token")
		userID := cookie.Value
		commentContent := r.FormValue("comment")
		commentedPostID := r.FormValue("commentPostID")
		addCommentToDb(commentContent, commentedPostID, userID)

		// Extract query parameters from the original URL
		redirectURL := "/postpage?id=" + commentedPostID
		http.Redirect(w, r, redirectURL, http.StatusFound)

	} else if r.Method == http.MethodGet {
		cookie, _ := r.Cookie("session_token")
		postID := r.FormValue("id")
		post, _ := getPostById(postID)
		var data struct {
			PostData   PostData
			IsLoggedIn bool
		}
		if cookie != nil {
			data = struct {
				PostData   PostData
				IsLoggedIn bool
			}{
				PostData:   post,
				IsLoggedIn: true,
			}
		} else {
			data = struct {
				PostData   PostData
				IsLoggedIn bool
			}{
				PostData:   post,
				IsLoggedIn: false,
			}
		}

		tmpl := template.Must(template.ParseFiles("static/html/postpage.html"))
		err := tmpl.Execute(w, data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func likeCommentHandler(w http.ResponseWriter, r *http.Request) {
	cookie, _ := r.Cookie("session_token")
	if cookie == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	commentID := r.FormValue("commentid")
	//postID := r.FormValue("")
	userID := cookie.Value
	isLiked := likesCommentController(userID, commentID)
	isDisliked := dislikesCommentController(userID, commentID)
	if isLiked {
		deleteLikedComments(userID, commentID)
		decreaseLikeCommentCount(commentID)
	} else if !isLiked && !isDisliked {
		insertLikedComment(userID, commentID)
		err := increaseLikeCommentCount(commentID)
		if err != nil {
			fmt.Println("Error liking post", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}
	http.Redirect(w, r, "/home", http.StatusFound)
}

func insertLikedComment(userID, commentID string) {
	connectDatabase()
	stmt, err := database.Prepare("insert into comment_likes (user_id, comment_id) values (?, ?)")
	checkError(err)
	defer stmt.Close()

	_, err = stmt.Exec(userID, commentID)
	checkError(err)
}

func likesCommentController(userID, commentID string) bool {
	connectDatabase()
	query := "SELECT COUNT(*) FROM comment_likes WHERE user_id = ? AND comment_id = ?"
	var count int
	err := database.QueryRow(query, userID, commentID).Scan(&count)
	if err != nil {
		fmt.Println("Error executing query:", err)
		return false
	}

	// Sonuçları kontrol et
	if count > 0 {
		fmt.Println("Record exists")
		return true
	} else {
		fmt.Println("Record does not exist")
		return false
	}
}

func deleteLikedComments(userID, commentID string) {
	connectDatabase()
	query := "DELETE FROM comment_likes WHERE user_id = ? AND comment_id = ?"

	result, err := database.Exec(query, userID, commentID)
	if err != nil {
		fmt.Println("Error executing delete query:", err)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		fmt.Println("Error getting rows affected:", err)
		return
	}

	if rowsAffected > 0 {
		fmt.Println("Like successfully removed")
	} else {
		fmt.Println("No like found to remove")
	}
}

func dislikeCommentHandler(w http.ResponseWriter, r *http.Request) {
	cookie, _ := r.Cookie("session_token")
	if cookie == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	commentID := r.FormValue("commentid")
	userID := cookie.Value
	isDisliked := dislikesCommentController(userID, commentID)
	isLiked := likesCommentController(userID, commentID)
	if isDisliked {
		deleteCommentDislikedPost(userID, commentID)
		decreaseDislikeCommentCount(commentID)
	} else if !isDisliked && !isLiked {
		insertCommentDislikedPost(userID, commentID)
		err := increaseDislikeCommentCount(commentID)
		if err != nil {
			fmt.Println("Error liking post", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}

	http.Redirect(w, r, "/home", http.StatusFound)
}

func dislikesCommentController(userID, commentID string) bool {
	connectDatabase()
	query := "SELECT COUNT(*) FROM comment_dislikes WHERE user_id = ? AND comment_id = ?"
	var count int
	err := database.QueryRow(query, userID, commentID).Scan(&count)
	if err != nil {
		fmt.Println("Error executing query:", err)
		return false
	}

	// Sonuçları kontrol et
	if count > 0 {
		fmt.Println("Record exists")
		return true
	} else {
		fmt.Println("Record does not exist")
		return false
	}
}

func insertCommentDislikedPost(userID, commentID string) {
	connectDatabase()
	stmt, err := database.Prepare("insert into comment_dislikes (user_id, comment_id) values (?, ?)")
	checkError(err)
	defer stmt.Close()

	_, err = stmt.Exec(userID, commentID)
	checkError(err)
}

func deleteCommentDislikedPost(userID, commentID string) {
	connectDatabase()
	query := "DELETE FROM comment_dislikes WHERE user_id = ? AND comment_id = ?"

	result, err := database.Exec(query, userID, commentID)
	if err != nil {
		fmt.Println("Error executing delete query:", err)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		fmt.Println("Error getting rows affected:", err)
		return
	}

	if rowsAffected > 0 {
		fmt.Println("Like successfully removed")
	} else {
		fmt.Println("No like found to remove")
	}
}
