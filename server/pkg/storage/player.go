package storage

import (
	"context"

	"cloud.google.com/go/datastore"
)

type Player struct {
	Name string
}

type PlayerStore interface {
	Create(ctx context.Context, id, name string) (Player, error)
	Get(ctx context.Context, id string) (Player, error)
	Set(ctx context.Context, p Player) error
}

type datastorePlayerStore struct {
	dsClient datastore.Client
}

func NewDatastorePlayerStore(dsClient datastore.Client) PlayerStore {
	return datastorePlayerStore{
		dsClient: dsClient,
	}
}

func (ds *datastorePlayerStore) Create(ctx context.Context, id, name string) (Player, error) {
	var (
		p   Player
		err error
		tx  *datastore.Transaction
	)

	// Lookup, set in transaction to ensure uniqueness
	k := playerKey(id)
	for i := 0; i < retries; i++ {
		tx, err = client.NewTransaction(ctx)
		if err != nil {
			break
		}

		found := true
		if err = tx.Get(k, &p); err != nil {
			if err == datastore.ErrNoSuchEntity {
				// This is good and what we're looking for.
				found = false
			} else {
				break
			}
		}
		if found {
			return nil, ErrNotUnique
		}

		p.Name = name
		if _, err = tx.Put(k, &p); err != nil {
			break
		}

		// Attempt to commit the transaction. If there's a conflict, try again.
		if _, err = tx.Commit(); err != datastore.ErrConcurrentTransaction {
			break
		}
	}
	if err != nil {
		return nil, err
	}
	return p
}

func (ds *datastorePlayerStore) Get(ctx context.Context, id string) (Player, error) {
	k := playerKey(id)
	var p Player
	if err := dsClient.Get(ctx, k, &p); err != nil {
		return nil, err
	}
	return p, nil
}

func (ds *datastorePlayerStore) Set(ctx context.Context, p Player) error {
	k := playerKey(id)
	if _, err := dsClient.Put(ctx, k, p); err != nil {
		return err
	}
	return nil
}

func playerKey(id string) *datastore.Key {
	return datastore.NameKey("Player", id, nil)
}
