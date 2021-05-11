package game

import (
	"fmt"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/squee1945/threespot/server/pkg/deck"
)

func TestNewPassingRoundFromEncoded(t *testing.T) {
	testCases := []struct {
		name        string
		encoded     string
		wantLeadPos int
		wantCards   []deck.Card
		wantErr     bool
	}{
		{
			name:    "empty string",
			encoded: "0|",
		},
		{
			name:    "lead pos not int",
			encoded: "?|7C",
			wantErr: true,
		},
		{
			name:    "lead pos too small",
			encoded: "-1|7C",
			wantErr: true,
		},
		{
			name:    "lead pos too big",
			encoded: "4|7C",
			wantErr: true,
		},
		{
			name:    "invalid card encoding",
			encoded: "0|?",
			wantErr: true,
		},
		{
			name:        "no cards",
			encoded:     "1|",
			wantLeadPos: 1,
		},
		{
			name:        "one passed card",
			encoded:     "0|7C",
			wantLeadPos: 0,
			wantCards:   []deck.Card{buildCard(t, "7C")},
		},
		{
			name:        "four passed cards",
			encoded:     "0|7C|8C|9C|TC",
			wantLeadPos: 0,
			wantCards:   []deck.Card{buildCard(t, "7C"), buildCard(t, "8C"), buildCard(t, "9C"), buildCard(t, "TC")},
		},
		{
			name:    "too many cards",
			encoded: "0|7C|8C|9C|TC|JC",
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r, err := NewPassingRoundFromEncoded(tc.encoded)

			if tc.wantErr && err == nil {
				t.Fatal("missing expected error")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tc.wantErr {
				return
			}

			if got, want := r.LeadPos(), tc.wantLeadPos; got != want {
				t.Errorf("LeadPos()=%d want=%d", got, want)
			}
			if diff := cmp.Diff(tc.wantCards, r.Cards(), compareCards); diff != "" {
				t.Errorf("cards mismatch (-want +got):\n%s", diff)
			}

			// Re-encode the passing round and make sure it matches.
			if got, want := r.Encoded(), strings.ToUpper(tc.encoded); got != want {
				t.Errorf("re-encoding does not match got=%q want=%q", got, want)
			}
		})
	}
}

func TestNewPassingRound(t *testing.T) {
	testCases := []struct {
		name    string
		leadPos int
		wantErr bool
	}{
		{
			name:    "lead pos too small",
			leadPos: -1,
			wantErr: true,
		},
		{
			name:    "lead pos too big",
			leadPos: 4,
			wantErr: true,
		},
		{
			name:    "valid lead pos",
			leadPos: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r, err := NewPassingRound(tc.leadPos)

			if tc.wantErr && err == nil {
				t.Fatal("missing expected error")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tc.wantErr {
				return
			}

			if got, want := r.LeadPos(), tc.leadPos; got != want {
				t.Errorf("LeadPos=%d want=%d", got, want)
			}
			if got := r.Cards(); got != nil {
				t.Errorf("Cards()=%v want=nil", got)
			}
		})
	}
}

func TestIsDoneAndNumPassed(t *testing.T) {
	testCases := []struct {
		encoded       string
		wantIsDone    bool
		wantNumPassed int
	}{
		{"0|", false, 0},
		{"0|7C", false, 1},
		{"0|7C|8C", false, 2},
		{"0|7C|8C|9C", false, 3},
		{"0|7C|8C|9C|TC", true, 4},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("case %d", i), func(t *testing.T) {
			r := buildPassingRound(t, tc.encoded)

			if got, want := r.IsDone(), tc.wantIsDone; got != want {
				t.Errorf("IsDone()=%t want=%t", got, want)
			}
			if got, want := r.NumPassed(), tc.wantNumPassed; got != want {
				t.Errorf("NumPassed()=%d want=%d", got, want)
			}
		})
	}
}

func TestCurrentTurnPosErrors(t *testing.T) {
	encoded := "0|7C|8C|9C|TC"
	r := buildPassingRound(t, encoded)

	_, err := r.CurrentTurnPos()
	if err == nil {
		t.Errorf("missing expected error")
	}
}

func TestCurrentTurnPos(t *testing.T) {
	testCases := []struct {
		name    string
		encoded string
		want    int
	}{
		{
			name:    "first card",
			encoded: "1|",
			want:    1,
		},
		{
			name:    "last catd",
			encoded: "1|7C|8C|9C",
			want:    0,
		},
		{
			name:    "wrap test",
			encoded: "3|7C|8C",
			want:    1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r := buildPassingRound(t, tc.encoded)

			got, err := r.CurrentTurnPos()
			if err != nil {
				t.Fatal(err)
			}

			if got != tc.want {
				t.Errorf("CurrentTurnPos()=%d want=%d", got, tc.want)
			}
		})
	}
}

func buildPassingRound(t *testing.T, encoded string) PassingRound {
	r, err := NewPassingRoundFromEncoded(encoded)
	if err != nil {
		t.Fatal(err)
	}
	return r
}
