package config

import (
	"testing"

	"sonar-survey-coverage-planner/backend/internal/model"
)

func TestOpenSQLiteDatabaseSeedsOperationalFixture(t *testing.T) {
	configuration := Config{DBDriver: "sqlite", DBDSN: "file:config-test?mode=memory&cache=shared", JWTSecret: "test-secret-with-more-than-thirty-two-characters", AutoMigrate: true, SeedData: true}
	db, err := OpenDatabase(configuration)
	if err != nil {
		t.Fatalf("open sqlite database: %v", err)
	}
	var users, areas, plans, runs int64
	if err := db.Model(&model.User{}).Count(&users).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Model(&model.SurveyArea{}).Count(&areas).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Model(&model.TransectPlan{}).Count(&plans).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Model(&model.SonarRun{}).Count(&runs).Error; err != nil {
		t.Fatal(err)
	}
	if users != 5 || areas != 1 || plans != 1 || runs != 1 {
		t.Fatalf("seed counts users=%d areas=%d plans=%d runs=%d", users, areas, plans, runs)
	}
}
