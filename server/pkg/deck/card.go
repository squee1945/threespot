package deck

import (
	"errors"
	"fmt"
	"strings"
)

// Card is a single card from a deck of Kaiser cards.
type Card interface {
	// Num returns the number of the card (including "T" (Ten), J" (Jack), "Q" (Queen), "K" (King), "A" (Ace)).
	Num() string
	// Suit returns the suit of the card.
	Suit() Suit
	// Human returns a human-readable form of the card.
	Human() string
	// Encoded returns an encoded form of the card.
	Encoded() string
	// IsSameAs returns true if the cards are the same.
	IsSameAs(Card) bool
}

type card struct {
	encoded string
}

var _ Card = (*card)(nil) // Ensure interface is implemented.

var (
	humanFromNum = map[string]string{
		"3": "3",
		"5": "5",
		"7": "7",
		"8": "8",
		"9": "9",
		"T": "10",
		"J": "Jack",
		"Q": "Queen",
		"K": "King",
		"A": "Ace",
	}
)

// NewCardFromEncoded builds a card from the Encoded() form.
func NewCardFromEncoded(encoded string) (Card, error) {
	if len(encoded) != 2 {
		return nil, errors.New("card string must be two characters")
	}
	suit, err := NewSuitFromEncoded(string(encoded[1]))
	if err != nil {
		return nil, err
	}
	return NewCard(string(encoded[0]), suit)
}

// NewCard builds a card.
func NewCard(num string, suit Suit) (Card, error) {
	if suit == NoTrump {
		return nil, fmt.Errorf("card cannot be no trump suit")
	}
	num = strings.ToUpper(num)
	if _, present := humanFromNum[num]; !present {
		return nil, fmt.Errorf("invalid num %q", num)
	}
	return &card{encoded: num + suit.Encoded()}, nil
}

func (c *card) Num() string {
	return string(c.encoded[0])
}

func (c *card) Suit() Suit {
	return suitFromEncoded[string(c.encoded[1])]
}

func (c *card) Human() string {
	return fmt.Sprintf("%s of %s", humanFromNum[c.Num()], c.Suit())
}

func (c *card) Encoded() string {
	return c.encoded
}

func (c *card) IsSameAs(other Card) bool {
	return c.Encoded() == other.Encoded()
}
