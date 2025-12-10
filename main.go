package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"sync"
	"time"
)

var tmpl = template.Must(template.ParseFiles("index.html"))

type BlogPost struct {
	Id        int       `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	Category  string    `json:"category"`
	Tags      []string  `json:"tags"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func (p *BlogPost) New(nextId int) *BlogPost {
	return &BlogPost{
		Id:        nextId,
		Title:     p.Title,
		Content:   p.Content,
		Tags:      p.Tags,
		Category:  p.Category,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func (p *BlogPost) Update(input BlogPost) {
	p.Title = input.Title
	p.Content = input.Content
	p.Category = input.Category
	p.Tags = input.Tags
	p.UpdatedAt = time.Now()
}

var dataStore = struct {
	sync.Mutex
	posts  []BlogPost
	nextId int
}{
	posts:  make([]BlogPost, 0),
	nextId: 1,
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		tmpl.Execute(w, nil)
	})

	mux.HandleFunc("POST /posts", createPost)
	mux.HandleFunc("PUT /posts/{id}", updatePost)

	fmt.Println("Server running on :8080")
	http.ListenAndServe(":8080", mux)
}

func createPost(w http.ResponseWriter, r *http.Request) {
	var input BlogPost

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid Json format", http.StatusBadRequest)
		return
	}

	if input.Title == "" || input.Content == "" {
		http.Error(w, "Title and content cannot be empty!", http.StatusBadRequest)
		return
	}

	dataStore.Lock()
	newPost := input.New(dataStore.nextId)

	dataStore.posts = append(dataStore.posts, *newPost)
	dataStore.nextId++

	dataStore.Unlock()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(newPost); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
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

	dataStore.Lock()
	defer dataStore.Unlock()

	// Dapetin data post dengan id == postId
	var foundIndex = -1
	for i, post := range dataStore.posts {
		if post.Id == postId {
			foundIndex = i
			break
		}
	}

	if foundIndex == -1 {
		http.Error(w, "Post not found", http.StatusNotFound)
		return
	}

	targetPost := &dataStore.posts[foundIndex]
	targetPost.Update(input)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(targetPost); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}
