package models

import (
	"encoding/json"
	"time"
)

type SurveyQuestion struct {
	ID           int64           `db:"id" json:"id"`
	Category     string          `db:"category" json:"category"`
	QuestionText string          `db:"question_text" json:"question_text"`
	Options      json.RawMessage `db:"options" json:"options"`
	SortOrder    int             `db:"sort_order" json:"sort_order"`
}

type SurveyResponse struct {
	ID         int64     `db:"id" json:"id"`
	UserID     int64     `db:"user_id" json:"user_id"`
	QuestionID int64     `db:"question_id" json:"question_id"`
	Answer     json.RawMessage `db:"answer" json:"answer"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
}

type SubmitSurveyRequest struct {
	Answers []SurveyAnswer `json:"answers"`
}

type SurveyAnswer struct {
	QuestionID int64           `json:"question_id"`
	Answer     json.RawMessage `json:"answer"`
}

type AIResult struct {
	ID                 int64            `db:"id" json:"id"`
	UserID             int64            `db:"user_id" json:"user_id"`
	RawResponse        json.RawMessage  `db:"raw_response" json:"-"`
	RecommendedCareers json.RawMessage  `db:"recommended_careers" json:"recommended_careers"`
	Status             string           `db:"status" json:"status"`
	ErrorMessage       *string          `db:"error_message" json:"error_message,omitempty"`
	CreatedAt          time.Time        `db:"created_at" json:"created_at"`
	CompletedAt        *time.Time       `db:"completed_at" json:"completed_at,omitempty"`
}

type ScoredCareer struct {
	CareerID  int64  `json:"career_id"`
	Title     string `json:"title"`
	Score     int    `json:"score"`
	Reasoning string `json:"reasoning"`
}

type CareerScore struct {
	ID        int64     `db:"id" json:"id"`
	UserID    int64     `db:"user_id" json:"user_id"`
	CareerID  int64     `db:"career_id" json:"career_id"`
	Score     int       `db:"score" json:"score"`
	Reasoning string    `db:"reasoning" json:"reasoning"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

type AIMentorRequest struct {
	CareerID int64  `json:"career_id"`
	Message  string `json:"message"`
}

type AIMentorChunk struct {
	Token string `json:"token"`
}

type ScoredCareerList struct {
	RecommendedCareers []ScoredCareer `json:"recommended_careers"`
}
