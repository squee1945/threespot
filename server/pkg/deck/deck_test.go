package deck

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestNewDeckHasCorrectCards(t *testing.T) {
	nd := NewDeck()
	d := nd.(*deck)

	if got, want := len(d.cards), 32; got != want {
		t.Fatalf("incorrect number of cards, got=%d, want=%d", got, want)
	}

	for _, wantSuit := range []Suit{Hearts, Diamonds, Spades, Clubs} {
		for _, wantNum := range []string{"8", "9", "T", "J", "Q", "K", "A"} {
			if !hasCard(t, d, wantNum, wantSuit) {
				t.Errorf("card %s of %s not found", wantNum, wantSuit)
			}
		}
	}

	// Check for special cards.
	for _, c := range []Card{
		{Num: "7", Suit: Clubs},
		{Num: "7", Suit: Diamonds},
		{Num: "3", Suit: Spades},
		{Num: "5", Suit: Hearts},
	} {
		if wantNum, wantSuit := c.Num, c.Suit; !hasCard(t, d, wantNum, wantSuit) {
			t.Errorf("card %s of %s not found", wantNum, wantSuit)
		}
	}

	// Verify all cards are unique.
	m := make(map[Card]bool)
	for i := 0; i < 32; i++ {
		if _, present := m[d.cards[i]]; present {
			t.Fatalf("card already present, %s", d.cards[i])
		}
		m[d.cards[i]] = true
	}
}

func TestShuffle(t *testing.T) {
	nd := NewDeck()
	d := nd.(*deck)
	original := make([]Card, 32)
	for i := 0; i < 32; i++ {
		original[i] = d.cards[i]
	}
	d.Shuffle()
	if diff := cmp.Diff(original, d.cards); diff == "" {
		t.Errorf("hand matches")
	}
	// Verify all cards are unique.
	m := make(map[Card]bool)
	for i := 0; i < 32; i++ {
		if _, present := m[d.cards[i]]; present {
			t.Fatalf("card already present, %s", d.cards[i])
		}
		m[d.cards[i]] = true
	}
}

func TestDeal(t *testing.T) {
	deck := NewDeck()
	hands := deck.Deal()

	if got, want := len(hands), 4; got != want {
		t.Fatalf("incorrect number of hands dealt, got=%d, want=%d", got, want)
	}
	for i := 0; i < 4; i++ {
		if got, want := len(hands[i]), 8; got != want {
			t.Fatalf("incorrect number of cards dealt, got=%d, want=%d", got, want)
		}
	}
	// Verify all cards are unique.
	m := make(map[Card]bool)
	for i := 0; i < 4; i++ {
		for j := 0; j < 8; j++ {
			if _, present := m[hands[i][j]]; present {
				t.Fatalf("card already present, %s", hands[i][j])
			}
			m[hands[i][j]] = true
		}
	}
}

func hasCard(t *testing.T, d *deck, wantNum string, wantSuit Suit) bool {
	t.Helper()
	for _, c := range d.cards {
		if c.Num == wantNum && c.Suit == wantSuit {
			return true
		}
	}
	return false
}
