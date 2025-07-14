package models

import (
	"time"
	"gorm.io/gorm"
)

type User struct {
    ID           uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
    FullName     string    `json:"full_name" gorm:"size:255;not null"`
    Email        string    `json:"email" gorm:"size:255;uniqueIndex;not null"`
    PasswordHash string    `json:"-" gorm:"size:255;not null"`
    CreatedAt    time.Time `json:"created_at"`
    UpdatedAt    time.Time `json:"updated_at"`
    DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
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