export interface ApiResponse<T> { data: T; request_id: string }
export interface PageMeta { page: number; page_size: number; total: number }
export interface PageResponse<T> extends ApiResponse<T> { meta: PageMeta }
export interface ApiErrorPayload { error: { code: string; message: string; details?: Record<string, unknown> }; request_id: string }

export type Position = [number, number]
export type GeoJSONGeometry =
  | { type: 'Polygon'; coordinates: Position[][] }
  | { type: 'MultiPolygon'; coordinates: Position[][][] }
  | { type: 'LineString'; coordinates: Position[] }
  | { type: 'MultiLineString'; coordinates: Position[][] }
export interface GeoJSONFeature { type: 'Feature'; properties: Record<string, unknown>; geometry: GeoJSONGeometry }

export interface AuditEvent {
  id: number; request_id: string; user_id: number; actor: string; role: string; action: string
  entity_type: string; entity_id: number; before_json: string; after_json: string; metadata: string; created_at: string
}
