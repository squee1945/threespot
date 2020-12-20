package deck

import (
	"fmt"
	"math/rand"
	"time"
)

type Suit string

const (
	Hearts   Suit = "hearts"
	Diamonds Suit = "diamonds"
	Spades   Suit = "spades"
	Clubs    Suit = "clubs"
)

type Card struct {
	Num  string
	Suit Suit
}

func (c Card) String() string {
	return fmt.Sprintf("%s of %s", c.Num, c.Suit)
}

type Deck interface {
	Shuffle()
	Deal() [][]Card // 4 hands of 8 cards
}

func NewDeck() Deck {
	// TODO
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
	// TODO
	a := d.cards
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(a), func(i, j int) { a[i], a[j] = a[j], a[i] })
	return
}

func (d *deck) Deal() [][]Card {
	// TODO
	result := make([][]Card, 4)
	for i := 0; i < 32; i++ {
		result[i%4] = append(result[i%4], d.cards[i])
	}
	return result
}
