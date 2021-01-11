package web

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/squee1945/threespot/server/pkg/game"
	"github.com/squee1945/threespot/server/pkg/util"
	"github.com/squee1945/threespot/server/pkg/web/api"
	"google.golang.org/appengine"
)

func (s *Server) Debug(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	playerID, err := util.PlayerID(r)
	if err != nil {
		if err != http.ErrNoCookie {
			sendServerError(w, "looking up player ID in cookie: %v", err)
			return
		}
		playerID = util.SetPlayerID(w)
	}

	player, err := game.GetPlayer(ctx, s.playerStore, playerID)
	if err != nil && err != game.ErrNotFound {
		sendServerError(w, "creating player: %v", err)
		return
	}

	args := debugArgs{
		PlayerName: player.Name(),
	}

	if strings.HasPrefix(r.URL.Path, "/debug/") {
		id := r.URL.Path[len("/debug/"):]
		g, err := game.GetGame(ctx, s.gameStore, s.playerStore, id)
		if err != nil {
			sendServerError(w, "unknown game %s", id)
			return
		}
		state, err := api.BuildGameState(g, player)
		if err != nil {
			sendServerError(w, "building game state: %v", err)
			return
		}
		pretty, err := json.MarshalIndent(state, "", "    ")
		if err != nil {
			sendServerError(w, "marshalling game state: %v", err)
			return
		}
		args.PrettyGameState = string(pretty)
	}

	s.render("debug.html", w, args)
}

type debugArgs struct {
	PlayerName      string
	PrettyGameState string
}
