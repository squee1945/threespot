package game

import (
	"context"
	"errors"
	"fmt"
	"regexp"

	"github.com/squee1945/threespot/server/pkg/storage"
)

// Player is a card player.
type Player interface {
	// ID is the unique ID of this player.
	ID() string
	// Name is the human-readable name of the player. It can be changed so it should not be used for any references.
	Name() string
	// SetName updates the name of the player.
	SetName(context.Context, string) (Player, error)
}

const (
	maxPlayerName = 100
)

var (
	validPlayerID = regexp.MustCompile(`^[A-Z0-9]{6,20}$`)
)

type player struct {
	store    storage.PlayerStore
	id, name string
}

var _ Player = (*player)(nil) // Ensure interface is implemented.

// NewPlayer creates a new player, storing it in the PlayerStore.
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
	return playerFromStorage(store, id, ps)
}

// GetPlayer fetches the player from the PlayerStore, returning ErrNotFound if not found.
func GetPlayer(ctx context.Context, store storage.PlayerStore, id string) (Player, error) {
	ps, err := store.Get(ctx, id)
	if err != nil {
		if err == storage.ErrNotFound {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("fetching player from storage: %v", err)
	}
	return playerFromStorage(store, id, ps)
}

func (p *player) ID() string {
	return p.id
}

func (p *player) Name() string {
	return p.name
}

func (p *player) SetName(ctx context.Context, name string) (Player, error) {
	p.name = name
	return p.save(ctx)
}

func (p *player) save(ctx context.Context) (Player, error) {
	ps := &storage.Player{
		Name: p.name,
	}
	if err := p.store.Set(ctx, p.id, ps); err != nil {
		return nil, fmt.Errorf("saving player: %v", err)
	}
	return p, nil
}

func playerFromStorage(store storage.PlayerStore, id string, ps *storage.Player) (Player, error) {
	if ps == nil {
		return nil, errors.New("nil player")
	}
	return &player{
		store: store,
		id:    id,
		name:  ps.Name,
	}, nil
}
