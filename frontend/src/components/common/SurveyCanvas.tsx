import { useEffect, useRef } from 'react'
import { Box } from '@mui/material'
import type { GeoJSONFeature, Position } from '../../types/api'
import { geometryBounds, geometryLines } from '../../utils/geometry'

export interface MapLayer{feature:GeoJSONFeature;label:string;color:string;fill?:string;width?:number;dash?:number[]}

export function SurveyCanvas({layers,ariaLabel='测绘几何画布'}:{layers:MapLayer[];ariaLabel?:string}){
  const canvasRef=useRef<HTMLCanvasElement>(null);const hostRef=useRef<HTMLDivElement>(null)
  useEffect(()=>{
    const canvas=canvasRef.current,host=hostRef.current;if(!canvas||!host)return
    const draw=()=>{
      const rect=host.getBoundingClientRect();const ratio=Math.min(window.devicePixelRatio||1,2);canvas.width=Math.max(1,Math.round(rect.width*ratio));canvas.height=Math.max(1,Math.round(rect.height*ratio));canvas.style.width=`${rect.width}px`;canvas.style.height=`${rect.height}px`
      const context=canvas.getContext('2d');if(!context)return;context.setTransform(ratio,0,0,ratio,0,0);const width=rect.width,height=rect.height
      context.fillStyle='#edf4f3';context.fillRect(0,0,width,height);context.strokeStyle='#d2dfdd';context.lineWidth=1
      for(let x=24;x<width;x+=48){context.beginPath();context.moveTo(x,0);context.lineTo(x,height);context.stroke()}for(let y=24;y<height;y+=48){context.beginPath();context.moveTo(0,y);context.lineTo(width,y);context.stroke()}
      const bounds=geometryBounds(layers.map(layer=>layer.feature));if(!bounds){context.fillStyle='#617475';context.font='14px sans-serif';context.fillText('等待几何证据',24,36);return}
      let{minX,maxX,minY,maxY}=bounds;if(maxX===minX){maxX+=1;minX-=1}if(maxY===minY){maxY+=1;minY-=1}
      const padding=32,scale=Math.min((width-padding*2)/(maxX-minX),(height-padding*2)/(maxY-minY));const project=(point:Position):Position=>[padding+(point[0]-minX)*scale,height-padding-(point[1]-minY)*scale]
      layers.forEach(layer=>{context.strokeStyle=layer.color;context.fillStyle=layer.fill??'transparent';context.lineWidth=layer.width??2;context.setLineDash(layer.dash??[]);geometryLines(layer.feature).forEach(line=>{if(!line.length)return;context.beginPath();line.forEach((point,index)=>{const [x,y]=project(point);if(index===0)context.moveTo(x,y);else context.lineTo(x,y)});if(layer.feature.geometry.type.includes('Polygon'))context.closePath();if(layer.fill)context.fill();context.stroke()})});context.setLineDash([])
      context.fillStyle='#415b5c';context.font='11px ui-monospace, monospace';context.fillText(`${minX.toFixed(0)} m`,padding,height-10);context.textAlign='right';context.fillText(`${maxX.toFixed(0)} m`,width-padding,height-10);context.textAlign='left'
    }
    draw();const observer=new ResizeObserver(draw);observer.observe(host);return()=>observer.disconnect()
  },[layers])
  return <Box ref={hostRef} className="survey-canvas" role="img" aria-label={ariaLabel}><canvas ref={canvasRef}/></Box>
}
