package deck

import (
	"math/rand"
	"time"
)

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
