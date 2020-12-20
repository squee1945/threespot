package deck

import "testing"

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

	// Check for special cards
	for _, c := range []Card{
		{Num: "7", Suit: Clubs},
		{Num: "7", Suit: Dimonds},
		{Num: "3", Suit: Spades},
		{Num: "5", Suit: Hearts},
	} {
		if wantNum, wantSuit := c.Num, c.Suit; !hasCard(t, d, wantNum, wantSuit) {
			t.Errorf("card %s of %s not found", wantNum, wantSuit)
		}
	}
}

func TestShuffle(t *testing.T) {
	// TODO
}

func TestDeal(t *testing.T) {
	// TODO
}

func hasCard(t *testing.T, d Deck, wantNum string, wantSuit Suit) bool {
	t.Helper()
	for _, c := range d.cards {
		if c.Num == wantNum && c.Suit == wantSuit {
			return true
		}
	}
	return false
}
