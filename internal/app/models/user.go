package models

import "time"

type User struct {
    ID           uint      `json:"id"`
    FullName     string    `json:"full_name"`
    Email        string    `json:"email"`
    PasswordHash string    `json:"-"`
    CreatedAt    time.Time `json:"created_at"`
    UpdatedAt    time.Time `json:"updated_at"`
}

type SignupRequest struct {
    FullName        string `json:"full_name" binding:"required"`
    Email          string `json:"email" binding:"required,email"`
    Password       string `json:"password" binding:"required,min=8"`
    ConfirmPassword string `json:"confirm_password" binding:"required,eqfield=Password"`
}

type LoginRequest struct {
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required"`
}

type AuthResponse struct {
    Token string `json:"token"`
    User  *User  `json:"user"`
}