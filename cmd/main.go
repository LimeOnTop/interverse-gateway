package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/LimeOnTop/interverse-gateway/cmd/config"
	"github.com/LimeOnTop/interverse-gateway/internal/api"
	"github.com/LimeOnTop/interverse-gateway/internal/clients"
	"github.com/LimeOnTop/interverse-gateway/internal/middleware"
)

func main() {
	cfg := config.Load()

	authClient := clients.NewAuthClient(cfg.AuthServiceURL)
	profileClient := clients.NewProfileClient(cfg.ProfileServiceURL)
	interviewClient := clients.NewInterviewClient(cfg.InterviewServiceURL)
	candidateClient := clients.NewCandidateClient(cfg.CandidateServiceURL)
	reportClient := clients.NewReportClient(cfg.ReportServiceURL)
	technologyClient := clients.NewTechnologyClient(cfg.TechnologyServiceURL)
	questionClient := clients.NewQuestionClient(cfg.QuestionServiceURL)

	authAPI := api.NewAuthAPI(authClient)
	profileAPI := api.NewProfileAPI(profileClient)
	interviewAPI := api.NewInterviewAPI(interviewClient)
	candidateAPI := api.NewCandidateAPI(candidateClient)
	reportAPI := api.NewReportAPI(reportClient)
	technologyAPI := api.NewTechnologyAPI(technologyClient)
	questionAPI := api.NewQuestionAPI(questionClient)

	router := gin.Default()
	router.Use(middleware.CORS())
	router.Use(middleware.Logger())

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	v1 := router.Group("/api/v1")
	{
		auth := v1.Group("/auth")
		{
			auth.POST("/register", authAPI.Register)
			auth.POST("/login", authAPI.Login)
			auth.POST("/refresh", authAPI.RefreshToken)
			auth.POST("/logout", middleware.AuthRequired(authClient), authAPI.Logout)
			auth.GET("/me", middleware.AuthRequired(authClient), authAPI.GetMe)
		}

		users := v1.Group("/users")
		users.Use(middleware.AuthRequired(authClient))
		{
			users.GET("/:id", authAPI.GetUser)
			users.PUT("/:id", authAPI.UpdateUser)
			users.DELETE("/:id", authAPI.DeleteUser)
			users.GET("/:id/profile", profileAPI.GetProfile)
			users.PUT("/:id/profile", profileAPI.UpdateProfile)
		}

		interviews := v1.Group("/interviews")
		interviews.Use(middleware.AuthRequired(authClient))
		{
			interviews.POST("/", interviewAPI.CreateInterview)
			interviews.GET("/", interviewAPI.GetInterviews)
			interviews.GET("/scheduled", interviewAPI.GetScheduledInterviews)
			interviews.POST("/generate-questions", interviewAPI.StartSessionFromBody)
			interviews.GET("/:id", interviewAPI.GetInterview)
			interviews.PUT("/:id", interviewAPI.UpdateInterview)
			interviews.DELETE("/:id", interviewAPI.DeleteInterview)
			interviews.POST("/:id/start", interviewAPI.StartSession)
			interviews.GET("/:id/session", interviewAPI.GetSessionContent)
			interviews.POST("/:id/questions", interviewAPI.GenerateQuestions)
		}

		candidates := v1.Group("/candidates")
		candidates.Use(middleware.AuthRequired(authClient))
		{
			candidates.POST("/", candidateAPI.CreateCandidate)
			candidates.GET("/", candidateAPI.GetCandidates)
			candidates.GET("/:id", candidateAPI.GetCandidate)
			candidates.PUT("/:id", candidateAPI.UpdateCandidate)
			candidates.DELETE("/:id", candidateAPI.DeleteCandidate)
			candidates.GET("/search", candidateAPI.SearchCandidates)
		}

		reports := v1.Group("/reports")
		reports.Use(middleware.AuthRequired(authClient))
		{
			reports.POST("/", reportAPI.CreateReport)
			reports.GET("/", reportAPI.GetReports)
			reports.GET("/:id", reportAPI.GetReport)
			reports.PUT("/:id", reportAPI.UpdateReport)
			reports.DELETE("/:id", reportAPI.DeleteReport)
		}

		technologies := v1.Group("/technologies")
		technologies.Use(middleware.AuthRequired(authClient))
		{
			technologies.GET("/", technologyAPI.GetTechnologies)
			technologies.GET("/:id", technologyAPI.GetTechnology)
			technologies.POST("/", technologyAPI.CreateTechnology)
			technologies.PUT("/:id", technologyAPI.UpdateTechnology)
			technologies.DELETE("/:id", technologyAPI.DeleteTechnology)
			technologies.GET("/search", technologyAPI.SearchTechnologies)
		}

		questions := v1.Group("/questions")
		questions.Use(middleware.AuthRequired(authClient))
		{
			questions.GET("/", questionAPI.GetQuestions)
			questions.GET("/:id", questionAPI.GetQuestion)
			questions.POST("/", questionAPI.CreateQuestion)
			questions.PUT("/:id", questionAPI.UpdateQuestion)
			questions.DELETE("/:id", questionAPI.DeleteQuestion)
			questions.GET("/search", questionAPI.SearchQuestions)
			questions.GET("/technology/:technology", questionAPI.GetQuestionsByTechnology)
		}
	}

	log.Printf("API Gateway starting on port %s", cfg.Port)
	router.Run(":" + cfg.Port)
}
