import { Box, Divider, Stack, Typography } from '@mui/material'
import type { ReactNode } from 'react'

export function PageHeader({eyebrow,title,description,actions}:{eyebrow:string;title:string;description:string;actions?:ReactNode}){
  return <><Stack direction={{xs:'column',sm:'row'}} justifyContent="space-between" alignItems={{xs:'stretch',sm:'flex-start'}} gap={2}>
    <Box><Typography variant="overline">{eyebrow}</Typography><Typography variant="h4" component="h1">{title}</Typography><Typography color="text.secondary" sx={{mt:.75,maxWidth:760}}>{description}</Typography></Box>{actions&&<Stack direction="row" gap={1}>{actions}</Stack>}
  </Stack><Divider sx={{my:2.5,borderColor:'#183f42',borderBottomWidth:2}}/></>
}
