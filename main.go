package main

import (
    "database/sql"
    "fmt"
    "log"
    "math/rand"
    "net/http"
    "os"
    "time"

    "github.com/gorilla/mux"
    "github.com/joho/godotenv"
    _ "github.com/lib/pq"
)

var db *sql.DB

func main() {
    // Load environment variables
    err := godotenv.Load()
    if err != nil {
        log.Println("⚠️  No .env file found, using system environment variables.")
    }

    connStr := os.Getenv("DATABASE_URL")
    if connStr == "" {
        log.Fatal("❌ DATABASE_URL environment variable not set")
    }

    db, err = sql.Open("postgres", connStr)
    if err != nil {
        log.Fatalf("Error opening database: %v", err)
    }

    if err = db.Ping(); err != nil {
        log.Fatalf("Could not connect to database: %v", err)
    }

    fmt.Println("✅ Connected to Neon PostgreSQL!")

    rand.Seed(time.Now().UnixNano()) // seed random once

    // Setup router
    r := mux.NewRouter()
    r.HandleFunc("/article/{id}", getArticleHandler).Methods("GET")
    r.HandleFunc("/articles", getAllArticlesHandler).Methods("GET")
    r.HandleFunc("/articles", addNewArticle).Methods("POST")
    r.HandleFunc("/article/{id}", updateArticleHandler).Methods("PUT")

    r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        fmt.Fprintln(w, "Welcome to the intro server!")
    })

    log.Println("🚀 Server running on port 8080")
    log.Fatal(http.ListenAndServe(":8080", r))
}

func addNewArticle(w http.ResponseWriter, r *http.Request) {
    title := r.URL.Query().Get("title") // for getting query params from URL
    content := r.URL.Query().Get("content")

    if title == "" || content == "" {
        http.Error(w, "Title and content are required", http.StatusBadRequest)
        return
    }

   

    _, err := db.Exec("INSERT INTO articles ( title, content) VALUES ( $1, $2)", title, content)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    fmt.Fprintln(w, "✅ Article added successfully")
}

func getArticleHandler(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    articleId := vars["id"] // get id from URL

    var title, content string
    err := db.QueryRow("SELECT title, content FROM articles WHERE id = $1", articleId).Scan(&title, &content)
    if err != nil {
        if err == sql.ErrNoRows {
            http.NotFound(w, r)
            return
        }
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    fmt.Fprintf(w, "📝 Article: %s\n📄 Content: %s\n", title, content)
}

func getAllArticlesHandler(w http.ResponseWriter, r *http.Request) {
    rows, err := db.Query("SELECT id, title, content FROM articles")
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    fmt.Fprintln(w, "📚 All Articles:")

    found := false
    for rows.Next() {
        found = true
        var id int
        var title, content string
        if err := rows.Scan(&id, &title, &content); err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        fmt.Fprintf(w, "- %d: %s — %s\n", id, title, content)
    }

    if !found {
        fmt.Fprintln(w, "No articles found.")
    }

    if err := rows.Err(); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}


func updateArticleHandler(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    articleId := vars["id"]
    title := r.URL.Query().Get("title")
    content := r.URL.Query().Get("content")

    if title == "" || content == "" {
        http.Error(w, "Title and content are required", http.StatusBadRequest)
        return
    }

    result, err := db.Exec("UPDATE articles SET title = $1, content = $2 WHERE id = $3", title, content, articleId)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    rowsAffected, err := result.RowsAffected()
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    if rowsAffected == 0 {
        http.NotFound(w, r)
        return
    }

    fmt.Fprintf(w, "✅ Article with ID %s updated\n", articleId)
}
