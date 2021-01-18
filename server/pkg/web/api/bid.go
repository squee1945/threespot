package api

import (
	"encoding/json"
	"net/http"

	"github.com/squee1945/threespot/server/pkg/game"
	"google.golang.org/appengine"
)

type PlaceBidRequest struct {
	ID  string
	Bid string
}

func (s *ApiServer) PlaceBid(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	if r.Method != "POST" {
		sendUserError(w, "Invalid method")
		return
	}

	player := s.lookupPlayer(ctx, w, r)
	if player == nil {
		return
	}

	var req PlaceBidRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		sendServerError(w, "decoding request: %v", err)
		return
	}

	g := s.lookupGame(ctx, w, req.ID)
	if g == nil {
		return
	}

	bid, err := game.NewBidFromEncoded(req.Bid)
	if err != nil {
		sendServerError(w, "creating bid: %v", err)
		return
	}

	newG, err := g.PlaceBid(ctx, player, bid)
	if err != nil {
		// TODO: return user errors with better details for invalid bids
		sendServerError(w, "placing bid: %v", err)
		return
	}

	s.sendGameState(ctx, w, newG, player)
}
