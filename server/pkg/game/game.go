package game

import (
	"context"
	"errors"
	"fmt"

	"github.com/squee1945/threespot/server/pkg/deck"
	"github.com/squee1945/threespot/server/pkg/storage"
)

var (
	ErrNotFound             = errors.New("Not found")
	ErrInvalidPosition      = errors.New("Invalid position")
	ErrPlayerPositionFilled = errors.New("Player position is already filled")
	ErrPlayerAlreadyAdded   = errors.New("Player is already added")
	ErrIncorrectBidOrder    = errors.New("Bidding out of order")
	ErrIncorrectPlayOrder   = errors.New("Playing out of order")
	ErrInvalidBid           = errors.New("Invalid bid")
	ErrNotBidding           = errors.New("Not currently bidding")
	ErrNotPlaying           = errors.New("Not currently playing cards")
	ErrMissingCard          = errors.New("Player does not have this card")
	ErrNotFollowingSuit     = errors.New("Must follow lead suit")
)

type GameState string

var (
	JoiningState  GameState = "JOINING"
	BiddingState  GameState = "BIDDING"
	PlayingState  GameState = "PLAYING"
	CompleteState GameState = "COMPLETE"
)

type Game interface {
	ID() string
	State() GameState
	Players() []Player
	PlacedBids() []Bid
	AvailableBids() []Bid
	PosToBid() int
	WinningBid() (string, int)
	PlayerHand(player Player) Hand

	AddPlayer(player Player, pos int) (Game, error)
	PlaceBid(player Player, bid Bid) (Game, error)
	PlayCard(player Player, card deck.Card) (Game, error)
}

type game struct {
	gameStore   storage.GameStore
	playerStore storage.PlayerStore

	id       string
	players  []Player // Position 0/2 are a team, 1/3 are a team; organizer is position 0.
	complete bool

	scoreToWin   int     // Either 52 or 62.
	score        []int   // 0-index is player 0/2 score; 1-index is player 1/3 score.
	scoreHistory [][]int // A list of pairs like above.

	currentDealerPos  int    // The position of the current dealer.
	currentBids       []Bid  // The bids for the current hand; empty string is 'pass'. The 0-index is the bid from the player clockwise from the currentDealerPos.
	currentWinningBid Bid    // The winning bid for the current hand.
	currentWinningPos int    // The position of the winning bidder (who will play first).
	currentHands      []Hand // Cards held by each player, parallel with the PlayerIDs above.
	currentTrick      Trick  // Cards played for current trick; 0-index is the lead player (i.e., the order the cards were played).
	currentTally      []int  // The running tally for the current hand; 0-index is player 0/2 tally; 1-index is player 1/3 tally.
}

func NewGame(ctx context.Context, gameStore storage.GameStore, playerStore storage.PlayerStore, id string, organizer Player) (Game, error) {
	if gs, err := gameStore.Create(ctx, id, organizer.ID()); err != nil {
		return Game{}, err
	}
	g, err := gameFromDatastore(ctx, gameStore, playerStore, id, gs)
	if err != nil {
		return Game{}, err
	}
	return g, nil
}

func GetGame(ctx context.Context, gameStore storage.GameStore, playerStore storage.PlayerStore, id string) (Game, error) {
	return nil, nil
	gs, err := gameStore.Get(ctx, id)
	if err != nil {
		if err == storage.ErrNotFound {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("fetching game from storage: %v", err)
	}
	g, err := gameFromDatastore(ctx, gameStore, playerStore, id, gs)
	if err != nil {
		return Game{}, err
	}
	return g, nil
}

func (g *game) ID() string {
	return g.id
}

func (g *game) State() GameState {
	if g.playerCount() < 4 {
		return JoiningState
	}
	if g.complete {
		return CompleteState
	}
	if g.currentWinningBid == nil {
		return BiddingState
	}
	return PlayingState
}

func (g *game) Players() []Player {
	return g.players
}

func (g *game) PlacedBids() []Bid {
	return g.currentBids
}

func (g *game) AvailableBids(Player player) []Bid {
	if len(g.currentBids) == 4 {
		return nil
	}
	pos, err := playerPos(player)
	if err != nil {
		return nil
	}
	return nextBids(g.currentBids, playerPos == g.currentDealerPos)
}

func (g *game) PosToBid() int {
	return (g.currentDealerPos + len(g.currentBids) + 1) % 4
}

func (g *game) WinningBid() Bid {
	return g.currentWinningBid
}

func (g *game) PlayerHand(player Player) Hand {
	pos, err := playerPos(player)
	if err != nil {
		return nil
	}
	return g.currentHands[pos]
}

func (g *game) AddPlayer(ctx context.Context, player Player, pos int) (Game, error) {
	gs, err := g.gameStore.AddPlayer(ctx, g.id, player.ID(), pos)
	if err != nil {
		if err == storage.ErrPlayerPositionFilled {
			return nil, ErrPlayerPositionFilled
		}
		if err == storage.ErrPlayerAlreadyAdded {
			return nil, ErrPlayerAlreadyAdded
		}
		return nil, fmt.Errorf("adding player: %v", err)
	}
	return gameFromDatastore(ctx, g.gameStore, g.playerStore, g.id, gs)
}

func (g *game) PlaceBid(ctx context.Context, player Player, bid Bid) (Game, error) {
	// Are we in bidding state?
	if g.State() != BiddingState {
		return nil, ErrNotBidding
	}

	// Can the player bid at the moment?
	pos, err := playerPos(player)
	if err != nil {
		return nil, err
	}
	if pos != g.PosToBid() {
		return nil, ErrIncorrectBidOrder
	}

	// Is bid in available bids?
	found := false
	for _, ab := range g.AvailableBids {
		if ab == bid {
			found = true
			break
		}
	}
	if !found {
		return nil, ErrInvalidBid
	}

	gs, err := g.gameStore.Get(ctx, g.id)
	if err != nil {
		return nil, fmt.Errorf("fetching game to update: %v", err)
	}
	gs.CurrentBids = append(gs.CurrentBids, string(Bid))

	if err = g.gameStore.Set(ctx, g.id, gs); err != nil {
		return nil, fmt.Errorf("saving game: %v", err)
	}
	return gameFromDatastore(ctx, g.gameStore, g.playerStore, g.id, gs)
}

func (g *game) PlayCard(player Player, card deck.Card) (Game, error) {
	// Are we in playing state?
	if g.State() != PlayingState {
		return nil, ErrNotPlaying
	}

	// Is it the player's turn?
	pos, err := playerPos(player)
	if err != nil {
		return nil, err
	}
	if g.trick.CurrentTurnPos() != pos {
		return nil, ErrIncorrectPlayOrder
	}

	playerHand := g.currentHands[pos]

	// Does the player have the card to play?
	if !playerHand.Contains(card) {
		return nil, ErrMissingCard
	}

	// Is the card a valid card to play (i.e., does it follow suit)?
	leadSuit := g.currentTrick.LeadSuit()
	if g.currentTrick.NumPlayed() > 0 && card.Suit() != leadSuit {
		if playerHand.ContainsSuit(leadSuit) {
			return nil, ErrNotFollowingSuit
		}
	}

	gs, err := g.gameStore.Get(ctx, g.id)
	if err != nil {
		return nil, fmt.Errorf("fetching game to update: %v", err)
	}

	// Remove the card from the player hand.
	newHand, err := hand.RemoveCard(card)
	if err != nil {
		return nil, err
	}
	gs.CurrentHands[pos] = newHand.Encoded()

	// Add the card to the current trick
	if err := g.currentTrick.PlayCard(pos, card); err != nil {
		return Game{}, err
	}
	gs.CurrentTrick = g.currentTrick.Encoded()

	// TODO: if the last card, compute the results
	if g.currentTrick.IsDone() {
		// Who won the trick?
		winningPos, err := g.currentTrick.WinningPos()
		if err != nil {
			return Game{}, err
		}
		// TODO: Adjust the tally.
		// TODO: If all cards are played, update the score.
		// TODO: If the score is a winning (note: bid out, etc.), complete the game.
		// TODO: Else deal a new hand and go to bidding round.
	}

	if err = g.gameStore.Set(ctx, g.id, gs); err != nil {
		return nil, fmt.Errorf("saving game: %v", err)
	}
	return gameFromDatastore(ctx, g.gameStore, g.playerStore, g.id, gs)
}

func (g *game) playerCount() int {
	c := 0
	for _, p := range g.players {
		if p != nil {
			c++
		}
	}
	return c
}

func (g *game) playerPos(player Player) (int, error) {
	for pos, p := range g.Players() {
		if p.ID() == player.ID() {
			return pos, nil
		}
	}
	return -1, errors.New("unknown player")
}

func gameFromDatastore(ctx context.Context, gameStore storage.GameStore, playerStore storage.PlayerStore, id string, gs storage.Game) (Game, error) {
	// Multi-get the players
	pss, err := playerStore.GetMulti(ctx, gs.PlayerIDs)
	if err != nil {
		return nil, fmt.Errorf("fetching player info from storage: %v", err)
	}
	players := make([]Player, 4)
	for i, ps := range pss {
		if ps == nil {
			continue
		}
		players[i] = playerFromStorage(playerStore, pss[i], ps)
	}

	var bids []Bid
	for _, bidStr := range gs.currentBids {
		bids = append(bids, NewBidFromEncoded(bs))
	}

	var hands []Hand
	for pos, encoded := range gs.CurrentHands {
		hand, err := NewHandFromEncoded(encoded)
		if err != nil {
			return nil, err
		}
	}
	for pos, cardStrs := range gs.CurrentHands {
		var cards []Card
		for _, cardStr := range cardStrs {
			cards = append(cards, deck.NewCardFromEncoded(cardStr))
		}
		hands = append(hands, NewHand(cards))
	}

	var winningBid Bid
	var trick Trick
	if gs.CurrentWinningBid != "" {
		winningBid = NewBidFromEncoded(gs.CurrentWinningBid)
		trick := NewTrickFromEncoded(gs.CurrentTrick)
	}

	g := Game{
		gameStore:         gameStore,
		playerStore:       playerStore,
		id:                id,
		players:           players,
		complete:          gs.Complete,
		scoreToWin:        gs.ScoreToWin,
		score:             gs.Score,
		scoreHistory:      gs.ScoreHistory,
		currentDealerPos:  gs.CurrentDealerPos,
		currentBids:       bids,
		currentWinningBid: winningBid,
		currentWinningPos: gs.CurrentWinningPos,
		currentHands:      hands,
		currentTrick:      trick,
		currentTally:      gs.CurrentTally,
	}
	return g, nil
}
