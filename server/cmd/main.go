package main

import (
	"log"
	"net/http"
	"os"

	"cloud.google.com/go/datastore"
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

	dsClient, err := datastore.NewClient(ctx, "my-project")
	if err != nil {
		log.Fatalf("Failed to create datastore client: %v", err)
	}

	server := web.Server{
		playerStore: storage.NewDatastorePlayerStore(dsClient),
	}

	http.Handle("/", server.Index)
	http.Handle("/setname", server.SetName)
	http.Handle("/new", server.NewGame)
	http.Handle("/join", server.JoinGame)
	http.Handle("/bid", server.PlaceBid)
	http.Handle("/play", server.PlayCard)
	http.Handle("/game", server.GameState)

	s := &http.Server{
		Addr:    ":" + port,
		Handler: nil,
		// ReadTimeout:    10 * time.Second,
		// WriteTimeout:   10 * time.Second,
		// MaxHeaderBytes: 1 << 20,
	}
	log.Fatal(s.ListenAndServe())
}
