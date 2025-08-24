package main

import (
	"log"
	"net/http"

	"github.com/rcarvalho-pb/flowforge-go/internal/api"
	"github.com/rcarvalho-pb/flowforge-go/internal/engine"
	"github.com/rcarvalho-pb/flowforge-go/internal/jobs"
	"github.com/rcarvalho-pb/flowforge-go/internal/repo"
	"github.com/rcarvalho-pb/flowforge-go/internal/web"
)

func main() {
	r := repo.NewMemory()
	e := engine.New(r, jobs.NoopScheduler{})
	apiServer := &api.Server{Eng: e}
	webServer := &web.Server{Eng: e}
	mux := http.NewServeMux()
	mux.Handle("/static/", http.StripPrefix("web/static/", http.FileServer(http.Dir("./web/static"))))
	mux.Handle("/api/", http.StripPrefix("/api", apiServer.Router()))
	mux.Handle("/web/", http.StripPrefix("/web", webServer.Router()))
	log.Println("listening on port :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
