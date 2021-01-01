package game

import (
	"fmt"
	"regexp"

	"github.com/squee1945/threespot/server/pkg/deck"
)

const (
	maxPlayerName = 100
)

var (
	validPlayerID = regexp.MustCompile(`^[a-z0-9]{6,20}$`)
)

// NewPlayer creates a new player.
func NewPlayer(id, name string) (Player, error) {
	if id == "" || name == "" {
		return nil, fmt.Errorf("id and name required")
	}
	if !validPlayerID.MatchString(id) {
		return nil, fmt.Errorf("invalid id %q", id)
	}
	if len(name) > maxPlayerName {
		return nil, fmt.Errorf("name %q is too long", name)
	}
	// TODO: store in datastore
	return &player{id: id, name: name}, nil
}

func GetPlayer(id string) (Player, error) {
	return nil, nil // TODO: fetch from datastore
}

// Player is a card player.
type Player interface {
	Name() string
	ID() string
	SetHand([]deck.Card)
	Hand() []deck.Card
}

type player struct {
	id, name string
	hand     []deck.Card
}

func (p *player) ID() string {
	return p.id
}

func (p *player) Name() string {
	return p.name
}

func (p *player) SetHand(hand []deck.Card) {
	p.hand = hand
}

func (p *player) Hand() []deck.Card {
	return p.hand
}
