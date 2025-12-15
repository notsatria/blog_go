package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/lib/pq"
	_ "github.com/lib/pq"
)

var tmpl = template.Must(template.ParseFiles("index.html"))

var db *sql.DB

type BlogPost struct {
	Id        int       `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	Category  string    `json:"category"`
	Tags      []string  `json:"tags"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func (p *BlogPost) Update(input BlogPost) {
	p.Title = input.Title
	p.Content = input.Content
	p.Category = input.Category
	p.Tags = input.Tags
	p.UpdatedAt = time.Now()
}

func main() {
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASS")
	dbHost := os.Getenv("DB_HOST")
	dbName := os.Getenv("DB_NAME")

	if dbUser == "" {
		dbUser = "postgres"
	}
	if dbPass == "" {
		dbPass = "postgresdb"
	}
	if dbHost == "" {
		dbHost = "localhost"
	}
	if dbName == "" {
		dbName = "blog_go"
	}

	var err error
	connStr := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable", dbUser, dbPass, dbHost, dbName)
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)

	// Ping to DB
	if err = db.Ping(); err != nil {
		log.Fatal("Failed to connect to DB:", err)
	}

	fmt.Println("Connected to DB!")

	mux := http.NewServeMux()

	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		tmpl.Execute(w, nil)
	})

	mux.HandleFunc("POST /posts", createPost)
	mux.HandleFunc("PUT /posts/{id}", updatePost)
	mux.HandleFunc("DELETE /posts/{id}", deletePost)
	mux.HandleFunc("GET /posts/{id}", getPost)
	mux.HandleFunc("GET /posts", getAllPost)

	fmt.Println("Server running on :8080")
	http.ListenAndServe(":8080", mux)
}

func createPost(w http.ResponseWriter, r *http.Request) {
	var input BlogPost

	var err error
	if err = json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid Json format", http.StatusBadRequest)
		return
	}

	if input.Title == "" || input.Content == "" {
		http.Error(w, "Title and content cannot be empty!", http.StatusBadRequest)
		return
	}

	query := `INSERT INTO posts (title, content, category, tags)
	VALUES($1, $2, $3, $4)
	RETURNING id, created_at, updated_at`

	newPost := input
	err = db.QueryRow(
		query,
		input.Title,
		input.Content,
		input.Category,
		pq.Array(input.Tags),
	).Scan(&newPost.Id, &newPost.CreatedAt, &newPost.UpdatedAt)

	if err != nil {
		log.Println(err)
		http.Error(w, "Error on creating a new post", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(newPost); err != nil {
		log.Println("Failed to encode post:", err)
	}
}

func updatePost(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	postId, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Id should be a number", http.StatusBadRequest)
		return
	}

	var input BlogPost

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid Json format", http.StatusBadRequest)
		return
	}

	updatedPost := input
	query := `UPDATE posts 
	SET title = $1, content = $2, category = $3, tags = $4, updated_at = $5
	WHERE id = $6
	RETURNING id, created_at, updated_at`

	err = db.QueryRow(query, input.Title, input.Content, input.Category, pq.Array(input.Tags), time.Now(), postId).Scan(
		&updatedPost.Id, &updatedPost.CreatedAt, &updatedPost.UpdatedAt,
	)

	if err != nil {
		log.Println("Error on update post:", err)
		if err == sql.ErrNoRows {
			// JIKA ID TIDAK DITEMUKAN
			http.Error(w, "Post not found", http.StatusNotFound) // 404
			return
		}

		http.Error(w, "Error on updating a post", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(updatedPost); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}

func deletePost(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	postId, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Id should be a number", http.StatusBadRequest)
		return
	}

	query := `DELETE FROM posts WHERE id = $1`

	res, err := db.Exec(query, postId)
	if err != nil {
		log.Println("Error on delete post:", err)
		http.Error(w, "Error on deleting a post", http.StatusInternalServerError)
		return
	}

	count, err := res.RowsAffected()
	if err != nil {
		http.Error(w, "Error checking rows affected", http.StatusInternalServerError)
		return
	}

	if count == 0 {
		http.Error(w, "Post not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func getPost(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	postId, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Id should be a number", http.StatusBadRequest)
		return
	}

	query := `
		SELECT id, title, content, category, tags, created_at, updated_at 
		FROM posts WHERE id = $1`

	var post BlogPost
	err = db.QueryRow(query, postId).Scan(
		&post.Id,
		&post.Title,
		&post.Content,
		&post.Category,
		pq.Array(&post.Tags),
		&post.CreatedAt,
		&post.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			// Beri tahu client bahwa data tidak ada (404)
			http.Error(w, "Post not found", http.StatusNotFound)
		} else {
			// Beri tahu client ada error server (500)
			log.Println("Error scanning post:", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(post); err != nil {
		log.Println(err)
	}
}

func getAllPost(w http.ResponseWriter, r *http.Request) {
	term := r.URL.Query().Get("term")

	var query string
	var args []interface{}

	if term != "" {
		query = `
		SELECT id, title, content, category, tags, created_at, updated_at 
		FROM posts
		WHERE title ILIKE $1
		OR content ILIKE $1 
		OR category ILIKE $1`

		searchPattern := "%" + term + "%"

		args = append(args, searchPattern)
	} else {
		query = `SELECT id, title, content, category, tags, created_at, updated_at 
		FROM posts`
	}

	var err error
	rows, err := db.Query(query, args...)

	if err != nil {
		log.Println("Error executing query:", err)
		// PERBAIKAN: Beri tahu client ada error 500
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	defer rows.Close()

	var posts []BlogPost = make([]BlogPost, 0)
	for rows.Next() {
		var post BlogPost
		if err = rows.Scan(&post.Id, &post.Title, &post.Content, &post.Category, pq.Array(&post.Tags), &post.CreatedAt, &post.UpdatedAt); err != nil {
			log.Println("Error on get posts: ", err)
			break
		}
		posts = append(posts, post)
	}

	// Cek error setelah loop selesai
	if err = rows.Err(); err != nil {
		log.Println("Error iterating rows:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err = json.NewEncoder(w).Encode(posts); err != nil {
		log.Println(err)
	}
}
