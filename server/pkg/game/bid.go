package game

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// Bid is a bid by a player.
type Bid interface {
	// Human is a human-readable version of this bid.
	Human() string
	// Encoded is the encoded form of this bid.
	Encoded() string
	// Pos is the player position of this bid.
	Pos() int
	// Valid is the value of this bid.
	Value() string
	// IsLessThan returns true if the other bid is smaller than this one.
	IsLessThan(other Bid) bool
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
	bidValue = map[string]int{
		pass: 0,
		"7":  70,
		"7N": 75,
		"8":  80,
		"8N": 85,
		"9":  90,
		"9N": 95,
		"A":  100,
		"AN": 105,
		"B":  110,
		"BN": 115,
		"C":  120,
		"CN": 125,
		// TODO: Kaiser bid
	}
	humanValues = map[string]string{}
)

type bid struct {
	value string
	pos   int
}

var _ Bid = (*bid)(nil) // Ensure interface is implemented.

// NewBidFromEncoded builds a bid from the Encoded() form.
func NewBidFromEncoded(encoded string) (Bid, error) {
	parts := strings.Split(strings.ToUpper(encoded), "|")
	if len(parts) != 2 {
		return nil, fmt.Errorf("encoding %q did not contain two parts", encoded)
	}
	pos, err := strconv.Atoi(parts[0])
	if err != nil {
		return nil, fmt.Errorf("encoding part[0] %q was not an int: %v", parts[0], err)
	}
	if pos < 0 || pos > 3 {
		return nil, fmt.Errorf("pos %d must be on interval [0,3]", pos)
	}
	value := parts[1]
	if _, present := humanFromEncoded[value]; !present {
		return nil, fmt.Errorf("unknown bid %q", value)
	}
	return &bid{pos: pos, value: value}, nil
}

func (b *bid) Encoded() string {
	return fmt.Sprintf("%d|%s", b.pos, b.value)
}

func (b *bid) Human() string {
	return humanFromEncoded[b.value]
}

func (b *bid) Pos() int {
	return b.pos
}

func (b *bid) Value() string {
	return b.value
}

func (b *bid) IsLessThan(other Bid) bool {
	return bidValue[b.Value()] < bidValue[other.Value()]
}

func nextBidValues(currentBids []Bid, isDealer bool) []string {
	highBidValue := 1 // skips "pass", will do it below
	for _, b := range currentBids {
		bv := bidValue[b.Value()]
		if bv > highBidValue {
			highBidValue = bv
		}
	}

	var available []string
	for k, v := range bidValue {
		if (isDealer && v >= highBidValue) || (!isDealer && v > highBidValue) {
			available = append(available, k)
		}
	}

	// if not dealer, can always pass
	// if dealer, can only pass if there is another bid
	if !isDealer || (isDealer && highBidValue > 1) {
		available = append(available, pass)
	}
	sort.SliceStable(available, func(i, j int) bool { return bidValue[available[i]] < bidValue[available[j]] })
	return available
}
