package models

import (
	"encoding/json"
	"time"

	"github.com/lib/pq"
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

type SalaryTiers struct {
	Entry      SalaryRange `json:"entry"`
	Average    SalaryRange `json:"average"`
	Experienced SalaryRange `json:"experienced"`
	Freelance  SalaryRange `json:"freelance,omitempty"`
}

type CitySalary struct {
	Entry      int `json:"entry"`
	Average    int `json:"average"`
	Experienced int `json:"experienced"`
}

type DemandData struct {
	Trend             string   `json:"trend"`
	GrowthForecast    string   `json:"growth_forecast"`
	Opportunities     string   `json:"opportunities"`
	TopHiringCompanies []string `json:"top_hiring_companies,omitempty"`
	FreelanceDemand   string   `json:"freelance_demand,omitempty"`
	NepalDemand       string   `json:"nepal_demand,omitempty"`
	GlobalOpportunity bool     `json:"global_opportunity"`
}

type SourceLabels struct {
	LastUpdated      string `json:"last_updated"`
	SalarySource     string `json:"salary_source"`
	DemandSource     string `json:"demand_source"`
	IsVerified       bool   `json:"is_verified"`
	ConfidenceScore  int    `json:"confidence_score"`
}

type CareerMetadata struct {
	AIProofScore       int    `json:"ai_proof_score,omitempty"`
	CreativeVsTechnical string `json:"creative_vs_technical,omitempty"`
	GovernmentVsPrivate string `json:"government_vs_private,omitempty"`
	FreelancePotential  int    `json:"freelance_potential,omitempty"`
	BurnoutRisk        int    `json:"burnout_risk,omitempty"`
	RemotePotential    int    `json:"remote_potential,omitempty"`
	GenderDiversity    string `json:"gender_diversity,omitempty"`
}

type Career struct {
	ID               int64           `db:"id" json:"id"`
	Title            string          `db:"title" json:"title"`
	Slug             string          `db:"slug" json:"slug"`
	Summary          string          `db:"summary" json:"summary"`
	Description      string          `db:"description" json:"description"`
	CategoryID       int64           `db:"category_id" json:"category_id"`
	DailyTasks       json.RawMessage `db:"daily_tasks" json:"daily_tasks"`
	Skills           pq.StringArray  `db:"skills" json:"skills"`
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

	SalaryTiers      json.RawMessage `db:"salary_tiers" json:"salary_tiers,omitempty"`
	CitySalaries     json.RawMessage `db:"city_salaries" json:"city_salaries,omitempty"`
	DemandData       json.RawMessage `db:"demand_data" json:"demand_data,omitempty"`
	SourceLabels     json.RawMessage `db:"source_labels" json:"source_labels,omitempty"`
	NeContent        json.RawMessage `db:"ne_content" json:"ne_content,omitempty"`
	CareerMetadata   json.RawMessage `db:"career_metadata" json:"career_metadata,omitempty"`
	WorkLifeBalance  int             `db:"work_life_balance" json:"work_life_balance"`
	StudyDuration    string          `db:"study_duration" json:"study_duration"`
	DegreeRequired   string          `db:"degree_required" json:"degree_required,omitempty"`
	CreativeScore    int             `db:"creative_score" json:"creative_score"`
	TechnicalScore   int             `db:"technical_score" json:"technical_score"`
	FreelancePotential int           `db:"freelance_potential" json:"freelance_potential"`
	IsGovernment     bool            `db:"is_government" json:"is_government"`
	IsRemoteOK       bool            `db:"is_remote_ok" json:"is_remote_ok"`
	ExamRequired     string          `db:"exam_required" json:"exam_required,omitempty"`

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

type CompareCareerRequest struct {
	CareerIDs []int64 `json:"career_ids"`
}

type CareerCompare struct {
	Career          Career          `json:"career"`
	SalaryRange     SalaryRange     `json:"salary_range"`
	Tags            []string        `json:"tags"`
	DemandData      DemandData      `json:"demand_data,omitempty"`
	SourceLabels    SourceLabels    `json:"source_labels,omitempty"`
}

type SavedComparison struct {
	ID        int64     `db:"id" json:"id"`
	UserID    int64     `db:"user_id" json:"user_id"`
	CareerIDs []int64   `db:"career_ids" json:"career_ids"`
	Name      string    `db:"name" json:"name"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

type CareerSource struct {
	ID             int64      `db:"id" json:"id"`
	CareerID       int64      `db:"career_id" json:"career_id"`
	SourceType     string     `db:"source_type" json:"source_type"`
	SourceName     string     `db:"source_name" json:"source_name"`
	SourceURL      string     `db:"source_url" json:"source_url"`
	IsVerified     bool       `db:"is_verified" json:"is_verified"`
	ConfidenceScore int       `db:"confidence_score" json:"confidence_score"`
	LastChecked    *time.Time `db:"last_checked" json:"last_checked,omitempty"`
	CreatedAt      time.Time  `db:"created_at" json:"created_at"`
}

type FilterOptions struct {
	Page               int    `json:"page"`
	Limit              int    `json:"limit"`
	Category           string `json:"category,omitempty"`
	Tag                string `json:"tag,omitempty"`
	Search             string `json:"search,omitempty"`
	MinSalary          int    `json:"min_salary,omitempty"`
	MaxSalary          int    `json:"max_salary,omitempty"`
	Difficulty         int    `json:"difficulty,omitempty"`
	MinFutureProof     int    `json:"min_future_proof,omitempty"`
	IsGovernment       *bool  `json:"is_government,omitempty"`
	IsRemoteOK         *bool  `json:"is_remote_ok,omitempty"`
	MinCreative        int    `json:"min_creative,omitempty"`
	MinTechnical       int    `json:"min_technical,omitempty"`
	MinFreelance       int    `json:"min_freelance,omitempty"`
	MinWorkLifeBalance int    `json:"min_work_life_balance,omitempty"`
	StudyDuration      string `json:"study_duration,omitempty"`
	DegreeRequired     string `json:"degree_required,omitempty"`
	Mode               string `json:"mode,omitempty"` // student or parent
}
