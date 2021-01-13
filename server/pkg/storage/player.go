package storage

import (
	"context"
	"fmt"

	"google.golang.org/appengine/datastore"
)

type Player struct {
	Name string `datastore:",noindex"`
}

type PlayerStore interface {
	Create(ctx context.Context, id, name string) (*Player, error)
	Get(ctx context.Context, id string) (*Player, error)
	GetMulti(ctx context.Context, ids []string) ([]*Player, error)
	Set(ctx context.Context, id string, p *Player) error
}

type datastorePlayerStore struct{}

var _ PlayerStore = (*datastorePlayerStore)(nil) // Ensure interface is implemented.

func NewDatastorePlayerStore() PlayerStore {
	return &datastorePlayerStore{}
}

func (s *datastorePlayerStore) Create(ctx context.Context, id, name string) (*Player, error) {
	if id == "" {
		return nil, fmt.Errorf("id required")
	}
	if name == "" {
		return nil, fmt.Errorf("name required")
	}
	k := playerKey(ctx, id)
	ps := &Player{}
	err := datastore.RunInTransaction(ctx, func(tc context.Context) error {
		found := true
		if err := datastore.Get(ctx, k, ps); err != nil {
			if err != datastore.ErrNoSuchEntity {
				return err
			}
			found = false
		}
		if found {
			return ErrNotUnique
		}

		ps = &Player{}
		ps.Name = name

		if _, err := datastore.Put(ctx, k, ps); err != nil {
			return err
		}

		return nil
	}, &datastore.TransactionOptions{Attempts: 3})

	if err != nil {
		return nil, err
	}
	return ps, nil
}

func (s *datastorePlayerStore) Get(ctx context.Context, id string) (*Player, error) {
	k := playerKey(ctx, id)
	ps := &Player{}
	if err := datastore.Get(ctx, k, ps); err != nil {
		if err == datastore.ErrNoSuchEntity {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return ps, nil
}

func (s *datastorePlayerStore) GetMulti(ctx context.Context, ids []string) ([]*Player, error) {
	var keys []*datastore.Key
	keyPos := make(map[string]int)
	for i, id := range ids {
		if id == "" {
			continue
		}
		keyPos[id] = i
		keys = append(keys, playerKey(ctx, id))
	}

	pss := make([]*Player, len(keys))
	if err := datastore.GetMulti(ctx, keys, pss); err != nil {
		return nil, err
	}

	result := make([]*Player, len(ids))
	for i := 0; i < len(keys); i++ {
		pos := keyPos[keys[i].StringID()]
		result[pos] = pss[i]
	}
	return result, nil
}

func (s *datastorePlayerStore) Set(ctx context.Context, id string, p *Player) error {
	k := playerKey(ctx, id)
	if _, err := datastore.Put(ctx, k, p); err != nil {
		return err
	}
	return nil
}

func playerKey(ctx context.Context, id string) *datastore.Key {
	return datastore.NewKey(ctx, "KaiserPlayer", id, 0, nil)
}
