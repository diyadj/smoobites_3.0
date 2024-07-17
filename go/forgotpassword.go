package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"fmt"
	"net/http"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

func generateToken() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func forgotPasswordHandler (db *sql.DB) http.HandlerFunc  {
	return func(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	email := r.FormValue("email")
	if email == "" {
		http.Error(w, "Email is required", http.StatusBadRequest)
        return
	}

	token, err := generateToken()
	if err != nil {
		http.Error(w, "Error generating token", http.StatusInternalServerError)
		return
	}

	expiresAt := time.Now().Add(1 * time.Hour)
	_, err = db.Exec("INSERT INTO password_resets (email, token, expires_at) VALUES (?, ?, ?)", email, token, expiresAt)
	if err != nil {
		http.Error(w, "Error saving token", http.StatusInternalServerError)
		return
	}

	resetLink := fmt.Sprintf("http://localhost:5500/reset-password?token=%s", token)
	fmt.Fprintf(w, "Password reset link: %s", resetLink)
	}
}

func resetPasswordHandler(db *sql.DB) http.HandlerFunc  {
	return func(w http.ResponseWriter, r *http.Request) {
	// if r.Method != http.MethodPost {
	// 	http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
	// 	return
	// }

	token := r.FormValue("token")
	newPassword := r.FormValue("password")
	if token == "" || newPassword == "" {
		http.Error(w, "Token and password are required", http.StatusBadRequest)
		return
	}

	var email string
	var expiresAt time.Time
	err := db.QueryRow("SELECT email, expires_at FROM password_resets WHERE token = ?", token).Scan(&email, &expiresAt)
	if err != nil {
		http.Error(w, "Invalid token", http.StatusBadRequest)
		return
	}

	if time.Now().After(expiresAt) {
		http.Error(w, "Token has expired", http.StatusBadRequest)
		return
	}

	hashedPassword, err := hashPassword(newPassword)
	if err != nil {
		http.Error(w, "Error hashing password", http.StatusInternalServerError)
		return
	}

	_, err = db.Exec("UPDATE users SET password = ? WHERE email = ?", hashedPassword, email)
	if err != nil {
		http.Error(w, "Error updating password", http.StatusInternalServerError)
		return
	}

	_, err = db.Exec("DELETE FROM password_resets WHERE token = ?", token)
	if err != nil {
		http.Error(w, "Error cleaning up token", http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "Password has been reset")
	}
}