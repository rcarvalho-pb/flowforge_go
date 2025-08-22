package main

import (
	"log"

	"github.com/rcarvalho-pb/flowforge-go/internal/dsl"
	"github.com/rcarvalho-pb/flowforge-go/internal/engine"
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

func main() {
	e, wfID := newEngineWithDef()
	doc, err := e.CreateDocument(wfID, map[string]any{"amount": 123})
	if err != nil {
		log.Fatal(err)
	}
	log.Println(doc)
}

func newEngineWithDef() (*engine.Engine, string) {
	r := repo.NewMemory()
	e := engine.New(r, jobs.NoopScheduler{})
	def, err := dsl.ParseDefinitionJSON(dslExample)
	if err != nil {
		log.Fatalf("parse dsl: %v", err)
	}
	id, err := e.CreateWorkflow(def)
	if err != nil {
		log.Fatalf("create workflow: %v", err)
	}
	return e, id
}
