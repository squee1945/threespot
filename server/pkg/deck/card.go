package deck

import (
	"errors"
	"fmt"
	"strings"
)

type Suit string

const (
	Hearts   Suit = "H"
	Diamonds Suit = "D"
	Spades   Suit = "S"
	Clubs    Suit = "C"
	NoTrump  Suit = "N"
)

var (
	validNums = map[string]bool{
		"7": true,
		"8": true,
		"9": true,
		"T": true,
		"J": true,
		"Q": true,
		"K": true,
		"A": true,
		"3": true,
		"5": true,
	}
	validSuits = map[Suit]bool{
		Hearts:   true,
		Diamonds: true,
		Spades:   true,
		Clubs:    true,
	}
	suitFromString = map[string]Suit{
		"H": Hearts,
		"D": Diamonds,
		"S": Spades,
		"C": Clubs,
	}
)

type Card string

func (c Card) Num() string {
	return string(c[0])
}

func (c Card) Suit() Suit {
	return suitFromString[strings.ToUpper(string(c[1]))]
}

func (c Card) String() string {
	n := c.Num()
	switch n {
	case "T":
		n = "10"
	case "J":
		n = "Jack"
	case "Q":
		n = "Queen"
	case "K":
		n = "King"
	case "A":
		n = "Ace"
	}
	s := ""
	switch c.Suit() {
	case Hearts:
		s = "Hearts"
	case Diamonds:
		s = "Diamonds"
	case Spades:
		s = "Spades"
	case Clubs:
		s = "Clubs"
	}
	return fmt.Sprintf("%s of %s", n, s)
}

func NewCardFromEncoded(encoded string) (Card, error) {
	if len(encoded) != 2 {
		return "", errors.New("card string must be two characters")
	}
	num := string(encoded[0])
	suitStr := string(encoded[1])
	suit, ok := suitFromString[strings.ToUpper(suitStr)]
	if !ok {
		return "", fmt.Errorf("unknown suit %q", suitStr)
	}
	return NewCard(num, suit)
}

func NewCard(num string, suit Suit) (Card, error) {
	if _, present := validNums[num]; !present {
		return Card(""), fmt.Errorf("invalid num %q", num)
	}
	if _, present := validSuits[suit]; !present {
		return Card(""), fmt.Errorf("invalid suit %q", suit)
	}
	return Card(num + string(suit)), nil
}
