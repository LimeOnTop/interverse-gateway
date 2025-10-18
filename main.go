package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/inter-verse/services/api-gateway/internal/config"
	"github.com/inter-verse/services/api-gateway/internal/handler"
	"github.com/inter-verse/services/api-gateway/internal/middleware"
	"github.com/inter-verse/services/api-gateway/internal/service"
)

func main() {
	cfg := config.Load()

	// Services will establish gRPC connections internally

	// Initialize services
	userService := service.NewUserService(cfg.UserServiceURL)
	interviewService := service.NewInterviewService(cfg.InterviewServiceURL)
	candidateService := service.NewCandidateService(cfg.CandidateServiceURL)
	reportService := service.NewReportService(cfg.ReportServiceURL)
	technologyService := service.NewTechnologyService(cfg.TechnologyServiceURL)
	questionService := service.NewQuestionService(cfg.QuestionServiceURL)

	// Initialize handlers
	userHandler := handler.NewUserHandler(userService)
	interviewHandler := handler.NewInterviewHandler(interviewService)
	candidateHandler := handler.NewCandidateHandler(candidateService)
	reportHandler := handler.NewReportHandler(reportService)
	technologyHandler := handler.NewTechnologyHandler(technologyService)
	questionHandler := handler.NewQuestionHandler(questionService)

	// Initialize Gin router
	router := gin.Default()

	// Middleware
	router.Use(middleware.CORS())
	router.Use(middleware.Logger())

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// API routes
	api := router.Group("/api/v1")
	{
		// Auth routes
		auth := api.Group("/auth")
		{
			auth.POST("/register", userHandler.Register)
			auth.POST("/login", userHandler.Login)
			auth.POST("/refresh", userHandler.RefreshToken)
			auth.POST("/logout", middleware.AuthRequired(userService), userHandler.Logout)
			auth.GET("/me", middleware.AuthRequired(userService), userHandler.GetMe)
		}

		// User routes
		users := api.Group("/users")
		users.Use(middleware.AuthRequired(userService))
		{
			users.GET("/:id", userHandler.GetUser)
			users.PUT("/:id", userHandler.UpdateUser)
			users.DELETE("/:id", userHandler.DeleteUser)
		}

		// Interview routes
		interviews := api.Group("/interviews")
		interviews.Use(middleware.AuthRequired(userService))
		{
			interviews.POST("/", interviewHandler.CreateInterview)
			interviews.GET("/", interviewHandler.GetInterviews)
			interviews.GET("/:id", interviewHandler.GetInterview)
			interviews.PUT("/:id", interviewHandler.UpdateInterview)
			interviews.DELETE("/:id", interviewHandler.DeleteInterview)
			interviews.GET("/scheduled", interviewHandler.GetScheduledInterviews)
			interviews.POST("/:id/questions", interviewHandler.GenerateQuestions)
		}

		// Candidate routes
		candidates := api.Group("/candidates")
		candidates.Use(middleware.AuthRequired(userService))
		{
			candidates.POST("/", candidateHandler.CreateCandidate)
			candidates.GET("/", candidateHandler.GetCandidates)
			candidates.GET("/:id", candidateHandler.GetCandidate)
			candidates.PUT("/:id", candidateHandler.UpdateCandidate)
			candidates.DELETE("/:id", candidateHandler.DeleteCandidate)
			candidates.GET("/search", candidateHandler.SearchCandidates)
		}

		// Report routes
		reports := api.Group("/reports")
		reports.Use(middleware.AuthRequired(userService))
		{
			reports.POST("/", reportHandler.CreateReport)
			reports.GET("/", reportHandler.GetReports)
			reports.GET("/:id", reportHandler.GetReport)
			reports.PUT("/:id", reportHandler.UpdateReport)
			reports.DELETE("/:id", reportHandler.DeleteReport)
		}

		// Technology routes
		technologies := api.Group("/technologies")
		technologies.Use(middleware.AuthRequired(userService))
		{
			technologies.GET("/", technologyHandler.GetTechnologies)
			technologies.GET("/:id", technologyHandler.GetTechnology)
			technologies.POST("/", technologyHandler.CreateTechnology)
			technologies.PUT("/:id", technologyHandler.UpdateTechnology)
			technologies.DELETE("/:id", technologyHandler.DeleteTechnology)
			technologies.GET("/search", technologyHandler.SearchTechnologies)
		}

		// Question routes
		questions := api.Group("/questions")
		questions.Use(middleware.AuthRequired(userService))
		{
			questions.GET("/", questionHandler.GetQuestions)
			questions.GET("/:id", questionHandler.GetQuestion)
			questions.POST("/", questionHandler.CreateQuestion)
			questions.PUT("/:id", questionHandler.UpdateQuestion)
			questions.DELETE("/:id", questionHandler.DeleteQuestion)
			questions.GET("/search", questionHandler.SearchQuestions)
			questions.GET("/technology/:technology", questionHandler.GetQuestionsByTechnology)
		}
	}

	log.Printf("API Gateway starting on port %s", cfg.Port)
	router.Run(":" + cfg.Port)
}
