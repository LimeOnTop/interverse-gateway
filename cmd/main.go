package main

import (
	"log"
	"net/http"

	"github.com/LimeOnTop/interverse-gateway/cmd/config"
	"github.com/LimeOnTop/interverse-gateway/internal/api"
	"github.com/LimeOnTop/interverse-gateway/internal/apperr"
	"github.com/LimeOnTop/interverse-gateway/internal/authjwt"
	"github.com/LimeOnTop/interverse-gateway/internal/clients"
	"github.com/LimeOnTop/interverse-gateway/internal/hh"
	"github.com/LimeOnTop/interverse-gateway/internal/middleware"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func main() {
	cfg := config.Load()
	apperr.Configure(cfg.DevMode)
	if !cfg.DevMode {
		gin.SetMode(gin.ReleaseMode)
	}

	authClient := clients.NewAuthClient(cfg.AuthServiceURL)
	profileClient := clients.NewProfileClient(cfg.ProfileServiceURL)
	interviewClient := clients.NewInterviewClient(cfg.InterviewServiceURL)
	reportClient := clients.NewReportClient(cfg.ReportServiceURL)
	technologyClient := clients.NewTechnologyClient(cfg.TechnologyServiceURL)
	questionClient := clients.NewQuestionClient(cfg.QuestionServiceURL)
	vacancyClient := clients.NewVacancyClient(cfg.VacancyServiceURL)

	jwtValidator := authjwt.NewValidator(cfg.JWTSecret)
	accessRedis := redis.NewClient(&redis.Options{
		Addr: cfg.RedisAccessAddr,
		DB:   cfg.RedisAccessDB,
	})
	defer accessRedis.Close()
	accessBlacklist := authjwt.NewAccessBlacklist(accessRedis)
	requireAuth := middleware.AuthRequired(jwtValidator, accessBlacklist)

	authAPI := api.NewAuthAPI(authClient, cfg.AdminUsername)
	profileAPI := api.NewProfileAPI(profileClient)
	hhClient := hh.NewClient(hh.Config{
		UserAgent: cfg.HHUserAgent,
	})
	hhImportAPI := api.NewHHImportAPI(hhClient)
	interviewAPI := api.NewInterviewAPI(interviewClient, authClient)
	reportAPI := api.NewReportAPI(reportClient)
	technologyAPI := api.NewTechnologyAPI(technologyClient)
	questionAPI := api.NewQuestionAPI(questionClient)
	contributionAPI := api.NewContributionAPI(questionClient)
	adminAPI := api.NewAdminAPI(questionClient)
	vacancyAPI := api.NewVacancyAPI(vacancyClient)

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
			auth.POST("/logout", requireAuth, authAPI.Logout)
			auth.GET("/me", requireAuth, authAPI.GetMe)
		}

		users := v1.Group("/users")
		users.Use(requireAuth)
		{
			users.GET("/:id", authAPI.GetUser)
			users.PUT("/:id", authAPI.UpdateUser)
			users.DELETE("/:id", authAPI.DeleteUser)
			users.GET("/:id/profile", profileAPI.GetProfile)
			users.PUT("/:id/profile", profileAPI.UpdateProfile)
			users.POST("/:id/profile/import/hh", hhImportAPI.ImportProfile)
		}

		interviews := v1.Group("/interviews")
		interviews.Use(requireAuth)
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

		reports := v1.Group("/reports")
		reports.Use(requireAuth)
		{
			reports.POST("/generate", reportAPI.GenerateReport)
			reports.POST("/:id/analyze", reportAPI.AnalyzeReport)
			reports.POST("/", reportAPI.CreateReport)
			reports.GET("/", reportAPI.GetReports)
			reports.GET("/:id", reportAPI.GetReport)
			reports.PUT("/:id", reportAPI.UpdateReport)
			reports.DELETE("/:id", reportAPI.DeleteReport)
		}

		technologies := v1.Group("/technologies")
		technologies.Use(requireAuth)
		{
			technologies.GET("/", technologyAPI.GetTechnologies)
			technologies.GET("/:id", technologyAPI.GetTechnology)
			technologies.POST("/", technologyAPI.CreateTechnology)
			technologies.PUT("/:id", technologyAPI.UpdateTechnology)
			technologies.DELETE("/:id", technologyAPI.DeleteTechnology)
			technologies.GET("/search", technologyAPI.SearchTechnologies)
		}

		questions := v1.Group("/questions")
		questions.Use(requireAuth)
		{
			questions.GET("/", questionAPI.GetQuestions)
			questions.GET("/:id", questionAPI.GetQuestion)
			questions.GET("/search", questionAPI.SearchQuestions)
			questions.GET("/technology/:technology", questionAPI.GetQuestionsByTechnology)
			questions.POST("/", middleware.AdminRequired(), questionAPI.CreateQuestion)
			questions.PUT("/:id", middleware.AdminRequired(), questionAPI.UpdateQuestion)
			questions.DELETE("/:id", middleware.AdminRequired(), questionAPI.DeleteQuestion)
		}

		contributions := v1.Group("/contributions")
		contributions.Use(requireAuth)
		{
			contributions.POST("/questions", contributionAPI.SubmitQuestion)
		}

		admin := v1.Group("/admin")
		admin.Use(requireAuth, middleware.AdminRequired())
		{
			admin.GET("/stats", adminAPI.Stats)
			admin.GET("/questions", adminAPI.ListQuestions)
			admin.GET("/moderation/questions", adminAPI.ListModerationQuestions)
			admin.POST("/moderation/questions/:id/approve", adminAPI.ApproveModerationQuestion)
			admin.POST("/moderation/questions/:id/reject", adminAPI.RejectModerationQuestion)
		}

		vacancies := v1.Group("/vacancies")
		vacancies.Use(requireAuth)
		{
			vacancies.GET("/", vacancyAPI.GetVacancies)
		}
	}

	log.Printf("API Gateway starting on port %s", cfg.Port)
	router.Run(":" + cfg.Port)
}
