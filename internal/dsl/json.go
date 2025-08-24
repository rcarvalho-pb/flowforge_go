package dsl

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/rcarvalho-pb/flowforge-go/internal/domain"
)

func ParseDefinitionJSON(b []byte) (*domain.WorkflowDefinition, error) {
	var def domain.WorkflowDefinition
	if err := json.Unmarshal(b, &def); err != nil {
		return nil, fmt.Errorf("invalid json: %w", err)
	}
	if def.Name == "" || len(def.States) == 0 {
		return nil, fmt.Errorf("%w: name/states required", domain.ErrWorkflowInvalid)
	}
	seenStates := map[string]bool{}
	var initialCount int
	for _, s := range def.States {
		if s.Name == "" {
			return nil, fmt.Errorf("%w: state without name", domain.ErrWorkflowInvalid)
		}
		if seenStates[s.Name] {
			return nil, fmt.Errorf("%w: duplicated state %s", domain.ErrWorkflowInvalid, s.Name)
		}
		seenStates[s.Name] = true
		if s.Initial {
			initialCount++
		}
	}
	if initialCount != 1 {
		return nil, fmt.Errorf("%w: there must be exactly 1 initial state", domain.ErrWorkflowInvalid)
	}
	for _, t := range def.Transitions {
		if t.From == "" || t.To == "" || t.Event == "" {
			return nil, fmt.Errorf("%w: transition missing fields", domain.ErrWorkflowInvalid)
		}
		if !seenStates[t.From] || !seenStates[t.To] {
			return nil, fmt.Errorf("%w: transition references unkown state", domain.ErrWorkflowInvalid)
		}
	}
	for state, dur := range def.SLA {
		if _, err := time.ParseDuration(dur); err != nil {
			return nil, fmt.Errorf("%w: invalid SLA duration for state %s", domain.ErrWorkflowInvalid, state)
		}
	}
	if def.Reapproval != nil {
		if def.Reapproval.ToState == "" {
			return nil, fmt.Errorf("%w: reapproval.toState unkown", domain.ErrWorkflowInvalid)
		}
		if _, ok := seenStates[def.Reapproval.ToState]; !ok {
			return nil, fmt.Errorf("%w: reapproval.toState unkown", domain.ErrWorkflowInvalid)
		}
		if _, err := time.ParseDuration(def.Reapproval.After); err != nil {
			return nil, fmt.Errorf("%w: invalid reapproval.after duration", domain.ErrWorkflowInvalid)
		}
	}
	def.CreatedAt = time.Now().UTC()
	return &def, nil
}
