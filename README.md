# Blogging Platform API (Go)

This project is a simple RESTful API for a personal blogging platform, built with Go. It provides basic CRUD (Create, Read, Update, Delete) operations for blog posts.

Project source URL: https://roadmap.sh/projects/blogging-platform-api

## Tech Stack

- **Language:** Go
- **Framework:** Net/Http (Standard Library)

## How to Run

1.  **Prerequisites:**

    - Go installed on your machine.
    - PostgreSQL installed and running.

2.  **Database Setup:**

    - Create a database named `blog_go`.
    - Create a table named `posts` with the following schema:

    ```sql
    CREATE TABLE posts (
        id SERIAL PRIMARY KEY,
        title VARCHAR(255) NOT NULL,
        content TEXT NOT NULL,
        category VARCHAR(100),
        tags TEXT[],
        created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
        updated_at TIMESTAMP WITH TIME ZONE DEFAULT now()
    );
    ```

3.  **Clone the repository:**

    ```bash
    git clone https://github.com/notsatria/blog_go.git
    cd blog_go
    ```

4.  **Run the server:**
    ```bash
    go run main.go
    ```
    The server will start on `http://localhost:8080`.

## API Endpoints

The API provides the following endpoints for managing blog posts.

### Create Blog Post

- **`POST /posts`**
- Creates a new blog post.

**Request Body:**

```json
{
  "title": "My First Blog Post",
  "content": "This is the content of my first blog post.",
  "category": "Technology",
  "tags": ["Tech", "Programming"]
}
```

**Success Response (201 Created):**

```json
{
  "id": 1,
  "title": "My First Blog Post",
  "content": "This is the content of my first blog post.",
  "category": "Technology",
  "tags": ["Tech", "Programming"],
  "createdAt": "2021-09-01T12:00:00Z",
  "updatedAt": "2021-09-01T12:00:00Z"
}
```

### Update Blog Post

- **`PUT /posts/:id`**
- Updates an existing blog post.

**Request Body:**

```json
{
  "title": "My Updated Blog Post",
  "content": "This is the updated content of my first blog post.",
  "category": "Technology",
  "tags": ["Tech", "Programming"]
}
```

**Success Response (200 OK):**

```json
{
  "id": 1,
  "title": "My Updated Blog Post",
  "content": "This is the updated content of my first blog post.",
  "category": "Technology",
  "tags": ["Tech", "Programming"],
  "createdAt": "2021-09-01T12:00:00Z",
  "updatedAt": "2021-09-01T12:30:00Z"
}
```

### Delete Blog Post

- **`DELETE /posts/:id`**
- Deletes an existing blog post. Returns a `204 No Content` on success.

### Get Blog Post

- **`GET /posts/:id`**
- Retrieves a single blog post.

**Success Response (200 OK):**

```json
{
  "id": 1,
  "title": "My First Blog Post",
  "content": "This is the content of my first blog post.",
  "category": "Technology",
  "tags": ["Tech", "Programming"],
  "createdAt": "2021-09-01T12:00:00Z",
  "updatedAt": "2021-09-01T12:00:00Z"
}
```

### Get All Blog Posts

- **`GET /posts`**
- Retrieves all blog posts. Can be filtered by a search term.

**Example Query:**

```
GET /posts?term=tech
```

**Success Response (200 OK):**

```json
[
  {
    "id": 1,
    "title": "My First Blog Post",
    "content": "This is the content of my first blog post.",
    "category": "Technology",
    "tags": ["Tech", "Programming"],
    "createdAt": "2021-09-01T12:00:00Z",
    "updatedAt": "2021-09-01T12:00:00Z"
  }
]
```
