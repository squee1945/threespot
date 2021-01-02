package web

import (
	"context"
	"net/http"

	"github.com/squee1945/threespot/server/pkg/game"
	"github.com/squee1945/threespot/server/pkg/util"
)

type NewGameResponse struct {
	ID string
}

func (s *Server) NewGame(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	player := s.lookupPlayer(ctx, w, r)
	if player == nil {
		return
	}

	id := util.RandString(7)
	g, err := game.NewGame(s.gameStore, id, player)
	if err != nil {
		sendServerError(w, "creating game: %v", err)
		return
	}

	if err := sendResponse(w, NewGameResponse{ID: g.ID()}); err != nil {
		sendServerError(w, "sending response: %v", err)
		return
	}
}
