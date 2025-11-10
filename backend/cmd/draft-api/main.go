package main

import (
	"log"
	"net/http"
	"time"

	"github.com/example/draftpractice/internal/draft"
	"github.com/example/draftpractice/internal/heroes"
	"github.com/example/draftpractice/internal/server"
)

func main() {
	heroCatalog := heroes.NewCatalog(nil, "", 30*time.Minute)
	draftService := draft.NewService(heroCatalog)

	handler := server.NewHandler(server.RouterConfig{
		DraftService: draftService,
		HeroSource:   heroCatalog,
	})

	if err := http.ListenAndServe(":8080", handler); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
