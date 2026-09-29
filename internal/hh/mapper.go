package hh

import (
	"encoding/json"
	"fmt"
	"strings"
)

type ImportedProfile struct {
	WorkExperience  string `json:"work_experience"`
	AvatarURL       string `json:"avatar_url"`
	AboutMe         string `json:"about_me"`
	HigherEducation string `json:"higher_education"`
	EnglishLevel    string `json:"english_level"`
	Skills          string `json:"skills"`
	ResumeID        string `json:"resume_id"`
	ResumeTitle     string `json:"resume_title"`
	Source          string `json:"source"`
}

func MapResume(resume Resume, avatarDataURL string) ImportedProfile {
	return ImportedProfile{
		WorkExperience:  formatExperience(resume),
		AvatarURL:       avatarDataURL,
		AboutMe:         formatAboutMe(resume),
		HigherEducation: formatEducation(resume),
		EnglishLevel:    mapEnglishLevel(resume),
		Skills:          formatSkills(resume),
		ResumeID:        resume.ID,
		ResumeTitle:     resume.Title,
		Source:          "hh.ru",
	}
}

func formatAboutMe(resume Resume) string {
	parts := make([]string, 0, 2)

	if title := strings.TrimSpace(resume.Title); title != "" {
		parts = append(parts, "Желаемая должность: "+title)
	}

	if skills := strings.TrimSpace(resume.Skills); skills != "" {
		parts = append(parts, skills)
	}

	return strings.Join(parts, "\n\n")
}

func formatSkills(resume Resume) string {
	names := make([]string, 0, len(resume.SkillSet))
	seen := make(map[string]struct{}, len(resume.SkillSet))
	for _, skill := range resume.SkillSet {
		name := strings.TrimSpace(skill.Name)
		if name == "" {
			continue
		}
		key := strings.ToLower(name)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		names = append(names, name)
	}
	if len(names) == 0 {
		return ""
	}
	encoded, err := json.Marshal(names)
	if err != nil {
		return ""
	}
	return string(encoded)
}

func formatExperience(resume Resume) string {
	if len(resume.Experience) == 0 {
		return ""
	}

	blocks := make([]string, 0, len(resume.Experience))
	for _, item := range resume.Experience {
		company := strings.TrimSpace(item.Company)
		position := strings.TrimSpace(item.Position)
		period := formatPeriod(item.Start, item.End)

		headerParts := make([]string, 0, 3)
		if position != "" {
			headerParts = append(headerParts, position)
		}
		if company != "" {
			headerParts = append(headerParts, company)
		}
		if period != "" {
			headerParts = append(headerParts, period)
		}

		block := strings.Join(headerParts, " — ")
		if description := strings.TrimSpace(item.Description); description != "" {
			block += "\n" + description
		}
		if block != "" {
			blocks = append(blocks, block)
		}
	}

	return strings.Join(blocks, "\n\n")
}

func formatPeriod(start string, end *string) string {
	start = strings.TrimSpace(start)
	endValue := "настоящее время"
	if end != nil && strings.TrimSpace(*end) != "" {
		endValue = strings.TrimSpace(*end)
	}
	if start == "" {
		return ""
	}
	return fmt.Sprintf("%s — %s", start, endValue)
}

func formatEducation(resume Resume) string {
	if resume.Education == nil {
		return ""
	}

	parts := make([]string, 0, len(resume.Education.Primary)+1)
	if resume.Education.Level != nil {
		if name := strings.TrimSpace(resume.Education.Level.Name); name != "" {
			parts = append(parts, name)
		}
	}

	for _, item := range resume.Education.Primary {
		chunks := make([]string, 0, 4)
		if name := firstNonEmpty(item.Name, item.Organization); name != "" {
			chunks = append(chunks, name)
		}
		if result := strings.TrimSpace(item.Result); result != "" {
			chunks = append(chunks, result)
		}
		if item.Year > 0 {
			chunks = append(chunks, fmt.Sprintf("%d", item.Year))
		}
		if len(chunks) > 0 {
			parts = append(parts, strings.Join(chunks, ", "))
		}
	}

	return strings.Join(parts, "; ")
}

func mapEnglishLevel(resume Resume) string {
	for _, language := range resume.Language {
		id := strings.ToLower(strings.TrimSpace(language.ID))
		name := strings.ToLower(strings.TrimSpace(language.Name))
		isEnglish := id == "eng" || id == "en" ||
			strings.Contains(name, "англий") ||
			strings.Contains(name, "english")
		if !isEnglish || language.Level == nil {
			continue
		}

		levelID := strings.ToLower(strings.TrimSpace(language.Level.ID))
		levelName := strings.ToLower(strings.TrimSpace(language.Level.Name))

		switch {
		case levelID == "a1" || strings.Contains(levelName, "a1"):
			return "A1"
		case levelID == "a2" || strings.Contains(levelName, "a2"):
			return "A2"
		case levelID == "b1" || strings.Contains(levelName, "b1"):
			return "B1"
		case levelID == "b2" || strings.Contains(levelName, "b2"):
			return "B2"
		case levelID == "c1" || strings.Contains(levelName, "c1"):
			return "C1"
		case levelID == "c2" || strings.Contains(levelName, "c2"):
			return "C2"
		case levelID == "l1" || strings.Contains(levelName, "родн") || strings.Contains(levelName, "native"):
			return "native"
		}
	}

	return ""
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func PickResumeID(items []resumeSummary, preferredID string) string {
	if preferredID != "" {
		for _, item := range items {
			if item.ID == preferredID {
				return preferredID
			}
		}
	}

	for _, item := range items {
		if strings.EqualFold(item.Status.ID, "published") || strings.EqualFold(item.Status.Name, "опубликовано") {
			return item.ID
		}
	}

	if len(items) > 0 {
		return items[0].ID
	}

	return ""
}

func (r Resume) PhotoURL() string {
	if r.Photo == nil {
		return ""
	}
	for _, candidate := range []string{r.Photo.Size500, r.Photo.Medium, r.Photo.Size100, r.Photo.Small, r.Photo.Large} {
		if strings.TrimSpace(candidate) != "" {
			return candidate
		}
	}
	return ""
}
