package main

import (
    "fmt"
    "golang.org/x/crypto/bcrypt"
)

func hashPassword(password string) (string, error) {
    bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    return string(bytes), err
}

func main() {
    password := "khoon"
    hashedPassword, err := hashPassword(password)
    if err != nil {
        fmt.Println("Error hashing password:", err)
        return
    }
    fmt.Println("Hashed password:", hashedPassword)
}