package game

import (
	"fmt"
	"sort"
	"strings"
)

var (
	pass      = "P"
	passBid   = &bid{encoded: pass}
	validBids = map[string]string{
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

type Bid interface {
	String() string
	Encoded() string
	IsGreaterThan(other Bid) bool
	IsGreaterThanOrEqualTo(other Bid) bool
	IsEqualTo(other Bid) bool
	IsLessThan(other Bid) bool
}

type bid struct {
	encoded string
}

func (b *bid) Encoded() string {
	return b.encoded
}

func (b *bid) String() string {
	return validBids[b.encoded]
}

func (b *bid) IsGreaterThan(other Bid) bool {
	return bidValue[b.encoded] > bidValue[other.Encoded()]
}

func (b *bid) IsGreaterThanOrEqualTo(other Bid) bool {
	return bidValue[b.encoded] >= bidValue[other.Encoded()]
}

func (b *bid) IsEqualTo(other Bid) bool {
	return bidValue[b.encoded] == bidValue[other.Encoded()]
}

func (b *bid) IsLessThan(other Bid) bool {
	return bidValue[b.encoded] < bidValue[other.Encoded()]
}

func NewBidFromEncoded(encoded string) (Bid, error) {
	encoded = strings.ToUpper(encoded)
	if _, present := validBids[encoded]; !present {
		return Bid{}, fmt.Errorf("unknown bid %q", encoded)
	}
	return &bid{encoded: encoded}, nil
}

func nextBids(currentBids []Bid, isDealer bool) []Bid {
	highBidValue := 1 // skips "pass", will do it below
	for _, b := range currentBids {
		bv := bidValue[b.Encoded()]
		if bv > highBidValue {
			highBidValue = bv
		}
	}

	var available []Bid
	for k, v := range bidValues {
		if (isDealer && v >= highBidValue) || (!isDealer && v > highBidValue) {
			available = append(available, &bid{encoded: k})
		}
	}

	// if not dealer, can always pass
	// if dealer, can only pass if there is another bid
	if !isDealer || (isDealer && currentHighBid != nil) {
		available = append(available, passBid)
	}
	sort.SliceStable(available, func(i, j Bid) bool { return i.IsLessThan(j) })
	return available
}
