package game

import (
	"context"
	"testing"

	"github.com/squee1945/threespot/server/pkg/storage"
)

func TestNewGame(t *testing.T) {
	ctx := context.Background()
	gameStore := storage.NewFakeGameStore(nil)
	playerStore := storage.NewFakePlayerStore()
	id := "ABC123"
	organizer := buildPlayer(t, playerStore, "FOOBAR", "Jake Cole")
	_, err := NewGame(ctx, gameStore, playerStore, id, organizer)
	if err != nil {
		t.Fatal(err)
	}
	// TODO
}
func TestGetGame(t *testing.T) {
	// TODO
}
func TestState(t *testing.T) {
	// TODO
}
func TestPlayerPos(t *testing.T) {
	//TODO
}
func TestPosToPlay(t *testing.T) {
	//TODO
}
func TestAvailableBids(t *testing.T) {
	//TODO
}
func TestPlayerHand(t *testing.T) {
	//TODO
}
func TestAddPlayer(t *testing.T) {
	// TODO
}
func TestPlaceBid(t *testing.T) {
	// TODO
}
func TestCallTrump(t *testing.T) {
	// TODO
}
func TestPlayCard(t *testing.T) {
	// TODO
}
func TestSave(t *testing.T) {
	// TODO
}
func TestGameFromStorage(t *testing.T) {
	//TODO
}

func buildPlayer(t *testing.T, playerStore storage.PlayerStore, id, name string) Player {
	t.Helper()
	p, err := NewPlayer(context.Background(), playerStore, id, name)
	if err != nil {
		t.Fatal(err)
	}
	return p
}
