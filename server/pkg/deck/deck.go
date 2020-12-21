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

func NewCard(num string, suit Suit) Card {
	return Card(num + string(suit))
}

type Deck interface {
	Shuffle()
	Deal() [][]Card // 4 hands of 8 cards
}

func NewDeck() Deck {
	d := &deck{}
	numset := []string{"8", "9", "T", "J", "Q", "K", "A"}
	suitset := []Suit{Hearts, Diamonds, Spades, Clubs}
	for j := 0; j < 4; j++ {
		for k := 0; k < 7; k++ {
			d.cards = append(d.cards, NewCard(numset[k], suitset[j]))
		}
	}
	d.cards = append(d.cards, NewCard("7", Clubs))
	d.cards = append(d.cards, NewCard("7", Diamonds))
	d.cards = append(d.cards, NewCard("3", Spades))
	d.cards = append(d.cards, NewCard("5", Hearts))
	return d
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
