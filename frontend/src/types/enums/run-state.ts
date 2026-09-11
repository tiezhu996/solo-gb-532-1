export type RunState = 'imported' | 'quality_checked' | 'processing' | 'processed' | 'rejected' | 'superseded'
export const RUN_STATE_LABEL: Record<RunState, string> = {
  imported: '已导入', quality_checked: '质量已检', processing: '处理中', processed: '已处理', rejected: '已驳回', superseded: '已替代'
}
export const RUN_TRANSITIONS: Partial<Record<RunState, RunState[]>> = {
  imported: ['quality_checked', 'rejected'], quality_checked: ['processing', 'rejected'], processing: ['processed', 'rejected'], processed: ['superseded']
}
