package storage

import (
	"context"
)

type fakePlayerStore struct {
	players map[string]*Player
}

var _ PlayerStore = (*fakePlayerStore)(nil) // Ensure interface is implemented.

// NewFakePlayerStore creates an in-memory player store, accepting an initial map of id->*Player.
func NewFakePlayerStore(players map[string]*Player) PlayerStore {
	f := &fakePlayerStore{
		players: make(map[string]*Player),
	}
	for k, v := range players {
		f.players[k] = v
	}
	return f
}

func (s *fakePlayerStore) Create(ctx context.Context, id, name string) (*Player, error) {
	for k := range s.players {
		if k == id {
			return nil, ErrNotUnique
		}
	}
	p := &Player{Name: name}
	s.players[id] = p
	return p, nil
}

func (s *fakePlayerStore) Get(ctx context.Context, id string) (*Player, error) {
	p, present := s.players[id]
	if !present {
		return nil, ErrNotFound
	}
	return p, nil
}

func (s *fakePlayerStore) GetMulti(ctx context.Context, ids []string) ([]*Player, error) {
	result := make([]*Player, len(ids))
	for i, id := range ids {
		result[i] = s.players[id]
	}
	return result, nil
}

func (s *fakePlayerStore) Set(ctx context.Context, id string, p *Player) error {
	s.players[id] = p
	return nil
}
