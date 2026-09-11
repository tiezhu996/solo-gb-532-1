import { create } from 'zustand'
import { authApi } from '../api/auth'
import { session } from '../api/client'
import type { UserIdentity } from '../types/auth'

const USER_KEY='sonar_coverage_user'
function savedUser():UserIdentity|null{try{return JSON.parse(sessionStorage.getItem(USER_KEY)??'null') as UserIdentity|null}catch{return null}}

interface AuthState{user:UserIdentity|null;loading:boolean;login:(username:string,password:string)=>Promise<void>;restore:()=>Promise<void>;logout:()=>void}
export const useAuthStore=create<AuthState>((set)=>({
  user:savedUser(),loading:false,
  login:async(username,password)=>{set({loading:true});try{const response=await authApi.login(username,password);session.set(response.data.token);sessionStorage.setItem(USER_KEY,JSON.stringify(response.data.user));set({user:response.data.user})}finally{set({loading:false})}},
  restore:async()=>{if(!session.token())return;try{const response=await authApi.me();sessionStorage.setItem(USER_KEY,JSON.stringify(response.data));set({user:response.data})}catch{session.clear();sessionStorage.removeItem(USER_KEY);set({user:null})}},
  logout:()=>{session.clear();sessionStorage.removeItem(USER_KEY);set({user:null})}
}))
