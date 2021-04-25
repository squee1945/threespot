package api

import (
	"net/http"
	"strings"

	"github.com/squee1945/threespot/server/pkg/game"
	"google.golang.org/appengine"
)

type JoinStateResponse struct {
	ID          string
	Version     string
	PlayerNames []string
	State       string
}

func BuildJoinState(g game.Game) *JoinStateResponse {
	var names []string
	for _, p := range g.Players() {
		if p == nil {
			names = append(names, "")
			continue
		}
		names = append(names, p.Name())
	}
	return &JoinStateResponse{
		ID:          g.ID(),
		Version:     g.Version(),
		PlayerNames: names,
		State:       string(g.State()),
	}
}

func (s *ApiServer) JoinGameState(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	if r.Method != "GET" {
		sendUserError(w, "Invalid method")
		return
	}

	var id string
	if strings.HasPrefix(r.URL.Path, "/api/join-state/") {
		id = r.URL.Path[len("/api/join-state/"):]
	} else {
		sendUserError(w, "Missing ID")
		return
	}

	// Check If-None-Modified against a cache entry.
	if etag := r.Header.Get("If-None-Match"); etag != "" {
		current := s.getGameStateVersion(ctx, id)
		if current != "" && strings.Contains(etag, current) {
			w.WriteHeader(http.StatusNotModified)
			return
		}
	}

	player := s.lookupPlayer(ctx, w, r)
	if player == nil {
		return
	}

	g := s.lookupGame(ctx, w, id)
	if g == nil {
		return
	}

	s.sendJoinState(ctx, w, g)
}
