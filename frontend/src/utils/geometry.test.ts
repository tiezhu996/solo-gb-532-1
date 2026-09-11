import { describe, expect, it } from 'vitest'
import { geometryBounds, geometryLines } from './geometry'
import type { GeoJSONFeature } from '../types/api'

describe('survey geometry utilities',()=>{
  const polygon:GeoJSONFeature={type:'Feature',properties:{},geometry:{type:'Polygon',coordinates:[[[0,0],[120,0],[120,80],[0,80],[0,0]]]}}
  const tracks:GeoJSONFeature={type:'Feature',properties:{},geometry:{type:'MultiLineString',coordinates:[[[10,20],[110,20]],[[10,60],[110,60]]]}}
  it('keeps polygon rings and multi-lines as separate drawable lines',()=>{expect(geometryLines(polygon)).toHaveLength(1);expect(geometryLines(tracks)).toHaveLength(2)})
  it('calculates stable projected bounds across layers',()=>{expect(geometryBounds([polygon,tracks])).toEqual({minX:0,minY:0,maxX:120,maxY:80,pointCount:9})})
})
