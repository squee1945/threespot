package game

type Score interface {
	Encoded() string
}

type score struct{}

func NewScoreFromEncoded(encoded string) (Score, error) {
	return nil, nil
}

func (s *score) Encoded() string {
	return ""
}
