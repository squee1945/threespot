package api

import (
	"encoding/json"
	"net/http"

	"github.com/squee1945/threespot/server/pkg/game"
	"google.golang.org/appengine"
)

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

	s.sendJoinState(ctx, w, newG)
}
