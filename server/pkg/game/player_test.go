package game

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/squee1945/threespot/server/pkg/deck"
)

func TestNewPlayer(t *testing.T) {
	testCases := []struct {
		name       string
		playerID   string
		playerName string
		wantErr    bool
	}{
		{
			name:       "id required",
			playerID:   "",
			playerName: "Some Name",
			wantErr:    true,
		},
		{
			name:       "name required",
			playerID:   "abc123",
			playerName: "",
			wantErr:    true,
		},
		{
			name:       "id must be valid",
			playerID:   "a 3",
			playerName: "",
			wantErr:    true,
		},
		{
			name:       "name must bevalid",
			playerID:   "abc123",
			playerName: strings.Repeat("a", 101),
			wantErr:    true,
		},
		{
			name:       "valid",
			playerID:   "abc123",
			playerName: "This is valid",
			wantErr:    false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			p, err := NewPlayer(tc.playerID, tc.playerName)

			if tc.wantErr && err == nil {
				t.Fatal("wanted error, got err=nil")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tc.wantErr {
				return
			}

			if got, want := p.ID(), tc.playerID; got != want {
				t.Errorf("incorrect ID, got=%q, want=%q", got, want)
			}

			if got, want := p.Name(), tc.playerName; got != want {
				t.Errorf("incorrect name, got=%q, want=%q", got, want)
			}
		})
	}
}

func TestSetHand(t *testing.T) {
	p, err := NewPlayer("abc123", "some-name")
	if err != nil {
		t.Fatal(err)
	}

	if len(p.Hand()) > 0 {
		t.Fatalf("hand must start empty, got=%v", p.Hand())
	}

	hand := []deck.Card{
		deck.Card{Num: "3", Suit: deck.Spades},
		deck.Card{Num: "5", Suit: deck.Hearts},
	}
	p.SetHand(hand)

	if diff := cmp.Diff(hand, p.Hand()); diff != "" {
		t.Errorf("hand mismatch (-want +got):\n%s", diff)
	}
}
