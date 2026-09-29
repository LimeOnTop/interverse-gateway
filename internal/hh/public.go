package hh

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"

	"golang.org/x/net/html"
	"golang.org/x/net/publicsuffix"
)

// FetchPublicResume loads a publicly viewable hh.ru resume page by URL and parses it.
// No OAuth is required. Fields that HH hides from anonymous viewers may be empty.
func (c *Client) FetchPublicResume(ctx context.Context, resumeURL string) (Resume, error) {
	resumeURL = strings.TrimSpace(resumeURL)
	if resumeURL == "" {
		return Resume{}, fmt.Errorf("укажите ссылку на резюме hh.ru")
	}

	resumeID := ExtractResumeID(resumeURL)
	if resumeID == "" {
		return Resume{}, fmt.Errorf("некорректная ссылка на резюме hh.ru")
	}
	if err := validateHHResumeHost(resumeURL); err != nil {
		return Resume{}, err
	}

	pageURL := "https://hh.ru/resume/" + url.PathEscape(resumeID)
	htmlBody, err := c.fetchPublicHTML(ctx, pageURL)
	if err != nil {
		return Resume{}, err
	}

	resume, err := ParsePublicResumeHTML(htmlBody, resumeID)
	if err != nil {
		return Resume{}, err
	}
	if strings.TrimSpace(resume.Title) == "" &&
		len(resume.SkillSet) == 0 &&
		len(resume.Experience) == 0 &&
		(resume.Education == nil) &&
		len(resume.Language) == 0 {
		return Resume{}, fmt.Errorf("не удалось прочитать резюме по ссылке — убедитесь, что оно открыто для просмотра")
	}
	return resume, nil
}

func validateHHResumeHost(raw string) error {
	if !strings.Contains(raw, "://") {
		return nil // bare id already accepted via ExtractResumeID
	}
	parsed, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("некорректная ссылка на резюме")
	}
	host := strings.ToLower(parsed.Hostname())
	if host == "hh.ru" || strings.HasSuffix(host, ".hh.ru") {
		return nil
	}
	return fmt.Errorf("ожидается ссылка вида https://hh.ru/resume/...")
}

func (c *Client) fetchPublicHTML(ctx context.Context, pageURL string) (string, error) {
	jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	if err != nil {
		return "", fmt.Errorf("cookie jar: %w", err)
	}

	client := &http.Client{
		Timeout: 30 * time.Second,
		Jar:     jar,
	}

	// Warm up cookies — without them hh.ru often returns a captcha shell.
	if err := c.doPublicGET(ctx, client, "https://hh.ru/"); err != nil {
		return "", fmt.Errorf("подготовка сессии hh.ru: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, pageURL, nil)
	if err != nil {
		return "", fmt.Errorf("create resume request: %w", err)
	}
	setBrowserHeaders(req, c.cfg.UserAgent)
	req.Header.Set("Referer", "https://hh.ru/")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("загрузка страницы резюме: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 6_000_000))
	if err != nil {
		return "", fmt.Errorf("чтение страницы резюме: %w", err)
	}
	if resp.StatusCode == http.StatusNotFound {
		return "", fmt.Errorf("резюме не найдено по этой ссылке")
	}
	if resp.StatusCode == http.StatusForbidden {
		return "", fmt.Errorf("доступ к резюме ограничен — откройте его для просмотра по ссылке в настройках hh.ru")
	}
	if resp.StatusCode >= http.StatusBadRequest {
		return "", fmt.Errorf("hh.ru вернул ошибку (%d)", resp.StatusCode)
	}

	htmlBody := string(body)
	if looksLikeCaptchaOnly(htmlBody) {
		return "", fmt.Errorf("hh.ru запросил капчу — повторите попытку позже")
	}
	return htmlBody, nil
}

func (c *Client) doPublicGET(ctx context.Context, client *http.Client, rawURL string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return err
	}
	setBrowserHeaders(req, c.cfg.UserAgent)
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 2_000_000))
	return nil
}

func setBrowserHeaders(req *http.Request, userAgent string) {
	if strings.TrimSpace(userAgent) == "" || strings.HasPrefix(userAgent, "InterVerse/") {
		userAgent = "Mozilla/5.0 (compatible; InterVerse/1.0; +https://inter-verse.ru)"
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "ru-RU,ru;q=0.9,en-US;q=0.8,en;q=0.7")
	req.Header.Set("Cache-Control", "no-cache")
}

func looksLikeCaptchaOnly(page string) bool {
	hasCaptcha := strings.Contains(strings.ToLower(page), "captcha")
	hasResume := strings.Contains(page, `data-qa="resume-block`) ||
		strings.Contains(page, `data-qa="resume-block-title-position"`)
	return hasCaptcha && !hasResume
}

// ParsePublicResumeHTML extracts profile fields from a public hh.ru resume HTML page.
func ParsePublicResumeHTML(page, resumeID string) (Resume, error) {
	doc, err := html.Parse(strings.NewReader(page))
	if err != nil {
		return Resume{}, fmt.Errorf("разбор HTML резюме: %w", err)
	}

	resume := Resume{ID: resumeID}

	titleNode := findByDataQA(doc, "resume-block-title-position")
	if titleNode == nil {
		titleNode = findByDataQA(doc, "bloko-header-1")
	}
	resume.Title = cleanText(textContent(titleNode))

	skillsRoot := findByDataQA(doc, "skills-table")
	skillNodes := []*html.Node{}
	if skillsRoot != nil {
		skillNodes = findAllByDataQA(skillsRoot, "bloko-tag__text")
		if len(skillNodes) == 0 {
			skillNodes = findAllByClass(skillsRoot, "bloko-tag__text")
		}
	}
	if len(skillNodes) == 0 {
		skillNodes = findAllByDataQA(doc, "bloko-tag__text")
	}
	if len(skillNodes) == 0 {
		skillNodes = findAllByClass(doc, "bloko-tag__text")
	}
	for _, tag := range skillNodes {
		name := cleanText(textContent(tag))
		if name == "" || looksLikeLanguageTag(name) {
			continue
		}
		resume.SkillSet = append(resume.SkillSet, struct {
			Name string `json:"name"`
		}{Name: name})
	}
	resume.SkillSet = uniqueSkillSet(resume.SkillSet)

	langRoot := findByDataQA(doc, "resume-block-languages")
	if langRoot != nil {
		for _, item := range findAllByDataQA(langRoot, "resume-block-language-item") {
			raw := cleanText(textContent(item))
			if raw == "" {
				continue
			}
			resume.Language = append(resume.Language, parseLanguageItem(raw))
		}
	}

	eduRoot := findByDataQA(doc, "resume-block-education")
	if eduRoot != nil {
		resume.Education = parseEducationBlock(eduRoot)
	}

	expRoot := findByDataQA(doc, "resume-block-experience")
	if expRoot != nil {
		resume.Experience = parseExperienceBlock(expRoot)
		if len(resume.Experience) == 0 {
			if summary := extractExperienceSummary(expRoot); summary != "" && resume.Skills == "" {
				resume.Skills = summary
			}
		}
	}

	if photo := findResumePhotoURL(doc); photo != "" {
		resume.Photo = &struct {
			Small   string `json:"small"`
			Medium  string `json:"medium"`
			Large   string `json:"40"`
			Size100 string `json:"100"`
			Size500 string `json:"500"`
		}{Medium: photo, Size500: photo}
	}

	return resume, nil
}

func parseLanguageItem(raw string) struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Level *struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"level"`
} {
	parts := splitDashParts(raw)
	item := struct {
		ID    string `json:"id"`
		Name  string `json:"name"`
		Level *struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"level"`
	}{Name: parts[0]}

	lower := strings.ToLower(parts[0])
	switch {
	case strings.Contains(lower, "англ") || strings.Contains(lower, "english"):
		item.ID = "eng"
	case strings.Contains(lower, "рус") || strings.Contains(lower, "russian"):
		item.ID = "rus"
	}

	if len(parts) >= 2 {
		levelRaw := strings.Join(parts[1:], " — ")
		levelID := ""
		for _, code := range []string{"a1", "a2", "b1", "b2", "c1", "c2"} {
			if strings.Contains(strings.ToLower(levelRaw), code) {
				levelID = code
				break
			}
		}
		if strings.Contains(strings.ToLower(levelRaw), "родн") || strings.Contains(strings.ToLower(levelRaw), "native") {
			levelID = "l1"
		}
		item.Level = &struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		}{ID: levelID, Name: levelRaw}
	}
	return item
}

func parseEducationBlock(root *html.Node) *struct {
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
} {
	texts := collectBlockParagraphs(root)
	edu := &struct {
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
	}{}

	for _, text := range texts {
		lower := strings.ToLower(text)
		if strings.Contains(lower, "образован") && edu.Level == nil {
			edu.Level = &struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			}{Name: text}
			continue
		}
		if text == "Образование" {
			continue
		}
		edu.Primary = append(edu.Primary, struct {
			Name         string `json:"name"`
			Organization string `json:"organization"`
			Result       string `json:"result"`
			Year         int    `json:"year"`
		}{Name: text})
	}
	if edu.Level == nil && len(edu.Primary) == 0 {
		return nil
	}
	return edu
}

func parseExperienceBlock(root *html.Node) []struct {
	Company     string  `json:"company"`
	Position    string  `json:"position"`
	Start       string  `json:"start"`
	End         *string `json:"end"`
	Description string  `json:"description"`
} {
	positions := findAllByDataQA(root, "resume-block-experience-position")
	if len(positions) == 0 {
		return nil
	}

	companies := findAllByDataQA(root, "resume-block-experience-company")
	descriptions := findAllByDataQA(root, "resume-block-experience-description")
	intervals := findAllByDataQA(root, "resume-block-experience-interval")
	if len(intervals) == 0 {
		intervals = findAllByDataQA(root, "resume-block-experience-period")
	}

	out := make([]struct {
		Company     string  `json:"company"`
		Position    string  `json:"position"`
		Start       string  `json:"start"`
		End         *string `json:"end"`
		Description string  `json:"description"`
	}, 0, len(positions))

	for i, posNode := range positions {
		item := struct {
			Company     string  `json:"company"`
			Position    string  `json:"position"`
			Start       string  `json:"start"`
			End         *string `json:"end"`
			Description string  `json:"description"`
		}{
			Position: cleanText(textContent(posNode)),
		}
		if i < len(companies) {
			item.Company = cleanText(textContent(companies[i]))
		}
		if i < len(descriptions) {
			item.Description = cleanText(textContent(descriptions[i]))
		}
		if i < len(intervals) {
			item.Start = cleanText(textContent(intervals[i]))
		}
		if item.Position != "" || item.Company != "" {
			out = append(out, item)
		}
	}
	return out
}

func extractExperienceSummary(root *html.Node) string {
	title := findByClass(root, "resume-block__title-text_sub")
	text := cleanText(textContent(title))
	if strings.HasPrefix(strings.ToLower(text), "опыт работы") && len([]rune(text)) > len("опыт работы") {
		return text
	}
	return ""
}

func findResumePhotoURL(doc *html.Node) string {
	for _, qa := range []string{"resume-photo-desktop", "resume-photo-mobile"} {
		node := findByDataQA(doc, qa)
		if node == nil {
			continue
		}
		img := findFirstTag(node, "img")
		if img == nil {
			continue
		}
		src := attr(img, "src")
		if src == "" || strings.Contains(src, "qr-code") || strings.HasSuffix(src, ".svg") {
			continue
		}
		if strings.HasPrefix(src, "//") {
			src = "https:" + src
		}
		return src
	}
	return ""
}

func looksLikeLanguageTag(name string) bool {
	lower := strings.ToLower(name)
	return strings.Contains(lower, "русский") ||
		strings.Contains(lower, "английский") ||
		strings.Contains(lower, " — a1") ||
		strings.Contains(lower, " — a2") ||
		strings.Contains(lower, " — b1") ||
		strings.Contains(lower, " — b2") ||
		strings.Contains(lower, " — c1") ||
		strings.Contains(lower, " — c2") ||
		strings.Contains(lower, "родной")
}

func splitDashParts(raw string) []string {
	raw = strings.ReplaceAll(raw, "–", "—")
	parts := strings.Split(raw, "—")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	if len(out) == 0 {
		return []string{strings.TrimSpace(raw)}
	}
	return out
}

func collectBlockParagraphs(root *html.Node) []string {
	var out []string
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode && (n.Data == "p" || n.Data == "li") {
			text := cleanText(textContent(n))
			if text != "" {
				out = append(out, text)
			}
			return
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(root)
	return uniqueStrings(out)
}

func findByDataQA(n *html.Node, qa string) *html.Node {
	var found *html.Node
	var walk func(*html.Node)
	walk = func(node *html.Node) {
		if found != nil || node == nil {
			return
		}
		if node.Type == html.ElementNode && attr(node, "data-qa") == qa {
			found = node
			return
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(n)
	return found
}

func findAllByDataQA(n *html.Node, qa string) []*html.Node {
	var out []*html.Node
	var walk func(*html.Node)
	walk = func(node *html.Node) {
		if node == nil {
			return
		}
		if node.Type == html.ElementNode && attr(node, "data-qa") == qa {
			out = append(out, node)
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(n)
	return out
}

func findByClass(n *html.Node, className string) *html.Node {
	var found *html.Node
	var walk func(*html.Node)
	walk = func(node *html.Node) {
		if found != nil || node == nil {
			return
		}
		if node.Type == html.ElementNode && hasClass(node, className) {
			found = node
			return
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(n)
	return found
}

func findAllByClass(n *html.Node, className string) []*html.Node {
	var out []*html.Node
	var walk func(*html.Node)
	walk = func(node *html.Node) {
		if node == nil {
			return
		}
		if node.Type == html.ElementNode && hasClass(node, className) {
			out = append(out, node)
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(n)
	return out
}

func findFirstTag(n *html.Node, tag string) *html.Node {
	var found *html.Node
	var walk func(*html.Node)
	walk = func(node *html.Node) {
		if found != nil || node == nil {
			return
		}
		if node.Type == html.ElementNode && node.Data == tag {
			found = node
			return
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(n)
	return found
}

func hasClass(n *html.Node, className string) bool {
	for _, part := range strings.Fields(attr(n, "class")) {
		if part == className {
			return true
		}
	}
	return false
}

func attr(n *html.Node, key string) string {
	for _, a := range n.Attr {
		if a.Key == key {
			return a.Val
		}
	}
	return ""
}

func textContent(n *html.Node) string {
	if n == nil {
		return ""
	}
	var b strings.Builder
	var walk func(*html.Node)
	walk = func(node *html.Node) {
		if node.Type == html.TextNode {
			b.WriteString(node.Data)
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(n)
	return b.String()
}

func cleanText(s string) string {
	s = strings.ReplaceAll(s, "\u00a0", " ")
	s = strings.ReplaceAll(s, "\u202f", " ")
	return strings.Join(strings.Fields(s), " ")
}

func uniqueStrings(items []string) []string {
	seen := make(map[string]struct{}, len(items))
	out := make([]string, 0, len(items))
	for _, item := range items {
		key := strings.ToLower(item)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, item)
	}
	return out
}

func uniqueSkillSet(items []struct {
	Name string `json:"name"`
}) []struct {
	Name string `json:"name"`
} {
	seen := make(map[string]struct{}, len(items))
	out := make([]struct {
		Name string `json:"name"`
	}, 0, len(items))
	for _, item := range items {
		key := strings.ToLower(strings.TrimSpace(item.Name))
		if key == "" {
			continue
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, item)
	}
	return out
}
