package api

import (
	"encoding/json"
	"net/http"

	"google.golang.org/appengine"
)

type DealCardsRequest struct {
	ID string
}

func (s *ApiServer) DealCards(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	if r.Method != "POST" {
		sendUserError(w, "Invalid method")
		return
	}

	player := s.lookupPlayer(ctx, w, r)
	if player == nil {
		return
	}

	var req DealCardsRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		sendServerError(w, "decoding request: %v", err)
		return
	}

	g := s.lookupGame(ctx, w, req.ID)
	if g == nil {
		return
	}

	newG, err := g.DealCards(ctx, player)
	if err != nil {
		sendUserError(w, "dealing cards: %v", err)
	}

	s.sendGameState(ctx, w, newG, player)
}
