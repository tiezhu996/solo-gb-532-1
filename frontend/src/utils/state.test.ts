import { describe, expect, it } from 'vitest'
import { RUN_STATE_LABEL, RUN_TRANSITIONS } from '../types/enums/run-state'
import { GAP_SEVERITY_LABEL, GAP_STATE_LABEL } from '../types/enums/gap-severity'
import { PLAN_SOURCE_LABEL } from '../types/enums/plan-source'

describe('shared workflow enumerations',()=>{
  it('prevents imported runs from skipping quality and processing',()=>{expect(RUN_TRANSITIONS.imported).toEqual(['quality_checked','rejected']);expect(RUN_TRANSITIONS.imported).not.toContain('processed')})
  it('provides operator-facing labels for every state and severity',()=>{expect(Object.keys(RUN_STATE_LABEL)).toHaveLength(6);expect(Object.keys(GAP_STATE_LABEL)).toHaveLength(6);expect(Object.keys(GAP_SEVERITY_LABEL)).toHaveLength(3)})
  it('labels every plan source including gap resurvey',()=>{expect(Object.keys(PLAN_SOURCE_LABEL)).toEqual(['manual','generated','resurvey']);expect(PLAN_SOURCE_LABEL.resurvey).toBe('缺口补测')})
})
