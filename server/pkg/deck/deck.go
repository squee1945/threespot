package deck

import (
	"fmt"
	"math/rand"
	"time"
)

type Suit string

const (
	Hearts   Suit = "Hearts"
	Diamonds Suit = "Diamonds"
	Spades   Suit = "Spades"
	Clubs    Suit = "Clubs"
)

type Card struct {
	Num  string
	Suit Suit
}

func (c Card) String() string {
	n := c.Num
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
	return fmt.Sprintf("%s of %s", n, c.Suit)
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
			d.cards = append(d.cards, Card{Num: numset[k], Suit: suitset[j]})
		}
	}
	d.cards = append(d.cards, Card{Num: "7", Suit: Clubs})
	d.cards = append(d.cards, Card{Num: "7", Suit: Diamonds})
	d.cards = append(d.cards, Card{Num: "3", Suit: Spades})
	d.cards = append(d.cards, Card{Num: "5", Suit: Hearts})
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
