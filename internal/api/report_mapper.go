package api

import (
	"encoding/json"
	"strconv"
)

type reportMetadata struct {
	OverallScore            int    `json:"overall_score"`
	AlgorithmScore          int    `json:"algorithm_score"`
	ArchitectureScore       int    `json:"architecture_score"`
	CodingScore             int    `json:"coding_score"`
	SoftSkillsScore         int    `json:"soft_skills_score"`
	Comments                string `json:"comments"`
	InterviewTitle          string `json:"interview_title"`
	InterviewLevel          string `json:"interview_level"`
	InterviewSpecialization string `json:"interview_specialization"`
	InterviewScheduledAt    string `json:"interview_scheduled_at"`
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

	return map[string]any{
		"id":                  report["id"],
		"interview_id":        report["interview_id"],
		"overall_score":       overallScore,
		"algorithm_score":     algorithmScore,
		"architecture_score":  metadata.ArchitectureScore,
		"coding_score":        codingScore,
		"soft_skills_score":   softSkillsScore,
		"comments":            comments,
		"recommendations":     report["recommendations"],
		"strengths":           report["strengths"],
		"weaknesses":          report["weaknesses"],
		"created_at":          report["created_at"],
		"interview": map[string]any{
			"id":             report["interview_id"],
			"title":          title,
			"specialization": metadata.InterviewSpecialization,
			"level":          metadata.InterviewLevel,
			"scheduled_at":   metadata.InterviewScheduledAt,
			"candidate": map[string]any{
				"name":       "Кандидат",
				"email":      "",
				"phone":      "",
				"experience": 0,
			},
		},
	}
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
