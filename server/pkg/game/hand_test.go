package game

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/squee1945/threespot/server/pkg/deck"
)

func TestNewHandFromEncoded(t *testing.T) {
	testCases := []struct {
		name    string
		encoded string
		want    *hand
		wantErr bool
	}{
		{
			name: "empty string",
			want: &hand{},
		},
		{
			name:    "single card",
			encoded: "3S",
			want:    &hand{cards: []deck.Card{card(t, "3", deck.Spades)}},
		},
		{
			name:    "multiple cards",
			encoded: "3S-5H-TC",
			want: &hand{
				cards: []deck.Card{card(t, "3", deck.Spades), card(t, "5", deck.Hearts), card(t, "T", deck.Clubs)},
			},
		},
		{
			name:    "lowercase cards",
			encoded: "ts",
			want:    &hand{cards: []deck.Card{card(t, "T", deck.Spades)}},
		},
		{
			name:    "invalid cards",
			encoded: "not-valid",
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := NewHandFromEncoded(tc.encoded)

			if tc.wantErr && err == nil {
				t.Fatal("missing expected error")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("unepxected error: %v", err)
			}
			if tc.wantErr {
				return
			}
			if diff := cmp.Diff(tc.want, got, cmp.Comparer(func(h1, h2 Hand) bool { return h1.Encoded() == h2.Encoded() })); diff != "" {
				t.Errorf("NewHandFromEncoded() mismatch (-want +got):\n%s", diff)
			}
		})
	}

}

func card(t *testing.T, val string, suit deck.Suit) deck.Card {
	t.Helper()
	c, err := deck.NewCard(val, suit)
	if err != nil {
		t.Fatal(err)
	}
	return c
}
