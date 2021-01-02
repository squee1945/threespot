package game

import (
	"fmt"
	"strings"

	"github.com/squee1945/threespot/server/pkg/deck"
)

const (
	orderedCards = "35789TJQKA"
)

type Trick interface {
	IsDone() bool
	PlayCard(playerPos int, card deck.Card) error
	WinningPos() (int, error)
}

func NewTrick(leadPos int, trump deck.Suit, played []deck.Card) Trick {
	return &trick{leadPos: leadPos, trump: trump, cards: played}, nil
}

type trick struct {
	// trump is the trump for the hand this trick belongs to.
	trump deck.Suit
	// leadPos is the position (0..3) of the leadoff player.
	leadPos int
	// plays are the cards played. plays[0] is the card played by the player in leadPos.
	cards []deck.Card
}

func (t *trick) PlayCard(playerPos int, card deck.Card) error {
	ord := t.toOrd(playerPos)
	if len(t.cards) != ord {
		return ErrIncorrectPlayOrder
	}
	t.cards = append(t.cards, card)
	return nil
}

func (t *trick) IsDone() bool {
	return len(t.cards) == 4
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
