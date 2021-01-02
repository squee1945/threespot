package web

import (
	"context"
	"html/template"
	"net/http"

	"github.com/squee1945/threespot/server/pkg/game"
)

func (s *Server) Index(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	playerID, err := playerID(r)
	if err != nil {
		if err == http.ErrNoCookie {
			playerID = setPlayerID(w)
		} else {
			sendServerError(w, "looking up player ID in cookie: %v", err)
			return
		}
	}

	player, err := game.GetPlayer(ctx, s.PlayerStore, playerID)
	if err != nil {
		if err == game.ErrNotFound {
			var err error
			player, err = game.NewPlayer(ctx, s.PlayerStore, playerID, playerID)
			if err != nil {
				sendServerError(w, "creating player: %v", err)
				return
			}
		} else {
			sendServerError(w, "looking up player: %v", err)
			return
		}
	}

	args := indexArgs{
		PlayerID:   player.ID(),
		PlayerName: player.Name(),
	}
	if err := indexPage.Execute(w, args); err != nil {
		sendServerError(w, "rending index page: %v", err)
	}
}

type indexArgs struct {
	PlayerID   string
	PlayerName string
}

var indexTemplateStr = `
<html>
<head><title>Kaiser</title></head>
<body>
<h1>Kaiser</h1>
<p>Welcome back {{.PlayerName}} (ID: {{.PlayerID}})</p>
<p>
  <form action='/setname' method='post'>
  Set your name: <input name='newname'>
  <br>
  <input type='submit'>
  </form>
</p>
</body>
</html>
`

var indexPage = template.Must(template.New("index").Parse(indexTemplateStr))
