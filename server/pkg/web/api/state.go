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

type ScoreEntry struct {
	Score02 int
	Note02  string
	Score13 int
	Note13  string
}

type GameStateResponse struct {
	ID      string
	Version string
	State   string // "JOINING", "BIDDING", "CALLING", "PLAYING", "COMPLETED"

	PlayerPosition int // player's original position
	PlayerNames    []string
	Score          []ScoreEntry
	CurrentScore   []int
	ToWin          int
	WinningTeam    int // 0 is players 0/2, 1 is players 1/3; -1 is neither.

	DealerPosition           int // last bidder
	PlayerHand               []string
	HandCounts               []int
	Trick                    []string
	TrickLeadPosition        int
	LastTrick                []string
	LastTrickLeadPosition    int
	LastTrickWinningPosition int

	PositionToPlay  int
	LeadBidPosition int
	BidsPlaced      []BidInfo

	AvailableBids      []BidInfo
	WinningBid         BidInfo
	WinningBidPosition int
	Trump              string
	TrickTally         []int
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

	s.sendGameState(ctx, w, g, player)
}

func BuildGameState(g game.Game, player game.Player) (*GameStateResponse, error) {
	var playerNames []string
	for _, p := range g.Players() {
		if p == nil {
			playerNames = append(playerNames, "")
			continue
		}
		playerNames = append(playerNames, p.Name())
	}

	state := &GameStateResponse{
		ID:          g.ID(),
		Version:     g.Version(),
		State:       string(g.State()),
		PlayerNames: playerNames,
	}
	if g.State() == game.JoiningState {
		return state, nil
	}

	playerPos, err := g.PlayerPos(player)
	if err != nil {
		return nil, err
	}

	playerHand, err := g.PlayerHand(player)
	if err != nil {
		return nil, err
	}

	positionToPlay, err := g.PosToPlay()
	if err != nil {
		return nil, err
	}

	var scores []ScoreEntry
	for _, round := range g.Score().Scores() {
		score := ScoreEntry{Score02: round[0], Score13: round[1]}
		scores = append(scores, score)
	}
	for _, note := range g.Score().Notes() {
		if note.Index >= len(scores) {
			continue
		}
		if note.Team == 0 {
			scores[note.Index].Note02 = note.Note
			continue
		}
		if note.Team == 1 {
			scores[note.Index].Note13 = note.Note
		}
	}

	state = &GameStateResponse{
		ID:             g.ID(),
		Version:        g.Version(),
		State:          string(g.State()),
		PlayerPosition: playerPos,
		PlayerNames:    playerNames,
		Score:          scores,
		CurrentScore:   g.Score().CurrentScore(),
		ToWin:          g.Score().ToWin(),
		WinningTeam:    g.Score().Winner(),
		DealerPosition: g.DealerPos(),
		PlayerHand:     cardsToStrings(playerHand.Cards()),
		HandCounts:     g.HandCounts(),
		PositionToPlay: positionToPlay,
	}

	if g.CurrentTrick() != nil {
		state.Trick = cardsToStrings(g.CurrentTrick().Cards())
		state.TrickLeadPosition = g.CurrentTrick().LeadPos()
		state.Trump = g.CurrentTrick().Trump().Encoded()
	}

	if g.LastTrick() != nil {
		lastTrickWinningPos, err := g.LastTrick().WinningPos()
		if err != nil {
			return nil, err
		}
		state.LastTrick = cardsToStrings(g.LastTrick().Cards())
		state.LastTrickLeadPosition = g.LastTrick().LeadPos()
		state.LastTrickWinningPosition = lastTrickWinningPos
	}

	if g.CurrentBidding() != nil {
		state.LeadBidPosition = g.CurrentBidding().LeadPos()
		if g.State() != game.PlayingState {
			state.BidsPlaced = bidsToBidInfos(g.CurrentBidding().Bids())
		}
		if g.State() == game.CallingState || g.State() == game.PlayingState {
			winningBid, winningPos, err := g.CurrentBidding().WinningBidAndPos()
			if err != nil {
				return nil, err
			}
			state.WinningBid = bidToBidInfo(winningBid)
			state.WinningBidPosition = winningPos
		}
	}

	if g.Tally() != nil {
		tally02, tally13 := g.Tally().Points()
		state.TrickTally = []int{tally02, tally13}
	}

	if g.State() == game.BiddingState && playerPos == positionToPlay {
		available, err := g.AvailableBids(player)
		if err != nil {
			return nil, err
		}
		state.AvailableBids = bidsToBidInfos(available)
	}

	return state, nil
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
