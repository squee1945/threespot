package api

import (
	"encoding/json"
	"net/http"

	"github.com/squee1945/threespot/server/pkg/game"
	"github.com/squee1945/threespot/server/pkg/util"
	"google.golang.org/appengine"
)

type NewGameRequest struct {
	PassCard bool
}

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

	var req NewGameRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendServerError(w, "decoding request: %v", err)
		return
	}

	rules := game.NewRules()
	rules.SetPassCard(req.PassCard)

	id := util.RandString(7)
	g, err := game.NewGame(ctx, s.gameStore, s.playerStore, id, player, rules)
	if err != nil {
		sendServerError(w, "creating game: %v", err)
		return
	}

	s.sendGameState(ctx, w, g, player)
}
