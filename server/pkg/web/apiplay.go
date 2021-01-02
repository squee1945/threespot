package web

import (
	"encoding/json"
	"net/http"

	"github.com/squee1945/threespot/server/pkg/deck"
)

type PlayCardRequest struct {
	ID   string
	Card string
}

func (s *Server) PlayCard(w http.ResponseWriter, r *http.Request) {
	player := lookupPlayer(w, r)
	if player == nil {
		return
	}

	var req PlayCardRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		sendServerError(w, "decoding request: %v", err)
		return
	}

	g := lookupGame(w, req.ID)
	if g == nil {
		return
	}

	card, err := deck.NewCardFromString(req.Card)
	if err != nil {
		sendServerError(w, "creating card: %v", err)
		return
	}

	if err := g.PlayCard(player, card); err != nil {
		sendServerError(w, "playing card: %v", err)
		return
	}

	sendGameState(w, req.ID, player)
}
