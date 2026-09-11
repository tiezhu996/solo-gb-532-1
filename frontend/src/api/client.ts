import type { ApiErrorPayload, ApiResponse, PageResponse } from '../types/api'

export class ApiError extends Error {
  constructor(readonly status:number, readonly code:string, message:string, readonly requestId:string, readonly details?:Record<string,unknown>) { super(message); this.name='ApiError' }
}

const TOKEN_KEY='sonar_coverage_token'
export const session={token:()=>sessionStorage.getItem(TOKEN_KEY),set:(value:string)=>sessionStorage.setItem(TOKEN_KEY,value),clear:()=>sessionStorage.removeItem(TOKEN_KEY)}

async function request<T>(path:string, init:RequestInit={}):Promise<T>{
  const headers=new Headers(init.headers); if(init.body&&!headers.has('Content-Type'))headers.set('Content-Type','application/json')
  const token=session.token(); if(token)headers.set('Authorization',`Bearer ${token}`)
  let response:Response
  try{response=await fetch(`/api/v1${path}`,{...init,headers})}catch{const message='无法连接覆盖分析服务，请检查服务状态。';window.dispatchEvent(new CustomEvent('api:error',{detail:message}));throw new Error(message)}
  const payload=await response.json() as ApiResponse<unknown>|ApiErrorPayload
  if(!response.ok){const errorPayload=payload as ApiErrorPayload;const error=new ApiError(response.status,errorPayload.error.code,errorPayload.error.message,errorPayload.request_id,errorPayload.error.details);if(response.status===401&&path!='/auth/login'){session.clear();window.dispatchEvent(new Event('auth:expired'))}window.dispatchEvent(new CustomEvent('api:error',{detail:`${error.code}：${error.message}`}));throw error}
  return payload as T
}

export const apiClient={
  get:<T>(path:string)=>request<ApiResponse<T>>(path),
  page:<T>(path:string)=>request<PageResponse<T>>(path),
  post:<T>(path:string,body:unknown,headers?:HeadersInit)=>request<ApiResponse<T>>(path,{method:'POST',body:JSON.stringify(body),headers}),
  put:<T>(path:string,body:unknown)=>request<ApiResponse<T>>(path,{method:'PUT',body:JSON.stringify(body)})
}
