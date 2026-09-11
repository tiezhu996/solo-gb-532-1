package config

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"sonar-survey-coverage-planner/backend/internal/constants"
	"sonar-survey-coverage-planner/backend/internal/model"
)

type Config struct {
	Port        string
	DBDriver    string
	DBDSN       string
	JWTSecret   string
	AutoMigrate bool
	SeedData    bool
	LogLevel    slog.Level
}

func Load() (Config, error) {
	config := Config{
		Port:        env("PORT", "8080"),
		DBDriver:    strings.ToLower(env("DB_DRIVER", "postgres")),
		DBDSN:       os.Getenv("DB_DSN"),
		JWTSecret:   os.Getenv("JWT_SECRET"),
		AutoMigrate: envBool("DB_AUTO_MIGRATE", true),
		SeedData:    envBool("SEED_DATA", true),
		LogLevel:    slog.LevelInfo,
	}
	if len(config.JWTSecret) < 24 {
		return Config{}, fmt.Errorf("JWT_SECRET must contain at least 24 characters")
	}
	if config.DBDriver != "postgres" && config.DBDriver != "sqlite" {
		return Config{}, fmt.Errorf("unsupported DB_DRIVER %q", config.DBDriver)
	}
	if config.DBDSN == "" && config.DBDriver == "postgres" {
		config.DBDSN = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=Asia/Shanghai",
			env("DB_HOST", "127.0.0.1"), env("DB_INTERNAL_PORT", "5432"), env("DB_USER", "sonar_app"),
			env("DB_PASSWORD", "change_this_database_password"), env("DB_NAME", "sonar_coverage"), env("DB_SSLMODE", "disable"))
	}
	if config.DBDSN == "" {
		config.DBDSN = "file:sonar-local.db?_busy_timeout=5000&_foreign_keys=on"
	}
	if strings.EqualFold(os.Getenv("LOG_LEVEL"), "debug") {
		config.LogLevel = slog.LevelDebug
	}
	return config, nil
}

func OpenDatabase(config Config) (*gorm.DB, error) {
	var dialector gorm.Dialector
	if config.DBDriver == "sqlite" {
		dialector = sqlite.Open(config.DBDSN)
	} else {
		dialector = postgres.Open(config.DBDSN)
	}
	db, err := gorm.Open(dialector, &gorm.Config{Logger: logger.Default.LogMode(logger.Warn), TranslateError: true})
	if err != nil {
		return nil, fmt.Errorf("open %s database: %w", config.DBDriver, err)
	}
	if config.AutoMigrate {
		if err := db.AutoMigrate(&model.User{}, &model.SurveyArea{}, &model.TransectPlan{}, &model.SonarRun{}, &model.CoverageGap{}, &model.AuditEvent{}); err != nil {
			return nil, fmt.Errorf("migrate database: %w", err)
		}
	}
	if config.SeedData {
		if err := seed(db); err != nil {
			return nil, fmt.Errorf("seed database: %w", err)
		}
	}
	return db, nil
}

func seed(db *gorm.DB) error {
	accounts := []struct{ Username, Display, Role string }{
		{"admin", "系统管理员", constants.RoleAdmin},
		{"planner", "测线规划员", constants.RoleSurveyPlanner},
		{"processor", "数据处理员", constants.RoleDataProcessor},
		{"reviewer", "覆盖复核员", constants.RoleReviewer},
		{"auditor", "审计观察员", constants.RoleAuditor},
	}
	for _, account := range accounts {
		hash, err := bcrypt.GenerateFromPassword([]byte("Sonar2026!"), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("hash seed password: %w", err)
		}
		user := model.User{Username: account.Username, DisplayName: account.Display, Role: account.Role, PasswordHash: string(hash), Active: true}
		if err := db.Where("username = ?", user.Username).FirstOrCreate(&user).Error; err != nil {
			return err
		}
	}
	var count int64
	if err := db.Model(&model.SurveyArea{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	boundary := []byte(`{"type":"Feature","properties":{"name":"示范测区边界"},"geometry":{"type":"Polygon","coordinates":[[[0,0],[1000,0],[1000,600],[0,600],[0,0]]]}}`)
	area := model.SurveyArea{AreaCode: "DEMO-A01", Name: "外海校准测区", BoundaryGeoJSON: boundary, TargetResolutionM: 20, CoordinateSystem: "EPSG:32650", DefaultSwathM: 180, OwnerTeam: "东海测绘一组", Status: constants.AreaActive, Version: 1}
	if err := db.Create(&area).Error; err != nil {
		return err
	}
	var planner model.User
	if err := db.Where("username = ?", "planner").First(&planner).Error; err != nil {
		return err
	}
	lines := []byte(`{"type":"Feature","properties":{"name":"基准平行测线"},"geometry":{"type":"MultiLineString","coordinates":[[[40,100],[960,100]],[[40,300],[960,300]],[[40,500],[960,500]]]}}`)
	plan := model.TransectPlan{SurveyAreaID: area.ID, Name: "基准东西向测线", LineGeoJSON: lines, PlannedHeading: 90, PlannedSwathM: 180, LineSpacingM: 200, PlanState: constants.PlanLocked, Version: 1, CreatedBy: planner.ID}
	if err := db.Create(&plan).Error; err != nil {
		return err
	}
	var processor model.User
	if err := db.Where("username = ?", "processor").First(&processor).Error; err != nil {
		return err
	}
	now := time.Now().UTC().Add(-2 * time.Hour)
	tracks := []byte(`{"type":"Feature","properties":{"source":"demo"},"geometry":{"type":"MultiLineString","coordinates":[[[30,100],[970,100]],[[30,300],[970,300]],[[30,500],[970,500]]]}}`)
	run := model.SonarRun{TransectPlanID: plan.ID, RunCode: "RUN-DEMO-001", TrackGeoJSON: tracks, ActualSwathM: 170, StartedAt: now, EndedAt: now.Add(70 * time.Minute), NavigationQuality: constants.NavGood, RunState: string(constants.RunProcessed), SourceChecksum: "2d447d9407dc470d8c728ea8ed7f5f138d03dddcc6577bfd730b1cb63867a57b", ImportedBy: processor.ID, Version: 1}
	return db.Create(&run).Error
}

func env(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func envBool(key string, fallback bool) bool {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}
	return parsed
}
