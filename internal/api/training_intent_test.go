package api

import (
	"context"
	"encoding/json"
	pb "github.com/LimeOnTop/interverse-contracts/interview/gen"
	"github.com/LimeOnTop/interverse-gateway/internal/middleware"
	"github.com/LimeOnTop/interverse-gateway/internal/usecase"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

type intentCache struct {
	redis.UniversalClient
	mu     sync.Mutex
	values map[string]string
	locked bool
	ttl    time.Duration
}

func (s *intentCache) Set(_ context.Context, k string, v interface{}, ttl time.Duration) *redis.StatusCmd {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.values[k] = string(v.([]byte))
	s.ttl = ttl
	return redis.NewStatusResult("OK", nil)
}
func (s *intentCache) Get(_ context.Context, k string) *redis.StringCmd {
	s.mu.Lock()
	defer s.mu.Unlock()
	v, ok := s.values[k]
	if !ok {
		return redis.NewStringResult("", redis.Nil)
	}
	return redis.NewStringResult(v, nil)
}
func (s *intentCache) Incr(context.Context, string) *redis.IntCmd { return redis.NewIntResult(1, nil) }
func (s *intentCache) Expire(context.Context, string, time.Duration) *redis.BoolCmd {
	return redis.NewBoolResult(true, nil)
}
func (s *intentCache) SetNX(context.Context, string, interface{}, time.Duration) *redis.BoolCmd {
	return redis.NewBoolResult(!s.locked, nil)
}
func (s *intentCache) Eval(context.Context, string, []string, ...interface{}) *redis.Cmd {
	return redis.NewCmdResult(1, nil)
}

type intentInterview struct {
	usecase.InterviewGateway
	creates, starts   int
	level, role, user string
	tech              []string
}

func (s *intentInterview) CreateInterview(_ context.Context, user, title, description, scheduled, level, role, plan string, tech []string, taskMode ...string) (*pb.CreateInterviewResponse, error) {
	s.creates++
	s.level = level
	s.role = role
	s.user = user
	s.tech = tech
	return &pb.CreateInterviewResponse{Response: &pb.Response{Success: true}, Interview: &pb.Interview{Id: "training-1"}}, nil
}
func (s *intentInterview) StartSession(_ context.Context, id, user string) (*pb.StartSessionResponse, error) {
	s.starts++
	return &pb.StartSessionResponse{Response: &pb.Response{Success: true}, Interview: &pb.Interview{Id: id, UserId: user, Level: s.level, Specialization: s.role, Technologies: s.tech}}, nil
}
func TestTrainingIntentRegistrationHandoff(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cache := &intentCache{values: map[string]string{}}
	upstream := &intentInterview{}
	api := NewTrainingIntentAPI(cache, upstream, nil)
	router := gin.New()
	router.POST("/intents", api.Save)
	router.POST("/intents/:id/resume", func(c *gin.Context) {
		if u := c.GetHeader("X-Test-User"); u != "" {
			c.Set("auth_user", middleware.User{ID: u, Role: "user"})
		}
	}, api.Resume)
	call := func(path, body, user string) *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("POST", path, strings.NewReader(body))
		r.Header.Set("Content-Type", "application/json")
		r.Header.Set("X-Test-User", user)
		router.ServeHTTP(w, r)
		return w
	}
	invalid := call("/intents", `{"specialization":"invalid","level":"middle","technologies":["React"]}`, "")
	if invalid.Code != 400 {
		t.Fatal(invalid.Body)
	}
	saved := call("/intents", `{"specialization":"frontend","level":"middle","technologies":["React"],"user_id":"intruder","interview_id":"foreign"}`, "")
	if saved.Code != 201 || cache.ttl != intentTTL {
		t.Fatalf("save %d %s ttl %s", saved.Code, saved.Body, cache.ttl)
	}
	var token struct {
		ID string `json:"id"`
	}
	json.Unmarshal(saved.Body.Bytes(), &token)
	path := "/intents/" + token.ID + "/resume"
	if call(path, "{}", "").Code != 401 {
		t.Fatal("anonymous resume allowed")
	}
	if call(path, "{}", "newly-verified-user").Code != 200 {
		t.Fatal("verified user cannot resume")
	}
	if upstream.user != "newly-verified-user" || upstream.role != "frontend" || upstream.level != "middle" || upstream.tech[0] != "React" {
		t.Fatalf("parameters lost: %+v", upstream)
	}
	if call(path, "{}", "newly-verified-user").Code != 200 || upstream.creates != 1 {
		t.Fatal("retry created duplicate")
	}
	if call(path, "{}", "other-user").Code != 404 {
		t.Fatal("foreign user can resume")
	}
	cache.locked = true
	if call(path, "{}", "newly-verified-user").Code != 409 {
		t.Fatal("concurrent resume allowed")
	}
	cache.locked = false
	delete(cache.values, "landing:intent:"+token.ID)
	if call(path, "{}", "newly-verified-user").Code != 410 {
		t.Fatal("expired intent resumed")
	}
}
