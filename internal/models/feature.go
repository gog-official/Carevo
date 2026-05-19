package models

import "time"

type Bookmark struct {
	UserID    int64     `db:"user_id" json:"user_id"`
	CareerID  int64     `db:"career_id" json:"career_id"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	Title     string    `db:"title" json:"title,omitempty"`
	Slug      string    `db:"slug" json:"slug,omitempty"`
	Summary   string    `db:"summary" json:"summary,omitempty"`
}

type BookmarkRequest struct {
	CareerID int64 `json:"career_id"`
}

type Challenge struct {
	ID            int64     `db:"id" json:"id"`
	UserID        int64     `db:"user_id" json:"user_id"`
	CareerID      int64     `db:"career_id" json:"career_id"`
	StartedAt     time.Time `db:"started_at" json:"started_at"`
	CurrentStreak int       `db:"current_streak" json:"current_streak"`
	LongestStreak int       `db:"longest_streak" json:"longest_streak"`
	IsActive      bool      `db:"is_active" json:"is_active"`
	CareerTitle   string    `db:"career_title" json:"career_title,omitempty"`
	CareerSlug    string    `db:"career_slug" json:"career_slug,omitempty"`
}

type CreateChallengeRequest struct {
	CareerID int64 `json:"career_id"`
}

type ChallengeCheckin struct {
	ID          int64     `db:"id" json:"id"`
	ChallengeID int64     `db:"challenge_id" json:"challenge_id"`
	CheckinDate time.Time `db:"checkin_date" json:"checkin_date"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
}

type ChallengeProgress struct {
	ChallengeID       int64 `json:"challenge_id"`
	DaysCompleted     int   `json:"days_completed"`
	CurrentStreak     int   `json:"current_streak"`
	LongestStreak     int   `json:"longest_streak"`
	PercentComplete   int   `json:"percent_complete"`
	DaysRemaining     int   `json:"days_remaining"`
}

type ProjectIdea struct {
	ID          int64     `db:"id" json:"id"`
	CareerID    int64     `db:"career_id" json:"career_id"`
	UserID      int64     `db:"user_id" json:"user_id"`
	Title       string    `db:"title" json:"title"`
	Description string    `db:"description" json:"description"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
}

type CreateProjectIdeaRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}
