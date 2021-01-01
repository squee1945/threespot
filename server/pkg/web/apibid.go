package web

import (
	"encoding/json"
	"net/http"

	"github.com/squee1945/threespot/server/pkg/game"
)

type PlaceBidRequest struct {
	ID  string
	Bid string
}

func PlaceBid(w http.ResponseWriter, r *http.Request) {
	player := lookupPlayer(w, r)
	if player == nil {
		return
	}

	var req PlaceBidRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		sendServerError(w, "decoding request: %v", err)
		return
	}

	g := lookupGame(w, req.ID)
	if g == nil {
		return
	}

	bid, err := game.NewBidFromString(req.Bid)
	if err != nil {
		sendServerError(w, "creating bid: %v", err)
		return
	}

	if err := g.PlaceBid(player, bid); err != nil {
		// TODO: return user errors with better details for invalid bids
		sendServerError(w, "placing bid: %v", err)
		return
	}

	sendGameState(w, req.ID, player)
}
