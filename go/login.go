package main

import (
    "database/sql"
    "fmt"
    "log"
    "net/http"
    "os"
    "path/filepath"

    "github.com/gorilla/sessions"
    _ "github.com/go-sql-driver/mysql"
    "golang.org/x/crypto/bcrypt"
)

var store = sessions.NewCookieStore([]byte("something-very-secret"))

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

func authHandler(db *sql.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        if r.Method != "POST" {
            http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
            return
        }

        if err := r.ParseForm(); err != nil {
            http.Error(w, "Error parsing the form", http.StatusBadRequest)
            return
        }

        action := r.FormValue("action")
        email := r.FormValue("email")
        password := r.FormValue("password")

        if email == "" || password == "" {
            http.Redirect(w, r, "/login.html?error=empty", http.StatusSeeOther)
            return
        }

        if action == "register" {
            // Register user
            name := r.FormValue("name")
            if name == "" {
                http.Error(w, "Name is required for registration", http.StatusBadRequest)
                return
            }

            hashedPassword, err := hashPassword(password)
            if err != nil {
                log.Printf("Failed to hash password: %v", err)
                http.Error(w, "Error processing the request", http.StatusInternalServerError)
                return
            }

            role := "user" // Default role, can be changed based on requirements

            _, err = db.Exec("INSERT INTO users (name, email, password, role) VALUES (?, ?, ?, ?)", name, email, hashedPassword, role)
            if err != nil {
                log.Printf("Failed to insert user: %v", err)
                http.Error(w, "Error saving the user", http.StatusInternalServerError)
                return
            }

            w.Header().Set("Content-Type", "text/html; charset=utf-8")
            fmt.Fprintf(w, "<h1>Registration successful!</h1>")
        } else if action == "login" {
            // Login user
            var storedPassword, role, name string
            var vendorId int
            err := db.QueryRow("SELECT password, role, name, id FROM users WHERE email = ?", email).Scan(&storedPassword, &role, &name, &vendorId)
            if err != nil {
                if err == sql.ErrNoRows {
                    http.Redirect(w, r, "/login.html?error=invalid", http.StatusSeeOther)
                } else {
                    log.Printf("Failed to query user: %v", err)
                    http.Error(w, "Error querying user", http.StatusInternalServerError)
                }
                return
            }

            if !checkPasswordHash(password, storedPassword) {
                http.Redirect(w, r, "/login.html?error=invalid", http.StatusSeeOther)
                return
            }

            // Create session
            session, _ := store.Get(r, "session-name")
            session.Values["user"] = name
            session.Values["role"] = role
            session.Values["vendorId"] = vendorId
            session.Save(r, w)

            if role == "vendor" {
                http.Redirect(w, r, "/vmainpage.html", http.StatusSeeOther)
            } else if role == "admin" {
                http.Redirect(w, r, "/adminpage.html", http.StatusSeeOther)
            }else {
                http.Redirect(w, r, "/user_landing.html", http.StatusSeeOther)
            }

        } else {
            http.Error(w, "Invalid action", http.StatusBadRequest)
        }
    }
}


func sessionInfoHandler() http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        session, _ := store.Get(r, "session-name")
        user := session.Values["user"]
        role := session.Values["role"]
        vendorId := session.Values["vendorId"]

        w.Header().Set("Content-Type", "application/json")
        if user != nil && role != nil && vendorId != nil {
            fmt.Fprintf(w, `{"user":"%s", "role":"%s", "vendorId":%d}`, user, role, vendorId)
        } else {
            fmt.Fprintf(w, `{"user":null, "role":null, "vendorId":null}`)
        }
    }
}

func logoutHandler() http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        session, _ := store.Get(r, "session-name")
        session.Options.MaxAge = -1 // This will clear the session
        session.Save(r, w)
        w.Header().Set("Content-Type", "application/json")
        fmt.Fprintf(w, `{"status":"logged out"}`)
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
    http.HandleFunc("/auth", authHandler(db))
    http.HandleFunc("/session-info", sessionInfoHandler())
    http.HandleFunc("/logout", logoutHandler())
    http.HandleFunc("/add-food-item", addFoodItemHandler(db))
    http.HandleFunc("/get-food-items", getFoodItemsHandler(db)) 
    http.HandleFunc("/get-food-details", getFoodDetailsHandler(db)) 
    http.HandleFunc("/update-food-item", updateFoodItemHandler(db))
    http.HandleFunc("/get-food-item", getFoodItemByIDHandler(db))
    http.HandleFunc("/set-food-session", setFoodSessionHandler())
    http.HandleFunc("/delete-food-item", deleteFoodItemHandler(db))


    log.Println("Server started on http://localhost:5500")
    if err := http.ListenAndServe(":5500", nil); err != nil {
        log.Fatal(err)
    }
}
