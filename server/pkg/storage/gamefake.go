package storage

import (
	"context"
)

type fakeGameStore struct {
	games map[string]*Game
}

var _ GameStore = (*fakeGameStore)(nil) // Ensure interface is implemented.

// NewFakeGameStore creates an in-memory game store, accepting an initial map of id->*Game.
func NewFakeGameStore(games map[string]*Game) GameStore {
	f := &fakeGameStore{
		games: make(map[string]*Game),
	}
	for k, v := range games {
		f.games[k] = v
	}
	return f
}

func (s *fakeGameStore) Create(ctx context.Context, id, organizingPlayerID string) (*Game, error) {
	for k := range s.games {
		if k == id {
			return nil, ErrNotUnique
		}
	}
	g := &Game{
		PlayerIDs: make([]string, 4),
	}
	g.PlayerIDs[0] = organizingPlayerID
	s.games[id] = g
	return g, nil
}

func (s *fakeGameStore) Get(ctx context.Context, id string) (*Game, error) {
	g, present := s.games[id]
	if !present {
		return nil, ErrNotFound
	}
	return g, nil
}

func (s *fakeGameStore) Set(ctx context.Context, id string, g *Game) error {
	s.games[id] = g
	return nil
}

func (s *fakeGameStore) AddPlayer(ctx context.Context, id, playerID string, pos int) (*Game, error) {
	g, present := s.games[id]
	if !present {
		return nil, ErrNotFound
	}
	for _, pid := range g.PlayerIDs {
		if pid == playerID {
			return nil, ErrPlayerAlreadyAdded
		}
	}
	if g.PlayerIDs[pos] != "" {
		return nil, ErrPlayerPositionFilled
	}
	g.PlayerIDs[pos] = playerID
	return g, nil
}
