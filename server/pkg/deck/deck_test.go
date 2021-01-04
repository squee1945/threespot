package deck

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestNewDeckHasCorrectCards(t *testing.T) {
	nd, err := NewDeck()
	if err != nil {
		t.Fatal(err)
	}
	d := nd.(*deck)

	if got, want := len(d.cards), 32; got != want {
		t.Fatalf("incorrect number of cards, got=%d, want=%d", got, want)
	}

	for _, wantSuit := range []Suit{Hearts, Diamonds, Spades, Clubs} {
		for _, wantNum := range []string{"8", "9", "T", "J", "Q", "K", "A"} {
			want, err := NewCard(wantNum, wantSuit)
			if err != nil {
				t.Fatal(err)
			}
			if !hasCard(t, d, want) {
				t.Errorf("card %s not found", want)
			}
		}
	}

	// Check for special cards.
	c1, err := NewCard("7", Clubs)
	if err != nil {
		t.Fatal(err)
	}
	c2, err := NewCard("7", Diamonds)
	if err != nil {
		t.Fatal(err)
	}
	c3, err := NewCard("3", Spades)
	if err != nil {
		t.Fatal(err)
	}
	c4, err := NewCard("5", Hearts)
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range []Card{c1, c2, c3, c4} {
		if !hasCard(t, d, c) {
			t.Errorf("card %s not found", c)
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
	nd, err := NewDeck()
	if err != nil {
		t.Fatal(err)
	}
	d := nd.(*deck)
	original := make([]Card, 32)
	for i := 0; i < 32; i++ {
		original[i] = d.cards[i]
	}
	d.Shuffle()
	if diff := cmp.Diff(original, d.cards, cmp.Comparer(func(c1, c2 Card) bool { return c1.Encoded() == c2.Encoded() })); diff == "" {
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
	deck, err := NewDeck()
	if err != nil {
		t.Fatal(err)
	}
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

func hasCard(t *testing.T, d *deck, want Card) bool {
	t.Helper()
	for _, c := range d.cards {
		if c.IsSameAs(want) {
			return true
		}
	}
	return false
}
