export type GapSeverity = 'minor' | 'major' | 'critical'
export type GapState = 'detected' | 'reviewed' | 'accepted' | 'false_positive' | 'resurveyed' | 'closed'
export const GAP_SEVERITY_LABEL: Record<GapSeverity,string> = { minor: '轻微', major: '主要', critical: '严重' }
export const GAP_STATE_LABEL: Record<GapState,string> = { detected:'已检测', reviewed:'已复核', accepted:'接受补测', false_positive:'误报', resurveyed:'已补测', closed:'已关闭' }
