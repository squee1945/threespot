package deck

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

type Deck interface {
	Shuffle()
	Deal() [][]Card // 4 hands of 8 cards
}

func NewDeck() Deck {
	// TODO
	return &deck{}
}

type deck struct {
	cards []Card
}

func (d *deck) Shuffle() {
	// TODO
	return
}

func (d *deck) Deal() [][]Card {
	// TODO
	return nil
}
