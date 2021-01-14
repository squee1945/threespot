package web

import (
	"log"
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
			http.Redirect(w, r, "/", 301)
			return
		}
		sendServerError(w, "fetching player cookie: %v", err)
		return
	}

	player, err := game.GetPlayer(ctx, s.playerStore, playerID)
	if err != nil {
		if err == game.ErrNotFound {
			http.Redirect(w, r, "/", 301)
			return
		}
		sendServerError(w, "getting player: %v", err)
		return
	}

	g, err := game.GetGame(ctx, s.gameStore, s.playerStore, id)
	if err != nil {
		if err == game.ErrNotFound {
			// TODO: do something better?
			http.Redirect(w, r, "/", 301)
			return
		}
		sendServerError(w, "getting game: %v", err)
		return
	}

	if g.State() == game.JoiningState {
		http.Redirect(w, r, "/join/"+id, 301)
		return
	}

	// Check that player is part of this game.
	inGame := false
	for _, p := range g.Players() {
		if player.ID() == p.ID() {
			inGame = true
			break
		}
	}
	if !inGame {
		log.Printf("Player %q is not in game %q", player.ID(), id)
		http.Redirect(w, r, "/?error=NOT_IN_GAME", 301)
		return
	}

	s.render("game.html", w, gameArgs{ID: id})
}

type gameArgs struct {
	ID string
}
