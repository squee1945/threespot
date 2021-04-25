package api

import (
	"net/http"
	"strings"

	"github.com/squee1945/threespot/server/pkg/deck"
	"github.com/squee1945/threespot/server/pkg/game"
	"google.golang.org/appengine"
)

type BidInfo struct {
	Code  string
	Human string
}

type JoiningInfo struct{}

type BiddingInfo struct {
	PositionToPlay        int
	DealerPosition        int      // last bidder
	PlayerHand            []string // own hand
	LeadBidPosition       int
	BidsPlaced            []BidInfo
	AvailableBids         []BidInfo
	LastTrick             []string
	LastTrickLeadPosition int
}

type CallingInfo struct {
	PositionToPlay        int // who's current turn
	DealerPosition        int
	WinningBid            BidInfo
	LeadBidPosition       int
	BidsPlaced            []BidInfo
	PlayerHand            []string // own hand
	LastTrick             []string
	LastTrickLeadPosition int
}

type PlayingInfo struct {
	PositionToPlay        int
	DealerPosition        int
	WinningBid            BidInfo
	WinningBidPosition    int
	Trump                 string
	PlayerHand            []string
	Trick                 []string
	TrickLeadPosition     int
	LastTrick             []string
	LastTrickLeadPosition int
	TrickTally            []int
}

type CompletedInfo struct {
	WinningTeam int // 0 is players 0/2, 1 is players 1/3
	LastTrick   []string
}

type GameStateResponse struct {
	ID      string
	Version string
	State   string // "JOINING", "BIDDING", "CALLING", "PLAYING", "COMPLETED"

	PlayerPosition int // player's original position
	PlayerNames    []string
	Score          [][]int
	CurrentScore   []int
	ToWin          int

	JoiningInfo   *JoiningInfo   `json:omitempty`
	BiddingInfo   *BiddingInfo   `json:omitempty`
	CallingInfo   *CallingInfo   `json:omitempty`
	PlayingInfo   *PlayingInfo   `json:omitempty`
	CompletedInfo *CompletedInfo `json:omitempty`
}

func (s *ApiServer) GameState(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	if r.Method != "GET" {
		sendUserError(w, "Invalid method")
		return
	}

	var id string
	if strings.HasPrefix(r.URL.Path, "/api/state/") {
		id = r.URL.Path[len("/api/state/"):]
	} else {
		sendUserError(w, "Missing ID")
		return
	}

	// Check If-None-Modified against a cache entry.
	if etag := r.Header.Get("If-None-Match"); etag != "" {
		current := s.getGameStateVersion(ctx, id)
		if current != "" && current == etag {
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

	s.sendGameState(ctx, w, g, player)
}

func BuildGameState(g game.Game, player game.Player) (*GameStateResponse, error) {
	playerPos, err := g.PlayerPos(player)
	if err != nil {
		return nil, err
	}

	var playerNames []string
	for _, p := range g.Players() {
		if p == nil {
			playerNames = append(playerNames, "")
			continue
		}
		playerNames = append(playerNames, p.Name())
	}

	state := &GameStateResponse{
		ID:             g.ID(),
		Version:        g.Version(),
		State:          string(g.State()),
		PlayerPosition: playerPos,
		PlayerNames:    playerNames,
		Score:          g.Score().Scores(),
		CurrentScore:   g.Score().CurrentScore(),
		ToWin:          g.Score().ToWin(),
	}
	switch g.State() {
	case game.JoiningState:
		info, err := buildJoiningInfo(g, player)
		if err != nil {
			return nil, err
		}
		state.JoiningInfo = info
	case game.BiddingState:
		info, err := buildBiddingInfo(g, player, playerPos)
		if err != nil {
			return nil, err
		}
		state.BiddingInfo = info
	case game.CallingState:
		info, err := buildCallingInfo(g, player, playerPos)
		if err != nil {
			return nil, err
		}
		state.CallingInfo = info
	case game.PlayingState:
		info, err := buildPlayingInfo(g, player, playerPos)
		if err != nil {
			return nil, err
		}
		state.PlayingInfo = info
	case game.CompletedState:
		info, err := buildCompletedInfo(g, player)
		if err != nil {
			return nil, err
		}
		state.CompletedInfo = info
	}
	return state, nil
}

func buildJoiningInfo(g game.Game, player game.Player) (*JoiningInfo, error) {
	return &JoiningInfo{}, nil
}

func buildBiddingInfo(g game.Game, player game.Player, playerPos int) (*BiddingInfo, error) {
	positionToPlay, err := g.PosToPlay()
	if err != nil {
		return nil, err
	}
	var availableBids []BidInfo
	if playerPos == positionToPlay {
		available, err := g.AvailableBids(player)
		if err != nil {
			return nil, err
		}
		availableBids = bidsToBidInfos(available)
	}
	playerHand, err := g.PlayerHand(player)
	if err != nil {
		return nil, err
	}

	info := &BiddingInfo{
		PositionToPlay:  positionToPlay,
		DealerPosition:  g.DealerPos(),
		PlayerHand:      cardsToStrings(playerHand.Cards()),
		LeadBidPosition: g.CurrentBidding().LeadPos(),
		BidsPlaced:      bidsToBidInfos(g.CurrentBidding().Bids()),
		AvailableBids:   availableBids,
	}
	if g.LastTrick() != nil {
		info.LastTrick = cardsToStrings(g.LastTrick().Cards())
		info.LastTrickLeadPosition = g.LastTrick().LeadPos()
	}
	return info, nil
}

func buildCallingInfo(g game.Game, player game.Player, playerPos int) (*CallingInfo, error) {
	positionToPlay, err := g.PosToPlay()
	if err != nil {
		return nil, err
	}
	winningBid, _, err := g.CurrentBidding().WinningBidAndPos()
	if err != nil {
		return nil, err
	}
	playerHand, err := g.PlayerHand(player)
	if err != nil {
		return nil, err
	}

	info := &CallingInfo{
		PositionToPlay:  positionToPlay,
		DealerPosition:  g.DealerPos(),
		WinningBid:      bidToBidInfo(winningBid),
		LeadBidPosition: g.CurrentBidding().LeadPos(),
		BidsPlaced:      bidsToBidInfos(g.CurrentBidding().Bids()),
		PlayerHand:      cardsToStrings(playerHand.Cards()),
	}
	if g.LastTrick() != nil {
		info.LastTrick = cardsToStrings(g.LastTrick().Cards())
		info.LastTrickLeadPosition = g.LastTrick().LeadPos()
	}
	return info, nil
}

func buildPlayingInfo(g game.Game, player game.Player, playerPos int) (*PlayingInfo, error) {
	positionToPlay, err := g.PosToPlay()
	if err != nil {
		return nil, err
	}
	winningBid, winningBidPos, err := g.CurrentBidding().WinningBidAndPos()
	if err != nil {
		return nil, err
	}
	playerHand, err := g.PlayerHand(player)
	if err != nil {
		return nil, err
	}
	tally02, tally13 := g.Tally().Points()

	info := &PlayingInfo{
		PositionToPlay:     positionToPlay,
		DealerPosition:     g.DealerPos(),
		WinningBid:         bidToBidInfo(winningBid),
		WinningBidPosition: winningBidPos,
		TrickLeadPosition:  g.CurrentTrick().LeadPos(),
		Trump:              g.CurrentTrick().Trump().Encoded(),
		PlayerHand:         cardsToStrings(playerHand.Cards()),
		TrickTally:         []int{tally02, tally13},
	}
	if g.CurrentTrick() != nil {
		info.Trick = cardsToStrings(g.CurrentTrick().Cards())
	}
	if g.LastTrick() != nil {
		info.LastTrick = cardsToStrings(g.LastTrick().Cards())
		info.LastTrickLeadPosition = g.LastTrick().LeadPos()
	}
	return info, nil
}

func buildCompletedInfo(g game.Game, player game.Player) (*CompletedInfo, error) {
	score := g.Score().CurrentScore()
	winner := 0
	if score[1] > score[0] {
		winner = 2
	}
	info := &CompletedInfo{WinningTeam: winner}
	if g.LastTrick() != nil {
		info.LastTrick = cardsToStrings(g.LastTrick().Cards())
	}
	return info, nil
}

func bidToBidInfo(b game.Bid) BidInfo {
	return BidInfo{Code: b.Encoded(), Human: b.Human()}
}

func bidsToBidInfos(bids []game.Bid) []BidInfo {
	var res []BidInfo
	for _, b := range bids {
		res = append(res, bidToBidInfo(b))
	}
	return res
}

func cardsToStrings(cards []deck.Card) []string {
	var res []string
	for _, c := range cards {
		res = append(res, c.Encoded())
	}
	return res
}
