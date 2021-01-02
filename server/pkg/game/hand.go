package game

import "github.com/squee1945/threespot/server/pkg/deck"

type Hand interface {
	AllCards() []deck.Card
	PlayedCards() []deck.Card
	HeldCards() []deck.Card
}

func NewHand(cards []Card) Hand {
	return nil
}
