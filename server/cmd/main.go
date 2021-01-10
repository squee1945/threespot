package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"cloud.google.com/go/datastore"
	"github.com/squee1945/threespot/server/pkg/storage"
	"github.com/squee1945/threespot/server/pkg/web"
)

const (
	defaultPort = "8080"
)

func main() {
	port, ok := os.LookupEnv("PORT")
	if !ok {
		port = defaultPort
	}

	ctx := context.Background()
	dsClient, err := datastore.NewClient(ctx, datastore.DetectProjectID)
	if err != nil {
		log.Fatalf("Failed to create datastore client: %v", err)
	}

	server := &web.Server{
		PlayerStore: storage.NewDatastorePlayerStore(dsClient),
	}

	// Pages for humans.
	http.HandleFunc("/", server.Index)
	http.HandleFunc("/game/", server.Game)

	// Pages for machines.
	http.HandleFunc("/api/user", server.UpdateUser)
	http.HandleFunc("/api/new", server.NewGame)
	http.HandleFunc("/api/join", server.JoinGame)
	http.HandleFunc("/api/bid", server.PlaceBid)
	http.HandleFunc("/api/trump", server.CallTrump)
	http.HandleFunc("/api/play", server.PlayCard)
	http.HandleFunc("/api/state", server.GameState)

	s := &http.Server{
		Addr:    ":" + port,
		Handler: nil,
		// ReadTimeout:    10 * time.Second,
		// WriteTimeout:   10 * time.Second,
		// MaxHeaderBytes: 1 << 20,
	}
	log.Fatal(s.ListenAndServe())
}
