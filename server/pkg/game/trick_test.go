package game

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/squee1945/threespot/server/pkg/deck"
)

var (
	compareCards = cmp.Comparer(func(c1, c2 deck.Card) bool { return c1.Encoded() == c2.Encoded() })
)

func TestNewTrick(t *testing.T) {
	trick, err := NewTrick(deck.Hearts, 1)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := trick.Trump(), deck.Hearts; !got.IsSameAs(want) {
		t.Errorf("incorrect trump got=%s want=%q", got, want)
	}
	if got, want := trick.LeadPos(), 1; got != want {
		t.Errorf("incorect lead pos got=%d want=%d", got, want)
	}

	_, err = NewTrick(deck.Spades, -1)
	if err == nil {
		t.Errorf("missing expected error for -1")
	}
	_, err = NewTrick(deck.Spades, 4)
	if err == nil {
		t.Errorf("missing expected error for 4")
	}
}

func TestNewTrickFromEncoded(t *testing.T) {
	testCases := []struct {
		name        string
		encoded     string
		wantTrump   string
		wantLeadPos int
		wantCards   []string
		wantErr     bool
	}{
		{
			name:    "empty string",
			wantErr: true,
		},
		{
			name:    "missing minimum elements",
			encoded: "1-",
			wantErr: true,
		},
		{
			name:    "bad trump",
			encoded: "1-?",
			wantErr: true,
		},
		{
			name:    "bad card",
			encoded: "0-N-??",
			wantErr: true,
		},
		{
			name:    "too many cards",
			encoded: "0-N-7D-8D-9D-TD-JD",
			wantErr: true,
		},
		{
			name:    "duplicate cards",
			encoded: "0-N-7D-7D",
			wantErr: true,
		},
		{
			name:        "valid with no cards",
			encoded:     "3-D",
			wantTrump:   "D",
			wantLeadPos: 3,
			wantCards:   []string{},
		},
		{
			name:        "valid with one card",
			encoded:     "2-H-5H",
			wantTrump:   "H",
			wantLeadPos: 2,
			wantCards:   []string{"5H"},
		},
		{
			name:        "valid with max cards",
			encoded:     "1-C-5H-8D-9C-TS",
			wantTrump:   "C",
			wantLeadPos: 1,
			wantCards:   []string{"5H", "8D", "9C", "TS"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			trick, err := NewTrickFromEncoded(tc.encoded)

			if tc.wantErr && err == nil {
				t.Fatal("missing expected error")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tc.wantErr {
				return
			}

			if got, want := trick.Trump(), buildSuit(t, tc.wantTrump); !got.IsSameAs(want) {
				t.Errorf("incorrect trump got=%s want=%q", got, want)
			}
			if got, want := trick.LeadPos(), tc.wantLeadPos; got != want {
				t.Errorf("incorect lead pos got=%d want=%d", got, want)
			}
			got := trick.Cards()
			var want []deck.Card
			for _, encoded := range tc.wantCards {
				want = append(want, buildCard(t, encoded))
			}
			if diff := cmp.Diff(want, got, compareCards); diff != "" {
				t.Errorf("trick.Cards() mismatch (-want +got):\n%s", diff)
			}

			// Re-encode the trick and make sure it matches.
			if got, want := trick.Encoded(), tc.encoded; got != want {
				t.Errorf("re-encoding does not match got=%q want=%q", got, want)
			}
		})
	}
}

func TestIsDone(t *testing.T) {
	testCases := []struct {
		cards []string
		want  bool
	}{
		{
			cards: nil,
		},
		{
			cards: []string{"7H"},
		},
		{
			cards: []string{"7H", "8H"},
		},
		{
			cards: []string{"7H", "8H", "9H"},
		},
		{
			cards: []string{"7H", "8H", "9H", "TH"},
			want:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%d cards", len(tc.cards)), func(t *testing.T) {
			trick := buildTrick(t, "N", 0, tc.cards...)

			if got, want := trick.IsDone(), tc.want; got != want {
				t.Errorf("trick.IsDone()=%t want=%t", got, want)
			}
		})
	}
}

func TestPlayCard(t *testing.T) {
	toPlay := "5H"
	testCases := []struct {
		name    string
		have    []string
		wantErr bool
	}{
		{
			name: "play first card",
		},
		{
			name: "play last card",
			have: []string{"8H", "9H", "TH"},
		},
		{
			name:    "card played must be unique",
			have:    []string{"8H", toPlay, "TH"},
			wantErr: true,
		},
		{
			name:    "cannot play more than 4 cards",
			have:    []string{"8H", "9H", "TH", "JH"},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			trick := buildTrick(t, "N", 0, tc.have...)

			err := trick.PlayCard(len(tc.have), buildCard(t, toPlay))

			if tc.wantErr && err == nil {
				t.Fatal("missing expected error")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tc.wantErr {
				return
			}

			if got, want := trick.NumPlayed(), len(tc.have)+1; got != want {
				t.Errorf("trick.NumPlayed()=%d, want=%d", got, want)
			}
			if got, want := trick.Cards()[len(tc.have)].Encoded(), toPlay; got != want {
				t.Errorf("last card played is incorrect got=%q, want=%q", got, want)
			}
		})
	}
}

func TestCurrentTurnPosition(t *testing.T) {
	testCases := []struct {
		leadPos   int
		playCards int
		wantPos   int
		wantErr   bool
	}{
		{0, 0, 0, false},
		{0, 1, 1, false},
		{0, 2, 2, false},
		{0, 3, 3, false},
		{0, 4, 0, true},
		{1, 0, 1, false},
		{1, 1, 2, false},
		{1, 2, 3, false},
		{1, 3, 0, false},
		{1, 4, 0, true},
		{2, 0, 2, false},
		{2, 1, 3, false},
		{2, 2, 0, false},
		{2, 3, 1, false},
		{2, 4, 0, true},
		{3, 0, 3, false},
		{3, 1, 0, false},
		{3, 2, 1, false},
		{3, 3, 2, false},
		{3, 4, 0, true},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("case %d", i), func(t *testing.T) {
			potentials := []string{"7H", "8H", "9H", "TH"}
			trick := buildTrick(t, "N", tc.leadPos, potentials[0:tc.playCards]...)

			got, err := trick.CurrentTurnPos()

			if tc.wantErr && err == nil {
				t.Fatal("missing expected error")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tc.wantErr {
				return
			}

			if got != tc.wantPos {
				t.Errorf("trick.CurrentTurnPos()=%d, want=%d", got, tc.wantPos)
			}
		})
	}
}

func TestWinningPos(t *testing.T) {
	testCases := []struct {
		trump   string
		cards   []string
		wantPos int
	}{
		{"N", []string{"8H", "9H", "3S", "7C"}, 1},
		{"N", []string{"8H", "9S", "3S", "7C"}, 0},
		{"N", []string{"8H", "5H", "AH", "KH"}, 2},
		{"D", []string{"8H", "9H", "3S", "7C"}, 1},
		{"D", []string{"8H", "9S", "3S", "7C"}, 0},
		{"D", []string{"8H", "5H", "AH", "KH"}, 2},
		{"H", []string{"7H", "3S", "8H", "AH"}, 3},
		{"H", []string{"3S", "AH", "7H", "KH"}, 1},
	}
	for _, tc := range testCases {
		trick := buildTrick(t, tc.trump, 0, tc.cards...)

		gotPos, err := trick.WinningPos()
		if err != nil {
			t.Fatal(err)
		}

		if gotPos != tc.wantPos {
			t.Errorf("trump=%q, cards=%v, got=%d, want=%d", tc.trump, tc.cards, gotPos, tc.wantPos)
		}
	}
}

func TestTrump(t *testing.T) {
	for _, want := range []string{"H", "S", "D", "C", "N"} {
		trick := buildTrick(t, want, 0)
		if got, want := trick.Trump().Encoded(), want; got != want {
			t.Errorf("wrong trump got=%q want=%q", got, want)
		}
	}
}

func TestLeadPos(t *testing.T) {
	for _, want := range []int{0, 1, 2, 3} {
		trick := buildTrick(t, "N", want)
		if got, want := trick.LeadPos(), want; got != want {
			t.Errorf("wrong lead pos got=%d want=%d", got, want)
		}
	}
}

func TestLeadSuit(t *testing.T) {
	testCases := []struct {
		name    string
		played  []string
		want    string
		wantErr bool
	}{
		{
			name:   "one card played",
			played: []string{"5H"},
			want:   "H",
		},
		{
			name:   "multiple cards played",
			played: []string{"5H", "8C", "9D"},
			want:   "H",
		},
		{
			name:    "no cards played",
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			trick := buildTrick(t, "N", 0, tc.played...)

			got, err := trick.LeadSuit()

			if tc.wantErr && err == nil {
				t.Fatal("missing expected error")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tc.wantErr {
				return
			}

			if got, want := got.Encoded(), tc.want; got != want {
				t.Errorf("trick.LeadSuit()=%q want=%q", got, want)
			}
		})
	}
}

func TestNumPlayed(t *testing.T) {
	for i := 0; i < 4; i++ {
		t.Run(fmt.Sprintf("cards played %d", i), func(t *testing.T) {
			potentials := []string{"7H", "8H", "9H", "TH"}
			trick := buildTrick(t, "N", 0, potentials[0:i]...)

			if got, want := trick.NumPlayed(), i; got != want {
				t.Errorf("trick.NumPlayed()=%d, want=%d", got, want)
			}
		})
	}
}

func TestCards(t *testing.T) {
	for i := 0; i < 4; i++ {
		t.Run(fmt.Sprintf("cards played %d", i), func(t *testing.T) {
			potentials := []string{"7H", "8H", "9H", "TH"}
			wantEncoded := potentials[0:i]
			trick := buildTrick(t, "N", 0, wantEncoded...)

			if got, want := trick.NumPlayed(), i; got != want {
				t.Errorf("trick.NumPlayed()=%d, want=%d", got, want)
			}

			got := trick.Cards()
			var want []deck.Card
			for _, encoded := range wantEncoded {
				want = append(want, buildCard(t, encoded))
			}
			if diff := cmp.Diff(want, got, compareCards); diff != "" {
				t.Errorf("trick.Cards() mismatch (-want +got):\n%s", diff)
			}
		})
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
		trickT := buildTrick(t, "N", tc.leadPos)
		trick := trickT.(*trick)

		if got, want := trick.toOrd(tc.playerPos), tc.wantOrd; got != want {
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
		trickT := buildTrick(t, "N", tc.leadPos)
		trick := trickT.(*trick)

		if got, want := trick.toPos(tc.playerOrd), tc.wantPos; got != want {
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
			trump: "N",
			a:     "7H",
			b:     "8H",
			want:  true,
		},
		{
			name:  "no trump, follow suit lower",
			lead:  "H",
			trump: "N",
			a:     "8H",
			b:     "7H",
			want:  false,
		},
		{
			name:  "no trump, suited beats non-suited",
			lead:  "H",
			trump: "N",
			a:     "8S",
			b:     "7H",
			want:  true,
		},
		{
			name:  "no trump, non-suited does not beats suited",
			lead:  "H",
			trump: "N",
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
			trickT := buildTrick(t, tc.trump, 0)
			trick := trickT.(*trick)
			leadSuit := buildSuit(t, tc.lead)

			cardA, cardB := buildCard(t, tc.a), buildCard(t, tc.b)
			if got, want := trick.isHigher(leadSuit, cardA, cardB), tc.want; got != want {
				t.Errorf("isHigher(lead=%q, trump=%q, %q, %q)=%t, want=%t", leadSuit, tc.trump, tc.a, tc.b, got, want)
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

func buildCard(t *testing.T, encodedCard string) deck.Card {
	t.Helper()
	c, err := deck.NewCardFromEncoded(encodedCard)
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func buildSuit(t *testing.T, encodedSuit string) deck.Suit {
	t.Helper()
	suit, err := deck.NewSuitFromEncoded(encodedSuit)
	if err != nil {
		t.Fatal(err)
	}
	return suit
}

func buildTrick(t *testing.T, encodedTrump string, leadPos int, encodedCards ...string) Trick {
	t.Helper()
	trump := buildSuit(t, encodedTrump)
	trick, err := NewTrick(trump, leadPos)
	if err != nil {
		t.Fatal(err)
	}
	for pos, c := range encodedCards {
		if err := trick.PlayCard(leadPos+pos, buildCard(t, c)); err != nil {
			t.Fatal(err)
		}
	}
	return trick
}
