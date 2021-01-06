package game

import (
	"context"
	"errors"
	"fmt"

	"github.com/squee1945/threespot/server/pkg/deck"
	"github.com/squee1945/threespot/server/pkg/storage"
)

type Game interface {
	ID() string
	State() GameState
	Players() []Player
	PlayerHand(player Player) Hand
	Score() Score
	PlayerPos(player Player) (int, error)
	DealerPos() int
	PosToPlay() (int, error)

	CurrentBidding() BiddingRound
	CurrentTrick() Trick
	AvailableBids(Player) []Bid

	AddPlayer(ctx context.Context, player Player, pos int) (Game, error)
	PlaceBid(ctx context.Context, player Player, bid Bid) (Game, error)
	CallTrump(ctx context.Context, player Player, trump deck.Suit) (Game, error)
	PlayCard(ctx context.Context, player Player, card deck.Card) (Game, error)
}

var (
	ErrNotFound             = errors.New("Not found")
	ErrInvalidPosition      = errors.New("Invalid position")
	ErrPlayerPositionFilled = errors.New("Player position is already filled")
	ErrPlayerAlreadyAdded   = errors.New("Player is already added")
	ErrIncorrectBidOrder    = errors.New("Bidding out of order")
	ErrIncorrectPlayOrder   = errors.New("Playing out of order")
	ErrInvalidBid           = errors.New("Invalid bid")
	ErrNotBidding           = errors.New("Not currently bidding")
	ErrNotCalling           = errors.New("Not currently calling trump")
	ErrNotPlaying           = errors.New("Not currently playing cards")
	ErrMissingCard          = errors.New("Player does not have this card")
	ErrNotFollowingSuit     = errors.New("Must follow lead suit")
)

type GameState string

var (
	JoiningState   GameState = "JOINING"
	BiddingState   GameState = "BIDDING"
	CallingState   GameState = "CALLING" // Trump
	PlayingState   GameState = "PLAYING"
	CompletedState GameState = "COMPLETED"
)

type game struct {
	gameStore   storage.GameStore
	playerStore storage.PlayerStore

	id       string
	players  []Player // Position 0/2 are a team, 1/3 are a team; organizer is position 0.
	complete bool

	score Score // The score of the game.

	currentDealerPos int          // The position of the current dealer.
	currentBidding   BiddingRound // The bids for the current hand. The 0-index is the bid from the player clockwise from the currentDealerPos.
	currentHands     []Hand       // Cards held by each player, parallel with the PlayerIDs above.
	currentTrick     Trick        // Cards played for current trick; 0-index is the lead player (i.e., the order the cards were played).
	currentTally     Tally        // The running tally for the current hand.
}

var _ Game = (*game)(nil) // Ensure interface is implemented.

func NewGame(ctx context.Context, gameStore storage.GameStore, playerStore storage.PlayerStore, id string, organizer Player) (Game, error) {
	gs, err := gameStore.Create(ctx, id, organizer.ID())
	if err != nil {
		return nil, err
	}
	g, err := gameFromDatastore(ctx, gameStore, playerStore, id, gs)
	if err != nil {
		return nil, err
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
		return nil, err
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
		return CompletedState
	}
	if !g.currentBidding.IsDone() {
		return BiddingState
	}
	if g.currentTrick == nil {
		return CallingState
	}
	return PlayingState
}

func (g *game) Players() []Player {
	return g.players
}

func (g *game) CurrentBidding() BiddingRound {
	return g.currentBidding
}

func (g *game) CurrentTrick() Trick {
	return g.currentTrick
}

// func (g *game) PlacedBids() []Bid {
// 	return g.currentBidding.Bids()
// }

// func (g *game) LeadBidPos() int {
// 	return g.currentBidding.LeadPos()
// }

func (g *game) DealerPos() int {
	return g.currentDealerPos
}

func (g *game) PosToPlay() (int, error) {
	switch g.State() {
	case BiddingState:
		return g.currentBidding.CurrentTurnPos()
	case CallingState:
		_, pos, err := g.currentBidding.WinningBidAndPos()
		if err != nil {
			return 0, err
		}
		return pos, nil
	case PlayingState:
		pos, err := g.currentTrick.CurrentTurnPos()
		if err != nil {
			return 0, err
		}
		return pos, nil
	}
	return 0, fmt.Errorf("no one plays in %s state", g.State())
}

func (g *game) Score() Score {
	return g.score
}

func (g *game) AvailableBids(player Player) []Bid {
	if g.currentBidding.IsDone() {
		return nil
	}
	pos, err := g.PlayerPos(player)
	if err != nil {
		return nil
	}
	return nextBidValues(g.currentBidding.Bids(), pos == g.currentDealerPos)
}

// func (g *game) WinningBid() (Bid, error) {
// 	bid, _, err := g.currentBidding.WinningBidAndPos()
// 	if err != nil {
// 		return nil, fmt.Errorf("bidding is not done")
// 	}
// 	return bid, nil
// }

func (g *game) PlayerHand(player Player) Hand {
	pos, err := g.PlayerPos(player)
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

	// Is bid in available bids?
	found := false
	for _, ab := range g.AvailableBids(player) {
		if ab.IsEqualTo(bid) {
			found = true
			break
		}
	}
	if !found {
		return nil, ErrInvalidBid
	}

	pos, err := g.PlayerPos(player)
	if err != nil {
		return nil, err
	}
	if err := g.currentBidding.placeBid(pos, bid); err != nil {
		return nil, err
	}

	gs, err := g.gameStore.Get(ctx, g.id)
	if err != nil {
		return nil, fmt.Errorf("fetching game to update: %v", err)
	}
	gs.CurrentBidding = g.currentBidding.Encoded()

	// If we have all the bids, start playing.
	if g.currentBidding.IsDone() {
		bid, pos, err := g.currentBidding.WinningBidAndPos()
		if err != nil {
			return nil, err
		}
		// If no-trump, we skip past trump selection.
		if bid.IsNoTrump() {
			if err := g.startTrick(gs, deck.NoTrump, pos); err != nil {
				return nil, err
			}
		}
	}

	if err = g.gameStore.Set(ctx, g.id, gs); err != nil {
		return nil, fmt.Errorf("saving game: %v", err)
	}

	return gameFromDatastore(ctx, g.gameStore, g.playerStore, g.id, gs)
}

func (g *game) CallTrump(ctx context.Context, player Player, trump deck.Suit) (Game, error) {
	// Are we in calling state?
	if g.State() != CallingState {
		return nil, ErrNotCalling
	}

	// Is this the right player to set trump?
	_, winningPos, err := g.currentBidding.WinningBidAndPos()
	if err != nil {
		return nil, err
	}
	pos, err := g.PlayerPos(player)
	if err != nil {
		return nil, err
	}
	if pos != winningPos {
		return nil, fmt.Errorf("incorrect player to select trump")
	}

	gs, err := g.gameStore.Get(ctx, g.id)
	if err != nil {
		return nil, fmt.Errorf("fetching game to update: %v", err)
	}

	if err := g.startTrick(gs, trump, winningPos); err != nil {
		return nil, err
	}

	if err = g.gameStore.Set(ctx, g.id, gs); err != nil {
		return nil, fmt.Errorf("saving game: %v", err)
	}

	return gameFromDatastore(ctx, g.gameStore, g.playerStore, g.id, gs)
}

func (g *game) PlayCard(ctx context.Context, player Player, card deck.Card) (Game, error) {
	// Are we in playing state?
	if g.State() != PlayingState {
		return nil, ErrNotPlaying
	}

	pos, err := g.PlayerPos(player)
	if err != nil {
		return nil, err
	}

	playerHand := g.currentHands[pos]

	// Does the player have the card to play?
	if !playerHand.Contains(card) {
		return nil, ErrMissingCard
	}

	// Is the card a valid card to play (i.e., does it follow suit)?
	leadSuit, err := g.currentTrick.LeadSuit()
	if err != nil {
		return nil, err
	}
	if g.currentTrick.NumPlayed() > 0 && card.Suit() != leadSuit {
		if playerHand.ContainsSuit(leadSuit, card) {
			return nil, ErrNotFollowingSuit
		}
	}

	gs, err := g.gameStore.Get(ctx, g.id)
	if err != nil {
		return nil, fmt.Errorf("fetching game to update: %v", err)
	}

	// Remove the card from the player hand.
	newHand, err := playerHand.removeCard(card)
	if err != nil {
		return nil, err
	}
	gs.CurrentHands[pos] = newHand.Encoded()

	// Add the card to the current trick
	if err := g.currentTrick.playCard(pos, card); err != nil {
		return nil, err
	}
	gs.CurrentTrick = g.currentTrick.Encoded()

	// If the last card, compute the results
	if g.currentTrick.IsDone() {

		// Who won the trick?
		winningPos, err := g.currentTrick.WinningPos()
		if err != nil {
			return nil, err
		}

		// Add the trick to the tally.
		if err := g.currentTally.addTrick(g.currentTrick); err != nil {
			return nil, err
		}
		gs.CurrentTally = g.currentTally.Encoded()

		// If all cards are played, update the score.
		if g.currentHands[0].IsEmpty() {
			if err := g.score.addTally(g.currentTally); err != nil {
				return nil, err
			}
			gs.Score = g.score.Encoded()

			// If the score is a winning (note: bid out, etc.), complete the game.
			// TODO: need better stuff here: consider bid-out, stealing the 5, etc.
			if g.score.CurrentScore()[0] > g.score.ToWin() || g.score.CurrentScore()[1] > g.score.ToWin() {
				g.complete = true
				gs.Complete = true
			} else {
				// Else deal a new hand and go to bidding round.
				if err := g.startHand(gs); err != nil {
					return nil, err
				}
			}
		} else {
			if err := g.startTrick(gs, g.currentTrick.Trump(), winningPos); err != nil {
				return nil, err
			}
		}
	}

	if err = g.gameStore.Set(ctx, g.id, gs); err != nil {
		return nil, fmt.Errorf("saving game: %v", err)
	}
	return gameFromDatastore(ctx, g.gameStore, g.playerStore, g.id, gs)
}

func (g *game) startHand(gs *storage.Game) error {
	deck, err := deck.NewDeck()
	if err != nil {
		return err
	}
	deck.Shuffle()

	g.currentHands = nil
	gs.CurrentHands = nil
	for _, h := range deck.Deal() {
		hand, err := NewHand(h)
		if err != nil {
			return err
		}
		g.currentHands = append(g.currentHands, hand)
		gs.CurrentHands = append(gs.CurrentHands, hand.Encoded())
	}

	g.currentDealerPos = (g.currentDealerPos + 1) % 4
	gs.CurrentDealerPos = g.currentDealerPos

	leadBidder := (g.currentDealerPos + 1) % 4 // to the left of the dealer
	biddingRound, err := NewBiddingRound(leadBidder)
	if err != nil {
		return err
	}
	g.currentBidding = biddingRound
	gs.CurrentBidding = g.currentBidding.Encoded()

	g.currentTrick = nil
	gs.CurrentTrick = ""

	g.currentTally = NewTally()
	gs.CurrentTally = g.currentTally.Encoded()
	return nil
}

func (g *game) startTrick(gs *storage.Game, trump deck.Suit, leadPos int) error {
	trick, err := NewTrick(trump, leadPos)
	if err != nil {
		return err
	}
	g.currentTrick = trick
	gs.CurrentTrick = g.currentTrick.Encoded()
	return nil
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

func (g *game) PlayerPos(player Player) (int, error) {
	for pos, p := range g.Players() {
		if p.ID() == player.ID() {
			return pos, nil
		}
	}
	return -1, errors.New("unknown player")
}

func gameFromDatastore(ctx context.Context, gameStore storage.GameStore, playerStore storage.PlayerStore, id string, gs *storage.Game) (Game, error) {
	if gs == nil {
		return nil, errors.New("nil game")
	}
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
		player, err := playerFromStorage(playerStore, gs.PlayerIDs[i], ps)
		if err != nil {
			return nil, err
		}
		players[i] = player
	}

	bidding, err := NewBiddingRoundFromEncoded(gs.CurrentBidding)
	if err != nil {
		return nil, err
	}

	var hands []Hand
	for _, encoded := range gs.CurrentHands {
		hand, err := NewHandFromEncoded(encoded)
		if err != nil {
			return nil, err
		}
		hands = append(hands, hand)
	}

	trick, err := NewTrickFromEncoded(gs.CurrentTrick)
	if err != nil {
		return nil, err
	}

	score, err := NewScoreFromEncoded(gs.Score)
	if err != nil {
		return nil, err
	}

	tally, err := NewTallyFromEncoded(gs.CurrentTally)
	if err != nil {
		return nil, err
	}

	g := &game{
		gameStore:        gameStore,
		playerStore:      playerStore,
		id:               id,
		players:          players,
		complete:         gs.Complete,
		score:            score,
		currentDealerPos: gs.CurrentDealerPos,
		currentBidding:   bidding,
		currentHands:     hands,
		currentTrick:     trick,
		currentTally:     tally,
	}
	return g, nil
}
