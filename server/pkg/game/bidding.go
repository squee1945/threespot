package game

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// BiddingRound round is a collection of bids.
type BiddingRound interface {
	// IsDone returns true if the bidding is complete (4 bids).
	IsDone() bool

	// placeBid makes a bid for the player in playerPos position. Bid validation must be performed before calling this method.
	placeBid(playerPos int, bid Bid) error

	// CurrentTurnPos returns the position of the player who's turn it is to bid. Returns error if IsDone().
	CurrentTurnPos() (int, error)

	// WinningBidAndPos returns the bid and position of the player that won the bidding. Returns error if !IsDone().
	WinningBidAndPos() (Bid, int, error)

	// LeadPos returns the position of the lead player.
	LeadPos() int

	// Bids returns the cards played; the first card is for the LeadBid, clockwise from there.
	Bids() []Bid

	// NumPlaced returns the number of bids placed.
	NumPlaced() int

	// Encoded returns the bids encoded into a single string.
	Encoded() string
}

type biddingRound struct {
	// leadPos is the position (0..3) of the leadoff bidder.
	leadPos int
	// bids are the bids placed. bids[0] is the bid placed by the player in leadPos.
	bids []Bid
}

var _ BiddingRound = (*biddingRound)(nil) // Ensure interface is implemented.

// NewBiddingRoundFromEncoded returns a set of bigs from the Encoded() form.
func NewBiddingRoundFromEncoded(encoded string) (BiddingRound, error) {
	// "{leadPos}|{bid0}|{bid1}|{bid2}|{bid3}"
	if encoded == "" {
		return nil, errors.New("empty string is invalid")
	}
	parts := strings.Split(encoded, "|")
	if len(parts) < 1 {
		return nil, fmt.Errorf("encoded %q has too few parts", encoded)
	}
	if len(parts) > 5 {
		return nil, fmt.Errorf("encoded %q has too many parts", encoded)
	}

	leadPos, err := strconv.Atoi(parts[0])
	if err != nil {
		return nil, fmt.Errorf("encoded part[0] %q was not an int: %v", parts[0], err)
	}

	brr, err := NewBiddingRound(leadPos)
	if err != nil {
		return nil, err
	}

	br := brr.(*biddingRound)

	for _, bstr := range parts[1:] {
		if bstr == "" {
			continue
		}
		bid, err := NewBidFromEncoded(bstr)
		if err != nil {
			return nil, err
		}
		br.bids = append(br.bids, bid)
	}
	return br, nil
}

// NewBiddingRound creates a new bidding round starting with the player in leadPos.
func NewBiddingRound(leadPos int) (BiddingRound, error) {
	if leadPos < 0 || leadPos > 3 {
		return nil, errors.New("leadPos must be on the interval [0,3]")
	}
	return &biddingRound{leadPos: leadPos}, nil
}

func (r *biddingRound) IsDone() bool {
	return r.NumPlaced() == 4
}

func (r *biddingRound) placeBid(playerPos int, bid Bid) error {
	ord := r.toOrd(playerPos)
	if len(r.bids) != ord {
		return ErrIncorrectBidOrder
	}
	r.bids = append(r.bids, bid)
	return nil

}

func (r *biddingRound) CurrentTurnPos() (int, error) {
	if r.IsDone() {
		return -1, fmt.Errorf("bidding is complete")
	}
	return (r.leadPos + len(r.bids)) % 4, nil
}

func (r *biddingRound) WinningBidAndPos() (Bid, int, error) {
	// The winning bid is the last non-pass bid.
	if !r.IsDone() {
		return nil, 0, fmt.Errorf("bidding is not done")
	}
	for i := 3; i >= 0; i-- {
		if r.bids[i].IsPass() {
			continue
		}
		return r.bids[i], r.toPos(i), nil
	}
	return nil, 0, fmt.Errorf("no non-pass bids found")
}

func (r *biddingRound) LeadPos() int {
	return r.leadPos
}

func (r *biddingRound) Bids() []Bid {
	return r.bids
}

func (r *biddingRound) NumPlaced() int {
	return len(r.bids)
}

func (r *biddingRound) Encoded() string {
	var bes []string
	for _, bid := range r.bids {
		bes = append(bes, bid.Encoded())
	}
	return strconv.Itoa(r.leadPos) + "|" + strings.Join(bes, "|")
}

// toOrd returns the player order for this bid (0..3), computed from the leadPos.
func (r *biddingRound) toOrd(playerPos int) int {
	return (playerPos + 4 - r.leadPos) % 4
}

// toPos returns the player position for this bid (0..3), computed from the leadPos.
func (r *biddingRound) toPos(playerOrd int) int {
	return (r.leadPos + playerOrd) % 4
}
