package deck

import (
	"fmt"
	"strings"
)

type Suit interface {
	Encoded() string
	String() string
}

type suit struct {
	encoded string
}

var _ Suit = (*suit)(nil) // Ensure interface is implemented.

var (
	Hearts   Suit = &suit{encoded: "H"}
	Diamonds Suit = &suit{encoded: "D"}
	Spades   Suit = &suit{encoded: "S"}
	Clubs    Suit = &suit{encoded: "C"}
	NoTrump  Suit = &suit{encoded: "N"}

	suitFromEncoded = map[string]Suit{
		"H": Hearts,
		"D": Diamonds,
		"S": Spades,
		"C": Clubs,
		"N": NoTrump,
	}

	stringFromEncoded = map[string]string{
		"H": "Hearts",
		"D": "Diamonds",
		"S": "Spades",
		"C": "Clubs",
		"N": "No Trump",
	}
)

func NewSuitFromEncoded(encoded string) (Suit, error) {
	s, present := suitFromEncoded[strings.ToUpper(encoded)]
	if !present {
		return nil, fmt.Errorf("invalid suit encoding %q", encoded)
	}
	return s, nil
}

func (s *suit) String() string {
	return stringFromEncoded[s.encoded]
}

func (s *suit) Encoded() string {
	return s.encoded
}
