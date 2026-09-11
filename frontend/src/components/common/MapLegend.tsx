import { Box, Stack, Typography } from '@mui/material'

export interface LegendItem{label:string;color:string;pattern?:'solid'|'dash'|'fill'}
export function MapLegend({items}:{items:LegendItem[]}){
  return <Stack component="ul" direction="row" useFlexGap flexWrap="wrap" gap={2} sx={{m:0,p:0,listStyle:'none'}} aria-label="几何图例">
    {items.map(item=><Stack component="li" direction="row" alignItems="center" gap={0.75} key={item.label}>
      <Box aria-hidden sx={{width:24,height:item.pattern==='fill'?10:2,border:item.pattern==='dash'?`1px dashed ${item.color}`:'none',bgcolor:item.pattern==='dash'?'transparent':item.color}} />
      <Typography variant="caption" color="text.secondary">{item.label}</Typography>
    </Stack>)}
  </Stack>
}
