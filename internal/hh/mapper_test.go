package hh

import "testing"

func TestExtractResumeID(t *testing.T) {
	cases := map[string]string{
		"https://hh.ru/resume/abc123def":        "abc123def",
		"https://hh.ru/resume/abc123def?query=1": "abc123def",
		"abc123def":                             "abc123def",
		"":                                      "",
	}

	for input, want := range cases {
		got := ExtractResumeID(input)
		if got != want {
			t.Fatalf("ExtractResumeID(%q)=%q, want %q", input, got, want)
		}
	}
}

func TestMapEnglishLevel(t *testing.T) {
	resume := Resume{
		Language: []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
			Level *struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"level"`
		}{
			{
				ID:   "eng",
				Name: "Английский",
				Level: &struct {
					ID   string `json:"id"`
					Name string `json:"name"`
				}{ID: "b2", Name: "B2 — Средний"},
			},
		},
	}

	if got := mapEnglishLevel(resume); got != "B2" {
		t.Fatalf("mapEnglishLevel()=%q, want B2", got)
	}
}

func TestMapResumeFields(t *testing.T) {
	end := "2023-01-01"
	resume := Resume{
		ID:     "r1",
		Title:  "Go Developer",
		Skills: "Пишу сервисы на Go",
		SkillSet: []struct {
			Name string `json:"name"`
		}{{Name: "Go"}, {Name: "PostgreSQL"}},
		Experience: []struct {
			Company     string  `json:"company"`
			Position    string  `json:"position"`
			Start       string  `json:"start"`
			End         *string `json:"end"`
			Description string  `json:"description"`
		}{
			{
				Company:     "InterVerse",
				Position:    "Backend",
				Start:       "2021-01-01",
				End:         &end,
				Description: "gRPC сервисы",
			},
		},
		Education: &struct {
			Level *struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"level"`
			Primary []struct {
				Name         string `json:"name"`
				Organization string `json:"organization"`
				Result       string `json:"result"`
				Year         int    `json:"year"`
			} `json:"primary"`
		}{
			Level: &struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			}{ID: "higher", Name: "Высшее"},
			Primary: []struct {
				Name         string `json:"name"`
				Organization string `json:"organization"`
				Result       string `json:"result"`
				Year         int    `json:"year"`
			}{
				{Name: "МГУ", Result: "Прикладная математика", Year: 2020},
			},
		},
	}

	mapped := MapResume(resume, "data:image/jpeg;base64,abc")
	if mapped.AvatarURL == "" || mapped.AboutMe == "" || mapped.WorkExperience == "" || mapped.HigherEducation == "" {
		t.Fatalf("mapped profile has empty fields: %+v", mapped)
	}
}
