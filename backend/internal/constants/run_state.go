package constants

import "fmt"

type RunState string

const (
	RunImported       RunState = "imported"
	RunQualityChecked RunState = "quality_checked"
	RunProcessing     RunState = "processing"
	RunProcessed      RunState = "processed"
	RunRejected       RunState = "rejected"
	RunSuperseded     RunState = "superseded"
)

var runTransitions = map[RunState]map[RunState]struct{}{
	RunImported:       {RunQualityChecked: {}, RunRejected: {}},
	RunQualityChecked: {RunProcessing: {}, RunRejected: {}},
	RunProcessing:     {RunProcessed: {}, RunRejected: {}},
	RunProcessed:      {RunSuperseded: {}},
	RunRejected:       {},
	RunSuperseded:     {},
}

func (s RunState) Valid() bool {
	_, ok := runTransitions[s]
	return ok
}

func (s RunState) CanTransition(target RunState) bool {
	allowed, ok := runTransitions[s]
	if !ok {
		return false
	}
	_, ok = allowed[target]
	return ok
}

func ParseRunState(value string) (RunState, error) {
	state := RunState(value)
	if !state.Valid() {
		return "", fmt.Errorf("unknown run state %q", value)
	}
	return state, nil
}

func RunStates() []RunState {
	return []RunState{RunImported, RunQualityChecked, RunProcessing, RunProcessed, RunRejected, RunSuperseded}
}
