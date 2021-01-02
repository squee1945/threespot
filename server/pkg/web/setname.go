package web

import (
	"context"
	"net/http"
	"strings"
)

func (s *Server) SetName(w http.ResponseWriter, r *http.Request) {
	context := context.Background()
	player := lookupPlayer(ctx, w, r)
	if player == nil {
		return
	}

	n := strings.TrimSpace(r.FormValue("newname"))
	if n == "" {
		sendUserError(w, "Name is required.")
		return
	}

	if err := player.SetName(ctx, n); err != nil {
		sendServerError(w, "setting player name: %v", err)
	}
	http.Redirect(w, r, "/", http.StatusFound)
}
