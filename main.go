package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
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
}

func (p *BlogPost) New(nextId int) *BlogPost {
	post := &BlogPost{
		Id:        nextId,
		Title:     p.Title,
		Content:   p.Content,
		Tags:      p.Tags,
		Category:  p.Category,
		CreatedAt: time.Now(),
	}
	return post
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
