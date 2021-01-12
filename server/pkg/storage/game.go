package storage

import (
	"context"
	"time"

	"google.golang.org/appengine/datastore"
)

type Game struct {
	PlayerIDs []string // Organizing player is 0-index; clockwise afterwards (e.g., 0 and 2 are partners).
	Created   time.Time
	Updated   time.Time
	Complete  bool

	Score string `datastore:",noindex"` // The running tally of the game.

	CurrentDealerPos int    `datastore:",noindex"` // The position of the current dealer.
	CurrentBidding   string `datastore:",noindex"` // The bids for the current hand; 0-index is the player clockwise from the CurrentDealerPos (one higher, wrapping at 4).
	CurrentHands     string `datastore:",noindex"` // Cards held by each player, parallel with the PlayerIDs above.
	CurrentTrick     string `datastore:",noindex"` // Cards played for current trick; 0-index is the lead player (i.e., the order the cards were played).
	CurrentTally     string `datastore:",noindex"` // The running tally for the current hand.
}

type GameStore interface {
	Create(ctx context.Context, id, organizingPlayerID string) (*Game, error)
	Get(ctx context.Context, id string) (*Game, error)
	Set(ctx context.Context, id string, g *Game) error
	AddPlayer(ctx context.Context, id, playerID string, pos int) (*Game, error)
}

type datastoreGameStore struct{}

var _ GameStore = (*datastoreGameStore)(nil) // Ensure interface is implemented.

func NewDatastoreGameStore() GameStore {
	return &datastoreGameStore{}
}

func (s *datastoreGameStore) Create(ctx context.Context, id, organizingPlayerID string) (*Game, error) {
	k := gameKey(ctx, id)
	gs := &Game{}
	err := datastore.RunInTransaction(ctx, func(tc context.Context) error {
		found := true
		if err := datastore.Get(ctx, k, gs); err != nil {
			if err != datastore.ErrNoSuchEntity {
				return err
			}
			found = false
		}
		if found {
			return ErrNotUnique
		}

		gs = &Game{}
		gs.PlayerIDs = make([]string, 4)
		gs.PlayerIDs[0] = organizingPlayerID
		gs.Created = time.Now().UTC()
		gs.Updated = gs.Created

		if _, err := datastore.Put(ctx, k, gs); err != nil {
			return err
		}

		return nil
	}, &datastore.TransactionOptions{Attempts: 3})

	if err != nil {
		return nil, err
	}
	return gs, nil
}

func (s *datastoreGameStore) Get(ctx context.Context, id string) (*Game, error) {
	k := gameKey(ctx, id)
	gs := &Game{}
	if err := datastore.Get(ctx, k, gs); err != nil {
		if err == datastore.ErrNoSuchEntity {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return gs, nil

}

func (s *datastoreGameStore) Set(ctx context.Context, id string, gs *Game) error {
	k := gameKey(ctx, id)
	gs.Updated = time.Now().UTC()
	if _, err := datastore.Put(ctx, k, gs); err != nil {
		return err
	}
	return nil
}

func (s *datastoreGameStore) AddPlayer(ctx context.Context, id, playerID string, pos int) (*Game, error) {
	k := gameKey(ctx, id)
	gs := &Game{}
	err := datastore.RunInTransaction(ctx, func(tc context.Context) error {
		if err := datastore.Get(ctx, k, gs); err != nil {
			if err == datastore.ErrNoSuchEntity {
				return ErrNotFound
			}
			return err
		}
		// Error if not unique
		for _, pid := range gs.PlayerIDs {
			if pid == "" {
				continue
			}
			if pid == playerID {
				return ErrPlayerAlreadyAdded
			}
		}

		if gs.PlayerIDs[pos] != "" {
			return ErrPlayerPositionFilled
		}

		gs.PlayerIDs[pos] = playerID
		gs.Updated = time.Now().UTC()
		if _, err := datastore.Put(ctx, k, gs); err != nil {
			return err
		}
		return nil
	}, &datastore.TransactionOptions{Attempts: 3})

	if err != nil {
		return nil, err
	}
	return gs, nil
}

func gameKey(ctx context.Context, id string) *datastore.Key {
	return datastore.NewKey(ctx, "KaiserGame", id, 0, nil)
}
