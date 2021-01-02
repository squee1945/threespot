package web

import (
	"net/http"

	"github.com/squee1945/threespot/server/pkg/game"
)

type NewGameResponse struct {
	ID string
}

func (s *Server) NewGame(w http.ResponseWriter, r *http.Request) {
	player := lookupPlayer(w, r)
	if player == nil {
		return
	}

	g, err := game.NewGame(player)
	if err != nil {
		sendServerError(w, "creating game: %v", err)
		return
	}

	if err := sendResponse(w, NewGameResponse{ID: g.ID()}); err != nil {
		sendServerError(w, "sending response: %v", err)
		return
	}
}
