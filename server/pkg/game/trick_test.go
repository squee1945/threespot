package game

import (
	"testing"

	"github.com/squee1945/threespot/server/pkg/deck"
)

func TestWinningPos(t *testing.T) {
	testCases := []struct {
		trump   string
		cards   []string
		wantPos int
	}{
		{"", []string{"8H", "9H", "3S", "7C"}, 1},
		{"", []string{"8H", "9S", "3S", "7C"}, 0},
		{"", []string{"8H", "5H", "AH", "KH"}, 2},
		{"D", []string{"8H", "9H", "3S", "7C"}, 1},
		{"D", []string{"8H", "9S", "3S", "7C"}, 0},
		{"D", []string{"8H", "5H", "AH", "KH"}, 2},
		{"H", []string{"7H", "3S", "8H", "AH"}, 3},
		{"H", []string{"3S", "AH", "7H", "KH"}, 1},
	}
	for _, tc := range testCases {
		tr, err := NewTrick(0, deck.Suit(tc.trump))
		if err != nil {
			t.Fatal(err)
		}

		for playerPos, c := range tc.cards {
			tr.PlayCard(playerPos, buildCard(t, c))
		}

		gotPos, err := tr.WinningPos()
		if err != nil {
			t.Fatal(err)
		}

		if gotPos != tc.wantPos {
			t.Errorf("trump=%q, cards=%v, got=%d, want=%d", tc.trump, tc.cards, gotPos, tc.wantPos)
		}
	}
}

func TestToOrd(t *testing.T) {
	testCases := []struct {
		leadPos   int
		playerPos int
		wantOrd   int
	}{
		{0, 0, 0},
		{0, 1, 1},
		{0, 2, 2},
		{0, 3, 3},
		{1, 0, 3},
		{1, 1, 0},
		{1, 2, 1},
		{1, 3, 2},
		{2, 0, 2},
		{2, 1, 3},
		{2, 2, 0},
		{2, 3, 1},
		{3, 0, 1},
		{3, 1, 2},
		{3, 2, 3},
		{3, 3, 0},
	}
	for _, tc := range testCases {
		trickT, err := NewTrick(tc.leadPos, deck.NoTrump)
		if err != nil {
			t.Fatal(err)
		}
		tr := trickT.(*trick)

		if got, want := tr.toOrd(tc.playerPos), tc.wantOrd; got != want {
			t.Errorf("toOrd(leadPos=%d, playerPos=%d)=%d, want=%d", tc.leadPos, tc.playerPos, got, want)
		}
	}
}

func TestToPos(t *testing.T) {
	testCases := []struct {
		leadPos   int
		playerOrd int
		wantPos   int
	}{
		{0, 0, 0},
		{0, 1, 1},
		{0, 2, 2},
		{0, 3, 3},
		{1, 0, 1},
		{1, 1, 2},
		{1, 2, 3},
		{1, 3, 0},
		{2, 0, 2},
		{2, 1, 3},
		{2, 2, 0},
		{2, 3, 1},
		{3, 0, 3},
		{3, 1, 0},
		{3, 2, 1},
		{3, 3, 2},
	}
	for _, tc := range testCases {
		trickT, err := NewTrick(tc.leadPos, deck.NoTrump)
		if err != nil {
			t.Fatal(err)
		}
		tr := trickT.(*trick)

		if got, want := tr.toPos(tc.playerOrd), tc.wantPos; got != want {
			t.Errorf("toPos(leadPos=%d, playerOrd=%d)=%d, want=%d", tc.leadPos, tc.playerOrd, got, want)
		}
	}
}

func TestIsHigher(t *testing.T) {
	testCases := []struct {
		name  string
		a, b  string // two-character encoded cards
		lead  string
		trump string
		want  bool
	}{
		{
			name:  "no trump, follow suit higher",
			lead:  "H",
			trump: "",
			a:     "7H",
			b:     "8H",
			want:  true,
		},
		{
			name:  "no trump, follow suit lower",
			lead:  "H",
			trump: "",
			a:     "8H",
			b:     "7H",
			want:  false,
		},
		{
			name:  "no trump, suited beats non-suited",
			lead:  "H",
			trump: "",
			a:     "8S",
			b:     "7H",
			want:  true,
		},
		{
			name:  "no trump, non-suited does not beats suited",
			lead:  "H",
			trump: "",
			a:     "7H",
			b:     "8S",
			want:  false,
		},
		{
			name:  "trump beats non-trump",
			lead:  "H",
			trump: "S",
			a:     "9H",
			b:     "8S",
			want:  true,
		},
		{
			name:  "suited does not beat trump",
			lead:  "H",
			trump: "S",
			a:     "8S",
			b:     "9H",
			want:  false,
		},
		{
			name:  "both trumps, a higher",
			lead:  "H",
			trump: "S",
			a:     "9S",
			b:     "8S",
			want:  false,
		},
		{
			name:  "both trumps, b higher",
			lead:  "H",
			trump: "S",
			a:     "QS",
			b:     "KS",
			want:  true,
		},
		{
			name:  "neither trump, a higher",
			lead:  "H",
			trump: "S",
			a:     "9H",
			b:     "5H",
			want:  false,
		},
		{
			name:  "neither trump, b higher",
			lead:  "H",
			trump: "S",
			a:     "5H",
			b:     "9H",
			want:  true,
		},
		{
			name:  "suited beats non-suited",
			lead:  "H",
			trump: "S",
			a:     "AC",
			b:     "5H",
			want:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			trump := deck.Suit(tc.trump)
			trickT, err := NewTrick(0, trump)
			if err != nil {
				t.Fatal(err)
			}
			tr := trickT.(*trick)

			cardA := buildCard(t, tc.a)
			cardB := buildCard(t, tc.b)
			if got, want := tr.isHigher(deck.Suit(tc.lead), cardA, cardB), tc.want; got != want {
				t.Errorf("isHigher(lead=%q, trump=%q, %q, %q)=%t, want=%t", deck.Suit(tc.lead), trump, tc.a, tc.b, got, want)
			}
		})

	}
}

func TestIsHigherNum(t *testing.T) {
	ordered := []string{"3", "5", "7", "8", "9", "T", "J", "Q", "K", "A"}
	for i, a := range ordered {
		// Everything before, and same as, a must return false.
		for _, b := range ordered[:i+1] {
			if got, want := isNumHigher(a, b), false; got != want {
				t.Errorf("isNumHigher(%q, %q)=%t, want=%t", a, b, got, want)
			}
		}
		// Everything after a must return true.
		for _, b := range ordered[i+1:] {
			if got, want := isNumHigher(a, b), true; got != want {
				t.Errorf("isNumHigher(%q, %q)=%t, want=%t", a, b, got, want)
			}
		}
	}
}

func buildCard(t *testing.T, s string) deck.Card {
	t.Helper()
	if len(s) != 2 {
		t.Fatalf("expected two-character card, got %q", s)
	}
	num := s[0:1]
	suit := deck.Suit(s[1:2])
	c, err := deck.NewCard(num, suit)
	if err != nil {
		t.Fatal(err)
	}
	return c
}
