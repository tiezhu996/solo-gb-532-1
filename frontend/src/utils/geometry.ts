import type { GeoJSONFeature, Position } from '../types/api'

export function geometryLines(feature:GeoJSONFeature):Position[][]{
  const geometry=feature.geometry
  if(geometry.type==='LineString')return[geometry.coordinates]
  if(geometry.type==='MultiLineString')return geometry.coordinates
  if(geometry.type==='Polygon')return geometry.coordinates
  return geometry.coordinates.flat()
}

export function geometryBounds(features:GeoJSONFeature[]){
  const points=features.flatMap(feature=>geometryLines(feature).flat())
  if(points.length===0)return null
  const xs=points.map(point=>point[0]),ys=points.map(point=>point[1])
  return{minX:Math.min(...xs),minY:Math.min(...ys),maxX:Math.max(...xs),maxY:Math.max(...ys),pointCount:points.length}
}
