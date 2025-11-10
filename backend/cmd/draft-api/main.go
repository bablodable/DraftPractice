package main

import (
	"log"
	"net/http"

	"github.com/example/draftpractice/internal/draft"
	"github.com/example/draftpractice/internal/heroes"
	"github.com/example/draftpractice/internal/server"
)

func main() {
	if err := heroes.Init(); err != nil {
		log.Fatalf("failed to load heroes: %v", err)
	}

	draftService := draft.NewService()

	handler := server.NewHandler(server.RouterConfig{
		DraftService: draftService,
	})

	if err := http.ListenAndServe(":8080", handler); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
