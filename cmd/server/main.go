package main

import (
	"log"
	"net/http"

	"github.com/rcarvalho-pb/flowforge-go/internal/api"
	"github.com/rcarvalho-pb/flowforge-go/internal/engine"
	"github.com/rcarvalho-pb/flowforge-go/internal/jobs"
	"github.com/rcarvalho-pb/flowforge-go/internal/repo"
)

func main() {
	r := repo.NewMemory()
	e := engine.New(r, jobs.NoopScheduler{})
	s := &api.Server{Eng: e}
	mux := http.NewServeMux()
	s.Routes(mux)
	fs := http.FileServer(http.Dir("./web"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))
	log.Println("listening on port :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
