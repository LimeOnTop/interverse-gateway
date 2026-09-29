package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Port                 string
	DevMode              bool
	AuthServiceURL       string
	ProfileServiceURL    string
	InterviewServiceURL  string
	ReportServiceURL     string
	TechnologyServiceURL string
	QuestionServiceURL   string
	VacancyServiceURL    string
	JWTSecret            string
	RedisAccessAddr      string
	RedisAccessDB        int
	AdminUsername        string
	HHUserAgent          string
}

func Load() *Config {
	godotenv.Load()

	return &Config{
		Port:                 getEnv("PORT", "8080"),
		DevMode:              getEnvBool("DEV_MODE", false),
		AuthServiceURL:       getEnv("AUTH_SERVICE_URL", "auth-service:50051"),
		ProfileServiceURL:    getEnv("PROFILE_SERVICE_URL", "user-service:50057"),
		InterviewServiceURL:  getEnv("INTERVIEW_SERVICE_URL", "interview-service:50052"),
		ReportServiceURL:     getEnv("REPORT_SERVICE_URL", "report-service:50054"),
		TechnologyServiceURL: getEnv("TECHNOLOGY_SERVICE_URL", "technology-service:50055"),
		QuestionServiceURL:   getEnv("QUESTION_SERVICE_URL", "question-service:50056"),
		VacancyServiceURL:    getEnv("VACANCY_SERVICE_URL", "vacancy-service:50058"),
		JWTSecret:            getEnv("JWT_SECRET", "your-secret-key"),
		RedisAccessAddr:      getEnv("REDIS_ACCESS_ADDR", "localhost:6379"),
		RedisAccessDB:        getEnvInt("REDIS_ACCESS_DB", 1),
		AdminUsername:        getEnv("ADMIN_USERNAME", ""),
		HHUserAgent:          getEnv("HH_USER_AGENT", "InterVerse/1.0 (dev@interverse.local)"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return defaultValue
	}
	return parsed
}
