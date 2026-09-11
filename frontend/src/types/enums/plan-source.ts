export type PlanSource = 'manual' | 'generated' | 'resurvey'
export const PLAN_SOURCE_LABEL: Record<PlanSource,string> = { manual:'手工创建', generated:'平行测线生成', resurvey:'缺口补测' }
