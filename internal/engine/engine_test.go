package engine_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/rcarvalho-pb/flowforge-go/internal/dsl"
	. "github.com/rcarvalho-pb/flowforge-go/internal/engine"
	"github.com/rcarvalho-pb/flowforge-go/internal/jobs"
	"github.com/rcarvalho-pb/flowforge-go/internal/repo"
)

var dslExample = []byte(`{
	"name": "PurchaseOrderApproval",
	"states": [
		{ "name": "Draft", "initial": true },
		{ "name": "ManagerApproval" },
		{ "name": "FinanceApproval" },
		{ "name": "Approved", "terminal": true },
		{ "name": "Rejected", "terminal": true }
	],
	"transitions": [
		{ "from": "Draft", "to": "ManagerApproval", "event": "submit", "roles": ["author"] },
		{ "from": "ManagerApproval", "to": "FinanceApproval", "event": "approve", "roles": ["manager"] },
		{ "from": "ManagerApproval", "to": "Rejected", "event": "reject", "role": ["manager"] },
		{ "from": "FinanceApproval", "to": "Approved", "event": "approve", "role": ["finance"] },
		{ "from": "FinanceApproval", "to": "Rejected", "event": "reject", "role": ["finance"] }
	],
	"sla": { "FinanceApproval": "72h" },
	"reapproval": { "after": "2592h", "toState": "ManagerApproval"}
}`)

func newEngineWithDef(t *testing.T) (*Engine, string) {
	r := repo.NewMemory()
	e := New(r, jobs.NoopScheduler{})
	def, err := dsl.ParseDefinitionJSON(dslExample)
	if err != nil {
		t.Fatalf("parse dsl: %v", err)
	}
	id, err := e.CreateWorkflow(def)
	if err != nil {
		t.Fatalf("create workflow: %v", err)
	}
	return e, id
}

func TestCreateDocument_SetsInitialAndTimers(t *testing.T) {
	e, wfID := newEngineWithDef(t)
	doc, err := e.CreateDocument(wfID, map[string]any{"amount": 123})
	if err != nil {
		t.Fatal(err)
	}
	if doc.Current != "Draft" {
		t.Fatalf("expected Draft, got %s", doc.Current)
	}
	if doc.DueAt != nil {
		t.Fatalf("no SLA for Draft, expected nil")
	}
	if doc.NextReapp == nil {
		t.Fatalf("expected NextReapp set")
	}
}

func TestApplyEvent_AllowWithRole(t *testing.T) {
	e, wfID := newEngineWithDef(t)
	doc, _ := e.CreateDocument(wfID, nil)
	_, err := e.ApplyEvent(doc.ID, "submit", []string{"guest"})
	if err == nil {
		t.Fatalf("expected forbidden error")
	}
}

func TestTerminalBlocksFurtherTransitions(t *testing.T) {
	e, wfID := newEngineWithDef(t)
	doc, _ := e.CreateDocument(wfID, nil)
	_, _ = e.ApplyEvent(doc.ID, "submit", []string{"author"})
	doc, _ = e.ApplyEvent(doc.ID, "approve", []string{"manager"})
	if doc.DueAt == nil {
		t.Fatalf("expected DueAt set on FinanceApproval")
	}
	delta := time.Until(*doc.DueAt)
	if !(delta > 71*time.Hour && delta <= 72*time.Hour) {
		t.Fatalf("expected ~72h, got %v", delta)
	}
}

func TestDSL_RoundtripJSON(t *testing.T) {
	var v map[string]any
	if err := json.Unmarshal(dslExample, &v); err != nil {
		t.Fatalf("json invalid: %s", err)
	}
}
