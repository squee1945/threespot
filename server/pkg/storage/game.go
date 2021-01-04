package storage

import (
	"context"
	"time"

	"cloud.google.com/go/datastore"
)

type Game struct {
	PlayerIDs []string // Organizing player is 0-index; clockwise afterwards (e.g., 0 and 2 are partners).
	Created   time.Time
	Complete  bool

	Score string `datastore:",noindex"`
	// Score        []int   `datastore:",noindex"` // 0-index is player 0/2 score; 1-index is player 1/3 score.
	// ScoreHistory [][]int `datastore:",noindex"` // A list of pairs like above.

	CurrentDealerPos  int      `datastore:",noindex"` // The position of the current dealer.
	CurrentBids       []string `datastore:",noindex"` // The bids for the current hand; empty string is 'pass'. 0-index is the player clockwise from the CurrentDealerPos.
	CurrentWinningBid string   `datastore:",noindex"` // The winning bid for the current hand.
	CurrentWinningPos int      `datastore:",noindex"` // The position of the winning bidder (who will play first).
	CurrentHands      []string `datastore:",noindex"` // Cards held by each player, parallel with the PlayerIDs above.
	CurrentTrick      string   `datastore:",noindex"` // Cards played for current trick; 0-index is the lead player (i.e., the order the cards were played).
	CurrentTally      string   `datastore:",noindex"` // The running tally for the current hand.
}

type GameStore interface {
	Create(ctx context.Context, id, organizingPlayerID string) (*Game, error)
	Get(ctx context.Context, id string) (*Game, error)
	Set(ctx context.Context, id string, g *Game) error
	AddPlayer(ctx context.Context, id, playerID string, pos int) (*Game, error)
}

type datastoreGameStore struct {
	dsClient *datastore.Client
}

var _ GameStore = (*datastoreGameStore)(nil) // Ensure interface is implemented.

func NewDatastoreGameStore(dsClient *datastore.Client) GameStore {
	return &datastoreGameStore{
		dsClient: dsClient,
	}
}

func (s *datastoreGameStore) Create(ctx context.Context, id, organizingPlayerID string) (*Game, error) {
	var (
		g   Game
		err error
		tx  *datastore.Transaction
	)

	// Lookup, set in transaction to ensure uniqueness
	k := gameKey(id)
	for i := 0; i < retries; i++ {
		tx, err = s.dsClient.NewTransaction(ctx)
		if err != nil {
			break
		}

		found := true
		if err = tx.Get(k, &g); err != nil {
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

		g.PlayerIDs = make([]string, 4)
		g.PlayerIDs[0] = organizingPlayerID
		g.Created = time.Now().UTC()
		if _, err = tx.Put(k, &g); err != nil {
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
	return &g, nil
}

func (s *datastoreGameStore) Get(ctx context.Context, id string) (*Game, error) {
	k := gameKey(id)
	var g Game
	if err := s.dsClient.Get(ctx, k, &g); err != nil {
		if err == datastore.ErrNoSuchEntity {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &g, nil

}

func (s *datastoreGameStore) Set(ctx context.Context, id string, g *Game) error {
	k := gameKey(id)
	if _, err := s.dsClient.Put(ctx, k, g); err != nil {
		return err
	}
	return nil
}

func (s *datastoreGameStore) AddPlayer(ctx context.Context, id, playerID string, pos int) (*Game, error) {
	var (
		g   Game
		err error
		tx  *datastore.Transaction
	)

	// Lookup, set in transaction to ensure only one player placed in a position.
	k := gameKey(id)
	for i := 0; i < retries; i++ {
		tx, err = s.dsClient.NewTransaction(ctx)
		if err != nil {
			break
		}

		if err = tx.Get(k, &g); err != nil {
			if err == datastore.ErrNoSuchEntity {
				err = ErrNotFound
				break
			} else {
				break
			}
		}

		// Error if not unique
		for _, pid := range g.PlayerIDs {
			if pid == "" {
				continue
			}
			if pid == playerID {
				err = ErrPlayerAlreadyAdded
				break
			}
		}

		if g.PlayerIDs[pos] != "" {
			err = ErrPlayerPositionFilled
			break
		}

		g.PlayerIDs[pos] = playerID
		if _, err = tx.Put(k, &g); err != nil {
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
	return &g, nil
}

func gameKey(id string) *datastore.Key {
	return datastore.NameKey("KaiserGame", id, nil)
}
