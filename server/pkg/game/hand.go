package game

import (
	"fmt"
	"strings"

	"github.com/squee1945/threespot/server/pkg/deck"
)

var (
	suitValues = map[deck.Suit]int{
		deck.Hearts:   4,
		deck.Spades:   3,
		deck.Diamonds: 2,
		deck.Clubs:    1,
	}
)

type Hand interface {
	Cards() []deck.Card
	Contains(card deck.Card) bool
	ContainsSuit(suit deck.Suit) bool
	RemoveCard(card deck.Card) (Hand, error)
	Encoded() []string
	IsEmpty() bool
	Encoded() string
}

type hand struct {
	cards []Card
}

func NewHandFromEncoded(encoded string) (Hand, error) {
	// "{card0}-{card1}-..."
	parts := strings.Split(strings.ToUpper(encoded), "-")
	var cards []Card
	for i, p := range parts {
		card, err := deck.NewCardFromEncoded(p)
		if err != nil {
			return nil, fmt.Errorf("part[%d] %q invalid card: %v", i, p, err)
		}
		cards = append(cards, card)
	}
	return &hand{cards: cards}, nil
}

func (h *hand) Contains(card deck.Card) bool {
	for _, c := range h.cards {
		if c == card {
			return true
		}
	}
	return false
}

func (h *hand) ContainsSuit(suit deck.Suit) bool {
	for _, c := range h.cards {
		if c.Suit() == suit {
			return true
		}
	}
	return false
}

func (h *hand) Cards() []deck.Card {
	return h.cards
}

func (h *hand) RemoveCard(card deck.Card) (Hand, error) {
	if !h.Contains(card) {
		return nil, ErrMissingCard
	}
	var newCards []deck.Card
	for _, c := range h.cards {
		if c == card {
			continue
		}
		newCards = append(newCards, c)
	}
	return NewHand(newCards)
}

func (h *hand) Encoded() []string {
	var encoded []string
	for _, c := range h.cards {
		encoded = append(encoded, string(c))
	}
	return encoded
}

func (h *hand) IsEmpty() bool {
	return len(h.cards) > 0
}

func handSorter(i, j Card) bool {
	if i.Suit() != j.Suit() {
		return suitValues[i.Suit()] < suitValues[j.Suit()]
	}
	return isNumHigher(a.Num(), b.Num())
}
