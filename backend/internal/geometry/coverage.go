package geometry

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/paulmach/orb"
	"github.com/paulmach/orb/geojson"
)

var (
	ErrInvalidGeometry  = errors.New("invalid survey geometry")
	ErrGeographicCRS    = errors.New("geographic coordinates cannot be used for metre calculations")
	ErrGeometryTooLarge = errors.New("geometry resolution creates too many analysis cells")
)

type Bounds struct {
	MinX float64 `json:"min_x"`
	MinY float64 `json:"min_y"`
	MaxX float64 `json:"max_x"`
	MaxY float64 `json:"max_y"`
}

type Track struct {
	RunID  uint
	Lines  []orb.LineString
	SwathM float64
}

type CoverageResult struct {
	AreaSquareM        float64
	CoverageRatio      float64
	OverlapRatio       float64
	GapRatio           float64
	GapGeoJSON         []byte
	RecommendedGeoJSON []byte
	FilteredFragments  int
	SampleCells        int
}

func ValidateProjectedCRS(value string) error {
	normalized := strings.ToUpper(strings.TrimSpace(value))
	if normalized == "" {
		return fmt.Errorf("%w: coordinate system is required", ErrInvalidGeometry)
	}
	if strings.Contains(normalized, "4326") || strings.Contains(normalized, "WGS84") || strings.Contains(normalized, "WGS 84") {
		return ErrGeographicCRS
	}
	return nil
}

func ParsePolygon(data []byte) (orb.Polygon, error) {
	feature, err := geojson.UnmarshalFeature(data)
	if err != nil {
		return nil, fmt.Errorf("%w: parse polygon feature: %v", ErrInvalidGeometry, err)
	}
	switch geometry := feature.Geometry.(type) {
	case orb.Polygon:
		if err := validatePolygon(geometry); err != nil {
			return nil, err
		}
		return geometry, nil
	case orb.MultiPolygon:
		if len(geometry) != 1 {
			return nil, fmt.Errorf("%w: survey boundary must contain one polygon", ErrInvalidGeometry)
		}
		if err := validatePolygon(geometry[0]); err != nil {
			return nil, err
		}
		return geometry[0], nil
	default:
		return nil, fmt.Errorf("%w: expected Polygon feature", ErrInvalidGeometry)
	}
}

func ParseLines(data []byte) ([]orb.LineString, error) {
	feature, err := geojson.UnmarshalFeature(data)
	if err != nil {
		return nil, fmt.Errorf("%w: parse line feature: %v", ErrInvalidGeometry, err)
	}
	var lines []orb.LineString
	switch geometry := feature.Geometry.(type) {
	case orb.LineString:
		lines = []orb.LineString{geometry}
	case orb.MultiLineString:
		lines = []orb.LineString(geometry)
	default:
		return nil, fmt.Errorf("%w: expected LineString or MultiLineString feature", ErrInvalidGeometry)
	}
	if len(lines) == 0 {
		return nil, fmt.Errorf("%w: at least one line is required", ErrInvalidGeometry)
	}
	for _, line := range lines {
		if len(line) < 2 {
			return nil, fmt.Errorf("%w: each line needs at least two positions", ErrInvalidGeometry)
		}
		for _, point := range line {
			if !finite(point[0]) || !finite(point[1]) {
				return nil, fmt.Errorf("%w: coordinate is not finite", ErrInvalidGeometry)
			}
		}
	}
	return lines, nil
}

func PolygonArea(polygon orb.Polygon) float64 {
	if len(polygon) == 0 {
		return 0
	}
	area := math.Abs(ringArea(polygon[0]))
	for _, hole := range polygon[1:] {
		area -= math.Abs(ringArea(hole))
	}
	return math.Max(area, 0)
}

func TrackLength(lines []orb.LineString) float64 {
	total := 0.0
	for _, line := range lines {
		for index := 1; index < len(line); index++ {
			total += distance(line[index-1], line[index])
		}
	}
	return total
}

func CalculateCoverage(boundary orb.Polygon, tracks []Track, resolutionM float64) (CoverageResult, error) {
	if resolutionM <= 0 || !finite(resolutionM) {
		return CoverageResult{}, fmt.Errorf("%w: resolution must be positive", ErrInvalidGeometry)
	}
	if len(tracks) == 0 {
		return CoverageResult{}, fmt.Errorf("%w: processed tracks are required", ErrInvalidGeometry)
	}
	bounds := polygonBounds(boundary)
	width, height := bounds.MaxX-bounds.MinX, bounds.MaxY-bounds.MinY
	if width <= 0 || height <= 0 {
		return CoverageResult{}, fmt.Errorf("%w: boundary has no area", ErrInvalidGeometry)
	}
	columns := int(math.Ceil(width / resolutionM))
	rows := int(math.Ceil(height / resolutionM))
	if columns*rows > 60000 {
		return CoverageResult{}, ErrGeometryTooLarge
	}
	inside, covered, overlap := 0, 0, 0
	gapMinX, gapMinY := math.Inf(1), math.Inf(1)
	gapMaxX, gapMaxY := math.Inf(-1), math.Inf(-1)
	for row := 0; row < rows; row++ {
		for column := 0; column < columns; column++ {
			point := orb.Point{
				bounds.MinX + (float64(column)+0.5)*resolutionM,
				bounds.MinY + (float64(row)+0.5)*resolutionM,
			}
			if !polygonContains(boundary, point) {
				continue
			}
			inside++
			passes := coveragePasses(point, tracks)
			if passes > 0 {
				covered++
			}
			if passes > 1 {
				overlap++
			}
			if passes == 0 {
				gapMinX = math.Min(gapMinX, point[0]-resolutionM/2)
				gapMinY = math.Min(gapMinY, point[1]-resolutionM/2)
				gapMaxX = math.Max(gapMaxX, point[0]+resolutionM/2)
				gapMaxY = math.Max(gapMaxY, point[1]+resolutionM/2)
			}
		}
	}
	if inside == 0 {
		return CoverageResult{}, fmt.Errorf("%w: no analysis cells fall inside boundary", ErrInvalidGeometry)
	}
	coverageRatio := float64(covered) / float64(inside)
	overlapRatio := float64(overlap) / float64(inside)
	gapRatio := 1 - coverageRatio
	gapJSON, recommendationJSON, fragments, err := gapArtifacts(bounds, gapMinX, gapMinY, gapMaxX, gapMaxY, resolutionM, gapRatio)
	if err != nil {
		return CoverageResult{}, err
	}
	return CoverageResult{
		AreaSquareM: PolygonArea(boundary), CoverageRatio: coverageRatio,
		OverlapRatio: overlapRatio, GapRatio: gapRatio, GapGeoJSON: gapJSON,
		RecommendedGeoJSON: recommendationJSON, FilteredFragments: fragments,
		SampleCells: inside,
	}, nil
}

func StableInputHash(areaID uint, coordinateSystem, algorithmVersion string, runChecksums []string, resolution float64) string {
	values := append([]string(nil), runChecksums...)
	sort.Strings(values)
	payload, _ := json.Marshal(struct {
		AreaID           uint     `json:"area_id"`
		CoordinateSystem string   `json:"coordinate_system"`
		Algorithm        string   `json:"algorithm"`
		Checksums        []string `json:"checksums"`
		Resolution       float64  `json:"resolution"`
	}{areaID, coordinateSystem, algorithmVersion, values, resolution})
	sum := sha256.Sum256(payload)
	return hex.EncodeToString(sum[:])
}

func MarshalFeature(geometry orb.Geometry) ([]byte, error) {
	feature := geojson.NewFeature(geometry)
	return feature.MarshalJSON()
}

func validatePolygon(polygon orb.Polygon) error {
	if len(polygon) == 0 || len(polygon[0]) < 4 {
		return fmt.Errorf("%w: polygon exterior requires four positions", ErrInvalidGeometry)
	}
	for _, ring := range polygon {
		for _, point := range ring {
			if !finite(point[0]) || !finite(point[1]) {
				return fmt.Errorf("%w: coordinate is not finite", ErrInvalidGeometry)
			}
		}
	}
	if PolygonArea(polygon) <= 0 {
		return fmt.Errorf("%w: polygon area must be positive", ErrInvalidGeometry)
	}
	return nil
}

func polygonBounds(polygon orb.Polygon) Bounds {
	bounds := Bounds{MinX: math.Inf(1), MinY: math.Inf(1), MaxX: math.Inf(-1), MaxY: math.Inf(-1)}
	for _, ring := range polygon {
		for _, point := range ring {
			bounds.MinX = math.Min(bounds.MinX, point[0])
			bounds.MinY = math.Min(bounds.MinY, point[1])
			bounds.MaxX = math.Max(bounds.MaxX, point[0])
			bounds.MaxY = math.Max(bounds.MaxY, point[1])
		}
	}
	return bounds
}

func ringArea(ring orb.Ring) float64 {
	area := 0.0
	for index := range ring {
		next := (index + 1) % len(ring)
		area += ring[index][0]*ring[next][1] - ring[next][0]*ring[index][1]
	}
	return area / 2
}

func polygonContains(polygon orb.Polygon, point orb.Point) bool {
	if !ringContains(polygon[0], point) {
		return false
	}
	for _, hole := range polygon[1:] {
		if ringContains(hole, point) {
			return false
		}
	}
	return true
}

func ringContains(ring orb.Ring, point orb.Point) bool {
	inside := false
	for current, previous := 0, len(ring)-1; current < len(ring); previous, current = current, current+1 {
		a, b := ring[current], ring[previous]
		intersects := (a[1] > point[1]) != (b[1] > point[1]) &&
			point[0] < (b[0]-a[0])*(point[1]-a[1])/(b[1]-a[1])+a[0]
		if intersects {
			inside = !inside
		}
	}
	return inside
}

func coveragePasses(point orb.Point, tracks []Track) int {
	passes := 0
	for _, track := range tracks {
		covered := false
		for _, line := range track.Lines {
			for index := 1; index < len(line); index++ {
				if pointSegmentDistance(point, line[index-1], line[index]) <= track.SwathM/2 {
					covered = true
					break
				}
			}
			if covered {
				break
			}
		}
		if covered {
			passes++
		}
	}
	return passes
}

func pointSegmentDistance(point, start, end orb.Point) float64 {
	dx, dy := end[0]-start[0], end[1]-start[1]
	if dx == 0 && dy == 0 {
		return distance(point, start)
	}
	t := ((point[0]-start[0])*dx + (point[1]-start[1])*dy) / (dx*dx + dy*dy)
	t = math.Max(0, math.Min(1, t))
	projection := orb.Point{start[0] + t*dx, start[1] + t*dy}
	return distance(point, projection)
}

func distance(a, b orb.Point) float64 {
	return math.Hypot(a[0]-b[0], a[1]-b[1])
}

func gapArtifacts(bounds Bounds, minX, minY, maxX, maxY, resolution, ratio float64) ([]byte, []byte, int, error) {
	if ratio <= 0 || math.IsInf(minX, 1) {
		center := orb.Point{(bounds.MinX + bounds.MaxX) / 2, (bounds.MinY + bounds.MaxY) / 2}
		gap, err := MarshalFeature(orb.Polygon{orb.Ring{center, center, center, center}})
		if err != nil {
			return nil, nil, 0, err
		}
		line, err := MarshalFeature(orb.LineString{center, center})
		return gap, line, 0, err
	}
	gap := orb.Polygon{orb.Ring{{minX, minY}, {maxX, minY}, {maxX, maxY}, {minX, maxY}, {minX, minY}}}
	gapJSON, err := MarshalFeature(gap)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("marshal gap geometry: %w", err)
	}
	centerX, centerY := (minX+maxX)/2, (minY+maxY)/2
	var recommendation orb.LineString
	if maxX-minX >= maxY-minY {
		recommendation = orb.LineString{{minX - resolution, centerY}, {maxX + resolution, centerY}}
	} else {
		recommendation = orb.LineString{{centerX, minY - resolution}, {centerX, maxY + resolution}}
	}
	recommendationJSON, err := MarshalFeature(recommendation)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("marshal recommendation geometry: %w", err)
	}
	filtered := 0
	if (maxX-minX)*(maxY-minY) < resolution*resolution*4 {
		filtered = 1
	}
	return gapJSON, recommendationJSON, filtered, nil
}

func finite(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}
