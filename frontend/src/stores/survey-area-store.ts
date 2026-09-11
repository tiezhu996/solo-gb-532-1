import { create } from 'zustand'
import { surveyAreaApi } from '../api/survey-area'
import type { CreateSurveyArea, SurveyAreaView } from '../types/survey-area'

interface AreaState{items:SurveyAreaView[];selected:SurveyAreaView|null;loading:boolean;fetch:()=>Promise<void>;select:(item:SurveyAreaView)=>void;create:(body:CreateSurveyArea)=>Promise<SurveyAreaView>}
export const useSurveyAreaStore=create<AreaState>((set)=>({items:[],selected:null,loading:false,
  fetch:async()=>{set({loading:true});try{const response=await surveyAreaApi.list('?page_size=100');set(state=>({items:response.data,selected:state.selected??response.data[0]??null}))}finally{set({loading:false})}},
  select:(selected)=>set({selected}),
  create:async(body)=>{const response=await surveyAreaApi.create(body);set(state=>({items:[response.data,...state.items],selected:response.data}));return response.data}
}))
