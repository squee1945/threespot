package game

import "github.com/squee1945/threespot/server/pkg/storage"

type Rules interface {
	SetPassCard(bool)
	PassCard() bool
}

type rules struct {
	passCard bool
}

var _ Rules = (*rules)(nil) // Ensure interface is implemented.

func NewRules() Rules {
	return &rules{}
}

func (r *rules) SetPassCard(passCard bool) {
	r.passCard = passCard
}

func (r *rules) PassCard() bool {
	return r.passCard
}

func rulesFromStorage(sr storage.Rules) Rules {
	return &rules{
		passCard: sr.PassCard,
	}
}
