import AssessmentOutlined from '@mui/icons-material/AssessmentOutlined'
import LogoutRounded from '@mui/icons-material/LogoutRounded'
import MapOutlined from '@mui/icons-material/MapOutlined'
import RouteOutlined from '@mui/icons-material/RouteOutlined'
import RadarOutlined from '@mui/icons-material/RadarOutlined'
import FactCheckOutlined from '@mui/icons-material/FactCheckOutlined'
import { Box, IconButton, Stack, Tooltip, Typography } from '@mui/material'
import { NavLink, Outlet } from 'react-router-dom'
import { useAuth } from '../../hooks/useAuth'

const links=[{to:'/areas',label:'测区',icon:MapOutlined},{to:'/plans',label:'测线',icon:RouteOutlined},{to:'/runs',label:'航迹',icon:RadarOutlined},{to:'/coverage',label:'覆盖',icon:AssessmentOutlined},{to:'/audit',label:'审计',icon:FactCheckOutlined,audit:true}]
export function AppLayout(){const{user,logout,hasRole}=useAuth();return <Box className="app-shell">
  <Box component="aside" className="sidebar"><Stack className="brand" direction="row" alignItems="center" gap={1.25}><Box className="brand-mark">SS</Box><Box><Typography fontWeight={800}>SonarScope</Typography><Typography variant="caption">Survey evidence desk</Typography></Box></Stack>
    <Box component="nav" aria-label="主导航" className="main-nav">{links.filter(item=>!item.audit||hasRole('admin','reviewer','auditor')).map(item=>{const Icon=item.icon;return <NavLink to={item.to} key={item.to} className={({isActive})=>isActive?'nav-link active':'nav-link'}><Icon fontSize="small"/><span>{item.label}</span></NavLink>})}</Box>
    <Stack className="operator" direction="row" alignItems="center" justifyContent="space-between"><Box><Typography variant="caption">{user?.display_name}</Typography><Typography variant="overline" display="block">{user?.role}</Typography></Box><Tooltip title="退出"><IconButton color="inherit" aria-label="退出" onClick={logout}><LogoutRounded/></IconButton></Tooltip></Stack>
  </Box>
  <Box className="workspace"><Box className="system-strip"><Typography variant="overline">OFFLINE HYDROGRAPHIC REVIEW</Typography><Typography variant="overline">NO VESSEL CONTROL</Typography></Box><Box className="boundary-note"><strong>离线决策边界</strong><span>几何结果仅供测绘人员复核，不连接船舶、声呐或自动驾驶设备，也不下发控制指令。</span></Box><Box component="main" className="page-content"><Outlet/></Box></Box>
  </Box>}
