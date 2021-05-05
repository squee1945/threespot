package game

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

const (
	// NoWinner means there is currently no winner.
	NoWinner int = -1
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

	// Winner returns 0 if team02 has won, 1 if team13 has one, and NoWinner otherwise.
	Winner() int

	// setWinner sets the winning team; use 0 for team02, 1 for team13.
	setWinner(int)

	// addTally adds a tally to the score, returning true if the game has been won.
	// An error is returned if the tally is not done.
	addTally(BiddingRound, Tally) (bool, error)
}

type score struct {
	scores [][]int
	toWin  int
	winner int
}

var _ Score = (*score)(nil) // Ensure interface is implemented.

// NewScoreFromEncoded builds a score sheet form the Encoded() representation.
func NewScoreFromEncoded(encoded string) (Score, error) {
	// "toWin-p02|p13||p02|p03||"
	// If there is a winning team, it will be encoded like this: "toWin||winningTeam-p02|p13||p02|p03"
	if encoded == "" {
		encoded = "52-"
	}

	score := NewScore().(*score)

	// Peel off the toWin from the front.
	topParts := strings.SplitN(encoded, "-", 2)
	if len(topParts) != 2 {
		return nil, fmt.Errorf("toWin score not encoded in %q correctly", encoded)
	}
	var toWinStr string
	var winner int
	if strings.Contains(topParts[0], "||") {
		winningParts := strings.SplitN(topParts[0], "||", 2)
		toWinStr = winningParts[0]
		var err error
		winner, err = strconv.Atoi(winningParts[1])
		if err != nil {
			return nil, fmt.Errorf("parsing winning team %q: %v", winningParts[1], err)
		}
	} else {
		toWinStr = topParts[0]
		winner = NoWinner
	}
	toWin, err := strconv.Atoi(toWinStr)
	if err != nil {
		return nil, fmt.Errorf("parsing toWin %q: %v", toWinStr, err)
	}
	score.toWin = toWin
	score.winner = winner

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
	return &score{toWin: 52, winner: NoWinner}
}

func (s *score) Encoded() string {
	var pairs []string
	for _, r := range s.scores {
		pairs = append(pairs, fmt.Sprintf("%d|%d", r[0], r[1]))
	}
	winningPart := strconv.Itoa(s.toWin)
	if s.winner != NoWinner {
		winningPart += "||" + strconv.Itoa(s.winner)
	}
	res := fmt.Sprintf("%s-%s", winningPart, strings.Join(pairs, "||"))
	return res
}

func (s *score) ToWin() int {
	return s.toWin
}

func (s *score) setTopScore62() {
	s.toWin = 62
}

// Returns 0 if team02 won, 1 if team13 won, and NoWinner if no one has won.
func (s *score) Winner() int {
	return s.winner
}

// team02 is 0, team13 is 1.
func (s *score) setWinner(winner int) {
	s.winner = winner
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

func (s *score) addTally(br BiddingRound, tally Tally) (bool, error) {
	if !tally.IsDone() {
		return false, errors.New("tally is not done")
	}

	bid, pos, err := br.WinningBidAndPos()
	if err != nil {
		return false, err
	}

	last02, last13 := 0, 0
	if len(s.scores) > 0 {
		last := s.scores[len(s.scores)-1]
		last02, last13 = last[0], last[1]
	}

	bidValue, err := bid.Value()
	if err != nil {
		return false, err
	}

	multiplier := 1
	if bid.IsNoTrump() {
		multiplier = 2
	}

	points02, points13 := tally.Points()
	madeBid02, madeBid13 := false, false
	sc := make([]int, 2)
	if pos == 0 || pos == 2 {
		if points02 >= bidValue {
			// Team02 made the bid.
			sc[0] = last02 + (points02 * multiplier)
			madeBid02 = true
		} else {
			sc[0] = last02 - (bidValue * multiplier)
		}
		sc[1] = last13 + points13
	} else {
		if points13 >= bidValue {
			// Team13 made the bid.
			sc[1] = last13 + (points13 * multiplier)
			madeBid13 = true
		} else {
			sc[1] = last13 - (bidValue * multiplier)
		}
		sc[0] = last02 + points02
	}
	s.scores = append(s.scores, sc)

	if multiplier == 2 && (madeBid02 || madeBid13) {
		s.setTopScore62()
	}

	// Check to see if there is a winner.
	if sc[0] >= s.ToWin() && madeBid02 {
		s.setWinner(0)
	} else if sc[1] >= s.ToWin() && madeBid13 {
		s.setWinner(1)
	}

	return s.Winner() != NoWinner, nil
}
