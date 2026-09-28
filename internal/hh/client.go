package hh

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	authorizeURL = "https://hh.ru/oauth/authorize"
	tokenURL     = "https://hh.ru/oauth/token"
	apiBaseURL   = "https://api.hh.ru"
)

type Config struct {
	ClientID     string
	ClientSecret string
	RedirectURI  string
	UserAgent    string
}

type Client struct {
	cfg        Config
	httpClient *http.Client
}

func NewClient(cfg Config) *Client {
	if cfg.UserAgent == "" {
		cfg.UserAgent = "InterVerse/1.0 (dev@interverse.local)"
	}

	return &Client{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) Enabled() bool {
	return c.cfg.ClientID != "" && c.cfg.ClientSecret != "" && c.cfg.RedirectURI != ""
}

func (c *Client) AuthURL(state string) (string, error) {
	if !c.Enabled() {
		return "", fmt.Errorf("hh.ru OAuth is not configured")
	}

	values := url.Values{}
	values.Set("response_type", "code")
	values.Set("client_id", c.cfg.ClientID)
	values.Set("redirect_uri", c.cfg.RedirectURI)
	if state != "" {
		values.Set("state", state)
	}

	return authorizeURL + "?" + values.Encode(), nil
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}

func (c *Client) ExchangeCode(ctx context.Context, code string) (TokenResponse, error) {
	if !c.Enabled() {
		return TokenResponse{}, fmt.Errorf("hh.ru OAuth is not configured")
	}

	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("client_id", c.cfg.ClientID)
	form.Set("client_secret", c.cfg.ClientSecret)
	form.Set("redirect_uri", c.cfg.RedirectURI)
	form.Set("code", code)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return TokenResponse{}, fmt.Errorf("create token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", c.cfg.UserAgent)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return TokenResponse{}, fmt.Errorf("exchange hh code: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return TokenResponse{}, fmt.Errorf("read token response: %w", err)
	}

	if resp.StatusCode >= http.StatusBadRequest {
		return TokenResponse{}, fmt.Errorf("hh token exchange failed (%d): %s", resp.StatusCode, string(body))
	}

	var token TokenResponse
	if err := json.Unmarshal(body, &token); err != nil {
		return TokenResponse{}, fmt.Errorf("parse token response: %w", err)
	}
	if token.AccessToken == "" {
		return TokenResponse{}, fmt.Errorf("hh token response has no access_token")
	}

	return token, nil
}

type resumeListResponse struct {
	Items []resumeSummary `json:"items"`
	Found int             `json:"found"`
}

type resumeSummary struct {
	ID     string `json:"id"`
	Title  string `json:"title"`
	Status struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"status"`
}

type Resume struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Skills   string `json:"skills"`
	SkillSet []struct {
		Name string `json:"name"`
	} `json:"skill_set"`
	Photo *struct {
		Small   string `json:"small"`
		Medium  string `json:"medium"`
		Large   string `json:"40"`
		Size100 string `json:"100"`
		Size500 string `json:"500"`
	} `json:"photo"`
	Experience []struct {
		Company     string  `json:"company"`
		Position    string  `json:"position"`
		Start       string  `json:"start"`
		End         *string `json:"end"`
		Description string  `json:"description"`
	} `json:"experience"`
	Education *struct {
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
	} `json:"education"`
	Language []struct {
		ID    string `json:"id"`
		Name  string `json:"name"`
		Level *struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"level"`
	} `json:"language"`
}

func (c *Client) ListMineResumes(ctx context.Context, accessToken string) ([]resumeSummary, error) {
	var payload resumeListResponse
	if err := c.getJSON(ctx, accessToken, "/resumes/mine", &payload); err != nil {
		return nil, err
	}
	return payload.Items, nil
}

func (c *Client) GetResume(ctx context.Context, accessToken, resumeID string) (Resume, error) {
	var resume Resume
	if err := c.getJSON(ctx, accessToken, "/resumes/"+url.PathEscape(resumeID), &resume); err != nil {
		return Resume{}, err
	}
	return resume, nil
}

func (c *Client) DownloadImageAsDataURL(ctx context.Context, accessToken, imageURL string) (string, error) {
	if imageURL == "" {
		return "", nil
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, imageURL, nil)
	if err != nil {
		return "", fmt.Errorf("create image request: %w", err)
	}
	req.Header.Set("User-Agent", c.cfg.UserAgent)
	if accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+accessToken)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("download image: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 2_000_000))
	if err != nil {
		return "", fmt.Errorf("read image: %w", err)
	}
	if resp.StatusCode >= http.StatusBadRequest {
		return "", fmt.Errorf("download image failed (%d)", resp.StatusCode)
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType == "" || !strings.HasPrefix(contentType, "image/") {
		contentType = "image/jpeg"
	}

	return "data:" + contentType + ";base64," + base64.StdEncoding.EncodeToString(body), nil
}

func (c *Client) getJSON(ctx context.Context, accessToken, path string, dest any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiBaseURL+path, nil)
	if err != nil {
		return fmt.Errorf("create hh request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("User-Agent", c.cfg.UserAgent)
	req.Header.Set("HH-User-Agent", c.cfg.UserAgent)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("call hh api: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read hh response: %w", err)
	}

	if resp.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("hh api %s failed (%d): %s", path, resp.StatusCode, string(body))
	}

	if err := json.Unmarshal(body, dest); err != nil {
		return fmt.Errorf("parse hh response: %w", err)
	}

	return nil
}

func ExtractResumeID(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}

	if !strings.Contains(raw, "/") && !strings.Contains(raw, "?") {
		return raw
	}

	parsed, err := url.Parse(raw)
	if err != nil {
		return ""
	}

	parts := strings.Split(strings.Trim(parsed.Path, "/"), "/")
	for i, part := range parts {
		if part == "resume" && i+1 < len(parts) {
			id := parts[i+1]
			if idx := strings.IndexAny(id, "?#"); idx >= 0 {
				id = id[:idx]
			}
			return id
		}
	}

	return ""
}
