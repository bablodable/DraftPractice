package main

import (
	"log"
	"net/http"

	"github.com/example/draftpractice/internal/draft"
	"github.com/example/draftpractice/internal/server"
)

func main() {
	draftService := draft.NewService()

	handler := server.NewHandler(server.RouterConfig{
		DraftService: draftService,
	})

	if err := http.ListenAndServe(":8080", handler); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
