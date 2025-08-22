package engine

import (
	"errors"
	"time"

	"github.com/rcarvalho-pb/flowforge-go/internal/domain"
	"github.com/rcarvalho-pb/flowforge-go/internal/jobs"
	"github.com/rcarvalho-pb/flowforge-go/internal/repo"
)

type Engine struct {
	repo      *repo.Memory
	scheduler jobs.Scheduler
}

type compiled struct {
	initial     string
	terminal    map[string]bool
	byFromEvent map[string]map[string]domain.Transition
	sla         map[string]time.Duration
	reappAfter  *time.Duration
	reappTo     string
}

func New(r *repo.Memory, s jobs.Scheduler) *Engine {
	return &Engine{
		repo:      r,
		scheduler: s,
	}
}

func compile(def *domain.WorkflowDefinition) (*compiled, error) {
	c := &compiled{
		terminal:    map[string]bool{},
		byFromEvent: map[string]map[string]domain.Transition{},
		sla:         map[string]time.Duration{},
	}
	for _, st := range def.States {
		if st.Initial {
			c.initial = st.Name
		}
		if st.Terminal {
			c.terminal[st.Name] = true
		}
	}
	for _, tr := range def.Transitions {
		if c.byFromEvent[tr.From] == nil {
			c.byFromEvent[tr.From] = map[string]domain.Transition{}
		}
		c.byFromEvent[tr.From][tr.Event] = tr
	}
	for state, dur := range def.SLA {
		d, _ := time.ParseDuration(dur)
		c.sla[state] = d
	}
	if def.Reapproval != nil {
		d, _ := time.ParseDuration(def.Reapproval.After)
		c.reappAfter = &d
		c.reappTo = def.Reapproval.ToState
	}
	if c.initial == "" {
		return nil, domain.ErrWorkflowInvalid
	}
	return c, nil
}

func (e *Engine) CreateWorkflow(def *domain.WorkflowDefinition) (string, error) {
	if _, err := compile(def); err != nil {
		return "", err
	}
	return e.repo.SaveWorkflow(def), nil
}

func (e *Engine) CreateDocument(workflowID string, data map[string]any) (*domain.Document, error) {
	def, ok := e.repo.GetWorkflow(workflowID)
	if !ok {
		return nil, errors.New("workflow not found")
	}

	c, err := compile(def)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	doc := &domain.Document{
		WorkflowID: workflowID,
		Current:    c.initial,
		Data:       data,
		CreatedAt:  now,
		UpdaterAt:  now,
	}
	if due, ok := c.sla[doc.Current]; ok {
		dt := now.Add(due)
		doc.DueAt = &dt
	}
	if c.reappAfter != nil {
		nr := now.Add(*c.reappAfter)
		doc.NextReapp = &nr
	}
	e.repo.CreateDoc(doc)
	return doc, nil
}

func (e *Engine) ApplyEvent(docID, event string, actorRoles []string) (*domain.Document, error) {
	doc, ok := e.repo.GetDoc(docID)
	if !ok {
		return nil, errors.New("document not found")
	}
	def, ok := e.repo.GetWorkflow(doc.WorkflowID)
	if !ok {
		return nil, errors.New("workflow not found")
	}
	c, err := compile(def)
	if err != nil {
		return nil, err
	}
	if c.terminal[doc.Current] {
		return nil, domain.ErrTerminalState
	}
	trs := c.byFromEvent[doc.Current]
	if trs == nil {
		return nil, domain.ErrInvalidTransition
	}
	tr, ok := trs[event]
	if !ok {
		return nil, domain.ErrInvalidTransition
	}
	if len(tr.Roles) > 0 && !hasAny(actorRoles, tr.Roles) {
		return nil, domain.ErrForbidden
	}
	doc.Current = tr.To
	now := time.Now().UTC()
	doc.UpdaterAt = now
	doc.DueAt = nil
	if due, ok := c.sla[doc.Current]; ok {
		dt := now.Add(due)
		doc.DueAt = &dt
	}
	if c.reappAfter != nil {
		nr := now.Add(*c.reappAfter)
		doc.NextReapp = &nr
	}
	e.repo.UpdateDoc(doc)
	return doc, nil
}

func hasAny(have, want []string) bool {
	set := map[string]struct{}{}
	for _, r := range have {
		set[r] = struct{}{}
	}
	for _, r := range want {
		if _, ok := set[r]; ok {
			return true
		}
	}
	return false
}
