package api

import (
	"sort"
	"strings"
)

// reportFocusGroup is a technology with the weak topics found in it.
type reportFocusGroup struct {
	Title  string   `json:"title"`
	Count  int      `json:"count"`
	Topics []string `json:"topics"`
}

// reportHeadline is the two-line verdict shown on the report and dashboard,
// derived from the section scores only (no LLM text, safe for Basic).
func reportHeadline(theory, coding int, hasTheory, hasTasks bool) string {
	const pass = 60
	switch {
	case hasTheory && hasTasks:
		switch {
		case theory >= 80 && coding >= 80:
			return "Уверенный результат.\nПопробуй следующий уровень."
		case theory >= pass && coding >= pass:
			return "Хороший результат.\nРазбери оставшиеся ошибки."
		case coding >= pass:
			return "Практика получается.\nЗакрепи теорию."
		case theory >= pass:
			return "Теория получается.\nЗакрепи её на практике."
		default:
			return "Есть над чем поработать.\nНачни с разбора ошибок."
		}
	case hasTheory:
		if theory >= 80 {
			return "Уверенная теория.\nДобавь практические задачи."
		}
		if theory >= pass {
			return "Хорошая теория.\nРазбери оставшиеся ошибки."
		}
		return "Закрепи теорию.\nНачни с разбора ошибок."
	case hasTasks:
		if coding >= pass {
			return "Практика получается.\nПроверь теоретическую базу."
		}
		return "Закрепи практику.\nНачни с разбора ошибок."
	default:
		return "Тренировка завершена.\nПопробуй следующую."
	}
}

// focusGroups groups weak points by technology, biggest group first, with up
// to three topics each.
func focusGroups(points []reportWeakPoint) []reportFocusGroup {
	byTitle := map[string]*reportFocusGroup{}
	order := []string{}
	for _, point := range points {
		title := strings.TrimSpace(point.Technology)
		if title == "" {
			title = "Общие темы"
		}
		group, ok := byTitle[title]
		if !ok {
			group = &reportFocusGroup{Title: title, Topics: []string{}}
			byTitle[title] = group
			order = append(order, title)
		}
		group.Count++
		topic := strings.TrimSpace(point.Topic)
		if topic != "" && len(group.Topics) < 3 && !containsFold(group.Topics, topic) {
			group.Topics = append(group.Topics, topic)
		}
	}

	groups := make([]reportFocusGroup, 0, len(order))
	for _, title := range order {
		groups = append(groups, *byTitle[title])
	}
	sort.SliceStable(groups, func(i, j int) bool { return groups[i].Count > groups[j].Count })
	return groups
}

func containsFold(values []string, value string) bool {
	for _, item := range values {
		if strings.EqualFold(item, value) {
			return true
		}
	}
	return false
}

// reviewTotals counts theory questions and practical tasks of the session and
// lists its technologies in order of appearance.
func reviewTotals(reviews []reportAnswerReview) (questions, tasks int, technologies []string) {
	technologies = []string{}
	for _, review := range reviews {
		if review.ItemType == "task" {
			tasks++
		} else {
			questions++
		}
		tech := strings.TrimSpace(review.Technology)
		if tech != "" && !containsFold(technologies, tech) {
			technologies = append(technologies, tech)
		}
	}
	return questions, tasks, technologies
}

func personalReportSummary(theory, coding int, hasTheory, hasTasks bool) string {
	switch {
	case hasTheory && hasTasks:
		switch {
		case theory >= 80 && coding >= 80:
			return "Ты уверенно справляешься с теорией и практикой. Попробуй тренировку уровнем выше или выбери новую технологию."
		case theory >= 60 && coding >= 60:
			return "У тебя хорошая база в теории и практике. Разбери ошибки этой тренировки и повтори темы, в которых ещё есть сомнения."
		case theory >= 60:
			return "С теорией у тебя получается лучше, чем с практикой. Запусти тренировку в режиме «Много задач» (4–6) и разбирай решения шаг за шагом — это поможет закрепить знания в коде."
		case coding >= 60:
			return "С практическими задачами ты справляешься увереннее, чем с теорией. Выбери «Только теория» и повтори вопросы с ошибками, чтобы лучше объяснять свои решения."
		default:
			return "Пока есть трудности и в теории, и в практике. Начни с разбора ошибок, затем выбери тренировку с 1–3 задачами и постепенно увеличивай нагрузку."
		}
	case hasTheory:
		if theory >= 60 {
			return "С теоретической частью у тебя хороший результат. В этой тренировке не было задач, поэтому практику мы пока не оценивали. Добавь 1–3 задачи в следующую тренировку, чтобы проверить, как применяешь знания в коде."
		}
		return "Теорию стоит закрепить. Разбери вопросы с ошибками и попробуй ещё одну тренировку в режиме «Только теория». Практических задач здесь не было — эта часть ещё не оценивалась."
	case hasTasks:
		if coding >= 60 {
			return "Ты хорошо справляешься с практическими задачами. Пройди тренировку с теоретическими вопросами, чтобы проверить и эту часть подготовки."
		}
		return "Практические задачи пока вызывают трудности. Начни с 1–3 задач, сверяй решения с разбором и повторяй их без подсказок."
	default:
		return "В этой тренировке нет ответов для оценки. Запусти новую тренировку и ответь на вопросы, чтобы получить разбор."
	}
}

// Legacy summaries are presented in the current coaching voice without rewriting stored scores.
func legacyReportVoice(text string) bool {
	text = strings.ToLower(text)
	return strings.Contains(text, "кандидат") || strings.Contains(text, "candidate") ||
		(strings.Contains(text, "уровень") && (strings.Contains(text, "подтвержд") || strings.Contains(text, "достигнут")))
}
