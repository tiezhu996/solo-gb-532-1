import { apiClient } from './client'
import type { AuditEvent } from '../types/api'
export const auditApi={list:(query='')=>apiClient.page<AuditEvent[]>(`/audits${query}`)}
