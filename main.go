package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
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

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid Json format", http.StatusBadRequest)
		return
	}

	if input.Title == "" || input.Content == "" {
		http.Error(w, "Title and content cannot be empty!", http.StatusBadRequest)
		return
	}

	dataStore.Lock()
	newPost := BlogPost{
		Id:        dataStore.nextId,
		Title:     input.Title,
		Content:   input.Content,
		Category:  input.Category,
		Tags:      input.Tags,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	dataStore.posts = append(dataStore.posts, newPost)
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

func deletePost(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	postId, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Id should be a number", http.StatusBadRequest)
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

	dataStore.posts = append(dataStore.posts[:foundIndex], dataStore.posts[foundIndex+1:]...)

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
