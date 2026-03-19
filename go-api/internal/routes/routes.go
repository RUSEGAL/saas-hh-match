package routes

import (
	"time"

	"go-api/internal/cache"
	handler404 "go-api/internal/handlers"
	handlers_admin "go-api/internal/handlers/admin"
	handlers_ai "go-api/internal/handlers/ai"
	handlers_auth "go-api/internal/handlers/auth"
	payments_handlers "go-api/internal/handlers/payments"
	resumes_handlers "go-api/internal/handlers/resumes"
	handlers_user "go-api/internal/handlers/user"
	handlers_vacancies "go-api/internal/handlers/vacancies"
	"go-api/internal/middleware/auth"
	"go-api/internal/middleware/ratelimit"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine) {
	r.NoRoute(handler404.Ret404)

	authRateLimiter := ratelimit.NewRateLimiter(cache.Client, 10, time.Minute, "auth")
	apiRateLimiter := ratelimit.NewRateLimiter(cache.Client, 100, time.Minute, "api")

	authGroup := r.Group("/auth")
	authGroup.Use(authRateLimiter.Limit())
	{
		authGroup.POST("/getToken", handlers_auth.GetToken)
	}

	ai := r.Group("/ai")
	{
		ai.POST("/webhook/analyze", handlers_ai.WebhookAnalyzeResume)
		ai.POST("/webhook/matches", handlers_ai.WebhookVacancyMatches)
	}

	internal := r.Group("/internal")
	{
		internal.GET("/resumes/:id", resumes_handlers.GetResumeInternal)
	}

	// TODO: Add ЮKassa webhook when ready
	// paymentsWebhook := r.Group("/payments")
	// {
	// 	paymentsWebhook.POST("/webhook/yookassa", payments_handlers.WebhookYookassa)
	// }

	api := r.Group("/api")
	api.Use(apiRateLimiter.Limit())
	{
		payments := api.Group("/payments")
		{
			payments.POST("/", payments_handlers.CreatePayment)
			payments.PATCH("/:id", payments_handlers.UpdatePayment)
			payments.GET("/me", payments_handlers.GetMyPayments)
		}

		resumes := api.Group("/resumes")
		{
			resumes.POST("/", resumes_handlers.AddResume)
			resumes.PATCH("/:id", resumes_handlers.UpdateResume)
			resumes.DELETE("/:id", resumes_handlers.DeleteResume)
			resumes.GET("/", resumes_handlers.GetUserResumes)
			resumes.GET("/me", resumes_handlers.GetMyResumes)
			resumes.GET("/:id", resumes_handlers.GetResumeByID)
		}

		vacancies := api.Group("/vacancies")
		{
			vacancies.POST("/match", handlers_vacancies.MatchVacancies)
			vacancies.GET("/matches/me", handlers_vacancies.GetUserMatches)
			vacancies.GET("/matches/:id", handlers_vacancies.GetMatchResults)
			vacancies.POST("/response", handlers_vacancies.SaveVacancyResponse)
			vacancies.POST("/view", handlers_vacancies.SaveVacancyView)
		}

		user := api.Group("/user")
		{
			user.GET("/stats", handlers_user.GetMyStats)
			user.GET("/payment-status", handlers_user.GetMyPaymentStatus)
		}

		admin := api.Group("/admin")
		admin.Use(middleware.RequireAdmin())
		{
			admin.GET("/users", handlers_admin.GetUsers)
			admin.GET("/stats", handlers_admin.GetStats)
			admin.GET("/users/:id/resumes", handlers_admin.GetUserResumes)
			admin.GET("/users/:id/payments", handlers_admin.GetUserPayments)
		}
	}
}
