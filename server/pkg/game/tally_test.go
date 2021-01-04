package game

import (
	"fmt"
	"testing"
)

func TestNewTally(t *testing.T) {
	tally := NewTally()
	got02, got13 := tally.Points()
	if got02 != 0 {
		t.Errorf("points02 is not 0")
	}
	if got13 != 0 {
		t.Errorf("points13 is not 0")
	}
	if tally.IsDone() {
		t.Errorf("IsDone must be false for new tally")
	}
}

func TestNewTallyFromEncoded(t *testing.T) {
	testCases := []struct {
		name             string
		encoded          string
		wantPoints02     int
		wantPoints13     int
		wantIsDone       bool
		wantErr          bool
		encodingOverride string
	}{
		{
			name:             "empty string",
			wantPoints02:     0,
			wantPoints13:     0,
			wantIsDone:       false,
			encodingOverride: "0|0|0",
		},
		{
			name:    "too short",
			encoded: "0|1",
			wantErr: true,
		},
		{
			name:    "too long",
			encoded: "0|1|2|3",
			wantErr: true,
		},
		{
			name:    "bad count",
			encoded: "A|2|3",
			wantErr: true,
		},
		{
			name:    "bad points02",
			encoded: "1|A|3",
			wantErr: true,
		},
		{
			name:    "bad points13",
			encoded: "1|2|A",
			wantErr: true,
		},
		{
			name:         "valid",
			encoded:      "1|2|3",
			wantPoints02: 2,
			wantPoints13: 3,
			wantIsDone:   false,
		},
		{
			name:         "valid, tally is done",
			encoded:      "8|2|3",
			wantPoints02: 2,
			wantPoints13: 3,
			wantIsDone:   true,
		},
		{
			name:         "valid can encode negative",
			encoded:      "1|-2|0",
			wantPoints02: -2,
			wantPoints13: 0,
			wantIsDone:   false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tally, err := NewTallyFromEncoded(tc.encoded)

			if tc.wantErr && err == nil {
				t.Fatal("missing expected error")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tc.wantErr {
				return
			}

			got02, got13 := tally.Points()
			if got, want := got02, tc.wantPoints02; got != want {
				t.Errorf("incorrect points02 got=%d want=%d", got, want)
			}
			if got, want := got13, tc.wantPoints13; got != want {
				t.Errorf("incorrect points13 got=%d want=%d", got, want)
			}
			if got, want := tally.IsDone(), tc.wantIsDone; got != want {
				t.Errorf("IsDone()=%t want=%t", got, want)
			}

			// Re-encode the tally and make sure it matches.
			wantEncoded := tc.encoded
			if tc.encodingOverride != "" {
				wantEncoded = tc.encodingOverride
			}
			if got, want := tally.Encoded(), wantEncoded; got != want {
				t.Errorf("re-encoding does not match got=%q want=%q", got, want)
			}
		})
	}
}

func TestTallyAddTrick(t *testing.T) {
	type want struct {
		isDone   bool
		points02 int
		points13 int
	}
	testCases := []struct {
		name    string
		tally   Tally
		trick   Trick
		want    want
		wantErr bool
	}{
		{
			name:  "02 wins, no specials",
			tally: buildTally(t, 0, 0, 0),
			trick: buildTrick(t, "N", 0, "AH", "8S", "7D", "7C"),
			want:  want{false, 1, 0},
		},
		{
			name:  "13 wins, no specials",
			tally: buildTally(t, 0, 0, 0),
			trick: buildTrick(t, "N", 0, "8H", "AH", "7D", "7C"),
			want:  want{false, 0, 1},
		},
		{
			name:  "02 wins, with 5",
			tally: buildTally(t, 0, 0, 0),
			trick: buildTrick(t, "N", 0, "5H", "8S", "7D", "7C"),
			want:  want{false, 6, 0},
		},
		{
			name:  "02 wins, with 3",
			tally: buildTally(t, 0, 0, 0),
			trick: buildTrick(t, "N", 0, "3S", "8H", "7D", "7C"),
			want:  want{false, -2, 0},
		},
		{
			name:  "02 wins, with 5 and 3",
			tally: buildTally(t, 0, 0, 0),
			trick: buildTrick(t, "N", 0, "3S", "5H", "7D", "7C"),
			want:  want{false, 3, 0},
		},
		{
			name:  "02 tally incremented",
			tally: buildTally(t, 1, 2, 0),
			trick: buildTrick(t, "N", 0, "AH", "8S", "7D", "7C"),
			want:  want{false, 3, 0},
		},
		{
			name:  "13 tally incremented",
			tally: buildTally(t, 2, 2, 5),
			trick: buildTrick(t, "N", 0, "8H", "AH", "7D", "7C"),
			want:  want{false, 2, 6},
		},
		{
			name:  "02 wins final trick",
			tally: buildTally(t, 7, 0, 9),
			trick: buildTrick(t, "N", 0, "AH", "8S", "7D", "7C"),
			want:  want{true, 1, 9},
		},
		{
			name:    "incomplete trick returns error",
			tally:   buildTally(t, 2, 2, 5),
			trick:   buildTrick(t, "N", 0, "8H", "AH"), // Incomplete.
			wantErr: true,
		},
		{
			name:    "tally can only accept 8 tricks",
			tally:   buildTally(t, 8, 2, 5), // Already done.
			trick:   buildTrick(t, "N", 0, "AH", "8S", "7D", "7C"),
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.tally.AddTrick(tc.trick)

			if tc.wantErr && err == nil {
				t.Fatal("missing expected error")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tc.wantErr {
				return
			}

			got02, got13 := tc.tally.Points()
			if got, want := got02, tc.want.points02; got != want {
				t.Errorf("incorrect points02 got=%d want=%d", got, want)
			}
			if got, want := got13, tc.want.points13; got != want {
				t.Errorf("incorrect points13 got=%d want=%d", got, want)
			}
			if got, want := tc.tally.IsDone(), tc.want.isDone; got != want {
				t.Errorf("IsDone()=%t want=%t", got, want)
			}
		})
	}
}

func TestTallyIsDone(t *testing.T) {
	testCases := []struct {
		count int
		want  bool
	}{
		{0, false},
		{1, false},
		{7, false},
		{8, true},
	}
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%d tricks", tc.count), func(t *testing.T) {
			if got, want := buildTally(t, tc.count, 0, 0).IsDone(), tc.want; got != want {
				t.Errorf("tally.IsDone()=%t want=%t", got, want)
			}
		})
	}
}

func buildTally(t *testing.T, count, points02, points13 int) Tally {
	t.Helper()
	return &tally{
		points02:   points02,
		points13:   points13,
		trickCount: count,
	}
}
