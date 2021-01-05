package web

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/squee1945/threespot/server/pkg/deck"
)

type PlayCardRequest struct {
	ID   string
	Card string
}

func (s *ApiServer) PlayCard(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	player := s.lookupPlayer(ctx, w, r)
	if player == nil {
		return
	}

	var req PlayCardRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		sendServerError(w, "decoding request: %v", err)
		return
	}

	g := s.lookupGame(ctx, w, req.ID)
	if g == nil {
		return
	}

	card, err := deck.NewCardFromEncoded(req.Card)
	if err != nil {
		sendServerError(w, "creating card: %v", err)
		return
	}

	newG, err := g.PlayCard(ctx, player, card)
	if err != nil {
		sendServerError(w, "playing card: %v", err)
		return
	}

	sendGameState(ctx, w, req.ID, newG, player)
}
