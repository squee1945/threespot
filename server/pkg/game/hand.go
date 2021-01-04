package game

import (
	"fmt"
	"sort"
	"strings"

	"github.com/squee1945/threespot/server/pkg/deck"
)

// Hand is a hand of cards held by a player.
type Hand interface {
	// Cards are the cards in the hand.
	Cards() []deck.Card
	// Contains returns true if a card is in the hand.
	Contains(card deck.Card) bool
	// ContainsSuit returns true if a card of the given suit is in the hand.
	// The ignoreCard is not considered when searching for the suit (i.e., it is the card that is being played).
	ContainsSuit(suit deck.Suit, ignoreCard deck.Card) bool
	// RemoveCard removes the given card from the hand, returning an error if it is not present.
	RemoveCard(card deck.Card) (Hand, error)
	// IsEmpty returns true if there are no cards left in the hand.
	IsEmpty() bool
	// Encoded returns the encoded form of the hand.
	Encoded() string
}

var (
	suitValues = map[string]int{
		deck.Hearts.Encoded():   4,
		deck.Spades.Encoded():   3,
		deck.Diamonds.Encoded(): 2,
		deck.Clubs.Encoded():    1,
	}
)

type hand struct {
	cards []deck.Card
}

var _ Hand = (*hand)(nil) // Ensure interface is implemented.

// NewHandFromEncoded builds a hand from the Encoded() form.
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
	return NewHand(cards)
}

// NewHand creates a hand from the given cards. An error is returned if there are duplicates.
func NewHand(cards []deck.Card) (Hand, error) {
	seen := make(map[string]bool, len(cards))
	for _, card := range cards {
		if _, present := seen[card.Encoded()]; present {
			return nil, fmt.Errorf("hand %v has duplicate card", cards)
		}
		seen[card.Encoded()] = true
	}
	return &hand{cards: cards}, nil
}

func (h *hand) Encoded() string {
	var cs []string
	for _, card := range h.cards {
		cs = append(cs, card.Encoded())
	}
	return strings.Join(cs, "-")
}

func (h *hand) Contains(card deck.Card) bool {
	for _, c := range h.cards {
		if c.Encoded() == card.Encoded() {
			return true
		}
	}
	return false
}

func (h *hand) ContainsSuit(suit deck.Suit, ignoreCard deck.Card) bool {
	for _, c := range h.cards {
		if c.IsSameAs(ignoreCard) {
			continue
		}
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
		if c.Encoded() == card.Encoded() {
			continue
		}
		newCards = append(newCards, c)
	}
	return &hand{cards: newCards}, nil
}

func (h *hand) IsEmpty() bool {
	return len(h.cards) == 0
}

func handSorter(cards []deck.Card) func(i, j int) bool {
	return func(i, j int) bool {
		card1, card2 := cards[i], cards[j]
		if !card1.Suit().IsSameAs(card2.Suit()) {
			return suitValues[card1.Suit().Encoded()] > suitValues[card2.Suit().Encoded()]
		}
		return !isNumHigher(card1.Num(), card2.Num())
	}
}
