package deck

import (
	"fmt"
	"strings"
)

// Suit is a suit for a deck of cards. For convenience in game code, "No Trump" is also possible.
type Suit interface {
	// Encoded returns the encoded form of the Suit.
	Encoded() string
	// Human returns a human-readable form of the Suit.
	Human() string
	// IsSameAs returns true if the suits are the same.
	IsSameAs(Suit) bool
}

type suit struct {
	encoded string
}

var _ Suit = (*suit)(nil) // Ensure interface is implemented.

var (
	// Hearts suit.
	Hearts Suit = &suit{encoded: "H"}
	// Diamonds suit.
	Diamonds Suit = &suit{encoded: "D"}
	// Spades suit.
	Spades Suit = &suit{encoded: "S"}
	// Clubs suit.
	Clubs Suit = &suit{encoded: "C"}
	// NoTrump suit; not suitable for cards, but useful in bidding and trick-tracking.
	NoTrump Suit = &suit{encoded: "N"}

	suitFromEncoded = map[string]Suit{
		"H": Hearts,
		"D": Diamonds,
		"S": Spades,
		"C": Clubs,
		"N": NoTrump,
	}

	humanFromEncoded = map[string]string{
		"H": "Hearts",
		"D": "Diamonds",
		"S": "Spades",
		"C": "Clubs",
		"N": "No Trump",
	}
)

// NewSuitFromEncoded builds a suit from the Encoded() form.
func NewSuitFromEncoded(encoded string) (Suit, error) {
	s, present := suitFromEncoded[strings.ToUpper(encoded)]
	if !present {
		return nil, fmt.Errorf("invalid suit encoding %q", encoded)
	}
	return s, nil
}

func (s *suit) Human() string {
	return humanFromEncoded[s.encoded]
}

func (s *suit) Encoded() string {
	return s.encoded
}

func (s *suit) IsSameAs(other Suit) bool {
	return s.Encoded() == other.Encoded()
}
