package game

import (
	"errors"

	"github.com/squee1945/threespot/server/pkg/deck"
)

var (
	NotFoundErr        = errors.New("Not found")
	InvalidPositionErr = errors.New("Invalid position")
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

	AddPlayer(player Player, pos int) error
	PlaceBid(player Player, bid Bid) error
	PlayCard(player Player, card deck.Card) error
	// // CurrentState() State
	// // Deal() error
	// // Bid() error
	// // PlayCard(Player, Card) (State, error)
	// WaitForPlayers() error
	// IsDone() bool
	// Deal() (Round, error)

}

type game struct {
	players []Player // Position 0/2 are a team, 1/3 are a team; organizer is position 0.
	// hands   []Hand
}

// type GameState interface{}

// type Hand interface {
// 	IsDone() bool
// 	Cards() []Card
// }

// type hand struct {
// 	dealerPos, bidPos int
// 	trump             *deck.Suit // nil for no-trump
// 	tricks            []Trick
// }

func NewGame(organizer Player) (Game, error) {
	return nil, nil
}

func GetGame(id string) (Game, error) {
	return nil, nil
}

// func DealHands(dealerPos int) ([][]Card, error) {
// 	return nil, nil
// }

// func Play(organizer Player) error {
// 	game, err := NewGame(organizer)
// 	if err != nil {
// 		return fmt.Errorf("creating game: %v", err)
// 	}
// 	if err := game.WaitForPlayers(); err != nil {
// 		return fmt.Errorf("waiting for players: %v", err)
// 	}

// 	for {
// 		if game.IsDone() {
// 			break
// 		}
// 		round, err := game.Deal()
// 		if err != nil {
// 			return fmt.Errorf("dealing round: %v", err)
// 		}
// 		if err := round.Bid(); err != nil {
// 			return fmt.Errorf("collecting bids: %v", err)
// 		}
// 		for {
// 			if round.IsDone() {
// 				break
// 			}
// 			if err := round.PlayTrick(); err != nil {
// 				return fmt.Errorf("playing trick: %v", err)
// 			}
// 		}
// 		if err != round.Complete(); err != nil {
// 			return fmt.Errorf("completing round: %v", err)
// 		}
// 	}
// 	if err := game.Complete(); err != nil {
// 		return fmt.Errorf("completing game: %v", err)
// 	}
// }
