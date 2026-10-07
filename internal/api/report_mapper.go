package api

import (
	"encoding/json"
	"net/url"
	"strconv"
	"strings"
)

type reportMetadata struct {
	OverallScore            int                  `json:"overall_score"`
	AlgorithmScore          int                  `json:"algorithm_score"`
	ArchitectureScore       int                  `json:"architecture_score"`
	CodingScore             int                  `json:"coding_score"`
	SoftSkillsScore         int                  `json:"soft_skills_score"`
	AlgorithmPassed         bool                 `json:"algorithm_passed"`
	ArchitecturePassed      bool                 `json:"architecture_passed"`
	CodingPassed            bool                 `json:"coding_passed"`
	SoftSkillsPassed        bool                 `json:"soft_skills_passed"`
	Comments                string               `json:"comments"`
	SummaryPublic           string               `json:"summary_public"`
	InterviewTitle          string               `json:"interview_title"`
	InterviewLevel          string               `json:"interview_level"`
	InterviewSpecialization string               `json:"interview_specialization"`
	InterviewScheduledAt    string               `json:"interview_scheduled_at"`
	AnswerReviews           []reportAnswerReview `json:"answer_reviews"`
	WeakPoints              []reportWeakPoint    `json:"weak_points"`
	AnalysisMode            string               `json:"analysis_mode"`
}

type reportWeakPoint struct {
	StepID        string `json:"step_id"`
	ItemType      string `json:"item_type"`
	Label         string `json:"label"`
	Prompt        string `json:"prompt"`
	Technology    string `json:"technology,omitempty"`
	UserAnswer    string `json:"user_answer"`
	CorrectAnswer string `json:"correct_answer"`
	Explanation   string `json:"explanation,omitempty"`
	Topic         string `json:"topic"`
	SourceURL     string `json:"source_url"`
}

type reportAnswerReview struct {
	StepID        string   `json:"step_id"`
	QuestionID    string   `json:"question_id"`
	ItemType      string   `json:"item_type"`
	SortOrder     int      `json:"sort_order"`
	Label         string   `json:"label"`
	Prompt        string   `json:"prompt"`
	Technology    string   `json:"technology,omitempty"`
	Options       []string `json:"options,omitempty"`
	SelectedIndex *int     `json:"selected_index,omitempty"`
	CorrectIndex  *int     `json:"correct_index,omitempty"`
	UserAnswer    string   `json:"user_answer"`
	CorrectAnswer string   `json:"correct_answer"`
	IsCorrect     *bool    `json:"is_correct,omitempty"`
}

// mapReportResponse flattens a stored report for the UI. Without full access
// (Basic plan) weak points, answer reviews and weaknesses are stripped and only
// the number of weak points is returned.
func mapReportResponse(report map[string]any, full bool) map[string]any {
	if report == nil {
		return nil
	}

	metadata := parseReportMetadata(stringValue(report["notes"]))
	overallScore := metadata.OverallScore
	if overallScore == 0 {
		if parsed, err := strconv.Atoi(stringValue(report["overall_rating"])); err == nil {
			overallScore = parsed
		}
	}

	algorithmScore := metadata.AlgorithmScore
	if algorithmScore == 0 {
		if parsed, err := strconv.Atoi(stringValue(report["technical_skills"])); err == nil {
			algorithmScore = parsed
		}
	}

	codingScore := metadata.CodingScore
	if codingScore == 0 {
		if parsed, err := strconv.Atoi(stringValue(report["problem_solving"])); err == nil {
			codingScore = parsed
		}
	}

	softSkillsScore := metadata.SoftSkillsScore
	if softSkillsScore == 0 {
		if parsed, err := strconv.Atoi(stringValue(report["communication_skills"])); err == nil {
			softSkillsScore = parsed
		}
	}

	comments := metadata.Comments
	if comments == "" {
		comments = stringValue(report["strengths"])
	}

	title := metadata.InterviewTitle
	if title == "" {
		title = "Тренировка"
	}

	// A Basic (no LLM) report can be re-analyzed once the user has Pro.
	aiAnalyzed := !isFallbackReport(comments) && !(full && metadata.AnalysisMode == "basic")

	algorithmPassed := resolveSectionPassed(metadata.AlgorithmPassed, algorithmScore)
	architecturePassed := resolveSectionPassed(metadata.ArchitecturePassed, metadata.ArchitectureScore)
	codingPassed := resolveSectionPassed(metadata.CodingPassed, codingScore)
	softSkillsPassed := resolveSectionPassed(metadata.SoftSkillsPassed, softSkillsScore)

	answerReviews := metadata.AnswerReviews
	if answerReviews == nil {
		answerReviews = []reportAnswerReview{}
	}
	if real, ok := realOverallScore(answerReviews, algorithmScore, codingScore); ok {
		overallScore = real
	}

	weakPoints := metadata.WeakPoints
	if weakPoints == nil {
		// Reports generated before weak points existed.
		weakPoints = weakPointsFromReviews(answerReviews, codingPassed)
	}

	weaknesses := report["weaknesses"]
	strengths := report["strengths"]
	recommendations := report["recommendations"]
	if !full {
		// Comments, strengths and recommendations name the weak topics, so
		// Basic gets only the topic-free public summary (empty for old reports).
		answerReviews = answeredOnly(answerReviews)
		weaknesses = ""
		strengths = ""
		recommendations = ""
		comments = metadata.SummaryPublic
	}
	visibleWeakPoints := weakPoints
	if !full {
		visibleWeakPoints = []reportWeakPoint{}
	}

	return map[string]any{
		"id":                  report["id"],
		"interview_id":        report["interview_id"],
		"user_id":             report["user_id"],
		"overall_score":       overallScore,
		"algorithm_score":     algorithmScore,
		"architecture_score":  metadata.ArchitectureScore,
		"coding_score":        codingScore,
		"soft_skills_score":   softSkillsScore,
		"algorithm_passed":    algorithmPassed,
		"architecture_passed": architecturePassed,
		"coding_passed":       codingPassed,
		"soft_skills_passed":  softSkillsPassed,
		"comments":            comments,
		"recommendations":     recommendations,
		"strengths":           strengths,
		"weaknesses":          weaknesses,
		"ai_analyzed":         aiAnalyzed,
		"created_at":          report["created_at"],
		"answer_reviews":      answerReviews,
		"weak_points":         visibleWeakPoints,
		"weak_points_count":   len(weakPoints),
		"locked":              !full,
		"interview": map[string]any{
			"id":             report["interview_id"],
			"title":          title,
			"specialization": metadata.InterviewSpecialization,
			"level":          metadata.InterviewLevel,
			"scheduled_at":   metadata.InterviewScheduledAt,
		},
	}
}

func resolveSectionPassed(stored bool, score int) bool {
	if stored {
		return true
	}
	if score > 0 {
		return score >= 60
	}
	return false
}

func parseReportMetadata(notes string) reportMetadata {
	if notes == "" {
		return reportMetadata{}
	}

	var metadata reportMetadata
	if err := json.Unmarshal([]byte(notes), &metadata); err != nil {
		return reportMetadata{}
	}

	return metadata
}

func stringValue(value any) string {
	if value == nil {
		return ""
	}
	if str, ok := value.(string); ok {
		return str
	}
	return ""
}

func protoReportToMap(report any) map[string]any {
	data, err := json.Marshal(report)
	if err != nil {
		return nil
	}

	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		return nil
	}

	return result
}

func isFallbackReport(comments string) bool {
	return strings.Contains(comments, "AI-анализ временно недоступен")
}

// realOverallScore recomputes the overall result as the mean of the sections
// the session had, so old reports that stored 0 for failed sections show the
// real percentage.
func realOverallScore(reviews []reportAnswerReview, theory, coding int) (int, bool) {
	hasQuestions, hasTasks := false, false
	for _, review := range reviews {
		if review.ItemType == "task" {
			hasTasks = true
		} else {
			hasQuestions = true
		}
	}
	scores := make([]int, 0, 2)
	if hasQuestions {
		scores = append(scores, theory)
	}
	if hasTasks {
		scores = append(scores, coding)
	}
	if len(scores) == 0 {
		return 0, false
	}
	sum := 0
	for _, score := range scores {
		sum += score
	}
	return int(float64(sum)/float64(len(scores)) + 0.5), true
}

func weakPointsFromReviews(reviews []reportAnswerReview, codingPassed bool) []reportWeakPoint {
	points := []reportWeakPoint{}
	for _, review := range reviews {
		if review.ItemType == "task" {
			if codingPassed && review.UserAnswer != "Нет ответа" {
				continue
			}
		} else if review.IsCorrect != nil && *review.IsCorrect {
			continue
		}
		topic := reviewTopic(review)
		points = append(points, reportWeakPoint{
			StepID:        review.StepID,
			ItemType:      review.ItemType,
			Label:         review.Label,
			Prompt:        review.Prompt,
			Technology:    review.Technology,
			UserAnswer:    review.UserAnswer,
			CorrectAnswer: review.CorrectAnswer,
			Topic:         topic,
			SourceURL:     "https://habr.com/ru/search/?q=" + url.QueryEscape(topic) + "&target_type=posts&order=relevance",
		})
	}
	return points
}

func reviewTopic(review reportAnswerReview) string {
	words := strings.Fields(review.Prompt)
	if len(words) > 8 {
		words = words[:8]
	}
	topic := strings.Trim(strings.Join(words, " "), " ?.:,")
	if review.Technology != "" && !strings.Contains(strings.ToLower(topic), strings.ToLower(review.Technology)) {
		topic = strings.TrimSpace(review.Technology + " " + topic)
	}
	return topic
}

// answeredOnly keeps the questions and the user's own answers for Basic: no
// correct answers, options or verdicts that would reveal the weak points.
func answeredOnly(reviews []reportAnswerReview) []reportAnswerReview {
	out := make([]reportAnswerReview, 0, len(reviews))
	for _, review := range reviews {
		out = append(out, reportAnswerReview{
			StepID:     review.StepID,
			ItemType:   review.ItemType,
			SortOrder:  review.SortOrder,
			Label:      review.Label,
			Prompt:     review.Prompt,
			Technology: review.Technology,
			UserAnswer: review.UserAnswer,
		})
	}
	return out
}
