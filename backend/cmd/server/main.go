package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"sonar-survey-coverage-planner/backend/internal/config"
	"sonar-survey-coverage-planner/backend/internal/handler"
	"sonar-survey-coverage-planner/backend/internal/repository"
	"sonar-survey-coverage-planner/backend/internal/router"
	"sonar-survey-coverage-planner/backend/internal/service"
)

func main() {
	configuration, err := config.Load()
	if err != nil {
		slog.Error("invalid configuration", "error", err)
		os.Exit(1)
	}
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: configuration.LogLevel}))
	db, err := config.OpenDatabase(configuration)
	if err != nil {
		log.Error("database initialization failed", "error", err)
		os.Exit(1)
	}

	supportRepository := repository.NewSupportRepository(db)
	areaRepository := repository.NewSurveyAreaRepository(db)
	planRepository := repository.NewTransectPlanRepository(db)
	runRepository := repository.NewSonarRunRepository(db)
	coverageRepository := repository.NewCoverageGapRepository(db)

	auditService := service.NewAuditService(supportRepository)
	authService := service.NewAuthService(supportRepository, configuration.JWTSecret)
	areaService := service.NewSurveyAreaService(areaRepository, auditService)
	planService := service.NewTransectPlanService(planRepository, areaRepository, auditService)
	runService := service.NewSonarRunService(runRepository, planRepository, auditService)
	coverageService := service.NewCoverageGapService(coverageRepository, areaRepository, runRepository, auditService)

	handlers := router.Handlers{
		Auth: handler.NewAuthHandler(authService, auditService), Area: handler.NewSurveyAreaHandler(areaService),
		Plan: handler.NewTransectPlanHandler(planService), Run: handler.NewSonarRunHandler(runService),
		Coverage: handler.NewCoverageGapHandler(coverageService), Audit: handler.NewAuditHandler(auditService),
	}
	engine := router.New(log, authService, handlers)
	server := &http.Server{Addr: ":" + configuration.Port, Handler: engine, ReadHeaderTimeout: 8 * time.Second, ReadTimeout: 20 * time.Second, WriteTimeout: 35 * time.Second, IdleTimeout: 60 * time.Second}

	go func() {
		log.Info("sonar coverage service listening", "port", configuration.Port, "database_driver", configuration.DBDriver)
		if serveErr := server.ListenAndServe(); serveErr != nil && !errors.Is(serveErr, http.ErrServerClosed) {
			log.Error("http server failed", "error", serveErr)
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Error("graceful shutdown failed", "error", err)
		os.Exit(1)
	}
	sqlDB, err := db.DB()
	if err == nil {
		_ = sqlDB.Close()
	}
	log.Info("service stopped")
}
