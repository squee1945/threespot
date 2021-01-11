package main

import (
	"log"
	"net/http"

	"github.com/squee1945/threespot/server/pkg/storage"
	"github.com/squee1945/threespot/server/pkg/web"
	"github.com/squee1945/threespot/server/pkg/web/api"
	"google.golang.org/appengine"
)

func main() {
	playerStore := storage.NewDatastorePlayerStore()
	gameStore := storage.NewDatastoreGameStore()
	server, err := web.NewServer(gameStore, playerStore)
	if err != nil {
		log.Fatal(err)
	}
	apiServer := api.NewServer(gameStore, playerStore)

	// Pages for humans.
	http.HandleFunc("/", server.Index)
	http.HandleFunc("/game/", server.Game)
	// http.HandleFunc("/debug", server.Debug)
	// http.HandleFunc("/debug/", server.Debug)

	// Pages for machines.
	http.HandleFunc("/api/user", apiServer.UpdateUser)
	http.HandleFunc("/api/new", apiServer.NewGame)
	http.HandleFunc("/api/join", apiServer.JoinGame)
	http.HandleFunc("/api/bid", apiServer.PlaceBid)
	http.HandleFunc("/api/trump", apiServer.CallTrump)
	http.HandleFunc("/api/play", apiServer.PlayCard)
	http.HandleFunc("/api/state/", apiServer.GameState)

	appengine.Main()
}
