package web

import (
	"context"
	"net/http"

	"github.com/squee1945/threespot/server/pkg/game"
	"github.com/squee1945/threespot/server/pkg/util"
)

func (s *ApiServer) NewGame(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	player := s.lookupPlayer(ctx, w, r)
	if player == nil {
		return
	}

	id := util.RandString(7)
	g, err := game.NewGame(ctx, s.gameStore, s.playerStore, id, player)
	if err != nil {
		sendServerError(w, "creating game: %v", err)
		return
	}

	sendGameState(ctx, w, id, g, player)
}
