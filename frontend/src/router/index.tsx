import { useEffect } from 'react'
import { Navigate, createBrowserRouter, useLocation } from 'react-router-dom'
import { Box, CircularProgress } from '@mui/material'
import { AppLayout } from '../components/common/AppLayout'
import { session } from '../api/client'
import { useAuthStore } from '../stores/auth-store'
import { LoginPage } from '../pages/LoginPage'
import { AreasPage } from '../pages/AreasPage'
import { PlansPage } from '../pages/PlansPage'
import { RunsPage } from '../pages/RunsPage'
import { CoveragePage } from '../pages/CoveragePage'
import { AuditPage } from '../pages/AuditPage'

function Protected(){const user=useAuthStore(state=>state.user);const restore=useAuthStore(state=>state.restore);const logout=useAuthStore(state=>state.logout);const location=useLocation();const token=session.token();useEffect(()=>{if(token&&!user)void restore()},[token,user,restore]);useEffect(()=>{window.addEventListener('auth:expired',logout);return()=>window.removeEventListener('auth:expired',logout)},[logout]);if(!token)return <Navigate to="/login" replace state={{from:location.pathname}}/>;if(!user)return <Box className="route-loading"><CircularProgress/><span>正在恢复离线工作会话</span></Box>;return <AppLayout/>}

export const router=createBrowserRouter([
  {path:'/login',element:<LoginPage/>},
  {path:'/',element:<Protected/>,children:[{index:true,element:<Navigate to="/areas" replace/>},{path:'areas',element:<AreasPage/>},{path:'plans',element:<PlansPage/>},{path:'runs',element:<RunsPage/>},{path:'coverage',element:<CoveragePage/>},{path:'audit',element:<AuditPage/>}]},
  {path:'*',element:<Navigate to="/" replace/>}
])
