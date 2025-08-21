package domain

import (
	"errors"
	"time"
)

var (
	ErrInvalidTransition = errors.New("invalid transition")
	ErrForbidden         = errors.New("forbidden: role not allowed for transition")
	ErrTerminalState     = errors.New("cannot transition from terminal")
	ErrWorkflowInvalid   = errors.New("workflow definition invalid")
)

type State struct {
	Name     string `json:"name"`
	Initial  bool   `json:"initial,omitempty"`
	Terminal bool   `json:"terminal,omitempty"`
}

type Transition struct {
	From  string   `jso:"from"`
	To    string   `jso:"to"`
	Event string   `json:"event"`
	Roles []string `json:"roles,omitempty"`
}

type Reapproval struct {
	After   string `json:"after"`
	ToState string `json:"to_state"`
}

type WorkflowDefinition struct {
	ID          string            `json:"id,omitempty"`
	Name        string            `json:"name,omitempty"`
	States      []State           `json:"states"`
	Transitions []Transition      `json:"transitions"`
	SLA         map[string]string `json:"sla,omitempty"`
	Reapproval  *Reapproval       `json:"reapproval,omitempty"`
	CreatedAt   time.Time         `json:"createt_ate,omitempty"`
}

type Document struct {
	ID         string
	WorkflowID string
	Current    string
	Data       map[string]any
	CreatedAt  time.Time
	UpdaterAt  time.Time
	DuaAt      *time.Time
	NextReapp  *time.Time
}
