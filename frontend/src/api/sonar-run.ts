import { apiClient } from './client'
import type { RunState } from '../types/enums/run-state'
import type { ImportSonarRun, RunQualityEvidence, SonarRun } from '../types/sonar-run'
export const sonarRunApi={list:()=>apiClient.page<SonarRun[]>('/runs?page_size=100'),get:(id:number)=>apiClient.get<SonarRun>(`/runs/${id}`),quality:(id:number)=>apiClient.get<RunQualityEvidence>(`/runs/${id}/quality`),import:(body:ImportSonarRun)=>apiClient.post<{run:SonarRun;idempotent:boolean}>('/runs/import',body),transition:(id:number,target:RunState,version:number,reason='')=>apiClient.post<SonarRun>(`/runs/${id}/transition`,{target_state:target,expected_version:version,reason})}
