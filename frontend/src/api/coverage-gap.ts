import { apiClient } from './client'
import type { DetectCoverage, CoverageGap, CoverageResult } from '../types/coverage-gap'
import type { GapState } from '../types/enums/gap-severity'
export const coverageGapApi={list:()=>apiClient.page<CoverageGap[]>('/coverage-gaps?page_size=100'),get:(id:number)=>apiClient.get<CoverageGap>(`/coverage-gaps/${id}`),detect:(body:DetectCoverage,key:string)=>apiClient.post<CoverageResult>('/coverage-gaps/detect',body,{'Idempotency-Key':key}),transition:(id:number,target:GapState,version:number,note:string)=>apiClient.post<CoverageGap>(`/coverage-gaps/${id}/transition`,{target_state:target,expected_version:version,review_note:note})}
