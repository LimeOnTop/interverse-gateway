package hh

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Config struct {
	UserAgent string
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
