export type Role = 'admin' | 'survey_planner' | 'data_processor' | 'reviewer' | 'auditor'
export interface UserIdentity { id: number; username: string; display_name: string; role: Role }
export interface LoginResponse { token: string; user: UserIdentity }
