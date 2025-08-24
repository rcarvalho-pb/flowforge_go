package web

import (
	"net/http"

	"github.com/rcarvalho-pb/flowforge-go/internal/engine"
)

type Server struct {
	Eng *engine.Engine
}

func (s *Server) Router() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/workflows", s.createWorkflow)
	mux.HandleFunc("/documents", s.createDocument)
	mux.HandleFunc("/documents/{id}/events", s.documentActions)

	return mux
}
