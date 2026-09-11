import { useMemo } from 'react'
import { useAuthStore } from '../stores/auth-store'
import type { Role } from '../types/auth'

export function useAuth(){
  const user=useAuthStore(state=>state.user);const logout=useAuthStore(state=>state.logout)
  return useMemo(()=>({user,logout,hasRole:(...roles:Role[])=>Boolean(user&&roles.includes(user.role)),readOnly:user?.role==='auditor'}),[user,logout])
}
