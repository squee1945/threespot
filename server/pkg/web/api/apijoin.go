package web

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/squee1945/threespot/server/pkg/game"
)

type JoinGameRequest struct {
	ID       string
	Position int
}

type JoinGameResponse struct {
	Players []GamePlayer
}

type GamePlayer struct {
	Name string
}

func (s *Server) JoinGame(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
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

	g := lookupGame(w, req.ID)
	if g == nil {
		return
	}

	if err := g.AddPlayer(player, req.Position); err != nil {
		if err == game.ErrInvalidPosition {
			sendUserError(w, "Invalid player position.")
			return
		}
		sendServerError(w, "adding player: %v", err)
		return
	}

	g = lookupGame(w, req.ID)
	if g == nil {
		return
	}

	var resp JoinGameResponse
	resp.Players = make([]GamePlayer, 4)
	for i, p := range g.Players() {
		if p == nil {
			continue
		}
		resp.Players[i].Name = p.Name()
	}

	if err := sendResponse(w, resp); err != nil {
		sendServerError(w, "sending response: %v", err)
		return
	}
}
