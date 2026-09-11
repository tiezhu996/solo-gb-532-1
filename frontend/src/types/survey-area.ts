import type { GeoJSONFeature } from './api'

export interface SurveyArea {
  id:number; area_code:string; name:string; boundary_geojson:GeoJSONFeature; target_resolution_m:number; coordinate_system:string
  default_swath_m:number; owner_team:string; status:'draft'|'active'|'archived'; version:number; created_at:string; updated_at:string
}
export interface AreaSummary { plan_count:number; processed_runs:number; open_gap_count:number; latest_coverage_ratio:number }
export interface SurveyAreaView extends SurveyArea { summary: AreaSummary }
export interface CreateSurveyArea {
  area_code:string; name:string; boundary_geojson:GeoJSONFeature; target_resolution_m:number; coordinate_system:string; default_swath_m:number; owner_team:string
}
