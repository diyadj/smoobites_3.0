package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"gopkg.in/gomail.v2"
)

func generateToken() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func sendResetEmail(to, resetLink string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", "smoobites@gmail.com")
	m.SetHeader("To", to)
	m.SetHeader("Subject", "Password Reset Request")
	m.SetBody("text/plain", fmt.Sprintf("To reset your password, click the following link: %s. This link expires in an hour. If you did not request to reset password, please ignore this email.", resetLink))

	d := gomail.NewDialer("smtp.gmail.com", 587, "smoobites@gmail.com", "yummyfoodinSMU")

	// Send the email
	if err := d.DialAndSend(m); err != nil {
		log.Printf("Failed to send email: %v", err)
		return err
	}
	return nil
}

func forgotPasswordHandler (db *sql.DB) http.HandlerFunc  {
	return func(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	email := r.FormValue("email")

	token, err := generateToken()
	fmt.Println(token)
	if err != nil {
		log.Printf("Failed to generate token: %v", err)
		http.Error(w, "Unable to generate token", http.StatusBadRequest)
	}

	expiresAt := time.Now().Add(1 * time.Hour).Format("2006-01-02 15:04:05")
	_, err = db.Exec("INSERT INTO password_resets (email, token, expires_at) VALUES (?, ?, ?)", email, token, expiresAt)
	if err != nil {
		http.Error(w, "Error saving token", http.StatusInternalServerError)
	}

	resetLink := fmt.Sprintf("http://localhost:8080/resetpassword.html?token=%s", token)
		err = sendResetEmail(email, resetLink)
		if err != nil {
			http.Error(w, "Failed to send email", http.StatusInternalServerError)
			return
		}

	fmt.Fprintln(w, "Password reset link has been sent to your email.")
	}
}

func resetPasswordHandler(db *sql.DB) http.HandlerFunc  {
	return func(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	token := r.FormValue("token")
	//fmt.Println(token)
	newPassword := r.FormValue("password")

	var email string
	var expiresAtStr string
	err := db.QueryRow("SELECT email, expires_at FROM password_resets WHERE token = ?", token).Scan(&email, &expiresAtStr)
	if err != nil {
		log.Printf("Error querying token: %v", err)
		//http.Error(w, "Error querying token:", http.StatusBadRequest)
		return
	}

	layout := "2006-01-02 15:04:05"
	expiresAt, err := time.Parse(layout, expiresAtStr)
	if err != nil {
		http.Error(w, "Error parsing time", http.StatusInternalServerError)
		log.Printf("Error parsing time: %v", err)
		return
	}


	if time.Now().After(expiresAt) {
		http.Error(w, "Expired token:", http.StatusBadRequest)
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

	http.Redirect(w, r, "/login.html", http.StatusSeeOther)
	}
}