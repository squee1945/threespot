package game

type Bid interface {
	String() string
	Suit() Suit
}

func NewBidFromString(bidStr string) (Bid, error) {
	return "", nil
}

func nextBids(currentBids []Bid, isDealer bool) []Bid {
	return nil
}
