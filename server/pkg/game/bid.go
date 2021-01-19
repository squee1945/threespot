package game

import (
	"fmt"
	"strings"
)

// Bid is a bid by a player.
type Bid interface {
	// Human is a human-readable version of this bid.
	Human() string
	// Encoded is the encoded form of this bid.
	Encoded() string
	// IsLessThan returns true if the other bid is smaller than this one.
	IsLessThan(other Bid) bool
	// IsEqualTo returns true if the other bid is the same value as this one.
	IsEqualTo(other Bid) bool
	// IsNoTrump returns true if the bid is no trump.
	IsNoTrump() bool
	// IsPass returns true if the bid is a pass.
	IsPass() bool
	// Value() returns the numeric value of the bid (e.g., "7" and "7N" return 7).
	Value() (int, error)
}

var (
	pass             = "P"
	humanFromEncoded = map[string]string{
		pass: "Pass",
		"7":  "7",
		"7N": "7 No Trump",
		"8":  "8",
		"8N": "8 No Trump",
		"9":  "9",
		"9N": "9 No Trump",
		"A":  "10",
		"AN": "10 No Trump",
		"B":  "11",
		"BN": "11 No Trump",
		"C":  "12",
		"CN": "12 No Trump",
		// TODO: Kaiser bid
	}
	bidFromEncoded = map[string]Bid{
		pass: &bid{value: pass},
		"7":  &bid{value: "7"},
		"7N": &bid{value: "7N"},
		"8":  &bid{value: "8"},
		"8N": &bid{value: "8N"},
		"9":  &bid{value: "9"},
		"9N": &bid{value: "9N"},
		"A":  &bid{value: "A"},
		"AN": &bid{value: "AN"},
		"B":  &bid{value: "B"},
		"BN": &bid{value: "BN"},
		"C":  &bid{value: "C"},
		"CN": &bid{value: "CN"},
		// TODO: Kaiser bid
	}
	bidValueFromEncoded = map[string]int{
		"7":  7,
		"7N": 7,
		"8":  8,
		"8N": 8,
		"9":  9,
		"9N": 9,
		"A":  10,
		"AN": 10,
		"B":  11,
		"BN": 11,
		"C":  12,
		"CN": 12,
		// TODO: Kaiser bid
	}
	orderedBids = []string{"P", "7", "7N", "8", "8N", "9", "9N", "A", "AN", "B", "BN", "C", "CN"}
	humanValues = map[string]string{}
)

type bid struct {
	value string
}

var _ Bid = (*bid)(nil) // Ensure interface is implemented.

// NewBidFromEncoded builds a bid from the Encoded() form.
func NewBidFromEncoded(encoded string) (Bid, error) {
	b, present := bidFromEncoded[encoded]
	if !present {
		return nil, fmt.Errorf("unknown bid %q", encoded)
	}
	return b, nil
}

func (b *bid) Encoded() string {
	return b.value
}

func (b *bid) Human() string {
	return humanFromEncoded[b.value]
}

func (b *bid) IsNoTrump() bool {
	return strings.HasSuffix(b.value, "N")
}

func (b *bid) IsPass() bool {
	return b.Encoded() == pass
}

func (b *bid) IsLessThan(other Bid) bool {
	return bidValue(b.Encoded()) < bidValue(other.Encoded())
}

func (b *bid) IsEqualTo(other Bid) bool {
	return bidValue(b.Encoded()) == bidValue(other.Encoded())
}

func (b *bid) Value() (int, error) {
	v, ok := bidValueFromEncoded[b.Encoded()]
	if !ok {
		return 0, fmt.Errorf("unknown bid value for %q", b.Encoded())
	}
	return v, nil
}

func nextBidValues(bids []Bid, isDealer bool) []Bid {
	if len(bids) == 0 {
		return valuesToBids(orderedBids)
	}

	highBid, highIndex := highestBid(bids)

	var encodeds []string
	if !isDealer {
		encodeds = []string{pass}
		encodeds = append(encodeds, orderedBids[highIndex+1:]...)
	} else {
		if highBid.IsPass() {
			encodeds = orderedBids[1:]
		} else {
			encodeds = []string{pass}
			encodeds = append(encodeds, orderedBids[highIndex:]...)
		}
	}
	return valuesToBids(encodeds)
}

func valuesToBids(encodeds []string) []Bid {
	var available []Bid
	for _, encoded := range encodeds {
		available = append(available, bidFromEncoded[encoded])
	}
	return available
}

func highestBid(bids []Bid) (Bid, int) {
	highIndex := 0
	highBid := bids[0]

	for _, bid := range bids {
		index := bidValue(bid.Encoded())
		if index > highIndex {
			highIndex = index
			highBid = bid
		}
	}
	return highBid, highIndex
}

func bidValue(encoded string) int {
	for i, b := range orderedBids {
		if b == encoded {
			return i
		}
	}
	return -1
}
