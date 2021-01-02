package game

import (
	"errors"

	"github.com/squee1945/threespot/server/pkg/deck"
)

var (
	ErrNotFound        = errors.New("Not found")
	ErrInvalidPosition = errors.New("Invalid position")
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
}

type game struct {
	players []Player // Position 0/2 are a team, 1/3 are a team; organizer is position 0.
}

func NewGame(organizer Player) (Game, error) {
	return nil, nil
}

func GetGame(id string) (Game, error) {
	return nil, nil
}
