package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
)

func setupDB() *sql.DB {
	db, err := sql.Open("mysql", "root:@tcp(localhost)/smoobites")
	if err != nil {
		log.Fatal(err)
	}
	if err = db.Ping(); err != nil {
		log.Fatal(err)
	}
	return db
}

func hashPassword(password string) (string, error) {
    bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    return string(bytes), err
}

func checkPasswordHash(password, hash string) bool {
    err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
    return err == nil
}

func registerHandler(db *sql.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        if r.Method != "POST" {
            http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
            return
        }

        if err := r.ParseForm(); err != nil {
            http.Error(w, "Error parsing the form", http.StatusBadRequest)
            return
        }

        email := r.FormValue("email")
        password := r.FormValue("password")
        if email == "" || password == "" {
            http.Error(w, "Email and password are required", http.StatusBadRequest)
            return
        }

        hashedPassword, err := hashPassword(password)
        if err != nil {
            log.Printf("Failed to hash password: %v", err)
            http.Error(w, "Error processing the request", http.StatusInternalServerError)
            return
        }

        _, err = db.Exec("INSERT INTO users (email, password) VALUES (?, ?)", email, hashedPassword)
        if err != nil {
            log.Printf("Failed to insert user: %v", err)
            http.Error(w, "Error saving the user", http.StatusInternalServerError)
            return
        }

        w.Header().Set("Content-Type", "text/html; charset=utf-8")
        fmt.Fprintf(w, "<h1>Registration successful!</h1>")
    }
}

func loginHandler(db *sql.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        if r.Method != "POST" {
            http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
            return
        }

        if err := r.ParseForm(); err != nil {
            http.Error(w, "Error parsing the form", http.StatusBadRequest)
            return
        }

        email := r.FormValue("email")
        password := r.FormValue("password")
        if email == "" || password == "" {
            http.Error(w, "Email and password are required", http.StatusBadRequest)
            return
        }

        var storedPassword string
        err := db.QueryRow("SELECT password FROM users WHERE email = ?", email).Scan(&storedPassword)
        if err != nil {
            if err == sql.ErrNoRows {
                http.Error(w, "Invalid email or password", http.StatusUnauthorized)
            } else {
                log.Printf("Failed to query user: %v", err)
                http.Error(w, "Error querying user", http.StatusInternalServerError)
            }
            return
        }

        if !checkPasswordHash(password, storedPassword) {
            http.Error(w, "Invalid email or password", http.StatusUnauthorized)
            return
        }

        w.Header().Set("Content-Type", "text/html; charset=utf-8")
        fmt.Fprintf(w, "<h1>Welcome, %s!</h1>", email)
    }
}

func main() {
	db := setupDB()
	defer db.Close()

	// Serve index.html for the root URL
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
            http.ServeFile(w, r, filepath.Join("..", "index.html"))
        } else {
            filePath := filepath.Join("..", r.URL.Path[1:]) // Remove the leading slash
            if _, err := os.Stat(filePath); os.IsNotExist(err) {
                http.NotFound(w, r)
                return
            }
            http.ServeFile(w, r, filePath)
        }
	})

	// API endpoints
	http.HandleFunc("/register", registerHandler(db))
	http.HandleFunc("/login", loginHandler(db))

	log.Println("Server started on http://localhost:5500")
	if err := http.ListenAndServe(":5500", nil); err != nil {
		log.Fatal(err)
	}
}
