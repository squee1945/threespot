package web

import (
	"net/http"

	"github.com/squee1945/threespot/server/pkg/game"
	"github.com/squee1945/threespot/server/pkg/util"
	"google.golang.org/appengine"
)

func (s *Server) Index(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		// Assume the rest of the url is the game id, and redirect to '/game/XXX'
		http.Redirect(w, r, "/game"+r.URL.Path, 302)
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

	player, err := game.GetPlayer(ctx, s.playerStore, playerID)
	if err != nil {
		if err != game.ErrNotFound {
			sendServerError(w, "looking up player: %v", err)
			return
		}
		args := indexArgs{
			Welcome: "Welcome to Online Kaiser! You must enter a name to play.",
		}
		s.render("index.html", w, args)
		return
	}

	args := indexArgs{
		Welcome:    "Welcome back " + player.Name(),
		Registered: true,
	}

	games, err := game.GetCurrentGames(ctx, s.gameStore, s.playerStore, playerID, 10)
	if err != nil {
		sendServerError(w, "lookup up current games: %v", err)
		return
	}
	for _, g := range games {
		gi := gameInfo{
			ID:    g.ID(),
			Score: g.Score().CurrentScore(),
		}
		for _, player := range g.Players() {
			name := "?"
			if player != nil {
				name = player.Name()
			}
			gi.PlayerNames = append(gi.PlayerNames, name)
		}
		args.CurrentGames = append(args.CurrentGames, gi)
	}
	s.render("index.html", w, args)
}

type indexArgs struct {
	Welcome      string
	Registered   bool
	CurrentGames []gameInfo
}

type gameInfo struct {
	ID          string
	PlayerNames []string
	Score       []int
}
