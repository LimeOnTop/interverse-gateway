package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port                 string
	UserServiceURL       string
	InterviewServiceURL  string
	CandidateServiceURL  string
	ReportServiceURL     string
	TechnologyServiceURL string
	QuestionServiceURL   string
	JWTSecret            string
}

func Load() *Config {
	// Load .env file if exists
	godotenv.Load()

	return &Config{
		Port:                 getEnv("PORT", "8080"),
		UserServiceURL:       getEnv("USER_SERVICE_URL", "user-service:50051"),
		InterviewServiceURL:  getEnv("INTERVIEW_SERVICE_URL", "interview-service:50052"),
		CandidateServiceURL:  getEnv("CANDIDATE_SERVICE_URL", "candidate-service:50053"),
		ReportServiceURL:     getEnv("REPORT_SERVICE_URL", "report-service:50054"),
		TechnologyServiceURL: getEnv("TECHNOLOGY_SERVICE_URL", "technology-service:50055"),
		QuestionServiceURL:   getEnv("QUESTION_SERVICE_URL", "question-service:50056"),
		JWTSecret:            getEnv("JWT_SECRET", "your-secret-key"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

