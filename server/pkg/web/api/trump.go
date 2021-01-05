package web

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/squee1945/threespot/server/pkg/deck"
)

type CallTrumpRequest struct {
	ID   string
	Suit string
}

func (s *ApiServer) CallTrump(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	player := s.lookupPlayer(ctx, w, r)
	if player == nil {
		return
	}

	var req CallTrumpRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		sendServerError(w, "decoding request: %v", err)
		return
	}

	g := s.lookupGame(ctx, w, req.ID)
	if g == nil {
		return
	}

	suit, err := deck.NewSuitFromEncoded(req.Suit)
	if err != nil {
		sendServerError(w, "creating suit: %v", err)
		return
	}

	newG, err := g.CallTrump(ctx, player, suit)
	if err != nil {
		sendServerError(w, "calling trump: %v", err)
		return
	}

	sendGameState(ctx, w, req.ID, newG, player)
}
