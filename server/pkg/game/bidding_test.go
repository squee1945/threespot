package game

import (
	"fmt"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestNewBiddingRoundFromEncoded(t *testing.T) {
	testCases := []struct {
		name        string
		encoded     string
		wantLeadPos int
		wantBids    []Bid
		wantErr     bool
	}{
		{
			name:    "empty string",
			encoded: "0|",
		},
		{
			name:    "lead pos not int",
			encoded: "?|7",
			wantErr: true,
		},
		{
			name:    "lead pos too small",
			encoded: "-1|7",
			wantErr: true,
		},
		{
			name:    "lead pos too big",
			encoded: "4|7",
			wantErr: true,
		},
		{
			name:    "invalid bid encoding",
			encoded: "0|?",
			wantErr: true,
		},
		{
			name:        "no bids",
			encoded:     "1|",
			wantLeadPos: 1,
		},
		{
			name:        "one bid",
			encoded:     "0|7N",
			wantLeadPos: 0,
			wantBids:    []Bid{buildBid(t, "7N")},
		},
		{
			name:        "four bids",
			encoded:     "0|7N|8|8N|8N",
			wantLeadPos: 0,
			wantBids:    []Bid{buildBid(t, "7N"), buildBid(t, "8"), buildBid(t, "8N"), buildBid(t, "8N")},
		},
		{
			name:    "too many bids",
			encoded: "0|7N|8|8N|8N|9",
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			br, err := NewBiddingRoundFromEncoded(tc.encoded)

			if tc.wantErr && err == nil {
				t.Fatal("missing expected error")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tc.wantErr {
				return
			}

			if got, want := br.LeadPos(), tc.wantLeadPos; got != want {
				t.Errorf("LeadPos()=%d want=%d", got, want)
			}
			if diff := cmp.Diff(tc.wantBids, br.Bids(), compareBids); diff != "" {
				t.Errorf("bids mismatch (-want +got):\n%s", diff)
			}

			// Re-encode the bidding round and make sure it matches.
			if got, want := br.Encoded(), strings.ToUpper(tc.encoded); got != want {
				t.Errorf("re-encoding does not match got=%q want=%q", got, want)
			}
		})
	}
}

func TestNewBiddingRound(t *testing.T) {
	testCases := []struct {
		name    string
		leadPos int
		wantErr bool
	}{
		{
			name:    "lead pos too small",
			leadPos: -1,
			wantErr: true,
		},
		{
			name:    "lead pos too big",
			leadPos: 4,
			wantErr: true,
		},
		{
			name:    "valid lead pos",
			leadPos: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			br, err := NewBiddingRound(tc.leadPos)

			if tc.wantErr && err == nil {
				t.Fatal("missing expected error")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tc.wantErr {
				return
			}

			if got, want := br.LeadPos(), tc.leadPos; got != want {
				t.Errorf("LeadPos=%d want=%d", got, want)
			}
			if got := br.Bids(); got != nil {
				t.Errorf("Bids()=%v want=nil", got)
			}
		})
	}
}

func TestBiddingRoundIsDoneAndNumPlaced(t *testing.T) {
	testCases := []struct {
		encoded       string
		wantIsDone    bool
		wantNumPlaced int
	}{
		{"0|", false, 0},
		{"0|7", false, 1},
		{"0|7|8", false, 2},
		{"0|7|8|9", false, 3},
		{"0|7|8|9|A", true, 4},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("case %d", i), func(t *testing.T) {
			br := buildBiddingRound(t, tc.encoded)

			if got, want := br.IsDone(), tc.wantIsDone; got != want {
				t.Errorf("IsDone()=%t want=%t", got, want)
			}
			if got, want := br.NumPlaced(), tc.wantNumPlaced; got != want {
				t.Errorf("NumPlaced()=%d want=%d", got, want)
			}
		})
	}
}

func TestBiddingRoundPlaceBid(t *testing.T) {
	testCases := []struct {
		name        string
		encoded     string
		playerPos   int
		bid         Bid
		wantErr     error
		wantEncoded string
	}{
		{
			name:        "first bid",
			encoded:     "1|",
			playerPos:   1,
			bid:         buildBid(t, "7N"),
			wantEncoded: "1|7N",
		},
		{
			name:        "last bid",
			encoded:     "2|7|8|9",
			playerPos:   1,
			bid:         buildBid(t, "9N"),
			wantEncoded: "2|7|8|9|9N",
		},
		{
			name:      "out of order bid",
			encoded:   "2|P|P|P",
			playerPos: 0,
			bid:       buildBid(t, "9N"),
			wantErr:   ErrIncorrectBidOrder,
		},
		{
			name:      "too many bids",
			encoded:   "0|7|8|9|9N",
			playerPos: 0,
			bid:       buildBid(t, "A"),
			wantErr:   ErrIncorrectBidOrder,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			br := buildBiddingRound(t, tc.encoded)

			err := br.placeBid(tc.playerPos, tc.bid)

			if tc.wantErr != nil && err == nil {
				t.Fatal("missing expected error")
			}
			if tc.wantErr == nil && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tc.wantErr != err {
				t.Fatalf("incorrect error got=%v want=%v", err, tc.wantErr)
			}
			if tc.wantErr != nil {
				return
			}

			if got, want := br.Encoded(), tc.wantEncoded; got != want {
				t.Errorf("incorrect encoding got=%q want=%q", got, want)
			}
		})
	}
}

func TestBiddingRoundCurrentTurnPosErrors(t *testing.T) {
	encoded := "0|P|P|P|7"
	br := buildBiddingRound(t, encoded)

	_, err := br.CurrentTurnPos()
	if err == nil {
		t.Errorf("missing expected error")
	}
}

func TestBiddingRoundCurrentTurnPos(t *testing.T) {
	testCases := []struct {
		name    string
		encoded string
		want    int
	}{
		{
			name:    "first bid",
			encoded: "1|",
			want:    1,
		},
		{
			name:    "last bid",
			encoded: "1|P|P|P",
			want:    0,
		},
		{
			name:    "wrap test",
			encoded: "3|P|P",
			want:    1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			br := buildBiddingRound(t, tc.encoded)

			got, err := br.CurrentTurnPos()
			if err != nil {
				t.Fatal(err)
			}

			if got != tc.want {
				t.Errorf("CurrentTurnPos()=%d want=%d", got, tc.want)
			}
		})
	}
}

func TestBiddingRoundWinningBidAndPosErrors(t *testing.T) {
	br := buildBiddingRound(t, "0|P") // Bidding not done.

	_, _, err := br.WinningBidAndPos()
	if err == nil {
		t.Errorf("missing expected error for incomplete bidding")
	}

	br = buildBiddingRound(t, "0|P|P|P|P") // Bidding done.

	_, _, err = br.WinningBidAndPos()
	if err == nil {
		t.Errorf("missing expected error for all passes")
	}
}

func TestBiddingRoundWinningBidAndPos(t *testing.T) {
	testCases := []struct {
		encoded string
		wantBid string
		wantPos int
	}{
		{"0|P|P|P|7", "7", 3},
		{"1|P|P|7|7", "7", 0},
		{"2|7|P|P|P", "7", 2},
		{"3|7|8|9|P", "9", 1},
	}

	for _, tc := range testCases {
		t.Run(tc.encoded, func(t *testing.T) {
			br := buildBiddingRound(t, tc.encoded)

			bid, pos, err := br.WinningBidAndPos()
			if err != nil {
				t.Fatal(err)
			}

			if got, want := bid.Encoded(), tc.wantBid; got != want {
				t.Errorf("incorrect bid got=%q want=%q", got, want)
			}
			if got, want := pos, tc.wantPos; got != want {
				t.Errorf("incorrect pos got=%d want=%d", got, want)
			}
		})
	}
}

func buildBiddingRound(t *testing.T, encoded string) BiddingRound {
	br, err := NewBiddingRoundFromEncoded(encoded)
	if err != nil {
		t.Fatal(err)
	}
	return br
}
