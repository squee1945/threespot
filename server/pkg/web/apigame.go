package web

import (
	"encoding/json"
	"net/http"

	"github.com/squee1945/threespot/server/pkg/deck"
	"github.com/squee1945/threespot/server/pkg/game"
)

type GameStateRequest struct {
	ID string
}

type GameStateResponse struct {
	State string // "BIDDING", "PLAYING"

	Bids               []string
	AvailableBids      []string
	PositionToBid      int
	WinningBid         string
	WinningBidPosition int

	PlayedCards []string
	HeldCards   []string
	// TODO: add players, score, current bid, current trick, trick count
}

func (s *Server) GameState(w http.ResponseWriter, r *http.Request) {
	player := lookupPlayer(w, r)
	if player == nil {
		return
	}

	var req GameStateRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		sendServerError(w, "decoding request: %v", err)
		return
	}

	sendGameState(w, req.ID, player)
}

func bidsToStrings(bids []game.Bid) []string {
	var res []string
	for _, b := range bids {
		res = append(res, string(b))
	}
	return res
}

func cardsToStrings(cards []deck.Card) []string {
	var res []string
	for _, c := range cards {
		res = append(res, string(c))
	}
	return res
}

func buildGameState(game game.Game, player game.Player) GameStateResponse {
	winningBid, winningBidPosition := game.WinningBid()
	return GameStateResponse{
		State:              string(game.State()),
		Bids:               bidsToStrings(game.PlacedBids()),
		AvailableBids:      bidsToStrings(game.AvailableBids()),
		PositionToBid:      game.PosToBid(),
		WinningBid:         string(winningBid),
		WinningBidPosition: winningBidPosition,
		PlayedCards:        cardsToStrings(game.PlayerHand(player).PlayedCards()),
		HeldCards:          cardsToStrings(game.PlayerHand(player).HeldCards()),
	}
}
