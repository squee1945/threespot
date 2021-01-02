package game

import (
	"fmt"
	"regexp"

	"github.com/squee1945/threespot/server/pkg/deck"
	"github.com/squee1945/threespot/server/pkg/storage"
)

const (
	maxPlayerName = 100
)

var (
	validPlayerID = regexp.MustCompile(`^[a-z0-9]{6,20}$`)
)

// NewPlayer creates a new player.
func NewPlayer(ctx context.Context, store storage.PlayerStorage, id, name string) (Player, error) {
	if id == "" || name == "" {
		return nil, fmt.Errorf("id and name required")
	}
	if !validPlayerID.MatchString(id) {
		return nil, fmt.Errorf("invalid id %q", id)
	}
	if len(name) > maxPlayerName {
		return nil, fmt.Errorf("name %q is too long", name)
	}

	p, err := store.Create(ctx, id, name)
	if err != nil {
		return nil, fmt.Errorf("creating player in storage: %v", err)
	}

	player = &player{
		store: store,
		id: id,
		name: name,
	}
	return player, nil
}

func GetPlayer(ctx context.Context, store storage.PlayerStorage, id string) (Player, error) {
	p, err := store.Get(ctx, id)
	if err != nil {
		if err == storage.ErrNotFound {
			return nil, game.ErrNotFound
		}
		return nil, fmt.Errorf("fetching player from storage: %v", err)
	player = &player{
		store: store,
		id: id,
		name: p.Name,
	}
	return player, nil
}

// Player is a card player.
type Player interface {
	Name() string
	ID() string
	SetHand([]deck.Card)
	Hand() []deck.Card
	SetName(string) error
}

type player struct {
	store storage.PlayerStore

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

func (p *player) SetName(ctx context.Context, name string) error {
	p.name = name
	ps := storage.Player{
		Name: name,
	}
	if err := p.store.Set(ctx, ps); err != nil {
		return fmt.Errorf("saving player in store: %v", err)
	}
	return nil
}
