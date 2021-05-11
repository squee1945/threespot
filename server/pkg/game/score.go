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

	// Notes is an array of notes to show next to the scores.
	Notes() []ScoreNote

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

// ScoreNote is a note to be attached to the score card.
type ScoreNote struct {
	// Team is the team, 0 for team02, 1 for team13.
	Team int
	// Index is the score index to attach the note to.
	Index int
	// Note is the note.
	Note string
}

type score struct {
	scores [][]int
	toWin  int
	winner int
	notes  []ScoreNote
}

var _ Score = (*score)(nil) // Ensure interface is implemented.

// NewScoreFromEncoded builds a score sheet form the Encoded() representation.
func NewScoreFromEncoded(encoded string) (Score, error) {
	// "toWin-p02|p13||p02|p03||"
	// If there is a winning team, it will be encoded like this: "toWin||winningTeam-p02|p13||p02|p03"
	// If there are score notes, they will be in a new top-level section, demarcated with an "**" like this:
	//      "toWin||winningTeam-p02|p13||p02|p13**[note1]||[note2]" where each note is "team|index|note"
	// That is, the notes should be split off the rest of the encoded string first.
	// Separate the scores from the notes
	cardAndNotes := strings.SplitN(encoded, "**", 2)
	encodedNotes := ""
	if len(cardAndNotes) == 2 {
		encodedNotes = cardAndNotes[1]
	}
	encoded = cardAndNotes[0]

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

	// Process the notes
	if encodedNotes != "" {
		var notes []ScoreNote
		noteParts := strings.Split(encodedNotes, "||")
		for _, encodedNote := range noteParts {
			parts := strings.Split(encodedNote, "|")
			if len(parts) != 3 {
				return nil, fmt.Errorf("each encoded note %q must have three parts", encodedNote)
			}
			note := ScoreNote{}
			if team, err := strconv.Atoi(parts[0]); err != nil {
				return nil, fmt.Errorf("parsing team from encoded note %q: %v", encodedNote, err)
			} else {
				if team != 0 && team != 1 {
					return nil, fmt.Errorf("team %q must be 1 or 0", encodedNote)
				}
				note.Team = team
			}

			if index, err := strconv.Atoi(parts[1]); err != nil {
				return nil, fmt.Errorf("parsing index from encoded note %q: %v", encodedNote, err)
			} else {
				if index >= len(score.scores) {
					return nil, fmt.Errorf("index %d is too large for number of scores %d", index, len(score.scores))
				}
				note.Index = index
			}

			note.Note = parts[2]
			notes = append(notes, note)
		}
		score.notes = notes
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
	if len(s.notes) > 0 {
		var parts []string
		for _, note := range s.notes {
			encodedNote := fmt.Sprintf("%d|%d|%s", note.Team, note.Index, note.Note)
			parts = append(parts, encodedNote)
		}
		res += "**" + strings.Join(parts, "||")
	}
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

func (s *score) Notes() []ScoreNote {
	return s.notes
}

func (s *score) attachNote(team, index int, note string) error {
	if team != 0 && team != 1 {
		return errors.New("team must be 0 or 1")
	}
	if index >= len(s.scores) {
		return fmt.Errorf("index %d is too large for number of scores %d", index, len(s.scores))
	}
	if strings.Contains(note, "-") || strings.Contains(note, "|") || strings.Contains(note, "*") {
		return fmt.Errorf("note %q cannot contain special characters", note)
	}
	for _, n := range s.notes {
		if n.Index == index && n.Team == team {
			return fmt.Errorf("note for team %d index %d already exists", team, index)
		}
	}
	s.notes = append(s.notes, ScoreNote{Team: team, Index: index, Note: note})
	return nil
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

	noteTrump := strconv.Itoa(bidValue)
	if bid.IsNoTrump() {
		noteTrump += "N"
	}

	note02, note13 := "", ""
	points02, points13 := tally.Points()
	madeBid02, madeBid13 := false, false
	sc := make([]int, 2)
	if pos == 0 || pos == 2 {
		if points02 >= bidValue {
			// Team02 made the bid.
			sc[0] = last02 + (points02 * multiplier)
			madeBid02 = true
			note02 = fmt.Sprintf("made %s bid", noteTrump)
		} else {
			// Team02 missed the bid.
			sc[0] = last02 - (bidValue * multiplier)
			note02 = fmt.Sprintf("miss %s bid", noteTrump)
		}
		sc[1] = last13 + points13
	} else {
		if points13 >= bidValue {
			// Team13 made the bid.
			sc[1] = last13 + (points13 * multiplier)
			madeBid13 = true
			note13 = fmt.Sprintf("made %s bid", noteTrump)
		} else {
			// Team13 missed the bid.
			sc[1] = last13 - (bidValue * multiplier)
			note13 = fmt.Sprintf("miss %s bid", noteTrump)
		}
		sc[0] = last02 + points02
	}
	s.scores = append(s.scores, sc)

	if multiplier == 2 && (madeBid02 || madeBid13) {
		s.setTopScore62()
	}

	// Check to see if there is a winner.
	if madeBid02 && (last02+(bidValue*multiplier) >= s.ToWin()) {
		s.setWinner(0)
		note02 = "bid out " + noteTrump
	} else if madeBid13 && (last13+(bidValue*multiplier) >= s.ToWin()) {
		s.setWinner(1)
		note13 = "bid out " + noteTrump
	}

	if note02 != "" {
		s.attachNote(0, len(s.scores)-1, note02)
	}
	if note13 != "" {
		s.attachNote(1, len(s.scores)-1, note13)
	}

	return s.Winner() != NoWinner, nil
}
