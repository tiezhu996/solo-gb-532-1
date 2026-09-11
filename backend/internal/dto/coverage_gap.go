package dto

type DetectCoverageRequest struct {
	SurveyAreaID     uint    `json:"survey_area_id" binding:"required,gt=0"`
	SourceRunIDs     []uint  `json:"source_run_ids" binding:"required,min=1,dive,gt=0"`
	AlgorithmVersion string  `json:"algorithm_version" binding:"required,min=3,max=40"`
	ResolutionM      float64 `json:"resolution_m" binding:"omitempty,gt=0,lte=100"`
}

type GapTransitionRequest struct {
	TargetState     string `json:"target_state" binding:"required,oneof=reviewed accepted false_positive resurveyed closed"`
	ExpectedVersion uint   `json:"expected_version" binding:"required,gt=0"`
	ReviewNote      string `json:"review_note" binding:"required,min=8,max=600"`
}

type CoverageGapQuery struct {
	SurveyAreaID uint
	State        string
	Severity     string
	Page         int
	PageSize     int
}

type CoverageEvidence struct {
	InputHash            string  `json:"input_hash"`
	CoordinateSystem     string  `json:"coordinate_system"`
	AlgorithmVersion     string  `json:"algorithm_version"`
	SourceRunCount       int     `json:"source_run_count"`
	CoverageRatio        float64 `json:"coverage_ratio"`
	OverlapRatio         float64 `json:"overlap_ratio"`
	GapRatio             float64 `json:"gap_ratio"`
	FilteredFragments    int     `json:"filtered_fragments"`
	ProcessingMillis     int64   `json:"processing_millis"`
	DecisionBoundaryNote string  `json:"decision_boundary_note"`
}
