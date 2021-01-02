package web

import (
	"context"
	"net/http"
	"strings"
)

func (s *Server) SetName(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	player := s.lookupPlayer(ctx, w, r)
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
		return
	}
	http.Redirect(w, r, "/", http.StatusFound)
}
