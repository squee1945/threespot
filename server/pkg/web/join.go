package web

import (
	"html/template"
	"net/http"
	"strings"

	"github.com/squee1945/threespot/server/pkg/game"
	"github.com/squee1945/threespot/server/pkg/util"
	"google.golang.org/appengine"
)

func (s *Server) Join(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	if !strings.HasPrefix(r.URL.Path, "/join/") { // /join/ABC123
		http.Redirect(w, r, "/", 301)
		return
	}

	id := r.URL.Path[len("/join/"):]
	args := joinArgs{ID: id, Address: util.Address(r, id)}

	playerID, err := util.PlayerID(r)
	if err != nil {
		if err == http.ErrNoCookie {
			playerID = util.SetPlayerID(w)
			http.Redirect(w, r, r.URL.Path, 302)
			return
		} else {
			sendServerError(w, "looking up player ID in cookie: %v", err)
			return
		}
	}

	g, err := game.GetGame(ctx, s.gameStore, s.playerStore, id)
	if err != nil {
		if err == game.ErrNotFound {
			http.Redirect(w, r, "/?error=GAME_NOT_FOUND", 301)
			return
		}
		sendServerError(w, "getting game: %v", err)
		return
	}
	args.Players = g.Players()

	if g.State() != game.JoiningState {
		http.Redirect(w, r, "/game/"+id, 301)
		return
	}

	p, err := game.GetPlayer(ctx, s.playerStore, playerID)
	if err != nil {
		if err != game.ErrNotFound {
			sendServerError(w, "getting player: %v", err)
			return
		}
	}

	args.HasName = "false"
	if p != nil {
		args.PlayerName = p.Name()
		if args.PlayerName != "" {
			args.HasName = "true"
		}
	}

	s.render("join.html", w, args)
}

type joinArgs struct {
	ID         string
	Players    []game.Player
	PlayerName string
	HasName    template.JS // "true" or "false"
	Address    string
}
