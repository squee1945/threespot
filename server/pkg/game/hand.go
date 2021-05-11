package game

import (
	"fmt"
	"sort"
	"strings"

	"github.com/squee1945/threespot/server/pkg/deck"
)

// Hands are the hands for a set of players.
type Hands interface {
	// Hand returns the hand for the player in playerPos.
	Hand(playerPos int) (Hand, error)
	// Encoded returns the encoded form for the hands.
	Encoded() string
}

// Hand is a hand of cards held by a player.
type Hand interface {
	// Cards are the cards in the hand.
	Cards() []deck.Card
	// Contains returns true if a card is in the hand.
	Contains(card deck.Card) bool
	// ContainsSuit returns true if a card of the given suit is in the hand.
	// The ignoreCard is not considered when searching for the suit (i.e., it is the card that is being played).
	ContainsSuit(suit deck.Suit, ignoreCard deck.Card) bool
	// removeCard removes the given card from the hand, returning an error if it is not present.
	removeCard(card deck.Card) error
	// addCard adds the given card, returning an error if the card is already in the hand.
	addCard(card deck.Card) error
	// IsEmpty returns true if there are no cards left in the hand.
	IsEmpty() bool
	// Encoded returns the encoded form of the hand.
	Encoded() string
}

const (
	handsDelim = "+"
	cardsDelim = "|"
)

var (
	suitValues = map[string]int{
		deck.Hearts.Encoded():   4,
		deck.Spades.Encoded():   3,
		deck.Diamonds.Encoded(): 2,
		deck.Clubs.Encoded():    1,
	}
)

type hands struct {
	hs []Hand
}

type hand struct {
	cards []deck.Card
}

var _ Hands = (*hands)(nil) // Ensure interface is implemented.
var _ Hand = (*hand)(nil)   // Ensure interface is implemented.

// NewHandsFromEncoded creates a set of hands from the Encoded() form.
func NewHandsFromEncoded(encoded string) (Hands, error) {
	if encoded == "" {
		hands := &hands{}
		for i := 0; i < 4; i++ {
			h, err := NewHandFromEncoded("")
			if err != nil {
				return nil, err
			}
			hands.hs = append(hands.hs, h)
		}
		return hands, nil
	}
	parts := strings.Split(encoded, handsDelim)
	if len(parts) != 4 {
		return nil, fmt.Errorf("encoded %q did not have 4 parts", encoded)
	}
	var hs []Hand
	for _, p := range parts {
		h, err := NewHandFromEncoded(p)
		if err != nil {
			return nil, err
		}
		hs = append(hs, h)
	}
	return &hands{hs: hs}, nil
}

// NewHands creates a new set of hands.
func NewHands(cardSets [][]deck.Card) (Hands, error) {
	if len(cardSets) != 4 {
		return nil, fmt.Errorf("must have 4 sets of cards")
	}
	var hs []Hand
	for _, set := range cardSets {
		h, err := NewHand(set)
		if err != nil {
			return nil, err
		}
		hs = append(hs, h)
	}
	return &hands{hs: hs}, nil
}

func (hs *hands) Hand(playerPos int) (Hand, error) {
	if playerPos < 0 || playerPos > 3 {
		return nil, fmt.Errorf("player position must be on interval [0,3]")
	}
	return hs.hs[playerPos], nil
}

func (hs *hands) Encoded() string {
	var encodes []string
	for _, h := range hs.hs {
		encodes = append(encodes, h.Encoded())
	}
	return strings.Join(encodes, handsDelim)
}

// NewHandFromEncoded builds a hand from the Encoded() form.
func NewHandFromEncoded(encoded string) (Hand, error) {
	// "{card0}|{card1}|..."
	if encoded == "" {
		return &hand{}, nil
	}
	parts := strings.Split(strings.ToUpper(encoded), cardsDelim)
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
	return strings.Join(cs, cardsDelim)
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

func (h *hand) removeCard(card deck.Card) error {
	if !h.Contains(card) {
		return ErrMissingCard
	}
	var newCards []deck.Card
	for _, c := range h.cards {
		if c.Encoded() == card.Encoded() {
			continue
		}
		newCards = append(newCards, c)
	}
	h.cards = newCards
	return nil
}

func (h *hand) addCard(card deck.Card) error {
	if h.Contains(card) {
		return fmt.Errorf("duplicate card %s", card)
	}
	h.cards = append(h.cards, card)
	return nil
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
