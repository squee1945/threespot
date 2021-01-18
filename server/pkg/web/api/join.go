package api

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/squee1945/threespot/server/pkg/game"
	"google.golang.org/appengine"
)

type JoinGameStateRequest struct {
	ID string
}

type JoinGameStateResponse struct {
	PlayerNames []string
	State       string
}

func (s *ApiServer) JoinGameState(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	if r.Method != "GET" {
		sendUserError(w, "Invalid method")
		return
	}

	var id string
	if strings.HasPrefix(r.URL.Path, "/api/join-state/") {
		id = r.URL.Path[len("/api/join-state/"):]
	} else {
		sendUserError(w, "Missing ID")
		return
	}

	// Check If-None-Modified against a cache entry.
	if etag := r.Header.Get("If-None-Match"); etag != "" {
		current := s.getGameStateVersion(ctx, id)
		if current != "" && current == etag {
			w.WriteHeader(http.StatusNotModified)
			return
		}
	}

	player := s.lookupPlayer(ctx, w, r)
	if player == nil {
		return
	}

	g := s.lookupGame(ctx, w, id)
	if g == nil {
		return
	}

	var names []string
	for _, p := range g.Players() {
		if p == nil {
			names = append(names, "")
			continue
		}
		names = append(names, p.Name())
	}

	resp := JoinGameStateResponse{
		PlayerNames: names,
		State:       string(g.State()),
	}

	log.Printf("Sending %#v\n", resp)
	if err := sendResponse(w, resp); err != nil {
		sendServerError(w, "sending response: %v", err)
	}
}

type JoinGameRequest struct {
	ID       string
	Position int
}

func (s *ApiServer) JoinGame(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	if r.Method != "POST" {
		sendUserError(w, "Invalid method")
		return
	}

	player := s.lookupPlayer(ctx, w, r)
	if player == nil {
		return
	}

	var req JoinGameRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		sendServerError(w, "decoding request: %v", err)
		return
	}

	g := s.lookupGame(ctx, w, req.ID)
	if g == nil {
		return
	}

	newG, err := g.AddPlayer(ctx, player, req.Position)
	if err != nil {
		if err == game.ErrInvalidPosition {
			sendUserError(w, "Invalid player position.")
			return
		}
		if err == game.ErrPlayerAlreadyAdded {
			sendUserError(w, "You're already in this game!")
			return
		}
		sendServerError(w, "adding player: %v", err)
		return
	}

	s.sendGameState(ctx, w, req.ID, newG, player)
}
