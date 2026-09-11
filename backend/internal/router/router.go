package router

import (
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"sonar-survey-coverage-planner/backend/internal/constants"
	"sonar-survey-coverage-planner/backend/internal/handler"
	"sonar-survey-coverage-planner/backend/internal/middleware"
	"sonar-survey-coverage-planner/backend/internal/service"
)

type Handlers struct {
	Auth     *handler.AuthHandler
	Area     *handler.SurveyAreaHandler
	Plan     *handler.TransectPlanHandler
	Run      *handler.SonarRunHandler
	Coverage *handler.CoverageGapHandler
	Audit    *handler.AuditHandler
}

func New(log *slog.Logger, auth *service.AuthService, handlers Handlers) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	engine.Use(middleware.RequestID(), middleware.Recovery(log), middleware.ErrorHandler(log), cors())
	engine.GET("/healthz", handlers.Auth.Health)
	engine.GET("/readyz", handlers.Auth.Ready)

	loginLimit := middleware.NewRateLimiter(12, time.Minute)
	importLimit := middleware.NewRateLimiter(20, time.Minute)
	coverageLimit := middleware.NewRateLimiter(10, time.Minute)
	apiV1 := engine.Group("/api/v1")
	apiV1.POST("/auth/login", loginLimit.Middleware("login"), handlers.Auth.Login)

	protected := apiV1.Group("")
	protected.Use(middleware.Auth(auth), middleware.AuditContextMiddleware())
	protected.GET("/auth/me", handlers.Auth.Me)

	protected.GET("/areas", handlers.Area.List)
	protected.GET("/areas/:id", handlers.Area.Get)
	protected.POST("/areas", middleware.RBAC(constants.RoleAdmin, constants.RoleSurveyPlanner), handlers.Area.Create)
	protected.PUT("/areas/:id", middleware.RBAC(constants.RoleAdmin, constants.RoleSurveyPlanner), handlers.Area.Update)

	protected.GET("/plans", handlers.Plan.List)
	protected.GET("/plans/:id", handlers.Plan.Get)
	protected.POST("/plans", middleware.RBAC(constants.RoleAdmin, constants.RoleSurveyPlanner), handlers.Plan.Create)
	protected.POST("/plans/generate", middleware.RBAC(constants.RoleAdmin, constants.RoleSurveyPlanner), handlers.Plan.Generate)
	protected.PUT("/plans/:id", middleware.RBAC(constants.RoleAdmin, constants.RoleSurveyPlanner), handlers.Plan.Update)
	protected.POST("/plans/:id/transition", middleware.RBAC(constants.RoleAdmin, constants.RoleSurveyPlanner), handlers.Plan.Lock)
	protected.POST("/plans/:id/copy", middleware.RBAC(constants.RoleAdmin, constants.RoleSurveyPlanner), handlers.Plan.Copy)

	protected.GET("/runs", handlers.Run.List)
	protected.GET("/runs/:id", handlers.Run.Get)
	protected.GET("/runs/:id/quality", handlers.Run.Quality)
	protected.POST("/runs/import", importLimit.Middleware("run-import"), middleware.RBAC(constants.RoleAdmin, constants.RoleDataProcessor), handlers.Run.Import)
	protected.POST("/runs/:id/transition", middleware.RBAC(constants.RoleAdmin, constants.RoleDataProcessor), handlers.Run.Transition)

	protected.GET("/coverage-gaps", handlers.Coverage.List)
	protected.GET("/coverage-gaps/:id", handlers.Coverage.Get)
	protected.POST("/coverage-gaps/detect", coverageLimit.Middleware("coverage-detect"), middleware.RBAC(constants.RoleAdmin, constants.RoleDataProcessor), handlers.Coverage.Detect)
	protected.POST("/coverage-gaps/:id/transition", middleware.RBAC(constants.RoleReviewer), handlers.Coverage.Transition)

	protected.GET("/audits", middleware.RBAC(constants.RoleAdmin, constants.RoleReviewer, constants.RoleAuditor), handlers.Audit.List)
	return engine
}

func cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := strings.TrimSpace(c.GetHeader("Origin"))
		if origin == "http://localhost:5173" || origin == "http://127.0.0.1:5173" {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Vary", "Origin")
			c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type, X-Request-ID, Idempotency-Key")
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, OPTIONS")
		}
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
