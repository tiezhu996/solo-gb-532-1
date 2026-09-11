import { apiClient } from './client'
import type { CreateSurveyArea, SurveyAreaView } from '../types/survey-area'
export const surveyAreaApi={list:(query='')=>apiClient.page<SurveyAreaView[]>(`/areas${query}`),get:(id:number)=>apiClient.get<SurveyAreaView>(`/areas/${id}`),create:(body:CreateSurveyArea)=>apiClient.post<SurveyAreaView>('/areas',body)}
