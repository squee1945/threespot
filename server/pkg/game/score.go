package game

type Score interface {
	Encoded() string
}

type score struct{}

var _ Score = (*score)(nil) // Ensure interface is implemented.

func NewScoreFromEncoded(encoded string) (Score, error) {
	return nil, nil
}

func (s *score) Encoded() string {
	return ""
}
