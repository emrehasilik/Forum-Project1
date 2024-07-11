package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	_ "github.com/mattn/go-sqlite3"
)

var database *sql.DB

func connectDatabase() error {
	var err error
	database, err = sql.Open("sqlite3", "database/forum.db")
	if err != nil {
		return err
	}
	return nil
}

// func InsertOrUpdateUser(id, email, picture string) error {
// 	query := `INSERT INTO users (id, email, image) VALUES (?, ?, ?)
// 			  ON CONFLICT(email) DO UPDATE SET id=excluded.id, image=excluded.image`
// 	_, err := database.Exec(query, id, email, picture)
// 	return err
// }
func getUser(cookie *http.Cookie) (User, bool) {
	err := connectDatabase()
	if err != nil {
		return User{}, false
	}
	defer database.Close()

	var user User
	var profileImg, bannerImg Image
	query := "SELECT username, email, name,surname,created_at,image,banner FROM users WHERE id = ?"
	err = database.QueryRow(query, cookie.Value).Scan(&user.Username, &user.Email, &user.Name, &user.Surname, &user.CreationDate, &profileImg.ImageData, &bannerImg.ImageData)
	user.ProfileImage = convertImg(profileImg)
	user.BannerImage = convertImg(bannerImg)
	if err != nil {
		fmt.Println(err)
		return User{}, false
	}
	return user, true
}

func getUserFullInfo(cookie *http.Cookie) (User, bool) {
	err := connectDatabase()
	if err != nil {
		fmt.Println(err)
		return User{}, false
	}

	var user User
	var profileImg, bannerImg Image
	query := "SELECT username, email,biography, password,name,surname,image,banner FROM users WHERE id = ?"
	err = database.QueryRow(query, cookie.Value).Scan(&user.Username, &user.Email, &user.Biography, &user.Password, &user.Name, &user.Surname, &profileImg.ImageData, &bannerImg.ImageData)
	user.ProfileImage = convertImg(profileImg)
	user.BannerImage = convertImg(bannerImg)
	if err != nil {
		fmt.Println(err)
		return User{}, false
	}
	fmt.Println(user)
	defer database.Close()

	return user, true
}

func updateUser(cookie *http.Cookie, user User) error {
	err := connectDatabase()
	if err != nil {
		fmt.Println(err)
		return err
	}
	// Hazırlık işlemi (Prepare statement)
	query, err := database.Prepare("UPDATE users SET username = ?, email = ?, biography = ?,name = ?, surname = ?, password = ? WHERE id = ?")
	if err != nil {
		return err
	}
	defer query.Close()

	// Parametreleri geçirerek sorguyu çalıştırma
	_, err = query.Exec(user.Username, user.Email, user.Biography, user.Name, user.Surname, user.Password, cookie.Value)
	if err != nil {
		return err
	}

	fmt.Println("Kullanıcı başarıyla güncellendi")
	return nil
}

func getPosts(userID string) ([]Post, error) {
	err := connectDatabase()
	if err != nil {
		return []Post{}, err
	}
	defer database.Close()

	query := `SELECT id, title, content, created_at, like_count, dislike_count, image FROM posts WHERE user_id = ? ORDER BY created_at DESC`
	rows, err := database.Query(query, userID)
	if err != nil {
		return []Post{}, err
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var post Post
		var postImage Image
		if err := rows.Scan(&post.PostID, &post.PostTitle, &post.PostContent, &post.PostCreatedAt, &post.PostLikeCount, &post.PostDislikeCount, &postImage.ImageData); err != nil {
			return []Post{}, err
		}
		post.PostImage = convertImg(postImage)
		posts = append(posts, post)
	}

	if err := rows.Err(); err != nil {
		return []Post{}, err
	}

	return posts, nil
}

func getPostById(postID string) (PostData, error) {
	err := connectDatabase()
	if err != nil {
		fmt.Println(err)
		return PostData{}, err
	}
	defer database.Close()

	query := `select users.image,users.username,users.name,users.surname, posts.id,posts.title, posts.content ,posts.created_at, posts.like_count, posts.dislike_count,posts.image
		from posts
		join users on posts.user_id = users.id
		where posts.id = ?`

	rows, err := database.Query(query, postID)
	if err != nil {
		fmt.Println(err)
		return PostData{}, err
	}
	defer rows.Close()

	var post PostData
	for rows.Next() {
		var postImage Image
		var userImage Image
		if err := rows.Scan(&userImage.ImageData, &post.UserData.Username, &post.UserData.Name, &post.UserData.Surname, &post.PostData.PostID, &post.PostData.PostTitle, &post.PostData.PostContent, &post.PostData.PostCreatedAt, &post.PostData.PostLikeCount, &post.PostData.PostDislikeCount, &postImage.ImageData); err != nil {
			fmt.Println(err)
			return PostData{}, err
		}
		post.PostData.PostImage = convertImg(postImage)
		post.CommentsData = getCommentsByPostID(postID)
	}
	if err := rows.Err(); err != nil {
		fmt.Println(err)
		return PostData{}, err
	}

	return post, nil
}

func getCommentsByPostID(postID string) []Comment {
	connectDatabase()
	query := `SELECT 
			comments.comment_id, comments.content, comments.created_at, comments.like_count, comments.dislike_count,
			users.username
		FROM 
			comments
		INNER JOIN 
			users ON comments.user_id = users.id
		WHERE 
			comments.post_id = ?`

	rows, err := database.Query(query, postID)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	defer rows.Close()

	var comments []Comment
	for rows.Next() {
		var comment Comment
		err := rows.Scan(&comment.CommentID, &comment.CommentContent, &comment.CommentCreatedAt, &comment.CommentLikeCount, &comment.CommentDislikeCount, &comment.CommentOwner)
		if err != nil {
			fmt.Println(err)
			return nil
		}
		comments = append(comments, comment)
	}
	return comments
}

func getAllPosts() []PostData {
	err := connectDatabase()
	checkError(err)
	defer database.Close()
	rows, err := database.Query(`
		select users.image,users.username,users.name,users.surname, posts.id,posts.title, posts.content ,posts.created_at, posts.like_count, posts.dislike_count,posts.image
		from posts
		join users on posts.user_id = users.id
		order by posts.created_at desc
	`)
	checkError(err)
	defer rows.Close()

	var posts []PostData
	for rows.Next() {
		var tempPostData PostData
		var postImage Image
		var userImage Image
		err = rows.Scan(&userImage.ImageData, &tempPostData.UserData.Username, &tempPostData.UserData.Name, &tempPostData.UserData.Surname, &tempPostData.PostData.PostID, &tempPostData.PostData.PostTitle, &tempPostData.PostData.PostContent, &tempPostData.PostData.PostCreatedAt, &tempPostData.PostData.PostLikeCount, &tempPostData.PostData.PostDislikeCount, &postImage.ImageData)
		tempPostData.PostData.PostImage = convertImg(postImage)
		tempPostData.UserData.ProfileImage = convertImg(userImage)
		checkError(err)
		posts = append(posts, tempPostData)
	}
	return posts
}

func checkError(err error) {
	if err != nil {
		panic(err)
	}
}

func insertPost(userID int, title, content, categories string, image []byte) {
	connectDatabase()
	var stmt *sql.Stmt
	var err error
	if image == nil {
		stmt, err = database.Prepare("insert into posts (user_id, title, content,category) values (?, ?, ?,?)")
		checkError(err)
		defer stmt.Close()

		_, err = stmt.Exec(userID, title, content, categories)
		checkError(err)
	} else {
		stmt, err = database.Prepare("insert into posts (user_id, title, content,category,image) values (?, ?, ?,?,?)")
		checkError(err)
		defer stmt.Close()

		_, err = stmt.Exec(userID, title, content, categories, image)
		checkError(err)
	}
}

func increaseLikeCount(postID string) error {
	err := connectDatabase()
	if err != nil {
		fmt.Println(err)
		return err
	}
	// Hazırlık işlemi (Prepare statement)
	query, err := database.Prepare("UPDATE posts SET like_count = like_count + 1 WHERE id = ?")
	if err != nil {
		return err
	}
	defer query.Close()

	// Parametreleri geçirerek sorguyu çalıştırma
	_, err = query.Exec(postID)
	if err != nil {
		return err
	}

	return nil
}

func decreaseLikeCount(postID string) error {
	err := connectDatabase()
	if err != nil {
		fmt.Println(err)
		return err
	}
	// Hazırlık işlemi (Prepare statement)
	query, err := database.Prepare("UPDATE posts SET like_count = like_count - 1 WHERE id = ?")
	if err != nil {
		return err
	}
	defer query.Close()

	// Parametreleri geçirerek sorguyu çalıştırma
	_, err = query.Exec(postID)
	if err != nil {
		return err
	}

	return nil
}

func increaseDislikeCount(postID string) error {
	err := connectDatabase()
	if err != nil {
		fmt.Println(err)
		return err
	}
	// Hazırlık işlemi (Prepare statement)
	query, err := database.Prepare("UPDATE posts SET dislike_count = dislike_count + 1 WHERE id = ?")
	if err != nil {
		return err
	}
	defer query.Close()

	// Parametreleri geçirerek sorguyu çalıştırma
	_, err = query.Exec(postID)
	if err != nil {
		return err
	}

	return nil
}

func decreaseDislikeCount(postID string) error {
	err := connectDatabase()
	if err != nil {
		fmt.Println(err)
		return err
	}
	// Hazırlık işlemi (Prepare statement)
	query, err := database.Prepare("UPDATE posts SET dislike_count = dislike_count - 1 WHERE id = ?")
	if err != nil {
		return err
	}
	defer query.Close()

	// Parametreleri geçirerek sorguyu çalıştırma
	_, err = query.Exec(postID)
	if err != nil {
		return err
	}

	return nil
}

func increaseLikeCommentCount(commentID string) error {
	err := connectDatabase()
	if err != nil {
		fmt.Println(err)
		return err
	}
	// Hazırlık işlemi (Prepare statement)
	query, err := database.Prepare("UPDATE comments SET like_count = like_count + 1 WHERE comment_id = ?")
	if err != nil {
		return err
	}
	defer query.Close()

	// Parametreleri geçirerek sorguyu çalıştırma
	_, err = query.Exec(commentID)
	if err != nil {
		return err
	}

	return nil
}

func decreaseLikeCommentCount(commentID string) error {
	err := connectDatabase()
	if err != nil {
		fmt.Println(err)
		return err
	}
	// Hazırlık işlemi (Prepare statement)
	query, err := database.Prepare("UPDATE comments SET like_count = like_count - 1 WHERE comment_id = ?")
	if err != nil {
		return err
	}
	defer query.Close()

	// Parametreleri geçirerek sorguyu çalıştırma
	_, err = query.Exec(commentID)
	if err != nil {
		return err
	}

	return nil
}

func increaseDislikeCommentCount(commentID string) error {
	err := connectDatabase()
	if err != nil {
		fmt.Println(err)
		return err
	}
	// Hazırlık işlemi (Prepare statement)
	query, err := database.Prepare("UPDATE comments SET dislike_count = dislike_count + 1 WHERE comment_id = ?")
	if err != nil {
		return err
	}
	defer query.Close()

	// Parametreleri geçirerek sorguyu çalıştırma
	_, err = query.Exec(commentID)
	if err != nil {
		return err
	}

	return nil
}

func decreaseDislikeCommentCount(commentID string) error {
	err := connectDatabase()
	if err != nil {
		fmt.Println(err)
		return err
	}
	// Hazırlık işlemi (Prepare statement)
	query, err := database.Prepare("UPDATE comments SET dislike_count = dislike_count - 1 WHERE comment_id = ?")
	if err != nil {
		return err
	}
	defer query.Close()

	// Parametreleri geçirerek sorguyu çalıştırma
	_, err = query.Exec(commentID)
	if err != nil {
		return err
	}

	return nil
}

func saveImageToDB(data []byte, userID string, picture string) {
	connectDatabase()
	updateImageSQL := fmt.Sprintf("UPDATE users SET %s = ? WHERE id = ?", picture)
	_, err := database.Exec(updateImageSQL, data, userID)
	if err != nil {
		log.Fatal(err)
		return
	}
	fmt.Println("success")
}

func getUsersTopPosts(userID string) []PostData {
	connectDatabase()
	query := `select users.username,users.name,users.surname, posts.id,posts.title, posts.content ,posts.created_at, posts.like_count, posts.dislike_count,posts.image
		from posts
		join users on posts.user_id = users.id
		WHERE user_id = ? ORDER BY like_count DESC LIMIT 3`
	rows, err := database.Query(query, userID)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var posts []PostData
	for rows.Next() {
		var post PostData
		var image Image
		err = rows.Scan(&post.UserData.Username, &post.UserData.Name, &post.UserData.Surname, &post.PostData.PostID, &post.PostData.PostTitle, &post.PostData.PostContent, &post.PostData.PostCreatedAt, &post.PostData.PostLikeCount, &post.PostData.PostDislikeCount, &image.ImageData)
		post.PostData.PostImage = convertImg(image)
		checkError(err)
		posts = append(posts, post)
	}
	return posts
}

func getTopPosts() ([]PostData, error) {
	connectDatabase()
	rows, err := database.Query(`select users.username,users.name,users.surname, posts.id,posts.title, posts.content ,posts.created_at, posts.like_count, posts.dislike_count,posts.image
from posts join users on posts.user_id = users.id ORDER BY like_count DESC LIMIT 3`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []PostData
	for rows.Next() {
		var post PostData
		var image Image
		err = rows.Scan(&post.UserData.Username, &post.UserData.Name, &post.UserData.Surname, &post.PostData.PostID, &post.PostData.PostTitle, &post.PostData.PostContent, &post.PostData.PostCreatedAt, &post.PostData.PostLikeCount, &post.PostData.PostDislikeCount, &image.ImageData)
		post.PostData.PostImage = convertImg(image)
		checkError(err)
		posts = append(posts, post)
	}
	return posts, nil
}

func addCommentToDb(content string, postID, userID string) {
	connectDatabase()
	stmt, err := database.Prepare("insert into comments (post_id,user_id, content) values (?, ?, ?)")
	checkError(err)
	defer stmt.Close()

	_, err = stmt.Exec(postID, userID, content)
	checkError(err)
}
