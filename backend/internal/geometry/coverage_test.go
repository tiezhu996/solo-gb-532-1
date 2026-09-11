package geometry

import (
	"math"
	"testing"

	"github.com/paulmach/orb"
)

func TestCalculateCoverageFixedFixture(t *testing.T) {
	boundary := orb.Polygon{orb.Ring{{0, 0}, {100, 0}, {100, 100}, {0, 100}, {0, 0}}}
	tracks := []Track{{RunID: 1, SwathM: 50, Lines: []orb.LineString{{{0, 25}, {100, 25}}, {{0, 75}, {100, 75}}}}}
	result, err := CalculateCoverage(boundary, tracks, 5)
	if err != nil {
		t.Fatalf("calculate coverage: %v", err)
	}
	if math.Abs(result.AreaSquareM-10000) > 0.001 {
		t.Fatalf("area = %.2f, want 10000", result.AreaSquareM)
	}
	if result.CoverageRatio < 0.95 {
		t.Fatalf("coverage ratio = %.3f, want near complete", result.CoverageRatio)
	}
	if result.SampleCells != 400 {
		t.Fatalf("sample cells = %d, want 400", result.SampleCells)
	}
	if len(result.GapGeoJSON) == 0 || len(result.RecommendedGeoJSON) == 0 {
		t.Fatal("expected explainable gap artifacts")
	}
}

func TestParseAndCoordinateGuard(t *testing.T) {
	if err := ValidateProjectedCRS("EPSG:4326"); err == nil {
		t.Fatal("geographic CRS must be rejected")
	}
	if err := ValidateProjectedCRS("EPSG:32650"); err != nil {
		t.Fatalf("projected CRS rejected: %v", err)
	}
	polygonJSON := []byte(`{"type":"Feature","properties":{},"geometry":{"type":"Polygon","coordinates":[[[0,0],[20,0],[20,10],[0,10],[0,0]]]}}`)
	polygon, err := ParsePolygon(polygonJSON)
	if err != nil {
		t.Fatalf("parse polygon: %v", err)
	}
	if PolygonArea(polygon) != 200 {
		t.Fatalf("area = %.1f, want 200", PolygonArea(polygon))
	}
	if _, err := ParseLines(polygonJSON); err == nil {
		t.Fatal("polygon must not parse as track lines")
	}
}

func TestStableInputHashIgnoresRunOrder(t *testing.T) {
	first := StableInputHash(4, "EPSG:32650", "grid-v1", []string{"beta", "alpha"}, 10)
	second := StableInputHash(4, "EPSG:32650", "grid-v1", []string{"alpha", "beta"}, 10)
	changed := StableInputHash(4, "EPSG:32650", "grid-v2", []string{"alpha", "beta"}, 10)
	if first != second {
		t.Fatal("hash should be stable across run order")
	}
	if first == changed {
		t.Fatal("algorithm version must affect input hash")
	}
}
