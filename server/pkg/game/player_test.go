package game

import (
	"context"
	"strings"
	"testing"

	"github.com/squee1945/threespot/server/pkg/storage"
)

func TestNewPlayer(t *testing.T) {
	testCases := []struct {
		name       string
		playerID   string
		playerName string
		wantErr    bool
	}{
		{
			name:       "id required",
			playerID:   "",
			playerName: "Some Name",
			wantErr:    true,
		},
		{
			name:       "name required",
			playerID:   "ABC123",
			playerName: "",
			wantErr:    true,
		},
		{
			name:       "id must be valid",
			playerID:   "A 3",
			playerName: "",
			wantErr:    true,
		},
		{
			name:       "name must be valid",
			playerID:   "ABC123",
			playerName: strings.Repeat("a", 101),
			wantErr:    true,
		},
		{
			name:       "valid",
			playerID:   "ABC123",
			playerName: "This is valid",
			wantErr:    false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			store := storage.NewFakePlayerStore()
			p, err := NewPlayer(ctx, store, tc.playerID, tc.playerName)

			if tc.wantErr && err == nil {
				t.Fatal("wanted error, got err=nil")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tc.wantErr {
				return
			}

			if got, want := p.ID(), tc.playerID; got != want {
				t.Errorf("incorrect ID, got=%q, want=%q", got, want)
			}

			if got, want := p.Name(), tc.playerName; got != want {
				t.Errorf("incorrect name, got=%q, want=%q", got, want)
			}

			// Get the player to make sure they were stored.
			lookup, err := GetPlayer(ctx, store, tc.playerID)
			if err != nil {
				t.Fatal()
			}
			if got, want := lookup.ID(), tc.playerID; got != want {
				t.Errorf("incorrect lookup ID, got=%q, want=%q", got, want)
			}

			if got, want := lookup.Name(), tc.playerName; got != want {
				t.Errorf("incorrect lookup name, got=%q, want=%q", got, want)
			}
		})
	}
}

func TestGetPlayer(t *testing.T) {
	id := "ABC123"
	name := "ABCABC"
	ctx := context.Background()
	store := storage.NewFakePlayerStore()
	_, err := NewPlayer(ctx, store, id, name)
	if err != nil {
		t.Fatal(err)
	}

	lookup, err := GetPlayer(ctx, store, id)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := lookup.ID(), id; got != want {
		t.Errorf("incorrect lookup ID, got=%q, want=%q", got, want)
	}

	if got, want := lookup.Name(), name; got != want {
		t.Errorf("incorrect lookup name, got=%q, want=%q", got, want)
	}

	_, err = GetPlayer(ctx, store, "UNKNOWN")
	if err == nil {
		t.Fatal("missing expected error")
	}
	if err != ErrNotFound {
		t.Fatalf("missing ErrNotFound, got=%v", err)
	}
}

func TestPlayerSetName(t *testing.T) {
	id := "ABC123"
	name := "ABCABC"
	ctx := context.Background()
	store := storage.NewFakePlayerStore()
	player, err := NewPlayer(ctx, store, id, name)
	if err != nil {
		t.Fatal(err)
	}

	player.SetName(ctx, "NEW NAME")

	lookup, err := GetPlayer(ctx, store, id)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := lookup.Name(), "NEW NAME"; got != want {
		t.Errorf("incorrect lookup name, got=%q, want=%q", got, want)
	}
}
