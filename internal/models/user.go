package models

import (
	"time"
	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID ` json:"id"`
	UserName     string    `json:"name"`
	Email        string    `json:"email"`
	AvatarUrl    string    `json:"avatar_url"`
	Role         string    `json:"role"`
	Password     string    `json:"-"`
	TokenVersion int32     `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type RegistrationRequest struct {
	UserName string `json:"user_name" binding:"required,min=6,max=8"`
	Email    string `json:"email" binding:"required,email,max=254"`
	Password string `json:"password" binding:"required,min=8,max=12"`
}
type LogInRequest struct {
	Email    string `json:"email" binding:"required,email,max=254"`
	Password string `json:"password" binding:"required,min=8,max=12"`
}

type UserResponse struct {
	UserName  string `json:"user_name"`
	Role      string `json:"role"`
	Email     string `json:"email"`
	AvatarUrl string `json:"avatar_url"`
}
