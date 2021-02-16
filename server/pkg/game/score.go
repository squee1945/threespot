package game

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// Score keeps track of the game store.
type Score interface {
	// ToWin is the score to win; either 52 or 62.
	ToWin() int

	// setTopScore62 sets the top score to be 62.
	setTopScore62()

	// Encoded is the encoded form of the score.
	Encoded() string

	// Scores returns a running score as a list of pairs. Each pair is the score of a hand, oldest first.
	// Each pair is (points02, points13). The current score is the last item in the list.
	Scores() [][]int

	// CurrentScore return a pair (points02, points13) representing the current score.
	CurrentScore() []int

	// addTally adds a tally to the score. An error is returned if the tally is not done.
	addTally(BiddingRound, Tally) error
}

type score struct {
	scores [][]int
	toWin  int
}

var _ Score = (*score)(nil) // Ensure interface is implemented.

// NewScoreFromEncoded builds a score sheet form the Encoded() representation.
func NewScoreFromEncoded(encoded string) (Score, error) {
	// "toWin-p02|p13||p02|p03||"
	if encoded == "" {
		encoded = "52-"
	}

	score := NewScore().(*score)

	// Peel off the toWin from the front.
	topParts := strings.SplitN(encoded, "-", 2)
	if len(topParts) != 2 {
		return nil, fmt.Errorf("toWin score not encoded in %q correctly", encoded)
	}
	toWin, err := strconv.Atoi(topParts[0])
	if err != nil {
		return nil, err
	}
	score.toWin = toWin

	if topParts[1] == "" {
		return score, nil
	}

	// Process the rest of encoded as a list of pairs.
	pairs := strings.Split(topParts[1], "||")
	for _, pair := range pairs {
		parts := strings.Split(pair, "|")
		if len(parts) != 2 {
			return nil, fmt.Errorf("each element of %q must be a pair", topParts[1])
		}
		points02, err := strconv.Atoi(parts[0])
		if err != nil {
			return nil, err
		}
		points13, err := strconv.Atoi(parts[1])
		if err != nil {
			return nil, err
		}
		r := []int{points02, points13}
		score.scores = append(score.scores, r)
	}
	return score, nil
}

// NewScore creates an empty score sheet.
func NewScore() Score {
	return &score{toWin: 52}
}

func (s *score) Encoded() string {
	var pairs []string
	for _, r := range s.scores {
		pairs = append(pairs, fmt.Sprintf("%d|%d", r[0], r[1]))
	}
	res := fmt.Sprintf("%d-%s", s.toWin, strings.Join(pairs, "||"))
	return res
}

func (s *score) ToWin() int {
	return s.toWin
}

func (s *score) setTopScore62() {
	s.toWin = 62
}

func (s *score) Scores() [][]int {
	return s.scores
}

func (s *score) CurrentScore() []int {
	if len(s.scores) == 0 {
		return []int{0, 0}
	}
	return s.scores[len(s.scores)-1]
}

func (s *score) addTally(br BiddingRound, tally Tally) error {
	if !tally.IsDone() {
		return errors.New("tally is not done")
	}

	bid, pos, err := br.WinningBidAndPos()
	if err != nil {
		return err
	}

	last02, last13 := 0, 0
	if len(s.scores) > 0 {
		last := s.scores[len(s.scores)-1]
		last02, last13 = last[0], last[1]
	}

	bidValue, err := bid.Value()
	if err != nil {
		return err
	}

	multiplier := 1
	if bid.IsNoTrump() {
		multiplier = 2
	}

	madeBid := false
	points02, points13 := tally.Points()
	sc := make([]int, 2)
	if pos == 0 || pos == 2 {
		// Team02 made the bid.
		if points02 >= bidValue {
			sc[0] = last02 + (points02 * multiplier)
			madeBid = true
		} else {
			sc[0] = last02 - (bidValue * multiplier)
		}
		sc[1] = last13 + points13
	} else {
		// Team13 made the bid.
		if points13 >= bidValue {
			sc[1] = last13 + (points13 * multiplier)
			madeBid = true
		} else {
			sc[1] = last13 - (bidValue * multiplier)
		}
		sc[0] = last02 + points02
	}
	s.scores = append(s.scores, sc)

	if multiplier == 2 && madeBid {
		s.setTopScore62()
	}

	return nil
}
