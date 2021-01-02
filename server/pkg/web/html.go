package web

import (
	"net/http"

	"github.com/squee1945/threespot/server/pkg/game"
	"github.com/squee1945/threespot/server/pkg/web/pages"
)

func (s *Server) Index(w http.ResponseWriter, r *http.Request) {
	playerID, err := playerID(r)
	var player game.Player
	if err != nil {
		if err == http.ErrNoCookie {
			playerID = setPlayerID(w)
			var err error
			player, err = game.NewPlayer(playerID, playerID)
			if err != nil {
				sendServerError(w, "creating player: %v", err)
				return
			}
		} else {
			sendServerError(w, "looking up player ID in cookie: %v", err)
			return
		}
	}

	args := pages.IndexArgs{
		PlayerID:   player.ID(),
		PlayerName: player.Name(),
	}
	if err := pages.IndexPage.Execute(w, args); err != nil {
		sendServerError(w, "rending index page: %v", err)
	}
}
