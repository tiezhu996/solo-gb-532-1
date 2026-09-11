import { create } from 'zustand'
import { transectPlanApi } from '../api/transect-plan'
import type { GenerateTransectPlan, TransectPlan } from '../types/transect-plan'

interface PlanState{items:TransectPlan[];selected:TransectPlan|null;loading:boolean;fetch:()=>Promise<void>;select:(item:TransectPlan)=>void;generate:(body:GenerateTransectPlan)=>Promise<TransectPlan>;lock:(item:TransectPlan)=>Promise<void>;copy:(item:TransectPlan)=>Promise<void>}
export const useTransectPlanStore=create<PlanState>((set)=>({items:[],selected:null,loading:false,
  fetch:async()=>{set({loading:true});try{const response=await transectPlanApi.list();set(state=>({items:response.data,selected:state.selected??response.data[0]??null}))}finally{set({loading:false})}},
  select:(selected)=>set({selected}),
  generate:async(body)=>{const response=await transectPlanApi.generate(body);set(state=>({items:[response.data,...state.items],selected:response.data}));return response.data},
  lock:async(item)=>{const response=await transectPlanApi.lock(item.id,item.version);set(state=>({items:state.items.map(value=>value.id===item.id?response.data:value),selected:response.data}));},
  copy:async(item)=>{const response=await transectPlanApi.copy(item.id);set(state=>({items:[response.data,...state.items],selected:response.data}));}
}))
