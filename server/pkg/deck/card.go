package deck

import (
	"errors"
	"fmt"
	"strings"
)

type Card interface {
	Num() string
	Suit() Suit
	String() string
	Encoded() string
}

type card struct {
	encoded string
}

var _ Card = (*card)(nil) // Ensure interface is implemented.

var (
	stringFromNum = map[string]string{
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

func NewCard(num string, suit Suit) (Card, error) {
	if suit == NoTrump {
		return nil, fmt.Errorf("card cannot be no trump suit")
	}
	num = strings.ToUpper(num)
	if _, present := stringFromNum[num]; !present {
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

func (c *card) String() string {
	return fmt.Sprintf("%s of %s", stringFromNum[c.Num()], c.Suit())
}

func (c *card) Encoded() string {
	return c.encoded
}
