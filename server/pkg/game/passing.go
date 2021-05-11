package game

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/squee1945/threespot/server/pkg/deck"
)

// PassingRound round is a collection of passed cards.
type PassingRound interface {
	// IsDone returns true if the passing is complete (4 cards).
	IsDone() bool

	// passCard passes a card for the player in playerPos position.
	passCard(playerPos int, card deck.Card) error

	// CurrentTurnPos returns the position of the player who's turn it is to pass. Returns error if IsDone().
	CurrentTurnPos() (int, error)

	// LeadPos returns the position of the lead player.
	LeadPos() int

	// Cards returns the cards passed; the first card is for the LeadPos, clockwise from there.
	Cards() []deck.Card

	// NumPassed returns the number of cards passed.
	NumPassed() int

	// FromPlayer returns the card that was passed from a given player. May only be called when passing IsDone.
	FromPlayer(playerPos int) (deck.Card, error)

	// Encoded returns the passed cards encoded into a single string.
	Encoded() string
}

type passingRound struct {
	// leadPos is the position (0..3) of the leadoff passer.
	leadPos int
	// cards are the cards passed. cards[0] is the card passed by the player in leadPos.
	cards []deck.Card
}

var _ PassingRound = (*passingRound)(nil) // Ensure interface is implemented.

// NewPassingRoundFromEncoded returns a set of bigs from the Encoded() form.
func NewPassingRoundFromEncoded(encoded string) (PassingRound, error) {
	// "{leadPos}|{card0}|{card1}|{card2}|{card3}"
	if encoded == "" {
		encoded = "0|"
	}
	parts := strings.Split(encoded, "|")
	if len(parts) < 1 {
		return nil, fmt.Errorf("encoded %q has too few parts", encoded)
	}
	if len(parts) > 5 {
		return nil, fmt.Errorf("encoded %q has too many parts", encoded)
	}

	leadPos, err := strconv.Atoi(parts[0])
	if err != nil {
		return nil, fmt.Errorf("encoded part[0] %q was not an int: %v", parts[0], err)
	}

	prr, err := NewPassingRound(leadPos)
	if err != nil {
		return nil, err
	}

	pr := prr.(*passingRound)

	for _, cstr := range parts[1:] {
		if cstr == "" {
			continue
		}
		card, err := deck.NewCardFromEncoded(cstr)
		if err != nil {
			return nil, err
		}
		pr.cards = append(pr.cards, card)
	}
	return pr, nil
}

// NewPassingRound creates a new passing round starting with the player in leadPos.
func NewPassingRound(leadPos int) (PassingRound, error) {
	if leadPos < 0 || leadPos > 3 {
		return nil, errors.New("leadPos must be on the interval [0,3]")
	}
	return &passingRound{leadPos: leadPos}, nil
}

func (r *passingRound) IsDone() bool {
	return r.NumPassed() == 4
}

func (r *passingRound) passCard(playerPos int, card deck.Card) error {
	ord := r.toOrd(playerPos)
	if len(r.cards) != ord {
		return ErrIncorrectPassOrder
	}
	r.cards = append(r.cards, card)
	return nil
}

func (r *passingRound) CurrentTurnPos() (int, error) {
	if r.IsDone() {
		return -1, fmt.Errorf("passing is complete")
	}
	return (r.leadPos + len(r.cards)) % 4, nil
}

func (r *passingRound) LeadPos() int {
	return r.leadPos
}

func (r *passingRound) Cards() []deck.Card {
	return r.cards
}

func (r *passingRound) NumPassed() int {
	return len(r.cards)
}

func (r *passingRound) FromPlayer(playerPos int) (deck.Card, error) {
	if !r.IsDone() {
		return nil, errors.New("passing is not complete")
	}
	for i, card := range r.cards {
		if (r.leadPos+i)%4 == playerPos {
			return card, nil
		}
	}
	// Should not happen.
	return nil, errors.New("unexpectedly reached end of passed cards")
}

func (r *passingRound) Encoded() string {
	var parts []string
	for _, card := range r.cards {
		parts = append(parts, card.Encoded())
	}
	return strconv.Itoa(r.leadPos) + "|" + strings.Join(parts, "|")
}

// toOrd returns the player order for this pass (0..3), computed from the leadPos.
func (r *passingRound) toOrd(playerPos int) int {
	return (playerPos + 4 - r.leadPos) % 4
}

// toPos returns the player position for this pass (0..3), computed from the leadPos.
func (r *passingRound) toPos(playerOrd int) int {
	return (r.leadPos + playerOrd) % 4
}
