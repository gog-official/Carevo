package models

import "time"

type User struct {
	ID             int64      `db:"id" json:"id"`
	Email          string     `db:"email" json:"email"`
	PasswordHash   string     `db:"password_hash" json:"-"`
	Name           string     `db:"name" json:"name"`
	Bio            string     `db:"bio" json:"bio"`
	AvatarURL      string     `db:"avatar_url" json:"avatar_url"`
	AuthProvider   string     `db:"auth_provider" json:"auth_provider"`
	AuthProviderID string     `db:"auth_provider_id" json:"-"`
	IsAdmin        bool       `db:"is_admin" json:"is_admin"`
	CreatedAt      time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt      time.Time  `db:"updated_at" json:"updated_at"`
}

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UpdateProfileRequest struct {
	Name      string `json:"name"`
	Bio       string `json:"bio"`
	AvatarURL string `json:"avatar_url"`
}

type OAuthLoginRequest struct {
	Provider   string `json:"provider"`
	ProviderID string `json:"provider_id"`
	Email      string `json:"email"`
	Name       string `json:"name"`
	AvatarURL  string `json:"avatar_url"`
}

type CareerSuggestion struct {
	ID          int64     `db:"id" json:"id"`
	UserID      int64     `db:"user_id" json:"user_id"`
	Title       string    `db:"title" json:"title"`
	Description string    `db:"description" json:"description"`
	Reason      string    `db:"reason" json:"reason"`
	Status      string    `db:"status" json:"status"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
	UserName    string    `db:"user_name" json:"user_name,omitempty"`
	UserEmail   string    `db:"user_email" json:"user_email,omitempty"`
}

type CreateSuggestionRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Reason      string `json:"reason"`
}
