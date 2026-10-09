package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/LimeOnTop/interverse-gateway/internal/apperr"
	"github.com/LimeOnTop/interverse-gateway/internal/usecase"
	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type GoogleOAuthAPI struct {
	authClient  usecase.AuthGateway
	oauthConfig *oauth2.Config
	frontendURL string
	enabled     bool
}

func NewGoogleOAuthAPI(
	authClient usecase.AuthGateway,
	clientID, clientSecret, redirectURL, frontendURL string,
) *GoogleOAuthAPI {
	enabled := strings.TrimSpace(clientID) != "" &&
		strings.TrimSpace(clientSecret) != "" &&
		!strings.Contains(clientID, "your-google") &&
		!strings.Contains(clientSecret, "your-google")

	cfg := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes:       []string{"openid", "email", "profile"},
		Endpoint:     google.Endpoint,
	}
	if frontendURL == "" {
		frontendURL = "https://inter-verse.ru"
	}
	return &GoogleOAuthAPI{
		authClient:  authClient,
		oauthConfig: cfg,
		frontendURL: strings.TrimRight(frontendURL, "/"),
		enabled:     enabled,
	}
}

func (a *GoogleOAuthAPI) Login(c *gin.Context) {
	if !a.enabled {
		apperr.Public(c, http.StatusServiceUnavailable, "Google OAuth не настроен")
		return
	}
	state := fmt.Sprintf("%d", time.Now().UnixNano())
	url := a.oauthConfig.AuthCodeURL(state, oauth2.AccessTypeOnline)
	c.Redirect(http.StatusFound, url)
}

func (a *GoogleOAuthAPI) Callback(c *gin.Context) {
	if !a.enabled {
		apperr.Public(c, http.StatusServiceUnavailable, "Google OAuth не настроен")
		return
	}
	if errMsg := c.Query("error"); errMsg != "" {
		c.Redirect(http.StatusFound, a.frontendURL+"/login?error="+url.QueryEscape(errMsg))
		return
	}
	code := c.Query("code")
	if code == "" {
		apperr.Public(c, http.StatusBadRequest, "missing code")
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 15*time.Second)
	defer cancel()

	token, err := a.oauthConfig.Exchange(ctx, code)
	if err != nil {
		apperr.Internal(c, err)
		return
	}

	email, name, err := fetchGoogleProfile(ctx, token.AccessToken)
	if err != nil {
		apperr.Internal(c, err)
		return
	}

	resp, err := a.authClient.OAuthLogin(ctx, email, name, "google")
	if err != nil {
		apperr.Internal(c, err)
		return
	}
	if resp.GetResponse() != nil && !resp.GetResponse().GetSuccess() {
		apperr.Upstream(c, http.StatusBadRequest, resp.GetResponse().GetError())
		return
	}

	redirect := fmt.Sprintf(
		"%s/oauth/callback?access_token=%s&refresh_token=%s",
		a.frontendURL,
		url.QueryEscape(resp.GetAccessToken()),
		url.QueryEscape(resp.GetRefreshToken()),
	)
	c.Redirect(http.StatusFound, redirect)
}

func fetchGoogleProfile(ctx context.Context, accessToken string) (email, name string, err error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://www.googleapis.com/oauth2/v2/userinfo", nil)
	if err != nil {
		return "", "", err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", "", err
	}
	defer res.Body.Close()
	body, err := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	if err != nil {
		return "", "", err
	}
	if res.StatusCode >= 300 {
		return "", "", fmt.Errorf("google userinfo status %d: %s", res.StatusCode, string(body))
	}
	var profile struct {
		Email string `json:"email"`
		Name  string `json:"name"`
	}
	if err := json.Unmarshal(body, &profile); err != nil {
		return "", "", err
	}
	if profile.Email == "" {
		return "", "", fmt.Errorf("google profile missing email")
	}
	return profile.Email, profile.Name, nil
}
