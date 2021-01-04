package deck

import "testing"

func TestNewCardFromEncoded(t *testing.T) {
	testCases := []struct {
		name     string
		encoded  string
		wantNum  string
		wantSuit string
		wantErr  bool
	}{
		{
			name:    "empty string",
			wantErr: true,
		},
		{
			name:    "too short",
			encoded: "3",
			wantErr: true,
		},
		{
			name:    "too long",
			encoded: "3Hearts",
			wantErr: true,
		},
		{
			name:    "bad num",
			encoded: "4S",
			wantErr: true,
		},
		{
			name:    "bad suit",
			encoded: "5?",
			wantErr: true,
		},
		{
			name:    "no-trump is bad suit",
			encoded: "9N",
			wantErr: true,
		},
		{
			name:     "valid num card",
			encoded:  "5H",
			wantNum:  "5",
			wantSuit: "H",
		},
		{
			name:     "valid face card",
			encoded:  "JS",
			wantNum:  "J",
			wantSuit: "S",
		},
		{
			name:     "valid ace",
			encoded:  "AC",
			wantNum:  "A",
			wantSuit: "C",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			card, err := NewCardFromEncoded(tc.encoded)
			if tc.wantErr && err == nil {
				t.Fatal("missing expected error")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tc.wantErr {
				return
			}

			if got, want := card.Num(), tc.wantNum; got != want {
				t.Errorf("incorrect num got=%q, want=%q", got, want)
			}
			if got, want := card.Suit(), buildSuit(t, tc.wantSuit); !got.IsSameAs(want) {
				t.Errorf("incorrect suit got=%q, want=%q", got, want)
			}

			// Re-encode the card and make sure it matches.
			if got, want := card.Encoded(), tc.encoded; got != want {
				t.Errorf("re-encoding does not match got=%q want=%q", got, want)
			}
		})
	}
}

func TestNewCard(t *testing.T) {
	testCases := []struct {
		name    string
		num     string
		wantErr bool
	}{
		{
			name:    "valid",
			num:     "7",
			wantErr: false,
		},
		{
			name:    "invalid num",
			num:     "-3",
			wantErr: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c, err := NewCard(tc.num, Diamonds)
			if tc.wantErr && err == nil {
				t.Fatal("missing expected error")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tc.wantErr {
				return
			}
			if got, want := c.Num(), tc.num; got != want {
				t.Errorf("incorrect num got=%q, want=%q", got, want)
			}
			if got, want := c.Suit(), Diamonds; !got.IsSameAs(want) {
				t.Errorf("incorrect suit got=%q, want=%q", got, want)
			}
		})
	}
}

func TestCardIsSameAs(t *testing.T) {
	card1 := buildCard(t, "5H")
	card2 := buildCard(t, "3S")
	card3 := buildCard(t, "5H") // Create another so we aren't using the same struct.

	if card1.IsSameAs(card2) != false {
		t.Errorf("cards %s and %s must be different", card1.Encoded(), card2.Encoded())
	}
	if card1.IsSameAs(card3) != true {
		t.Errorf("cards %s and %s must be same", card1.Encoded(), card3.Encoded())
	}
}

func TestCardNum(t *testing.T) {
	for _, num := range []string{"8", "9", "T", "J", "Q", "K", "A"} {
		card := buildCard(t, num+"D")

		if got, want := card.Num(), num; got != want {
			t.Errorf("card.Num()=%q want=%q", got, want)
		}
	}
}

func TestCardSuit(t *testing.T) {
	for _, suit := range []string{"C", "D", "H", "S"} {
		card := buildCard(t, "8"+suit)

		if got, want := card.Suit(), buildSuit(t, suit); !got.IsSameAs(want) {
			t.Errorf("card.Suit()=%q want=%q", got, want)
		}
	}
}

func buildCard(t *testing.T, encodedCard string) Card {
	t.Helper()
	c, err := NewCardFromEncoded(encodedCard)
	if err != nil {
		t.Fatal(err)
	}
	return c
}
