package game

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestNewScore(t *testing.T) {
	score := NewScore().(*score)
	if score.toWin != 52 {
		t.Errorf("toWin is not 52")
	}
	if len(score.scores) != 0 {
		t.Errorf("len(scores) != 0")
	}
}

func TestNewScoreFromEncoded(t *testing.T) {
	testCases := []struct {
		name        string
		encoded     string
		wantToWin   int
		wantScores  [][]int
		wantErr     bool
		encOverride string
	}{
		{
			name:        "empty string",
			encoded:     "",
			wantToWin:   52,
			encOverride: "52-",
		},
		{
			name:    "bad toWin",
			encoded: "A-0|1",
			wantErr: true,
		},
		{
			name:    "missing toWin",
			encoded: "-0|1",
			wantErr: true,
		},
		{
			name:    "bad point02",
			encoded: "52-A|1",
			wantErr: true,
		},
		{
			name:    "bad point13",
			encoded: "52-0|A",
			wantErr: true,
		},
		{
			name:    "incorrect pairs",
			encoded: "52-0|1||2|",
			wantErr: true,
		},
		{
			name:       "valid",
			encoded:    "52-0|1||2|3||4|5||6|7",
			wantToWin:  52,
			wantScores: [][]int{{0, 1}, {2, 3}, {4, 5}, {6, 7}},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			scoreS, err := NewScoreFromEncoded(tc.encoded)

			if tc.wantErr && err == nil {
				t.Fatal("missing expected error")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tc.wantErr {
				return
			}

			score := scoreS.(*score)

			if got, want := score.toWin, tc.wantToWin; got != want {
				t.Errorf("score.toWin=%d want=%d", got, want)
			}
			if diff := cmp.Diff(tc.wantScores, score.Scores()); diff != "" {
				t.Errorf("hand.Scores() mismatch (-want +got):\n%s", diff)
			}

			// Re-encode the score and make sure it matches.
			encoded := tc.encoded
			if tc.encOverride != "" {
				encoded = tc.encOverride
			}
			if got, want := score.Encoded(), encoded; got != want {
				t.Errorf("re-encoding does not match got=%q want=%q", got, want)
			}
		})
	}
}

func TestScoreSetTopScore62(t *testing.T) {
	score := NewScore()
	if got, want := score.ToWin(), 52; got != want {
		t.Errorf("got=%d want=%d", got, want)
	}
	score.SetTopScore62()
	if got, want := score.ToWin(), 62; got != want {
		t.Errorf("got=%d want=%d", got, want)
	}
}

func TestScoreScores(t *testing.T) {
	score := NewScore()
	score.AddTally(buildTally(t, 8, 1, 2))
	score.AddTally(buildTally(t, 8, 3, 4))
	score.AddTally(buildTally(t, 8, 5, 6))
	// Check that we're keeping a running score.
	want := [][]int{
		{1, 2},
		{4, 6},
		{9, 12},
	}
	if diff := cmp.Diff(want, score.Scores()); diff != "" {
		t.Errorf("score.Scores() mismatch (-want +got):\n%s", diff)
	}
}

func TestScoreCurrentScore(t *testing.T) {
	testCases := []struct {
		name    string
		tallies [][]int
		want    []int
	}{
		{
			name: "empty is 0-0",
			want: []int{0, 0},
		},
		{
			name:    "one entry",
			tallies: [][]int{{1, 2}},
			want:    []int{1, 2},
		},
		{
			name: "multiple tallies, current is running total",
			tallies: [][]int{
				{1, 2},
				{3, 4},
				{-3, 6},
			},
			want: []int{1, 12},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			score := NewScore()
			for _, s := range tc.tallies {
				tally := buildTally(t, 8, s[0], s[1])
				if err := score.AddTally(tally); err != nil {
					t.Fatal(err)
				}
			}

			if diff := cmp.Diff(tc.want, score.CurrentScore()); diff != "" {
				t.Errorf("score.CurrentScore() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestScoreAddTally(t *testing.T) {
	testCases := []struct {
		name    string
		tallies []Tally
		want    []int
		wantErr bool
	}{
		{
			name:    "zero tallies",
			tallies: []Tally{},
			want:    []int{0, 0},
		},
		{
			name:    "one tally",
			tallies: []Tally{buildTally(t, 8, 1, 2)},
			want:    []int{1, 2},
		},
		{
			name: "mulitple tallies",
			tallies: []Tally{
				buildTally(t, 8, 1, 2),
				buildTally(t, 8, -3, 12),
			},
			want: []int{-2, 14},
		},
		{
			name:    "incomplete tally",
			tallies: []Tally{buildTally(t, 7, 1, 2)}, // 7 is incomplete
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			score := NewScore()
			for _, tally := range tc.tallies {

				err := score.AddTally(tally)

				if tc.wantErr && err == nil {
					t.Fatal("missing expected error")
				}
				if !tc.wantErr && err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if tc.wantErr {
					return
				}
			}

			if diff := cmp.Diff(tc.want, score.CurrentScore()); diff != "" {
				t.Errorf("score.CurrentScore() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
