import { apiClient } from './client'
import type { GenerateTransectPlan, TransectPlan } from '../types/transect-plan'
export const transectPlanApi={list:()=>apiClient.page<TransectPlan[]>('/plans?page_size=100'),get:(id:number)=>apiClient.get<TransectPlan>(`/plans/${id}`),generate:(body:GenerateTransectPlan)=>apiClient.post<TransectPlan>('/plans/generate',body),lock:(id:number,version:number)=>apiClient.post<TransectPlan>(`/plans/${id}/transition`,{target_state:'locked',expected_version:version}),copy:(id:number)=>apiClient.post<TransectPlan>(`/plans/${id}/copy`,{})}
