package game

import (
	"fmt"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestNewBidFromEncoded(t *testing.T) {
	testCases := []struct {
		name    string
		encoded string
		wantVal string
		wantErr bool
	}{
		{
			name:    "empty string",
			wantErr: true,
		},
		{
			name:    "bad value",
			encoded: "?",
			wantErr: true,
		},
		{
			name:    "too long",
			encoded: "78",
			wantErr: true,
		},
		{
			name:    "pass bid",
			encoded: "P",
			wantVal: "P",
		},
		{
			name:    "regular bid",
			encoded: "7",
			wantVal: "7",
		},
		{
			name:    "no trump bid",
			encoded: "7N",
			wantVal: "7N",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			bid, err := NewBidFromEncoded(tc.encoded)

			if tc.wantErr && err == nil {
				t.Fatal("missing expected error")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tc.wantErr {
				return
			}

			// Re-encode the bid and make sure it matches.
			if got, want := bid.Encoded(), strings.ToUpper(tc.encoded); got != want {
				t.Errorf("re-encoding does not match got=%q want=%q", got, want)
			}
		})
	}
}

func TestBidIsLessThan(t *testing.T) {
	testCases := []struct {
		b1, b2 string
		want   bool
	}{
		{
			b1:   "7",
			b2:   "7",
			want: false,
		},
		{
			b1:   "7",
			b2:   "8",
			want: true,
		},
		{
			b1:   "9",
			b2:   "A", // "A" is a 10-bid.
			want: true,
		},
		{
			b1:   "A",
			b2:   "A",
			want: false,
		},
		{
			b1:   "A",
			b2:   "B", // "B" is an 11-bid.
			want: true,
		},
		{
			b1:   "B",
			b2:   "C", // "C" is a 12-bid.
			want: true,
		},
		{
			b1:   "7",
			b2:   "7N", // "7N" is 7 No Trump.
			want: true,
		},
		{
			b1:   "C",
			b2:   "CN", // "CN" is 12 No Trump.
			want: true,
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%s-%s", tc.b1, tc.b2), func(t *testing.T) {
			bid1 := buildBid(t, tc.b1)
			bid2 := buildBid(t, tc.b2)
			if got, want := bid1.IsLessThan(bid2), tc.want; got != want {
				t.Errorf("bid.IsLessThan()=%t want=%t", got, want)
			}
		})
	}
}

func TestNextBidValues(t *testing.T) {
	testCases := []struct {
		highBid  string
		isDealer bool
		want     []string
	}{
		{
			highBid: "",
			want:    []string{"P", "7", "7N", "8", "8N", "9", "9N", "A", "AN", "B", "BN", "C", "CN"},
		},
		{
			highBid: "P",
			want:    []string{"P", "7", "7N", "8", "8N", "9", "9N", "A", "AN", "B", "BN", "C", "CN"},
		},
		{
			highBid: "7",
			want:    []string{"P", "7N", "8", "8N", "9", "9N", "A", "AN", "B", "BN", "C", "CN"},
		},
		{
			highBid: "8",
			want:    []string{"P", "8N", "9", "9N", "A", "AN", "B", "BN", "C", "CN"},
		},
		{
			highBid: "C",
			want:    []string{"P", "CN"},
		},
		{
			highBid: "CN",
			want:    []string{"P"},
		},
		{
			highBid:  "P",
			isDealer: true,
			want:     []string{"7", "7N", "8", "8N", "9", "9N", "A", "AN", "B", "BN", "C", "CN"},
		},
		{
			highBid:  "7",
			isDealer: true,
			want:     []string{"P", "7", "7N", "8", "8N", "9", "9N", "A", "AN", "B", "BN", "C", "CN"},
		},
		{
			highBid:  "8",
			isDealer: true,
			want:     []string{"P", "8", "8N", "9", "9N", "A", "AN", "B", "BN", "C", "CN"},
		},
		{
			highBid:  "C",
			isDealer: true,
			want:     []string{"P", "C", "CN"},
		},
		{
			highBid:  "CN",
			isDealer: true,
			want:     []string{"P", "CN"},
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("high=%s isDealer=%t", tc.highBid, tc.isDealer), func(t *testing.T) {
			var priorBids []Bid
			if tc.highBid != "" {
				priorBids = []Bid{buildBid(t, tc.highBid)}
			}
			bids := nextBidValues(priorBids, tc.isDealer)

			var got []string
			for _, b := range bids {
				got = append(got, b.Encoded())
			}

			if diff := cmp.Diff(tc.want, got, cmp.Comparer(func(b1, b2 Bid) bool { return b1.Encoded() == b2.Encoded() })); diff != "" {
				t.Errorf("nextBidValues() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func buildBid(t *testing.T, val string) Bid {
	t.Helper()
	bid, err := NewBidFromEncoded(val)
	if err != nil {
		t.Fatal(err)
	}
	return bid
}
