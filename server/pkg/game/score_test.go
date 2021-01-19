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
	score.setTopScore62()
	if got, want := score.ToWin(), 62; got != want {
		t.Errorf("got=%d want=%d", got, want)
	}
}

func TestScoreScores(t *testing.T) {
	score := NewScore()
	score.addTally(buildBiddingRound(t, "0|7|P|P|P"), buildTally(t, 8, 7, 3))
	score.addTally(buildBiddingRound(t, "0|7|P|P|P"), buildTally(t, 8, 8, 2))
	score.addTally(buildBiddingRound(t, "0|7|P|P|P"), buildTally(t, 8, 9, 1))
	// Check that we're keeping a running score.
	want := [][]int{
		{7, 3},
		{15, 5},
		{24, 6},
	}
	if diff := cmp.Diff(want, score.Scores()); diff != "" {
		t.Errorf("score.Scores() mismatch (-want +got):\n%s", diff)
	}
}

func TestScoreCurrentScore(t *testing.T) {
	testCases := []struct {
		name    string
		bids    []string
		tallies [][]int
		want    []int
	}{
		{
			name: "empty is 0-0",
			want: []int{0, 0},
		},
		{
			name:    "one entry",
			tallies: [][]int{{9, 1}},
			bids:    []string{"0|7|P|P|P"},
			want:    []int{9, 1},
		},
		{
			name: "multiple tallies, current is running total",
			bids: []string{
				"0|7|P|P|P",
				"0|7|P|P|P",
				"0|7|P|P|P",
			},
			tallies: [][]int{
				{7, 3},
				{8, 2},
				{10, -2},
			},
			want: []int{25, 3},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if len(tc.bids) != len(tc.tallies) {
				t.Fatalf("bids and tallies must be same length")
			}
			score := NewScore()
			for i := range tc.tallies {
				tally := buildTally(t, 8, tc.tallies[i][0], tc.tallies[i][1])
				if err := score.addTally(buildBiddingRound(t, tc.bids[i]), tally); err != nil {
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
		bid     string
		tally   Tally
		want    []int
		wantErr bool
	}{
		{
			name:  "team02 makes bid",
			bid:   "0|P|P|9|P",
			tally: buildTally(t, 8, 9, 1),
			want:  []int{9, 1},
		},
		{
			name:  "team02 misses bid",
			bid:   "0|P|P|9|P",
			tally: buildTally(t, 8, 8, 1),
			want:  []int{-9, 1},
		},
		{
			name:  "team02 makes no trump bid",
			bid:   "0|P|P|9N|P",
			tally: buildTally(t, 8, 10, 0),
			want:  []int{20, 0},
		},
		{
			name:  "team02 misses no trump bid",
			bid:   "0|P|P|9N|P",
			tally: buildTally(t, 8, 8, 2),
			want:  []int{-18, 2},
		},
		{
			name:  "team13 makes bid",
			bid:   "0|P|9|P|P",
			tally: buildTally(t, 8, 1, 9),
			want:  []int{1, 9},
		},
		{
			name:  "team13 misses bid",
			bid:   "0|P|9|P|P",
			tally: buildTally(t, 8, 1, 8),
			want:  []int{1, -9},
		},
		{
			name:  "team13 makes no trump bid",
			bid:   "0|P|9N|P|P",
			tally: buildTally(t, 8, 0, 10),
			want:  []int{0, 20},
		},
		{
			name:  "team13 misses no trump bid",
			bid:   "0|P|9N|P|P",
			tally: buildTally(t, 8, 2, 8),
			want:  []int{2, -18},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			score := NewScore()

			err := score.addTally(buildBiddingRound(t, tc.bid), tc.tally)

			if tc.wantErr && err == nil {
				t.Fatal("missing expected error")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tc.wantErr {
				return
			}

			if diff := cmp.Diff(tc.want, score.CurrentScore()); diff != "" {
				t.Errorf("score.CurrentScore() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
