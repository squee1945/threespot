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
	PositionToPlay int
	DealerPosition int
	PlayerHand     []string
	LeadBid        int
	BidsPlaced     []BidInfo
	AvailableBids  []BidInfo
}

type CallingInfo struct {
	PositionToPlay int
	DealerPosition int
	WinningBid     BidInfo
}

type PlayingInfo struct {
	PositionToPlay int
	DealerPosition int
	WinningBid     BidInfo
	WinningBidPos  int
	Trump          string
	PlayerHand     []string
	Trick          []string
}

type CompletedInfo struct {
	WinningTeam int // 0 is players 0/2, 1 is players 1/3
}

type GameStateResponse struct {
	ID      string
	Version int64
	State   string // "JOINING", "BIDDING", "CALLING", "PLAYING", "COMPLETED"

	PlayerPosition int
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

	player := s.lookupPlayer(ctx, w, r)
	if player == nil {
		return
	}

	g := s.lookupGame(ctx, w, id)
	if g == nil {
		return
	}

	sendGameState(ctx, w, id, g, player)
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
		PositionToPlay: positionToPlay,
		DealerPosition: g.DealerPos(),
		PlayerHand:     cardsToStrings(playerHand.Cards()),
		LeadBid:        g.CurrentBidding().LeadPos(),
		BidsPlaced:     bidsToBidInfos(g.CurrentBidding().Bids()),
		AvailableBids:  availableBids,
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

	info := &CallingInfo{
		PositionToPlay: positionToPlay,
		DealerPosition: g.DealerPos(),
		WinningBid:     bidToBidInfo(winningBid),
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

	info := &PlayingInfo{
		PositionToPlay: positionToPlay,
		DealerPosition: g.DealerPos(),
		WinningBid:     bidToBidInfo(winningBid),
		WinningBidPos:  winningBidPos,
		Trump:          g.CurrentTrick().Trump().Encoded(),
		PlayerHand:     cardsToStrings(playerHand.Cards()),
		Trick:          cardsToStrings(g.CurrentTrick().Cards()),
	}
	return info, nil
}

func buildCompletedInfo(g game.Game, player game.Player) (*CompletedInfo, error) {
	score := g.Score().CurrentScore()
	winner := 0
	if score[1] > score[0] {
		winner = 2
	}
	info := &CompletedInfo{
		WinningTeam: winner,
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
