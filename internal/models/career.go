package models

import (
	"encoding/json"
	"time"
)

type Category struct {
	ID          int64     `db:"id" json:"id"`
	Name        string    `db:"name" json:"name"`
	Slug        string    `db:"slug" json:"slug"`
	Description string    `db:"description" json:"description"`
	Icon        string    `db:"icon" json:"icon"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
}

type SalaryRange struct {
	Min      int    `json:"min"`
	Max      int    `json:"max"`
	Currency string `json:"currency"`
	Period   string `json:"period"`
}

type Career struct {
	ID               int64           `db:"id" json:"id"`
	Title            string          `db:"title" json:"title"`
	Slug             string          `db:"slug" json:"slug"`
	Summary          string          `db:"summary" json:"summary"`
	Description      string          `db:"description" json:"description"`
	CategoryID       int64           `db:"category_id" json:"category_id"`
	DailyTasks       json.RawMessage `db:"daily_tasks" json:"daily_tasks"`
	Skills           []string        `db:"skills" json:"skills"`
	SalaryMin        int             `db:"salary_min" json:"-"`
	SalaryMax        int             `db:"salary_max" json:"-"`
	SalaryCurrency   string          `db:"salary_currency" json:"-"`
	SalaryPeriod     string          `db:"salary_period" json:"-"`
	Difficulty       int             `db:"difficulty" json:"difficulty"`
	FutureProofScore int             `db:"future_proof_score" json:"future_proof_score"`
	EducationReq     string          `db:"education_required" json:"education_required"`
	Outlook          string          `db:"outlook" json:"outlook"`
	CreatedAt        time.Time       `db:"created_at" json:"created_at"`
	UpdatedAt        time.Time       `db:"updated_at" json:"updated_at"`
	Rank             float64         `db:"rank" json:"-"`

	CategoryName string `db:"category_name" json:"category_name,omitempty"`
	CategorySlug string `db:"category_slug" json:"category_slug,omitempty"`
}

func (c *Career) SalaryRange() SalaryRange {
	return SalaryRange{
		Min:      c.SalaryMin,
		Max:      c.SalaryMax,
		Currency: c.SalaryCurrency,
		Period:   c.SalaryPeriod,
	}
}

type CareerTag struct {
	ID       int64  `db:"id" json:"id"`
	CareerID int64  `db:"career_id" json:"career_id"`
	Tag      string `db:"tag" json:"tag"`
}

type Resource struct {
	ID          int64     `db:"id" json:"id"`
	CareerID    int64     `db:"career_id" json:"career_id"`
	Title       string    `db:"title" json:"title"`
	URL         string    `db:"url" json:"url"`
	Description string    `db:"description" json:"description"`
	IsFree      bool      `db:"is_free" json:"is_free"`
	SortOrder   int       `db:"sort_order" json:"sort_order"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
}

type RoadmapStep struct {
	ID          int64           `db:"id" json:"id"`
	CareerID    int64           `db:"career_id" json:"career_id"`
	StepNumber  int             `db:"step_number" json:"step_number"`
	Title       string          `db:"title" json:"title"`
	Description string          `db:"description" json:"description"`
	Duration    string          `db:"duration" json:"duration"`
	SortOrder   int             `db:"sort_order" json:"sort_order"`
	Links       json.RawMessage `db:"links" json:"links"`
	CreatedAt   time.Time       `db:"created_at" json:"created_at"`
}

type CategoryWithCount struct {
	Category
	CareerCount int `db:"career_count" json:"career_count"`
}
