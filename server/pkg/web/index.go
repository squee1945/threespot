package web

import (
	"context"
	"html/template"
	"net/http"

	"github.com/squee1945/threespot/server/pkg/game"
)

func (s *Server) Index(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	// Give each visitor a unique random cookie ("player ID").
	playerID, err := playerID(r)
	if err != nil {
		if err != http.ErrNoCookie {
			sendServerError(w, "looking up player ID in cookie: %v", err)
			return
		}
		playerID = setPlayerID(w)
	}

	var args indexArgs
	if player, err := game.GetPlayer(ctx, s.PlayerStore, playerID); err != nil {
		if err != game.ErrNotFound {
			sendServerError(w, "looking up player: %v", err)
			return
		}
		args.Welcome = "Welcome to Online Kaiser! You must enter a name to play."
	} else {
		args.Welcome = "Welcome back " + player.Name()
		args.Registered = true
	}

	if err := indexPage.Execute(w, args); err != nil {
		sendServerError(w, "rending index page: %v", err)
	}
}

type indexArgs struct {
	Welcome    string
	Registered bool
}

var indexTemplateStr = `
<html>
<head><title>Online Kaiser</title></head>
<body>
<h1>Kaiser</h1>
<p>{{.Welcome}}</p>
{{if .Registered }}
	<p>TODO: Show form to start new game, or join existing game.</p>
	<p>TODO: Show links to active games.</p>
{{ else }}
	<p>TODO: Show form to set user info.</p>
{{ end }}
</body>
</html>
`

var indexPage = template.Must(template.New("index").Parse(indexTemplateStr))
