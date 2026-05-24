package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/guruorgoru/carevo/internal/models"
	"github.com/jmoiron/sqlx"
)

type Worker struct {
	db       *sqlx.DB
	provider Provider
	jobs     chan int64
}

func NewWorker(db *sqlx.DB, provider Provider) *Worker {
	return &Worker{
		db:       db,
		provider: provider,
		jobs:     make(chan int64, 100),
	}
}

func (w *Worker) Start(ctx context.Context) {
	go w.run(ctx)
}

func (w *Worker) Submit(userID int64) {
	select {
	case w.jobs <- userID:
	default:
		log.Printf("worker queue full, dropping job for user %d", userID)
	}
}

func (w *Worker) run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case userID := <-w.jobs:
			w.process(ctx, userID)
		}
	}
}

type surveyQA struct {
	Question string `json:"question"`
	Answer   any    `json:"answer"`
}

func (w *Worker) process(ctx context.Context, userID int64) {
	_, err := w.db.Exec(
		`INSERT INTO ai_results (user_id, status) VALUES ($1, 'processing')
		 ON CONFLICT (user_id) DO UPDATE SET status = 'processing', raw_response = NULL, recommended_careers = NULL, completed_at = NULL, error_message = NULL`,
		userID,
	)
	if err != nil {
		log.Printf("worker: create ai_results for user %d: %v", userID, err)
		return
	}

	var responses []struct {
		QuestionText string          `db:"question_text"`
		Category     string          `db:"category"`
		Answer       json.RawMessage `db:"answer"`
	}
	err = w.db.Select(&responses, `
		SELECT sq.question_text, sq.category, sr.answer
		FROM survey_responses sr
		JOIN survey_questions sq ON sq.id = sr.question_id
		WHERE sr.user_id = $1
		ORDER BY sq.sort_order`, userID)
	if err != nil {
		errMsg := err.Error()
		w.db.Exec(`UPDATE ai_results SET status = 'error', error_message = $2 WHERE user_id = $1`, userID, errMsg)
		log.Printf("worker: fetch responses for user %d: %v", userID, err)
		return
	}

	if len(responses) == 0 {
		errMsg := "no survey responses found"
		w.db.Exec(`UPDATE ai_results SET status = 'error', error_message = $2 WHERE user_id = $1`, userID, errMsg)
		return
	}

	var qas []surveyQA
	for _, r := range responses {
		var answer any
		json.Unmarshal(r.Answer, &answer)
		qas = append(qas, surveyQA{Question: fmt.Sprintf("[%s] %s", r.Category, r.QuestionText), Answer: answer})
	}
	qaJSON, _ := json.MarshalIndent(qas, "", "  ")

	var careers []struct {
		ID    int64  `db:"id"`
		Title string `db:"title"`
		Slug  string `db:"slug"`
	}
	w.db.Select(&careers, "SELECT id, title, slug FROM careers ORDER BY title")

	var careerNames []string
	careerMap := make(map[string]int64)
	careerTitles := make(map[int64]string)
	for _, c := range careers {
		careerNames = append(careerNames, fmt.Sprintf("%s (slug: %s)", c.Title, c.Slug))
		careerMap[strings.ToLower(c.Title)] = c.ID
		careerTitles[c.ID] = c.Title
	}

	systemPrompt := `You are a career counselor for Nepal. Your job is to match users to careers based on their survey answers.

Return ONLY valid JSON. Every array element inside "recommended_careers" MUST start with {"title":...}. This is the exact format:

{"recommended_careers":[{"title":"Exact Career Title","score":85,"reasoning":"2-3 sentence explanation why this career fits"}]}

CRITICAL RULES — follow these exactly:
- EVERY element in the recommended_careers array MUST start with {"title" — never omit the opening brace
- Score must be 0-100
- Return 5-10 career matches sorted by score descending
- Use the EXACT career title from the provided list — copy it character for character
- Be realistic about Nepal job market conditions
- Consider education requirements, skills, and local demand
- Return ONLY the JSON object, no markdown, no code fences, no extra text`

	userPrompt := fmt.Sprintf(`Here are the user's survey answers:
%s

Here are the available careers:
%s

Return the top career matches as JSON.`, string(qaJSON), strings.Join(careerNames, "\n"))

	result, err := w.provider.GenerateJSON(ctx, systemPrompt, userPrompt)
	if err != nil {
		errMsg := fmt.Sprintf("AI provider error: %v", err)
		w.db.Exec(`UPDATE ai_results SET status = 'error', error_message = $2 WHERE user_id = $1`, userID, errMsg)
		log.Printf("worker: ai generate for user %d: %v", userID, err)
		return
	}

	if strings.TrimSpace(result) == "" {
		w.db.Exec(`UPDATE ai_results SET status = 'error', error_message = 'AI returned empty response' WHERE user_id = $1`, userID)
		log.Printf("worker: empty response for user %d", userID)
		return
	}

	var scored models.ScoredCareerList
	if err := json.Unmarshal([]byte(result), &scored); err != nil {
		log.Printf("worker: parse response for user %d (retrying): %v", userID, err)
		log.Printf("worker: raw response (%d chars): %.400s", len(result), result)

		result, err = w.provider.GenerateJSON(ctx, systemPrompt, userPrompt)
		if err != nil {
			errMsg := fmt.Sprintf("AI provider error (retry): %v", err)
			w.db.Exec(`UPDATE ai_results SET status = 'error', error_message = $2 WHERE user_id = $1`, userID, errMsg)
			log.Printf("worker: ai generate retry for user %d: %v", userID, err)
			return
		}

		if strings.TrimSpace(result) == "" {
			w.db.Exec(`UPDATE ai_results SET status = 'error', error_message = 'AI returned empty response (retry)' WHERE user_id = $1`, userID)
			log.Printf("worker: empty response on retry for user %d", userID)
			return
		}

		if err := json.Unmarshal([]byte(result), &scored); err != nil {
			log.Printf("worker: raw retry response (%d chars): %.400s", len(result), result)

			fixed := fixMalformedJSON([]byte(result))
			log.Printf("worker: fixed response (%d chars): %.400s", len(fixed), string(fixed))

			if err := json.Unmarshal(fixed, &scored); err == nil && len(scored.RecommendedCareers) > 0 {
				log.Printf("worker: JSON repair succeeded for user %d", userID)
				goto done
			}

			if parsed := extractJSON(fixed); parsed != nil {
				if err := json.Unmarshal(parsed, &scored); err == nil && len(scored.RecommendedCareers) > 0 {
					log.Printf("worker: extractJSON after repair succeeded for user %d", userID)
					goto done
				}
			}

			errMsg := fmt.Sprintf("failed to parse AI response after retry: %v", err)
			w.db.Exec(`UPDATE ai_results SET status = 'error', error_message = $2 WHERE user_id = $1`, userID, errMsg)
			log.Printf("worker: parse response for user %d after retry: %v", userID, err)
			return
		}
	}

done:

	for i := range scored.RecommendedCareers {
		title := strings.TrimSpace(strings.ToLower(scored.RecommendedCareers[i].Title))
		if id, ok := careerMap[title]; ok {
			scored.RecommendedCareers[i].CareerID = id
			scored.RecommendedCareers[i].Title = careerTitles[id]
			continue
		}
		for dbTitle, id := range careerMap {
			if strings.Contains(dbTitle, title) || strings.Contains(title, dbTitle) {
				scored.RecommendedCareers[i].CareerID = id
				scored.RecommendedCareers[i].Title = careerTitles[id]
				break
			}
		}
	}

	scoredJSON, _ := json.Marshal(scored.RecommendedCareers)

	tx, err := w.db.Beginx()
	if err != nil {
		w.db.Exec(`UPDATE ai_results SET status = 'error', error_message = 'transaction error' WHERE user_id = $1`, userID)
		return
	}
	defer tx.Rollback()

	_, err = tx.Exec(
		`UPDATE ai_results SET raw_response = $2, recommended_careers = $3, status = 'completed', completed_at = NOW() WHERE user_id = $1`,
		userID, result, scoredJSON,
	)
	if err != nil {
		log.Printf("worker: update ai_results for user %d: %v", userID, err)
		return
	}

	for _, sc := range scored.RecommendedCareers {
		if sc.CareerID == 0 {
			continue
		}
		_, err = tx.Exec(
			`INSERT INTO career_scores (user_id, career_id, score, reasoning) VALUES ($1, $2, $3, $4)
			 ON CONFLICT (user_id, career_id) DO UPDATE SET score = $3, reasoning = $4`,
			userID, sc.CareerID, sc.Score, sc.Reasoning,
		)
		if err != nil {
			log.Printf("worker: insert career_score for user %d career %d: %v", userID, sc.CareerID, err)
			return
		}
	}

	if err := tx.Commit(); err != nil {
		log.Printf("worker: commit for user %d: %v", userID, err)
		return
	}

	log.Printf("worker: completed for user %d with %d career matches", userID, len(scored.RecommendedCareers))
}

func extractJSON(in []byte) []byte {
	s := string(in)

	start := strings.Index(s, "{")
	if start < 0 {
		return nil
	}
	s = s[start:]

	depth := 0
	for i, c := range s {
		switch c {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return []byte(s[:i+1])
			}
		}
	}
	return nil
}

func fixMalformedJSON(in []byte) []byte {
	s := string(in)

	lines := strings.Split(s, "\n")
	var fixed []string
	inArray := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if strings.Contains(trimmed, `"recommended_careers"`) {
			inArray = true
			fixed = append(fixed, line)
			continue
		}

		if inArray && trimmed == "]" {
			inArray = false
			fixed = append(fixed, line)
			continue
		}

		if inArray && trimmed == "[" {
			fixed = append(fixed, line)
			continue
		}

		if inArray {
			hasBrace := strings.Contains(trimmed, "{")
			startsWithQuote := strings.HasPrefix(trimmed, `"`)

			if startsWithQuote && !hasBrace {
				indent := line[:len(line)-len(strings.TrimLeft(line, " \t"))]
				fixed = append(fixed, indent+"{")
				fixed = append(fixed, line)
				continue
			}

			if trimmed == "}" || trimmed == "}," {
				continue
			}

			if trimmed == `,` || trimmed == `` {
				continue
			}

			if hasBrace && strings.HasPrefix(trimmed, "{") {
				fixed = append(fixed, line)
				continue
			}

			fixed = append(fixed, line)
		} else {
			fixed = append(fixed, line)
		}
	}

	result := strings.Join(fixed, "\n")

	if strings.HasSuffix(strings.TrimSpace(result), ",") {
		result = strings.TrimRight(result, " \t,")
	}
	result = strings.TrimSpace(result)
	if !strings.HasSuffix(result, "}") {
		result += "\n}"
	}

	return []byte(result)
}
