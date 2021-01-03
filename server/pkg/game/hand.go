package game

import (
	"fmt"
	"sort"
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
	IsEmpty() bool
	Encoded() string
}

type hand struct {
	cards []deck.Card
}

func NewHandFromEncoded(encoded string) (Hand, error) {
	// "{card0}-{card1}-..."
	if encoded == "" {
		return &hand{}, nil
	}
	parts := strings.Split(strings.ToUpper(encoded), "-")
	var cards []deck.Card
	for i, p := range parts {
		card, err := deck.NewCardFromEncoded(p)
		if err != nil {
			return nil, fmt.Errorf("part[%d] %q invalid card: %v", i, p, err)
		}
		cards = append(cards, card)
	}
	return &hand{cards: cards}, nil
}

func (h *hand) Encoded() string {
	var cs []string
	for _, card := range h.cards {
		cs = append(cs, string(card))
	}
	return strings.Join(cs, "-")
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
	sort.SliceStable(h.cards, handSorter(h.cards))
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
	return &hand{cards: newCards}, nil
}

func (h *hand) IsEmpty() bool {
	return len(h.cards) > 0
}

func handSorter(cards []deck.Card) func(i, j int) bool {
	return func(i, j int) bool {
		card1, card2 := cards[i], cards[j]
		if card1.Suit() != card2.Suit() {
			return suitValues[card1.Suit()] < suitValues[card2.Suit()]
		}
		return isNumHigher(card1.Num(), card2.Num())
	}
}
