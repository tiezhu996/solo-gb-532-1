import { create } from 'zustand'
import { sonarRunApi } from '../api/sonar-run'
import type { RunState } from '../types/enums/run-state'
import type { ImportSonarRun, RunQualityEvidence, SonarRun } from '../types/sonar-run'

interface RunStore{items:SonarRun[];selected:SonarRun|null;quality:RunQualityEvidence|null;loading:boolean;fetch:()=>Promise<void>;select:(item:SonarRun)=>Promise<void>;import:(body:ImportSonarRun)=>Promise<SonarRun>;transition:(item:SonarRun,target:RunState,reason?:string)=>Promise<void>}
export const useSonarRunStore=create<RunStore>((set)=>({items:[],selected:null,quality:null,loading:false,
  fetch:async()=>{set({loading:true});try{const response=await sonarRunApi.list();set(state=>({items:response.data,selected:state.selected??response.data[0]??null}))}finally{set({loading:false})}},
  select:async(selected)=>{set({selected,quality:null});const response=await sonarRunApi.quality(selected.id);set({quality:response.data})},
  import:async(body)=>{const response=await sonarRunApi.import(body);const item=response.data.run;set(state=>({items:state.items.some(value=>value.id===item.id)?state.items:[item,...state.items],selected:item}));return item},
  transition:async(item,target,reason='')=>{const response=await sonarRunApi.transition(item.id,target,item.version,reason);set(state=>({items:state.items.map(value=>value.id===item.id?response.data:value),selected:response.data}));}
}))
