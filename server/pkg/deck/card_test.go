package deck

import "testing"

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
				t.Fatal("wanted error, got err=nil")
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
			if got, want := c.Suit(), Diamonds; got != want {
				t.Errorf("incorrect suit got=%q, want=%q", got, want)
			}
		})
	}
}

func TestIsSameAs(t *testing.T) {
	card1, err := NewCardFromEncoded("5H")
	if err != nil {
		t.Fatal(err)
	}
	card2, err := NewCardFromEncoded("3S")
	if err != nil {
		t.Fatal(err)
	}
	card3, err := NewCardFromEncoded("5H") // Create another so we aren't using the same struct.
	if err != nil {
		t.Fatal(err)
	}

	if card1.IsSameAs(card2) != false {
		t.Errorf("cards %s and %s must be different", card1.Encoded(), card2.Encoded())
	}
	if card1.IsSameAs(card3) != true {
		t.Errorf("cards %s and %s must be different", card1.Encoded(), card3.Encoded())
	}
}
