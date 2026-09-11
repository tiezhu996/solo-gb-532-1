import type { GeoJSONFeature } from './api'
import type { TransectPlan } from './transect-plan'
import type { RunState } from './enums/run-state'

export interface SonarRun {
  id:number; transect_plan_id:number; run_code:string; track_geojson:GeoJSONFeature; actual_swath_m:number; started_at:string; ended_at:string
  navigation_quality:'good'|'degraded'|'invalid'; run_state:RunState; source_checksum:string; imported_by:number; version:number; created_at:string; updated_at:string
  transect_plan?:TransectPlan
}
export interface ImportSonarRun { transect_plan_id:number; run_code:string; track_geojson:GeoJSONFeature; actual_swath_m:number; started_at:string; ended_at:string; navigation_quality:'good'|'degraded'|'invalid' }
export interface RunQualityEvidence { coordinate_count:number; track_length_m:number; duration_minutes:number; average_speed_mps:number; warnings:string[] }
