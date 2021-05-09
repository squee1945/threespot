package api

import (
	"net/http"

	"github.com/squee1945/threespot/server/pkg/game"
	"github.com/squee1945/threespot/server/pkg/util"
	"google.golang.org/appengine"
)

func (s *ApiServer) NewGame(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	if r.Method != "POST" {
		sendUserError(w, "Invalid method")
		return
	}

	player := s.lookupPlayer(ctx, w, r)
	if player == nil {
		return
	}

	id := util.RandString(7)
	g, err := game.NewGame(ctx, s.gameStore, s.playerStore, id, player, game.NewRules())
	if err != nil {
		sendServerError(w, "creating game: %v", err)
		return
	}

	s.sendGameState(ctx, w, g, player)
}
