import { useMemo } from 'react'
import type { CoverageGap } from '../types/coverage-gap'
import type { SonarRun } from '../types/sonar-run'
import type { SurveyArea } from '../types/survey-area'
import type { TransectPlan } from '../types/transect-plan'
import type { MapLayer } from '../components/common/SurveyCanvas'

export function useCoverageLayers(area?:SurveyArea|null,plan?:TransectPlan|null,runs:SonarRun[]=[],gap?:CoverageGap|null){
  return useMemo<MapLayer[]>(()=>{
    const layers:MapLayer[]=[]
    if(area)layers.push({feature:area.boundary_geojson,label:'测区边界',color:'#185b63',fill:'rgba(34, 122, 126, 0.08)',width:2.2})
    if(plan)layers.push({feature:plan.line_geojson,label:'计划测线',color:'#255d86',width:1.5,dash:[7,4]})
    runs.forEach((run,index)=>layers.push({feature:run.track_geojson,label:`航迹 ${run.run_code}`,color:index%2===0?'#2c7873':'#647a38',width:2.2}))
    if(gap){layers.push({feature:gap.gap_geojson,label:'覆盖缺口',color:'#b8443c',fill:'rgba(184, 68, 60, 0.22)',width:2});layers.push({feature:gap.recommended_line_geojson,label:'建议补测线',color:'#d39a20',width:3,dash:[3,3]})}
    return layers
  },[area,plan,runs,gap])
}
