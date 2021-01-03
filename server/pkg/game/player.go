package game

import (
	"context"
	"fmt"
	"regexp"

	"github.com/squee1945/threespot/server/pkg/storage"
)

const (
	maxPlayerName = 100
)

var (
	validPlayerID = regexp.MustCompile(`^[A-Z0-9]{6,20}$`)
)

// NewPlayer creates a new player.
func NewPlayer(ctx context.Context, store storage.PlayerStore, id, name string) (Player, error) {
	if id == "" || name == "" {
		return nil, fmt.Errorf("id and name required")
	}
	if !validPlayerID.MatchString(id) {
		return nil, fmt.Errorf("invalid id %q", id)
	}
	if len(name) > maxPlayerName {
		return nil, fmt.Errorf("name %q is too long", name)
	}

	ps, err := store.Create(ctx, id, name)
	if err != nil {
		return nil, fmt.Errorf("creating player in storage: %v", err)
	}
	return playerFromStorage(store, id, ps), nil
}

func GetPlayer(ctx context.Context, store storage.PlayerStore, id string) (Player, error) {
	ps, err := store.Get(ctx, id)
	if err != nil {
		if err == storage.ErrNotFound {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("fetching player from storage: %v", err)
	}
	return playerFromStorage(store, id, ps), nil
}

// Player is a card player.
type Player interface {
	ID() string
	Name() string
	// SetHand([]deck.Card)
	// Hand() []deck.Card
	SetName(context.Context, string) error
}

type player struct {
	store storage.PlayerStore

	id, name string
	// hand     []deck.Card
}

func (p *player) ID() string {
	return p.id
}

func (p *player) Name() string {
	return p.name
}

// func (p *player) SetHand(hand []deck.Card) {
// 	p.hand = hand
// }

// func (p *player) Hand() []deck.Card {
// 	return p.hand
// }

func (p *player) SetName(ctx context.Context, name string) error {
	p.name = name
	ps := storage.Player{
		Name: name,
	}
	if err := p.store.Set(ctx, p.id, ps); err != nil {
		return fmt.Errorf("saving player in store: %v", err)
	}
	return nil
}

func playerFromStorage(store storage.PlayerStore, id string, ps storage.Player) Player {
	return player{
		store: store,
		id:    id,
		name:  ps.Name,
	}
}
