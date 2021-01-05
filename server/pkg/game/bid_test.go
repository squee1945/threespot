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
		wantPos int
		wantVal string
		wantErr bool
	}{
		{
			name:    "empty string",
			wantErr: true,
		},
		{
			name:    "bad pos",
			encoded: "?|7",
			wantErr: true,
		},
		{
			name:    "pos too high",
			encoded: "4|7",
			wantErr: true,
		},
		{
			name:    "bad value",
			encoded: "0|?",
			wantErr: true,
		},
		{
			name:    "too short",
			encoded: "0",
			wantErr: true,
		},
		{
			name:    "too long",
			encoded: "0|7|8",
			wantErr: true,
		},
		{
			name:    "pass bid",
			encoded: "0|P",
			wantPos: 0,
			wantVal: "P",
		},
		{
			name:    "regular bid",
			encoded: "1|7",
			wantPos: 1,
			wantVal: "7",
		},
		{
			name:    "no trump bid",
			encoded: "3|7N",
			wantPos: 3,
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

			if got, want := bid.Pos(), tc.wantPos; got != want {
				t.Errorf("bid.Pos()=%d, want=%d", got, want)
			}
			if got, want := bid.Value(), tc.wantVal; got != want {
				t.Errorf("bid.Pos()=%s, want=%s", got, want)
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
			bid1 := buildBid(t, 0, tc.b1)
			bid2 := buildBid(t, 1, tc.b2)
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
			got := nextBidValues([]Bid{buildBid(t, 0, tc.highBid)}, tc.isDealer)

			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("nextBidValues() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func buildBid(t *testing.T, pos int, val string) Bid {
	t.Helper()
	bid, err := NewBidFromEncoded(fmt.Sprintf("%d|%s", pos, val))
	if err != nil {
		t.Fatal(err)
	}
	return bid
}
