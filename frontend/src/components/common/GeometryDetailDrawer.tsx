import CloseRounded from '@mui/icons-material/CloseRounded'
import ContentCopyRounded from '@mui/icons-material/ContentCopyRounded'
import { Box, Divider, Drawer, IconButton, Stack, Tooltip, Typography } from '@mui/material'
import type { GeoJSONFeature } from '../../types/api'

export function GeometryDetailDrawer({open,onClose,title,geometry,metadata}:{open:boolean;onClose:()=>void;title:string;geometry?:GeoJSONFeature|null;metadata?:Record<string,string|number>}){
  const copy=()=>geometry&&navigator.clipboard.writeText(JSON.stringify(geometry,null,2))
  return <Drawer anchor="right" open={open} onClose={onClose} PaperProps={{sx:{width:{xs:'100%',sm:480},maxWidth:'100%'}}}>
    <Stack direction="row" alignItems="center" justifyContent="space-between" sx={{p:2}}><Box><Typography variant="overline">GEOMETRY EVIDENCE</Typography><Typography variant="h6">{title}</Typography></Box><Stack direction="row"><Tooltip title="复制 GeoJSON"><span><IconButton aria-label="复制 GeoJSON" onClick={copy} disabled={!geometry}><ContentCopyRounded/></IconButton></span></Tooltip><IconButton aria-label="关闭几何详情" onClick={onClose}><CloseRounded/></IconButton></Stack></Stack>
    <Divider/>{metadata&&<Box component="dl" className="detail-grid" sx={{m:0,p:2}}>{Object.entries(metadata).map(([key,value])=><Box key={key}><Typography component="dt" variant="caption" color="text.secondary">{key}</Typography><Typography component="dd" sx={{m:0,fontFamily:'ui-monospace, SFMono-Regular, Menlo, monospace',fontSize:13,wordBreak:'break-word'}}>{value}</Typography></Box>)}</Box>}
    <Box component="pre" sx={{m:0,p:2,bgcolor:'#f2f6f5',borderTop:'1px solid',borderColor:'divider',fontSize:12,lineHeight:1.55,whiteSpace:'pre-wrap',wordBreak:'break-word',overflow:'auto',flex:1}}>{geometry?JSON.stringify(geometry,null,2):'未选择几何证据'}</Box>
  </Drawer>
}
