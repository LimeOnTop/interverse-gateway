package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	pb "github.com/LimeOnTop/interverse-contracts/vacancy/gen"
	"github.com/LimeOnTop/interverse-gateway/internal/usecase"
	"github.com/gin-gonic/gin"
)

type publicVacancyStub struct {
	usecase.VacancyGateway
	calls int
	query string
	fail  bool
}

func (s *publicVacancyStub) SearchVacancies(_ context.Context, text, area string, page, perPage int32, sources []string) (*pb.SearchVacanciesResponse, error) {
	s.calls++
	s.query = text
	if area != "" || page != 0 || perPage != 50 || len(sources) != 0 {
		panic("unbounded public search")
	}
	if s.fail {
		return nil, errors.New("provider unavailable")
	}
	return &pb.SearchVacanciesResponse{Response: &pb.Response{Success: true}, Items: []*pb.Vacancy{{Id: "1", Title: "Python <script>alert(1)</script>", Url: "https://hh.ru/vacancy/1", CompanyName: "Test"}, {Id: "bad", Title: "Bad", Url: "javascript:alert(1)"}}}, nil
}
func publicRouter(a *PublicJobsAPI) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/jobs/:track", a.Get)
	r.GET("/pages/:track", a.Page)
	return r
}
func TestPublicJobsWhitelistAndCache(t *testing.T) {
	cache := &intentCache{values: map[string]string{}}
	stub := &publicVacancyStub{}
	r := publicRouter(NewPublicJobsAPI(cache, stub, t.TempDir()))
	bad := httptest.NewRecorder()
	r.ServeHTTP(bad, httptest.NewRequest("GET", "/jobs/arbitrary?user_id=private", nil))
	if bad.Code != 404 || stub.calls != 0 {
		t.Fatal("unknown track reached upstream")
	}
	for i := 0; i < 2; i++ {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest("GET", "/jobs/python?text=private&per_page=99999", nil))
		if w.Code != 200 {
			t.Fatal(w.Code, w.Body.String())
		}
		var f publicFeed
		if json.Unmarshal(w.Body.Bytes(), &f) != nil || f.Track != "python" || len(f.Items) != 1 || f.Items[0].Id != "1" {
			t.Fatal("invalid public projection", w.Body.String())
		}
	}
	if stub.calls != 1 || stub.query != "Python developer" {
		t.Fatal("public search was not fixed and cached", stub.calls, stub.query)
	}
}
func TestPublicJobsStaleFallback(t *testing.T) {
	raw, _ := json.Marshal(publicFeed{Track: "java", Items: []*pb.Vacancy{{Title: "Java", Url: "https://hh.ru/vacancy/2"}}, UpdatedAt: "2026-10-10T00:00:00Z"})
	cache := &intentCache{values: map[string]string{"public:jobs:v2:java:stale": string(raw)}}
	r := publicRouter(NewPublicJobsAPI(cache, &publicVacancyStub{fail: true}, t.TempDir()))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/jobs/java", nil))
	var f publicFeed
	_ = json.Unmarshal(w.Body.Bytes(), &f)
	if w.Code != 200 || !f.Stale || len(f.Items) != 1 {
		t.Fatal("cached fallback lost or unlabelled", w.Body.String())
	}
}
func TestPublicJobsHTMLIsEscapedAndContainsData(t *testing.T) {
	dir := t.TempDir()
	_ = os.WriteFile(filepath.Join(dir, "python.html"), []byte(`<main><!--PUBLIC_JOB_CARDS--></main><!--PUBLIC_JOB_DATA-->`), 0600)
	r := publicRouter(NewPublicJobsAPI(&intentCache{values: map[string]string{}}, &publicVacancyStub{}, dir))
	warm := httptest.NewRecorder()
	r.ServeHTTP(warm, httptest.NewRequest("GET", "/jobs/python", nil))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/pages/python", nil))
	html := w.Body.String()
	if w.Code != 200 || !strings.Contains(html, "public-jobs-data") || !strings.Contains(html, "&lt;script&gt;") || strings.Contains(html, "<script>alert(1)") || strings.Contains(html, "javascript:") {
		t.Fatal("unsafe or missing server-rendered data", html)
	}
}

func TestPublicJobRelevance(t *testing.T) {
	cases := []struct {
		track, title, snippet string
		skills                []string
		want                  bool
	}{
		{"react", "Golang разработчик", "", []string{"Go", "PostgreSQL"}, false},
		{"react", "QA Engineer", "React", nil, false},
		{"react", "Frontend разработчик", "Опыт React и TypeScript", nil, true},
		{"react", "Fullstack разработчик", "", []string{"Python", "React"}, true},
		{"java", "JavaScript developer", "", nil, false},
		{"java", "Java-разработчик", "", nil, true},
		{"frontend", "Фронтенд-разработчик", "", nil, true},
		{"python", "Python developer", "", nil, true},
	}
	for _, c := range cases {
		if publicJobRelevant(c.track, &pb.Vacancy{Title: c.title, Snippet: c.snippet, Skills: c.skills}) != c.want {
			t.Errorf("wrong relevance: %s %s", c.track, c.title)
		}
	}
}

type slowPublicVacancy struct {
	usecase.VacancyGateway
	started, release chan struct{}
}

func (s *slowPublicVacancy) SearchVacancies(ctx context.Context, text, area string, page, perPage int32, sources []string) (*pb.SearchVacanciesResponse, error) {
	close(s.started)
	select {
	case <-s.release:
	case <-ctx.Done():
	}
	return nil, errors.New("slow provider")
}

func TestPublicPageDoesNotWaitForProviders(t *testing.T) {
	dir := t.TempDir()
	_ = os.WriteFile(filepath.Join(dir, "python.html"), []byte(`<header>Navigation</header><main><!--PUBLIC_JOB_CARDS--></main><!--PUBLIC_JOB_DATA-->`), 0600)
	stub := &slowPublicVacancy{started: make(chan struct{}), release: make(chan struct{})}
	defer close(stub.release)
	a := NewPublicJobsAPI(&intentCache{values: map[string]string{}}, stub, dir)
	w := httptest.NewRecorder()
	done := make(chan struct{})
	go func() { publicRouter(a).ServeHTTP(w, httptest.NewRequest("GET", "/pages/python", nil)); close(done) }()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("page waited for vacancy providers")
	}
	if w.Code != 200 || !strings.Contains(w.Body.String(), "Navigation") || !strings.Contains(w.Body.String(), "role=\"status\"") {
		t.Fatal("page shell missing", w.Body.String())
	}
	select {
	case <-stub.started:
	case <-time.After(time.Second):
		t.Fatal("background vacancy load not started")
	}
}
