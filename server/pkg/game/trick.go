package game

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/squee1945/threespot/server/pkg/deck"
)

const (
	orderedCards = "35789TJQKA"
)

var (
	validTrickSuits = map[deck.Suit]bool{
		deck.Hearts:   true,
		deck.Diamonds: true,
		deck.Spades:   true,
		deck.Clubs:    true,
		deck.NoTrump:  true,
	}
)

type Trick interface {
	// IsDone returns true if the trick is complete (4 cards played).
	IsDone() bool

	// PlayCard adds a card to the trick for the player in playerPos position.
	PlayCard(playerPos int, card deck.Card) error

	// CurrentTurnPos returns the position of the player who's turn it is to play. Returns error if IsDone().
	CurrentTurnPos() (int, error)

	// WinningPos returns the position of the player that won the trick. Returns error if !IsDone().
	WinningPos() (int, error)

	// Trump returns the trump suit for this trick.
	Trump() deck.Suit

	// LeadPos returns the position of the lead player.
	LeadPos() int

	// LeadSuit returns the suit that was lead. Returns error if no card has been played.
	LeadSuit() (deck.Suit, error)

	// NumPlayed returns the number of cards played so far.
	NumPlayed() int

	// Cards returns the cards played; the first card is for the LeadPos, clockwise from there.
	Cards() []deck.Card

	// Encoded returns the entire trick encoded into a single string.
	Encoded() string
}

type trick struct {
	// trump is the trump for the hand this trick belongs to.
	trump deck.Suit
	// leadPos is the position (0..3) of the leadoff player.
	leadPos int
	// cards are the cards played. plays[0] is the card played by the player in leadPos.
	cards []deck.Card
}

var _ Trick = (*trick)(nil) // Ensure interface is implemented.

func NewTrickFromEncoded(encoded string) (Trick, error) {
	// "{leadPos}-{trump}-{card0}-{card1}-{card2}-{card3}"
	parts := strings.Split(encoded, "-")
	if len(parts) < 2 {
		return nil, fmt.Errorf("encoded string %q must have at least two parts", encoded)
	}
	leadPos, err := strconv.Atoi(parts[0])
	if err != nil {
		return nil, fmt.Errorf("encoded string part[0] %q not int: %v", parts[0], err)
	}
	if leadPos < 0 || leadPos > 3 {
		return nil, fmt.Errorf("encoded string part[0] %q not in range", parts[0])
	}
	trump, err := deck.NewSuitFromEncoded(strings.ToUpper(parts[1]))
	if err != nil {
		return nil, fmt.Errorf("encoded string part[1] %q not suit: %v", parts[1], err)
	}
	var cards []deck.Card
	for i := 2; i < len(parts); i++ {
		card, err := deck.NewCardFromEncoded(strings.ToUpper(parts[i]))
		if err != nil {
			return nil, fmt.Errorf("encoded string part[%d] %q not card: %v", i, parts[i], err)
		}
		cards = append(cards, card)
	}
	return &trick{
		trump:   trump,
		leadPos: leadPos,
		cards:   cards,
	}, nil
}

func (t *trick) Encoded() string {
	s := fmt.Sprintf("%d-%s", t.leadPos, t.trump)
	for _, c := range t.cards {
		s += fmt.Sprintf("-%s", c)
	}
	return s
}

func (t *trick) PlayCard(playerPos int, card deck.Card) error {
	ord := t.toOrd(playerPos)
	if len(t.cards) != ord {
		return ErrIncorrectPlayOrder
	}
	t.cards = append(t.cards, card)
	return nil
}

func (t *trick) Trump() deck.Suit {
	return t.trump
}

func (t *trick) LeadPos() int {
	return t.leadPos
}

func (t *trick) LeadSuit() (deck.Suit, error) {
	if len(t.cards) == 0 {
		return nil, errors.New("no cards have been played")
	}
	return t.cards[0].Suit(), nil
}

func (t *trick) NumPlayed() int {
	return len(t.cards)
}

func (t *trick) CurrentTurnPos() (int, error) {
	if t.IsDone() {
		return -1, fmt.Errorf("trick is complete")
	}
	return (t.leadPos + len(t.cards)) % 4, nil
}

func (t *trick) IsDone() bool {
	return len(t.cards) == 4
}

func (t *trick) Cards() []deck.Card {
	return t.cards
}

func (t *trick) WinningPos() (int, error) {
	if !t.IsDone() {
		return 0, fmt.Errorf("trick is incomplete")
	}
	highOrd := 0
	for i := 1; i < 4; i++ {
		if t.isHigher(t.cards[0].Suit(), t.cards[highOrd], t.cards[i]) {
			highOrd = i
		}
	}
	return t.toPos(highOrd), nil
}

// toOrd returns the player order for this trick (0..3), computed from the leadPos.
func (t *trick) toOrd(playerPos int) int {
	return (playerPos + 4 - t.leadPos) % 4
}

func (t *trick) toPos(playerOrd int) int {
	return (t.leadPos + playerOrd) % 4
}

// isHigher returns true if b is higher than a, considering the lead suit and the trump.
func (t *trick) isHigher(lead deck.Suit, a, b deck.Card) bool {
	// First, consider the trump suit.
	if t.trump != deck.NoTrump {
		// If a and b are trump, return true if b is higher.
		if a.Suit() == t.trump && b.Suit() == t.trump {
			return isNumHigher(a.Num(), b.Num())
		}
		// If a is trump and b is not, return false.
		if a.Suit() == t.trump && b.Suit() != t.trump {
			return false
		}
		// If b is trump and a is not, return true.
		if a.Suit() != t.trump && b.Suit() == t.trump {
			return true
		}
	}

	// Otherwise, no trumps are in play.

	// If both are suited, return true if b is higher.
	if a.Suit() == lead && b.Suit() == lead {
		return isNumHigher(a.Num(), b.Num())
	}
	// If a is suited, but b is not, return false.
	if a.Suit() == lead && b.Suit() != lead {
		return false
	}
	// If b is suited, but a is not, return true.
	if a.Suit() != lead && b.Suit() == lead {
		return true
	}
	// Otherwise, neither are suited, just return false.
	return false
}

// isNumHigher returns true if b is a higher "num" than a.
// NOTE: this method is only valid if cards are the same suit.
func isNumHigher(a, b string) bool {
	return strings.Index(orderedCards, a) < strings.Index(orderedCards, b)
}
