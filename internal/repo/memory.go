package repo

import (
	"fmt"
	"sync"

	"github.com/rcarvalho-pb/flowforge-go/internal/domain"
)

type Memory struct {
	mu        sync.RWMutex
	workflows map[string]*domain.WorkflowDefinition
	docs      map[string]*domain.Document
	seq       int64
}

func NewMemory() *Memory {
	return &Memory{
		workflows: map[string]*domain.WorkflowDefinition{},
		docs:      map[string]*domain.Document{},
	}
}

func (m *Memory) nextID(prefix string) string {
	m.seq++
	return fmt.Sprintf("%s_%d", prefix, m.seq)
}

func (m *Memory) SaveWorkflow(def *domain.WorkflowDefinition) string {
	m.mu.Lock()
	defer m.mu.Unlock()

	if def.ID == "" {
		def.ID = m.nextID("wf")
	}
	m.workflows[def.ID] = def
	return def.ID
}

func (m *Memory) GetWorkflow(id string) (*domain.WorkflowDefinition, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	def, ok := m.workflows[id]
	return def, ok
}

func (m *Memory) CreateDoc(d *domain.Document) string {
	m.mu.Lock()
	defer m.mu.Unlock()

	if d.ID == "" {
		d.ID = m.nextID("doc")
	}
	m.docs[d.ID] = d
	return d.ID
}

func (m *Memory) GetDoc(id string) (*domain.Document, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	d, ok := m.docs[id]
	return d, ok
}

func (m *Memory) UpdateDoc(d *domain.Document) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.docs[d.ID] = d
}
