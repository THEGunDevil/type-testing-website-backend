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
type UserResponse struct {
	UserName  string `json:"user_name"`
	Role      string `json:"role"`
	Email     string `json:"email"`
	AvatarUrl string `json:"avatar_url"`
}
