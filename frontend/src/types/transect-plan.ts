import type { GeoJSONFeature } from './api'
import type { CoverageGap } from './coverage-gap'
import type { GapState } from './enums/gap-severity'
import type { PlanSource } from './enums/plan-source'
import type { SurveyArea } from './survey-area'

export interface TransectPlan {
  id:number; survey_area_id:number; name:string; line_geojson:GeoJSONFeature; planned_heading:number; planned_swath_m:number; line_spacing_m:number
  plan_state:'draft'|'locked'; plan_source:PlanSource; source_gap_id:number|null; source_plan_id:number|null; source_input_hash:string
  source_gap_version:number|null; source_gap_state:''|GapState
  version:number; created_by:number; created_at:string; updated_at:string; survey_area?:SurveyArea; source_gap?:CoverageGap; source_plan?:TransectPlan
}
export interface GenerateTransectPlan { survey_area_id:number; name:string; heading:number; line_spacing_m:number; planned_swath_m:number }
export interface GenerateResurveyPlan { survey_area_id:number; name?:string; planned_swath_m?:number; line_spacing_m?:number }
