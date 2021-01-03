package game

import (
	"context"
	"strings"
	"testing"

	"github.com/squee1945/threespot/server/pkg/storage"
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
			playerID:   "ABC123",
			playerName: "",
			wantErr:    true,
		},
		{
			name:       "id must be valid",
			playerID:   "A 3",
			playerName: "",
			wantErr:    true,
		},
		{
			name:       "name must be valid",
			playerID:   "ABC123",
			playerName: strings.Repeat("a", 101),
			wantErr:    true,
		},
		{
			name:       "valid",
			playerID:   "ABC123",
			playerName: "This is valid",
			wantErr:    false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			p, err := NewPlayer(ctx, storage.NewFakePlayerStore(nil), tc.playerID, tc.playerName)

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

// func TestSetHand(t *testing.T) {
// 	p, err := NewPlayer("abc123", "some-name")
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	if len(p.Hand()) > 0 {
// 		t.Fatalf("hand must start empty, got=%v", p.Hand())
// 	}

// 	c1, err := deck.NewCard("3", deck.Spades)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	c2, err := deck.NewCard("5", deck.Hearts)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	hand := []deck.Card{c1, c2}
// 	p.SetHand(hand)

// 	if diff := cmp.Diff(hand, p.Hand()); diff != "" {
// 		t.Errorf("hand mismatch (-want +got):\n%s", diff)
// 	}
// }
