import { Chip } from '@mui/material'
import { RUN_STATE_LABEL, type RunState } from '../../types/enums/run-state'

const colors:Record<RunState,{bg:string,fg:string,border:string}>={
  imported:{bg:'#eef2f2',fg:'#42575a',border:'#bdc9ca'},quality_checked:{bg:'#e7f1f0',fg:'#245b59',border:'#9dc4c0'},processing:{bg:'#fff4d8',fg:'#765611',border:'#dfc16c'},processed:{bg:'#e4f1e8',fg:'#285f3d',border:'#9ac3a6'},rejected:{bg:'#f9e6e4',fg:'#8d312c',border:'#dda19c'},superseded:{bg:'#ece9f0',fg:'#5a4e65',border:'#bbb2c3'}
}
export function RunStateBadge({state}:{state:RunState}){const style=colors[state];return <Chip size="small" label={RUN_STATE_LABEL[state]} sx={{height:24,borderRadius:1,bgcolor:style.bg,color:style.fg,border:`1px solid ${style.border}`,fontWeight:700}}/>}
