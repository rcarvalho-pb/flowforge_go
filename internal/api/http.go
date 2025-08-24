package api

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/rcarvalho-pb/flowforge-go/internal/dsl"
	"github.com/rcarvalho-pb/flowforge-go/internal/engine"
)

type Server struct {
	Eng *engine.Engine
}

func (s *Server) Routes(mux *http.ServeMux) {
	mux.HandleFunc("/api/workflows", s.createWorkflow)
	mux.HandleFunc("/api/documents", s.createDocument)
	mux.HandleFunc("/api/documents/", s.documentActions)
}

func (s *Server) createWorkflow(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	// b, _ := io.ReadAll(r.Body)
	// defer r.Body.Close()
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	b := []byte(r.FormValue("workflow"))
	def, err := dsl.ParseDefinitionJSON(b)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	id, err := s.Eng.CreateWorkflow(def)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	json.NewEncoder(w).Encode(map[string]string{"id": id})
}

func (s *Server) createDocument(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	workflowId := r.FormValue("workflowId")
	rawData := r.FormValue("data")
	log.Println(workflowId)
	log.Println(rawData)
	var data map[string]any
	if err := json.Unmarshal([]byte(rawData), &data); err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	doc, err := s.Eng.CreateDocument(workflowId, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	json.NewEncoder(w).Encode(doc)
}

func (s *Server) documentActions(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/documents/"), "/")
	if len(parts) == 1 && r.Method == http.MethodGet {
		id := parts[0]
		doc, err := s.Eng.ApplyEvent(id, "", nil)
		_ = doc
		_ = err
		http.Error(w, "not implemented", http.StatusNotImplemented)
		return
	}

	if len(parts) == 2 && parts[1] == "events" && r.Method == http.MethodPost {
		id := parts[0]
		var req struct {
			Event string   `json:"event"`
			Roles []string `json:"roles"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		doc, err := s.Eng.ApplyEvent(id, req.Event, req.Roles)
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		json.NewEncoder(w).Encode(doc)
		return
	}
	http.Error(w, "not found", http.StatusNotFound)
}
