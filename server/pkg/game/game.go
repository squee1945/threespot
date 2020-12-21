package game

import (
	"github.com/squee1945/threespot/server/pkg/deck"
)

type Game interface{}

type game struct {
	players []Player // Position 0/2 are a team, 1/3 are a team.
	hands   []Hand
}

type Hand interface{}

type hand struct {
	dealerPos, bidPos int
	trump             *deck.Suit // nil for no-trump
	tricks            []Trick
}

func NewGame(players []Player) (Game, error) {
	return nil, nil
}

func NewHand(dealerPos int) (Hand, error) {
	return nil, nil
}
