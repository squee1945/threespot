package deck

import "testing"

func TestNewSuitFromEncoded(t *testing.T) {
	testCases := []struct {
		name    string
		encoded string
		want    string
		wantErr bool
	}{
		{
			name:    "empty string",
			wantErr: true,
		},
		{
			name:    "invalid suit",
			encoded: "?",
			wantErr: true,
		},
		{
			name:    "too long",
			encoded: "HH",
			wantErr: true,
		},
		{
			name:    "valid diamond",
			encoded: "D",
			want:    "D",
		},
		{
			name:    "valid heart",
			encoded: "H",
			want:    "H",
		},
		{
			name:    "valid spades",
			encoded: "S",
			want:    "S",
		},
		{
			name:    "valid clubs",
			encoded: "C",
			want:    "C",
		},
		{
			name:    "valid no trump",
			encoded: "N",
			want:    "N",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			suit, err := NewSuitFromEncoded(tc.encoded)

			if tc.wantErr && err == nil {
				t.Fatal("missing expected error")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tc.wantErr {
				return
			}

			if got, want := suit, buildSuit(t, tc.want); !got.IsSameAs(want) {
				t.Errorf("incorrect suit got=%v want=%v", got, want)
			}

			// Re-encode the suit and make sure it matches.
			if got, want := suit.Encoded(), tc.encoded; got != want {
				t.Errorf("re-encoding does not match got=%q want=%q", got, want)
			}
		})
	}
}

func TestSuitIsSameAs(t *testing.T) {
	suit1 := buildSuit(t, "H")
	suit2 := buildSuit(t, "S")
	suit3 := buildSuit(t, "H") // Create another so we aren't using the same struct.

	if suit1.IsSameAs(suit2) != false {
		t.Errorf("suits %s and %s must be different", suit1.Encoded(), suit2.Encoded())
	}
	if suit1.IsSameAs(suit3) != true {
		t.Errorf("suits %s and %s must be same", suit1.Encoded(), suit3.Encoded())
	}
}

func buildSuit(t *testing.T, encodedSuit string) Suit {
	t.Helper()
	suit, err := NewSuitFromEncoded(encodedSuit)
	if err != nil {
		t.Fatal(err)
	}
	return suit
}
