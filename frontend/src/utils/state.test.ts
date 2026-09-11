import { describe, expect, it } from 'vitest'
import { RUN_STATE_LABEL, RUN_TRANSITIONS } from '../types/enums/run-state'
import { GAP_SEVERITY_LABEL, GAP_STATE_LABEL } from '../types/enums/gap-severity'

describe('shared workflow enumerations',()=>{
  it('prevents imported runs from skipping quality and processing',()=>{expect(RUN_TRANSITIONS.imported).toEqual(['quality_checked','rejected']);expect(RUN_TRANSITIONS.imported).not.toContain('processed')})
  it('provides operator-facing labels for every state and severity',()=>{expect(Object.keys(RUN_STATE_LABEL)).toHaveLength(6);expect(Object.keys(GAP_STATE_LABEL)).toHaveLength(6);expect(Object.keys(GAP_SEVERITY_LABEL)).toHaveLength(3)})
})
