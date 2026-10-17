package api

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	authv1 "github.com/LimeOnTop/interverse-contracts/auth/gen"
	"github.com/LimeOnTop/interverse-gateway/internal/usecase"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"golang.org/x/oauth2"
)

type securityOAuthAuth struct{ usecase.AuthGateway }

func (securityOAuthAuth) OAuthLogin(context.Context, string, string, string) (*authv1.OAuthLoginResponse, error) {
	return &authv1.OAuthLoginResponse{AccessToken: "fixture-access", RefreshToken: "fixture-refresh", NewRegistration: true, User: &authv1.User{Id: "fixture-user"}}, nil
}

type securityTransport func(*http.Request) (*http.Response, error)

func (f securityTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
func TestOAuthStatePKCEAndSingleUseExchange(t *testing.T) {
	addr := os.Getenv("REDIS_TEST_ADDR")
	if addr == "" {
		t.Skip("requires isolated Redis")
	}
	cache := redis.NewClient(&redis.Options{Addr: addr, DB: 1})
	defer cache.Close()
	gin.SetMode(gin.TestMode)
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		if r.Form.Get("code_verifier") == "" {
			t.Error("PKCE verifier missing")
		}
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"access_token":"google-fixture","token_type":"Bearer"}`)
	}))
	defer provider.Close()
	original := http.DefaultClient
	http.DefaultClient = &http.Client{Transport: securityTransport(func(r *http.Request) (*http.Response, error) {
		if r.URL.Host == "www.googleapis.com" {
			return &http.Response{StatusCode: 200, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(`{"email":"oauth@example.invalid","name":"Fixture","verified_email":true}`))}, nil
		}
		return http.DefaultTransport.RoundTrip(r)
	})}
	defer func() { http.DefaultClient = original }()
	a := NewGoogleOAuthAPI(securityOAuthAuth{}, "fixture", "fixture", "http://localhost/callback", "http://localhost", cache)
	a.oauthConfig.Endpoint = oauth2.Endpoint{AuthURL: provider.URL + "/authorize", TokenURL: provider.URL + "/token"}
	router := gin.New()
	router.GET("/login", a.Login)
	router.GET("/callback", a.Callback)
	router.POST("/exchange", a.Exchange)
	request := func(method, path, body string, cookie *http.Cookie, origin string) *httptest.ResponseRecorder {
		r := httptest.NewRequest(method, path, strings.NewReader(body))
		r.Header.Set("Content-Type", "application/json")
		if cookie != nil {
			r.AddCookie(cookie)
		}
		if origin != "" {
			r.Header.Set("Origin", origin)
		}
		w := httptest.NewRecorder()
		router.ServeHTTP(w, r)
		return w
	}
	login := request("GET", "/login", "", nil, "")
	if login.Code != 302 {
		t.Fatalf("login %d", login.Code)
	}
	redirect, _ := url.Parse(login.Header().Get("Location"))
	state := redirect.Query().Get("state")
	if state == "" || redirect.Query().Get("code_challenge_method") != "S256" || redirect.Query().Get("code_challenge") == "" {
		t.Fatal("state/PKCE missing")
	}
	stateCookie := login.Result().Cookies()[0]
	if !stateCookie.HttpOnly || stateCookie.SameSite != http.SameSiteLaxMode {
		t.Fatal("unsafe state cookie")
	}
	if w := request("GET", "/callback?code=fixture&state=wrong", "", stateCookie, ""); w.Code != 400 {
		t.Fatal("invalid state accepted")
	}
	callback := request("GET", "/callback?code=fixture&state="+state, "", stateCookie, "")
	if callback.Code != 302 {
		t.Fatalf("callback %d %s", callback.Code, callback.Body.String())
	}
	if w := request("GET", "/callback?code=fixture&state="+state, "", stateCookie, ""); w.Code != 400 {
		t.Fatal("state reused")
	}
	location := callback.Header().Get("Location")
	if strings.Contains(location, "fixture-access") || strings.Contains(location, "refresh_token") {
		t.Fatal("JWT leaked into URL")
	}
	dest, _ := url.Parse(location)
	code := dest.Query().Get("code")
	var binding *http.Cookie
	for _, c := range callback.Result().Cookies() {
		if c.Name == "iv_oauth_exchange" {
			binding = c
		}
	}
	if binding == nil {
		t.Fatal("binding missing")
	}
	body := `{"code":"` + code + `"}`
	wrong := *binding
	wrong.Value = "wrong"
	if request("POST", "/exchange", body, &wrong, "").Code != 400 {
		t.Fatal("wrong binding accepted")
	}
	if request("POST", "/exchange", body, binding, "https://attacker.invalid").Code != 403 {
		t.Fatal("foreign origin accepted")
	}
	exchange := request("POST", "/exchange", body, binding, "http://localhost")
	if exchange.Code != 200 {
		t.Fatalf("exchange %d", exchange.Code)
	}
	var tokens oauthExchange
	json.Unmarshal(exchange.Body.Bytes(), &tokens)
	if !tokens.NewRegistration || tokens.UserID != "fixture-user" {
		t.Fatal("registration metadata lost in exchange")
	}
	if tokens.AccessToken != "fixture-access" {
		t.Fatal("wrong token")
	}
	if request("POST", "/exchange", body, binding, "").Code != 400 {
		t.Fatal("exchange replay accepted")
	}
	expired, _ := oauthRandom()
	raw, _ := json.Marshal(oauthExchange{Binding: binding.Value})
	cache.Set(context.Background(), oauthKey("exchange", expired), raw, time.Millisecond)
	time.Sleep(5 * time.Millisecond)
	if request("POST", "/exchange", `{"code":"`+expired+`"}`, binding, "").Code != 400 {
		t.Fatal("expired code accepted")
	}
}
