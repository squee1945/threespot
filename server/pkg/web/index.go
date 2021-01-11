package web

import (
	"net/http"

	"github.com/squee1945/threespot/server/pkg/game"
	"github.com/squee1945/threespot/server/pkg/util"
	"google.golang.org/appengine"
)

func (s *Server) Index(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	ctx := appengine.NewContext(r)

	// Give each visitor a unique random cookie ("player ID").
	playerID, err := util.PlayerID(r)
	if err != nil {
		if err != http.ErrNoCookie {
			sendServerError(w, "looking up player ID in cookie: %v", err)
			return
		}
		playerID = util.SetPlayerID(w)
	}

	var args indexArgs
	if player, err := game.GetPlayer(ctx, s.playerStore, playerID); err != nil {
		if err != game.ErrNotFound {
			sendServerError(w, "looking up player: %v", err)
			return
		}
		args.Welcome = "Welcome to Online Kaiser! You must enter a name to play."
	} else {
		args.Welcome = "Welcome back " + player.Name()
		args.Registered = true
	}

	s.render("index.html", w, args)
}

type indexArgs struct {
	Welcome    string
	Registered bool
}
