package api

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/redis/go-redis/v9"
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
	cache       *redis.Client
	authClient  usecase.AuthGateway
	oauthConfig *oauth2.Config
	frontendURL string
	enabled     bool
}

func NewGoogleOAuthAPI(
	authClient usecase.AuthGateway,
	clientID, clientSecret, redirectURL, frontendURL string,
	caches ...*redis.Client,
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
	var cache *redis.Client
	if len(caches) > 0 {
		cache = caches[0]
	}
	return &GoogleOAuthAPI{
		cache:       cache,
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
	if a.cache == nil {
		apperr.Public(c, http.StatusServiceUnavailable, "OAuth unavailable")
		return
	}
	state, err := oauthRandom()
	if err != nil {
		apperr.Internal(c, err)
		return
	}
	verifier := oauth2.GenerateVerifier()
	if err = a.cache.Set(c.Request.Context(), oauthKey("state", state), verifier, 10*time.Minute).Err(); err != nil {
		apperr.Internal(c, err)
		return
	}
	a.cookie(c, "iv_oauth_state", state, 600)
	c.Header("Cache-Control", "no-store")
	c.Redirect(http.StatusFound, a.oauthConfig.AuthCodeURL(state, oauth2.AccessTypeOnline, oauth2.S256ChallengeOption(verifier)))
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

	state := c.Query("state")
	cookie, err := c.Cookie("iv_oauth_state")
	if err != nil || state == "" || subtle.ConstantTimeCompare([]byte(state), []byte(cookie)) != 1 || a.cache == nil {
		apperr.Public(c, http.StatusBadRequest, "Invalid OAuth state")
		return
	}
	verifier, err := a.cache.GetDel(c.Request.Context(), oauthKey("state", state)).Result()
	a.cookie(c, "iv_oauth_state", "", -1)
	if err != nil {
		apperr.Public(c, http.StatusBadRequest, "Expired OAuth state")
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 15*time.Second)
	defer cancel()

	token, err := a.oauthConfig.Exchange(ctx, code, oauth2.VerifierOption(verifier))
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

	exchangeCode, err := oauthRandom()
	if err != nil {
		apperr.Internal(c, err)
		return
	}
	binding, err := oauthRandom()
	if err != nil {
		apperr.Internal(c, err)
		return
	}
	record, _ := json.Marshal(oauthExchange{NewRegistration: resp.GetNewRegistration(), UserID: resp.GetUser().GetId(), Binding: binding, AccessToken: resp.GetAccessToken(), RefreshToken: resp.GetRefreshToken()})
	if err = a.cache.Set(ctx, oauthKey("exchange", exchangeCode), record, time.Minute).Err(); err != nil {
		apperr.Internal(c, err)
		return
	}
	a.cookie(c, "iv_oauth_exchange", binding, 60)
	c.Header("Cache-Control", "no-store")
	c.Redirect(http.StatusFound, a.frontendURL+"/oauth/callback?code="+url.QueryEscape(exchangeCode))
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
		Verified bool   `json:"verified_email"`
		Email    string `json:"email"`
		Name     string `json:"name"`
	}
	if err := json.Unmarshal(body, &profile); err != nil {
		return "", "", err
	}
	if profile.Email == "" || !profile.Verified {
		return "", "", fmt.Errorf("google profile missing email")
	}
	return profile.Email, profile.Name, nil
}

type oauthExchange struct {
	NewRegistration bool   `json:"new_registration"`
	UserID          string `json:"user_id"`
	Binding         string `json:"binding"`
	AccessToken     string `json:"access_token"`
	RefreshToken    string `json:"refresh_token"`
}

func oauthRandom() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b), err
}
func oauthKey(kind, value string) string {
	return fmt.Sprintf("oauth:%s:%x", kind, sha256.Sum256([]byte(value)))
}
func (a *GoogleOAuthAPI) cookie(c *gin.Context, name, value string, age int) {
	u, _ := url.Parse(a.frontendURL)
	secure := u != nil && u.Scheme == "https"
	http.SetCookie(c.Writer, &http.Cookie{Name: name, Value: value, Path: "/", MaxAge: age, HttpOnly: true, Secure: secure, SameSite: http.SameSiteLaxMode})
}
func (a *GoogleOAuthAPI) Exchange(c *gin.Context) {
	c.Header("Cache-Control", "no-store")
	u, _ := url.Parse(a.frontendURL)
	origin := ""
	if u != nil {
		origin = u.Scheme + "://" + u.Host
	}
	if requestOrigin := c.GetHeader("Origin"); requestOrigin != "" && requestOrigin != origin {
		apperr.Public(c, http.StatusForbidden, "Invalid origin")
		return
	}
	var req struct {
		Code string `json:"code" binding:"required,len=43"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apperr.Bind(c, err)
		return
	}
	binding, err := c.Cookie("iv_oauth_exchange")
	if err != nil || a.cache == nil {
		apperr.Public(c, http.StatusBadRequest, "Invalid OAuth exchange")
		return
	}
	raw, err := a.cache.Get(c.Request.Context(), oauthKey("exchange", req.Code)).Bytes()
	if err != nil {
		apperr.Public(c, http.StatusBadRequest, "Expired OAuth exchange")
		return
	}
	var record oauthExchange
	if json.Unmarshal(raw, &record) != nil || subtle.ConstantTimeCompare([]byte(record.Binding), []byte(binding)) != 1 {
		apperr.Public(c, http.StatusBadRequest, "Invalid OAuth exchange")
		return
	}
	// Compare-and-delete makes the exchange single-use even with concurrent requests.
	used, err := a.cache.Eval(c.Request.Context(), `if redis.call('GET',KEYS[1])==ARGV[1] then redis.call('DEL',KEYS[1]);return 1 end;return 0`, []string{oauthKey("exchange", req.Code)}, string(raw)).Int()
	if err != nil || used != 1 {
		apperr.Public(c, http.StatusBadRequest, "Expired OAuth exchange")
		return
	}
	a.cookie(c, "iv_oauth_exchange", "", -1)
	c.JSON(http.StatusOK, gin.H{"access_token": record.AccessToken, "refresh_token": record.RefreshToken, "new_registration": record.NewRegistration, "user_id": record.UserID})
}
