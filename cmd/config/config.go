package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port                 string
	AuthServiceURL       string
	ProfileServiceURL    string
	InterviewServiceURL  string
	ReportServiceURL     string
	TechnologyServiceURL string
	QuestionServiceURL   string
	JWTSecret            string
	AdminUsername        string
	HHClientID           string
	HHClientSecret       string
	HHRedirectURI        string
	HHUserAgent          string
}

func Load() *Config {
	godotenv.Load()

	return &Config{
		Port:                 getEnv("PORT", "8080"),
		AuthServiceURL:       getEnv("AUTH_SERVICE_URL", "auth-service:50051"),
		ProfileServiceURL:    getEnv("PROFILE_SERVICE_URL", "user-service:50057"),
		InterviewServiceURL:  getEnv("INTERVIEW_SERVICE_URL", "interview-service:50052"),
		ReportServiceURL:     getEnv("REPORT_SERVICE_URL", "report-service:50054"),
		TechnologyServiceURL: getEnv("TECHNOLOGY_SERVICE_URL", "technology-service:50055"),
		QuestionServiceURL:   getEnv("QUESTION_SERVICE_URL", "question-service:50056"),
		JWTSecret:            getEnv("JWT_SECRET", "your-secret-key"),
		AdminUsername:        getEnv("ADMIN_USERNAME", ""),
		HHClientID:           getEnv("HH_CLIENT_ID", ""),
		HHClientSecret:       getEnv("HH_CLIENT_SECRET", ""),
		HHRedirectURI:        getEnv("HH_REDIRECT_URI", "http://localhost:3000/profile"),
		HHUserAgent:          getEnv("HH_USER_AGENT", "InterVerse/1.0 (dev@interverse.local)"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
