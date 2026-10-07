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
			return "Уверенный результат.\nМожно повышать уровень."
		case theory >= pass && coding >= pass:
			return "Хороший результат.\nОстались точечные пробелы."
		case coding >= pass:
			return "Сильная практика.\nТеория требует закрепления."
		case theory >= pass:
			return "Хорошая теория.\nПрактику стоит подтянуть."
		default:
			return "Есть над чем поработать.\nНачните с разбора ошибок."
		}
	case hasTheory:
		if theory >= 80 {
			return "Уверенная теория.\nДобавьте практические задачи."
		}
		if theory >= pass {
			return "Хорошая теория.\nОстались точечные пробелы."
		}
		return "Теория требует закрепления.\nНачните с разбора ошибок."
	case hasTasks:
		if coding >= pass {
			return "Сильная практика.\nЗакрепите теорию."
		}
		return "Практику стоит подтянуть.\nНачните с разбора ошибок."
	default:
		return "Тренировка завершена.\nНачните следующую."
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
