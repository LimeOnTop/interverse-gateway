package api

import (
	"encoding/json"
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
	InterviewTitle          string               `json:"interview_title"`
	InterviewLevel          string               `json:"interview_level"`
	InterviewSpecialization string               `json:"interview_specialization"`
	InterviewScheduledAt    string               `json:"interview_scheduled_at"`
	AnswerReviews           []reportAnswerReview `json:"answer_reviews"`
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

func mapReportResponse(report map[string]any) map[string]any {
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

	aiAnalyzed := !isFallbackReport(comments)

	algorithmPassed := resolveSectionPassed(metadata.AlgorithmPassed, algorithmScore)
	architecturePassed := resolveSectionPassed(metadata.ArchitecturePassed, metadata.ArchitectureScore)
	codingPassed := resolveSectionPassed(metadata.CodingPassed, codingScore)
	softSkillsPassed := resolveSectionPassed(metadata.SoftSkillsPassed, softSkillsScore)

	answerReviews := metadata.AnswerReviews
	if answerReviews == nil {
		answerReviews = []reportAnswerReview{}
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
		"recommendations":     report["recommendations"],
		"strengths":           report["strengths"],
		"weaknesses":          report["weaknesses"],
		"ai_analyzed":         aiAnalyzed,
		"created_at":          report["created_at"],
		"answer_reviews":      answerReviews,
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
