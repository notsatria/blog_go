package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"sync"
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

var dataStore = struct {
	sync.RWMutex
	posts  []BlogPost
	nextId int
}{
	posts:  make([]BlogPost, 0),
	nextId: 1,
}

func main() {
	var err error

	connStr := "user=postgres password=postgresdb dbname=blog_go sslmode=disable"

	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}

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

	dataStore.RLock()
	defer dataStore.RUnlock()

	// Dapetin data post dengan id == postId
	var foundPost BlogPost
	var found = false
	for _, post := range dataStore.posts {
		if post.Id == postId {
			foundPost = post
			found = true
			break
		}
	}

	if !found {
		http.Error(w, "Post not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(foundPost); err != nil {
		log.Println(err)
	}
}

func getAllPost(w http.ResponseWriter, r *http.Request) {
	dataStore.RLock()
	defer dataStore.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(dataStore.posts); err != nil {
		log.Println(err)
	}
}

func getPostAndReturnValue(id int) (BlogPost, error) {
	query := `
		SELECT id, title, content, category, tags, created_at, updated_at 
		FROM posts WHERE id = $1`

	var post BlogPost
	err := db.QueryRow(query, id).Scan(
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
			log.Printf("No post with id %d found", id)
		} else {
			log.Println("Error scanning post:", err)
		}
		return BlogPost{}, err
	}

	return post, nil
}
