package deck

import (
	"fmt"
	"math/rand"
	"time"
)

type Suit string

const (
	Hearts   Suit = "H"
	Diamonds Suit = "D"
	Spades   Suit = "S"
	Clubs    Suit = "C"
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
)

type Card string

func (c Card) Num() string {
	return string(c[0])
}

func (c Card) Suit() Suit {
	switch string(c[1]) {
	case "H":
		return Hearts
	case "D":
		return Diamonds
	case "S":
		return Spades
	case "C":
		return Clubs
	}
	return ""
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

func NewCard(num string, suit Suit) (Card, error) {
	if _, present := validNums[num]; !present {
		return Card(""), fmt.Errorf("invalid num %q", num)
	}
	if _, present := validSuits[suit]; !present {
		return Card(""), fmt.Errorf("invalid suit %q", suit)
	}
	return Card(num + string(suit)), nil
}

type Deck interface {
	Shuffle()
	Deal() [][]Card // 4 hands of 8 cards
}

func NewDeck() (Deck, error) {
	d := &deck{}
	numset := []string{"8", "9", "T", "J", "Q", "K", "A"}
	suitset := []Suit{Hearts, Diamonds, Spades, Clubs}
	for j := 0; j < 4; j++ {
		for k := 0; k < 7; k++ {
			c, err := NewCard(numset[k], suitset[j])
			if err != nil {
				return nil, err
			}
			d.cards = append(d.cards, c)
		}
	}
	c, err := NewCard("7", Clubs)
	if err != nil {
		return nil, err
	}
	d.cards = append(d.cards, c)
	c, err = NewCard("7", Diamonds)
	if err != nil {
		return nil, err
	}
	d.cards = append(d.cards, c)
	c, err = NewCard("3", Spades)
	if err != nil {
		return nil, err
	}
	d.cards = append(d.cards, c)
	c, err = NewCard("5", Hearts)
	if err != nil {
		return nil, err
	}
	d.cards = append(d.cards, c)
	return d, nil
}

type deck struct {
	cards []Card
}

func (d *deck) Shuffle() {
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(d.cards), func(i, j int) { d.cards[i], d.cards[j] = d.cards[j], d.cards[i] })
	return
}

func (d *deck) Deal() [][]Card {
	result := make([][]Card, 4)
	for i := 0; i < 32; i++ {
		result[i%4] = append(result[i%4], d.cards[i])
	}
	return result
}
