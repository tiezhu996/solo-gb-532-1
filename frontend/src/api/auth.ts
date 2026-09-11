import { apiClient } from './client'
import type { LoginResponse, UserIdentity } from '../types/auth'
export const authApi={login:(username:string,password:string)=>apiClient.post<LoginResponse>('/auth/login',{username,password}),me:()=>apiClient.get<UserIdentity>('/auth/me')}
