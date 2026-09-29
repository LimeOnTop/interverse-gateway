package hh

import "testing"

func TestParsePublicResumeHTML(t *testing.T) {
	const page = `
<html><body>
  <span data-qa="resume-block-title-position">Go Developer</span>
	<div data-qa="skills-table">
    <span data-qa="bloko-tag__text">Go</span>
    <span data-qa="bloko-tag__text">PostgreSQL</span>
    <span data-qa="bloko-tag__text">Русский — Родной</span>
  </div>
  <div data-qa="resume-block-experience">
    <span class="resume-block__title-text resume-block__title-text_sub">Опыт работы 3 года</span>
    <span data-qa="resume-block-experience-position">Backend</span>
    <span data-qa="resume-block-experience-company">InterVerse</span>
    <span data-qa="resume-block-experience-interval">2021 — 2024</span>
    <div data-qa="resume-block-experience-description">gRPC сервисы</div>
  </div>
  <div data-qa="resume-block-education"><p>Высшее образование</p><p>МГУ, Прикладная математика, 2020</p></div>
  <div data-qa="resume-block-languages">
    <span data-qa="resume-block-language-item">Английский — B2 — Средний</span>
  </div>
</body></html>`

	resume, err := ParsePublicResumeHTML(page, "abc123")
	if err != nil {
		t.Fatal(err)
	}
	if resume.Title != "Go Developer" {
		t.Fatalf("title=%q", resume.Title)
	}
	if len(resume.SkillSet) != 2 {
		t.Fatalf("skills=%v", resume.SkillSet)
	}
	if len(resume.Experience) != 1 || resume.Experience[0].Company != "InterVerse" {
		t.Fatalf("experience=%v", resume.Experience)
	}
	if resume.Education == nil || resume.Education.Level == nil || resume.Education.Level.Name != "Высшее образование" {
		t.Fatalf("education=%v", resume.Education)
	}
	if got := mapEnglishLevel(resume); got != "B2" {
		t.Fatalf("english=%q", got)
	}
}

func TestExtractResumeID(t *testing.T) {
	cases := map[string]string{
		"https://hh.ru/resume/abc123def":         "abc123def",
		"https://hh.ru/resume/abc123def?query=1": "abc123def",
		"abc123def": "abc123def",
		"":          "",
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
			ID    string `json:"id"`
			Name  string `json:"name"`
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
	if mapped.Skills != `["Go","PostgreSQL"]` {
		t.Fatalf("skills=%q", mapped.Skills)
	}
}
