package web

import (
	"net/http"
	"strings"

	"github.com/squee1945/threespot/server/pkg/game"
	"github.com/squee1945/threespot/server/pkg/util"
	"google.golang.org/appengine"
)

func (s *Server) Game(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	if !strings.HasPrefix(r.URL.Path, "/game/") {
		http.Redirect(w, r, "/", 302)
		return
	}

	id := r.URL.Path[len("/game/"):]

	playerID, err := util.PlayerID(r)
	if err != nil {
		if err == http.ErrNoCookie {
			http.Redirect(w, r, "/", 302)
			return
		}
		sendServerError(w, "fetching player cookie: %v", err)
		return
	}

	_, err = game.GetPlayer(ctx, s.playerStore, playerID)
	if err != nil {
		if err == game.ErrNotFound {
			http.Redirect(w, r, "/", 302)
			return
		}
		sendServerError(w, "getting player: %v", err)
		return
	}

	_, err = game.GetGame(ctx, s.gameStore, s.playerStore, id)
	if err != nil {
		if err == game.ErrNotFound {
			// TODO: do something better?
			http.Redirect(w, r, "/", 302)
			return
		}
		sendServerError(w, "getting game: %v", err)
		return
	}

	// TODO: check that player is part of this game.

	s.render("game.html", w, gameArgs{ID: id})
}

type gameArgs struct {
	ID string
}
