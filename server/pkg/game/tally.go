package game

import (
	"fmt"
	"strconv"
	"strings"
)

// Tally keeps track of the score during an individial hand.
type Tally interface {
	// addTrick adds a trick to the tally.
	addTrick(Trick) error
	// IsDone returns true if the tally is complete.
	IsDone() bool
	// Points returns the points of the tally. first int is players 0/2; second int is players 1/3.
	Points() (int, int)
	// Encoded returns the encoded tally.
	Encoded() string
}

type tally struct {
	points02   int
	points13   int
	trickCount int
}

var _ Tally = (*tally)(nil) // Ensure interface is implemented.

// NewTallyFromEncoded builds a tally from the Encoded() form.
func NewTallyFromEncoded(encoded string) (Tally, error) {
	// "cardCount|points02|points13"
	if encoded == "" {
		encoded = "0|0|0"
	}
	parts := strings.Split(encoded, "|")
	if len(parts) != 3 {
		return nil, fmt.Errorf("encoded %q does not contain 3 parts", encoded)
	}
	count, err := strconv.Atoi(parts[0])
	if err != nil {
		return nil, fmt.Errorf("count %q is not an int", parts[0])
	}
	points02, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, fmt.Errorf("points02 %q is not an int", parts[1])
	}
	points13, err := strconv.Atoi(parts[2])
	if err != nil {
		return nil, fmt.Errorf("points13 %q is not an int", parts[2])
	}
	return &tally{
		points02:   points02,
		points13:   points13,
		trickCount: count,
	}, nil
}

// NewTally builds an empty tally.
func NewTally() Tally {
	return &tally{}
}

func (t *tally) addTrick(trick Trick) error {
	if !trick.IsDone() {
		return fmt.Errorf("trick is not complete")
	}
	if t.IsDone() {
		return fmt.Errorf("tally is already done")
	}

	winningPos, err := trick.WinningPos()
	if err != nil {
		return err
	}

	score := 1
	if trick.ContainsThreeOfSpades() {
		score -= 3
	}
	if trick.ContainsFiveOfHearts() {
		score += 5
	}

	switch winningPos {
	case 0, 2:
		t.points02 += score
	case 1, 3:
		t.points13 += score
	}
	t.trickCount++
	return nil
}

func (t *tally) IsDone() bool {
	return t.trickCount == 8
}

func (t *tally) Points() (int, int) {
	return t.points02, t.points13
}

func (t *tally) Encoded() string {
	return fmt.Sprintf("%d|%d|%d", t.trickCount, t.points02, t.points13)
}
