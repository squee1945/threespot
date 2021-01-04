package game

import (
	"strings"
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
			want:    &hand{cards: []deck.Card{buildCard(t, "3S")}},
		},
		{
			name:    "multiple cards",
			encoded: "3S-5H-TC",
			want: &hand{
				cards: []deck.Card{buildCard(t, "3S"), buildCard(t, "5H"), buildCard(t, "TC")},
			},
		},
		{
			name:    "lowercase cards",
			encoded: "ts",
			want:    &hand{cards: []deck.Card{buildCard(t, "TS")}},
		},
		{
			name:    "invalid cards",
			encoded: "not-valid",
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			hand, err := NewHandFromEncoded(tc.encoded)

			if tc.wantErr && err == nil {
				t.Fatal("missing expected error")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tc.wantErr {
				return
			}
			if diff := cmp.Diff(tc.want, hand, cmp.Comparer(func(h1, h2 Hand) bool { return h1.Encoded() == h2.Encoded() })); diff != "" {
				t.Errorf("NewHandFromEncoded() mismatch (-want +got):\n%s", diff)
			}

			// Re-encode the hand and make sure it matches.
			if got, want := hand.Encoded(), strings.ToUpper(tc.encoded); got != want {
				t.Errorf("re-encoding does not match got=%q want=%q", got, want)
			}
		})
	}
}

func TestNewHand(t *testing.T) {
	testCases := []struct {
		name    string
		cards   []string
		wantErr bool
	}{
		{
			name: "empty hand",
		},
		{
			name:  "one card",
			cards: []string{"3S"},
		},
		{
			name:  "multiple cards",
			cards: []string{"7D", "3S", "5H", "7S"},
		},
		{
			name:    "duplicate cards",
			cards:   []string{"3S", "3S"},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cards := buildCards(t, tc.cards)

			hand, err := NewHand(cards)

			if tc.wantErr && err == nil {
				t.Fatal("missing expected error")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tc.wantErr {
				return
			}

			if diff := cmp.Diff(cards, hand.Cards(), compareCards); diff != "" {
				t.Errorf("hand.Cards() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestHandCards(t *testing.T) {
	testCases := []struct {
		name  string
		cards []string
	}{
		{
			name: "empty hand",
		},
		{
			name:  "single card",
			cards: []string{"3S"},
		},
		{
			name:  "full hand",
			cards: []string{"3S", "8S", "9S", "TS", "JS", "QS", "KS", "AS"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cards := buildCards(t, tc.cards)
			hand := buildHand(t, tc.cards)

			if diff := cmp.Diff(cards, hand.Cards(), compareCards, ignoreCardOrder); diff != "" {
				t.Errorf("hand.Cards() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestHandCardsOrdered(t *testing.T) {
	testCases := []struct {
		name  string
		cards []string
		want  []string
	}{
		{
			name:  "suits ordered",
			cards: []string{"8C", "8D", "8S", "8H"},
			want:  []string{"8H", "8S", "8D", "8C"},
		},
		{
			name:  "numbers ordered",
			cards: []string{"9S", "TS", "JS", "QS", "KS", "AS"},
			want:  []string{"AS", "KS", "QS", "JS", "TS", "9S"},
		},
		{
			name:  "full test",
			cards: []string{"3S", "5H", "TH", "9C", "QC", "TS", "7D", "AH"},
			want:  []string{"AH", "TH", "5H", "TS", "3S", "7D", "QC", "9C"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			hand := buildHand(t, tc.cards)

			if diff := cmp.Diff(buildCards(t, tc.want), hand.Cards(), compareCards); diff != "" {
				t.Errorf("hand.Cards() out of order (-want +got):\n%s", diff)
			}
		})
	}
}

func TestHandContains(t *testing.T) {
	hand := buildHand(t, []string{"3S", "5H", "8D"})
	testCases := []struct {
		card string
		want bool
	}{
		{"3S", true},
		{"5H", true},
		{"8D", true},
		{"JC", false},
	}

	for _, tc := range testCases {
		t.Run(tc.card, func(t *testing.T) {
			if got, want := hand.Contains(buildCard(t, tc.card)), tc.want; got != want {
				t.Errorf("hand.Contains()=%t want=%t", got, want)
			}
		})
	}
}

func TestHandContainsSuit(t *testing.T) {
	hand := buildHand(t, []string{"3S", "5H", "8S"})
	testCases := []struct {
		ignore string
		suit   string
		want   bool
		name   string
	}{
		{"5H", "H", false, "No other hearts"},
		{"5H", "D", false, "No diamonds"},
		{"5H", "S", true, "Has a spade"},
		{"3S", "H", true, "Has a heart"},
		{"3S", "S", true, "Has another spade"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if got, want := hand.ContainsSuit(buildSuit(t, tc.suit), buildCard(t, tc.ignore)), tc.want; got != want {
				t.Errorf("hand.ContainsSuit()=%t, want=%t", got, want)
			}
		})
	}
}
func TestHandRemoveCard(t *testing.T) {
	testCases := []struct {
		name    string
		cards   []string
		remove  string
		want    []string
		wantErr bool
	}{
		{
			name:    "remove from empty hand",
			cards:   []string{},
			remove:  "3S",
			wantErr: true,
		},
		{
			name:   "remove only card",
			cards:  []string{"3S"},
			remove: "3S",
			want:   []string{},
		},
		{
			name:   "remove some card",
			cards:  []string{"3S", "5H", "7D", "TC"},
			remove: "7D",
			want:   []string{"3S", "5H", "TC"},
		},
		{
			name:    "remove non-existent card",
			cards:   []string{"3S", "5H", "7D", "TC"},
			remove:  "AH",
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			hand := buildHand(t, tc.cards)

			hand, err := hand.RemoveCard(buildCard(t, tc.remove))

			if tc.wantErr && err == nil {
				t.Fatal("missing expected error")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tc.wantErr {
				return
			}

			if diff := cmp.Diff(buildCards(t, tc.want), hand.Cards(), compareCards, ignoreCardOrder); diff != "" {
				t.Errorf("hand.Cards() incorrect (-want +got):\n%s", diff)
			}
		})
	}
}
func TestHandIsEmpty(t *testing.T) {
	if got, want := buildHand(t, []string{}).IsEmpty(), true; got != want {
		t.Errorf("empty hand not empty")
	}
	if got, want := buildHand(t, []string{"3S"}).IsEmpty(), false; got != want {
		t.Errorf("non-empty hand is empty")
	}
}

func buildCards(t *testing.T, encodedCards []string) []deck.Card {
	t.Helper()
	var cards []deck.Card
	for _, encoded := range encodedCards {
		cards = append(cards, buildCard(t, encoded))
	}
	return cards
}

func buildHand(t *testing.T, encodedCards []string) Hand {
	t.Helper()
	cards := buildCards(t, encodedCards)
	hand, err := NewHand(cards)
	if err != nil {
		t.Fatal(err)
	}
	return hand
}
