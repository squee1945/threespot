package game

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/squee1945/threespot/server/pkg/storage"
)

var (
	comparePlayers = cmp.Comparer(func(p1, p2 Player) bool {
		return p1 == nil && p2 == nil || (p1 != nil && p2 != nil && p1.ID() == p2.ID())
	})

	ignoreDates = cmpopts.IgnoreFields(storage.Game{}, "Created", "Updated")
	ignoreHands = cmpopts.IgnoreFields(storage.Game{}, "CurrentHands")
)

func TestNewGame(t *testing.T) {
	ctx := context.Background()
	gameStore := storage.NewFakeGameStore(nil)
	playerStore := storage.NewFakePlayerStore()
	id := "ABC123"
	organizer := buildPlayer(t, playerStore, "PLAYERID")
	g, err := NewGame(ctx, gameStore, playerStore, id, organizer, NewRules())
	if err != nil {
		t.Fatal(err)
	}

	if got, want := g.ID(), id; got != want {
		t.Errorf("ID()=%q want=%q", got, want)
	}
	if diff := cmp.Diff([]Player{organizer, nil, nil, nil}, g.Players(), comparePlayers); diff != "" {
		t.Errorf("Players() mismatch (-want +got):\n%s", diff)
	}
	if got, want := g.State(), JoiningState; got != want {
		t.Errorf("State=%v want=%v", got, want)
	}

	// Duplicate must raise error.
	_, err = NewGame(ctx, gameStore, playerStore, id, organizer, NewRules())
	if err == nil {
		t.Errorf("missing expected error")
	}
}

func TestGetGame(t *testing.T) {
	ctx := context.Background()
	gameStore := storage.NewFakeGameStore(nil)
	playerStore := storage.NewFakePlayerStore()
	id := "ABC123"
	organizer := buildPlayer(t, playerStore, "PLAYERID")
	_, err := NewGame(ctx, gameStore, playerStore, id, organizer, NewRules())
	if err != nil {
		t.Fatal(err)
	}

	g, err := GetGame(ctx, gameStore, playerStore, id)
	if err != nil {
		t.Fatal(err)
	}

	if got, want := g.ID(), id; got != want {
		t.Errorf("ID()=%q want=%q", got, want)
	}
	if diff := cmp.Diff([]Player{organizer, nil, nil, nil}, g.Players(), comparePlayers); diff != "" {
		t.Errorf("Players() mismatch (-want +got):\n%s", diff)
	}

	// Not found.
	_, err = GetGame(ctx, gameStore, playerStore, "UNKNOWN")
	if err != ErrNotFound {
		t.Errorf("incorrect error got=%v want=%v", err, ErrNotFound)
	}
}

func TestState(t *testing.T) {
	testCases := []struct {
		want GameState
		gs   *storage.Game
	}{
		{
			want: JoiningState,
			gs: &storage.Game{
				PlayerIDs: []string{"ABE", "BOB", "CAL"},
			},
		},
		{
			want: CompletedState,
			gs: &storage.Game{
				PlayerIDs:      []string{"ABE", "BOB", "CAL", "DON"},
				Complete:       true,
				CurrentBidding: "0|P|P|P",
				CurrentTrick:   "3|D",
			},
		},
		{
			want: DealingState,
			gs: &storage.Game{
				PlayerIDs:      []string{"ABE", "BOB", "CAL", "DON"},
				CurrentBidding: "0|P|P|P",
			},
		},
		{
			want: BiddingState,
			gs: &storage.Game{
				PlayerIDs:      []string{"ABE", "BOB", "CAL", "DON"},
				CurrentHands:   "AH+AS+AD+AC", // Non-empty hand
				CurrentBidding: "0|P|P|P",
			},
		},
		{
			want: CallingState,
			gs: &storage.Game{
				PlayerIDs:      []string{"ABE", "BOB", "CAL", "DON"},
				CurrentHands:   "AH+AS+AD+AC", // Non-empty hand
				CurrentBidding: "0|P|P|P|7",
			},
		},
		{
			want: PlayingState,
			gs: &storage.Game{
				PlayerIDs:      []string{"ABE", "BOB", "CAL", "DON"},
				CurrentBidding: "0|P|P|P|7",
				CurrentHands:   "AH+AS+AD+AC", // Non-empty hand
				CurrentTrick:   "3|D",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.want.String(), func(t *testing.T) {
			g, _, _ := buildGame(t, tc.gs)

			if got, want := g.State(), tc.want; got != want {
				t.Errorf("State()=%s want=%s", got, want)
			}
		})
	}
}

func TestPlayerPos(t *testing.T) {
	testCases := []struct {
		name    string
		gs      *storage.Game
		pid     string
		want    int
		wantErr bool
	}{
		{
			name: "first player",
			gs: &storage.Game{
				PlayerIDs:      []string{"ABE", "BOB", "CAL", "DON"},
				CurrentBidding: "0|P|P|P",
			},
			pid:  "ABE",
			want: 0,
		},
		{
			name: "last player",
			gs: &storage.Game{
				PlayerIDs:      []string{"ABE", "BOB", "CAL", "DON"},
				CurrentBidding: "0|P|P|P",
			},
			pid:  "DON",
			want: 3,
		},
		{
			name: "unknown player",
			gs: &storage.Game{
				PlayerIDs:      []string{"ABE", "BOB", "CAL", "DON"},
				CurrentBidding: "0|P|P|P",
			},
			pid:     "NOTINGAME",
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			g, _, playerStore := buildGame(t, tc.gs)
			player, err := GetPlayer(ctx, playerStore, tc.pid)
			if err != nil {
				t.Fatal(err)
			}

			pos, err := g.PlayerPos(player)

			if tc.wantErr && err == nil {
				t.Fatal("missing expected error")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tc.wantErr {
				return
			}

			if got, want := pos, tc.want; got != want {
				t.Errorf("PlayerPos()=%d want=%d", got, want)
			}
		})
	}
}

func TestPosToPlay(t *testing.T) {
	testCases := []struct {
		name string
		gs   *storage.Game
		want int
	}{
		{
			name: "dealing pos",
			gs: &storage.Game{
				CurrentDealerPos: 3,
				PlayerIDs:        []string{"ABE", "BOB", "CAL", "DON"},
				CurrentBidding:   "2|P|P|P",
			},
			want: 3,
		},
		{
			name: "bidding pos",
			gs: &storage.Game{
				PlayerIDs:      []string{"ABE", "BOB", "CAL", "DON"},
				CurrentBidding: "2|P|P|P",
				CurrentHands:   "AH+AS+AD+AC",
			},
			want: 1,
		},
		{
			name: "calling trump pos",
			gs: &storage.Game{
				PlayerIDs:      []string{"ABE", "BOB", "CAL", "DON"},
				CurrentBidding: "2|P|P|7|P",
				CurrentHands:   "AH+AS+AD+AC",
			},
			want: 0,
		},
		{
			name: "playing pos",
			gs: &storage.Game{
				PlayerIDs:      []string{"ABE", "BOB", "CAL", "DON"},
				CurrentBidding: "0|P|P|P|7",
				CurrentTrick:   "3|D",
				CurrentHands:   "AH+AS+AD+AC",
			},
			want: 3,
		},
		{
			name: "joining pos",
			gs: &storage.Game{
				PlayerIDs: []string{"ABE", "", "CAL", ""},
			},
			want: -1,
		},
		{
			name: "completed pos",
			gs: &storage.Game{
				PlayerIDs: []string{"ABE", "BOB", "CAL", "DON"},
				Complete:  true,
			},
			want: -1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			g, _, _ := buildGame(t, tc.gs)

			pos, err := g.PosToPlay()
			if err != nil {
				t.Fatal(err)
			}

			if got, want := pos, tc.want; got != want {
				t.Errorf("PosToPlay()=%d want=%d", got, want)
			}
		})
	}
}

func TestAvailableBids(t *testing.T) {
	testCases := []struct {
		name    string
		gs      *storage.Game
		pid     string
		want    []string
		wantErr bool
	}{
		{
			name: "not bidding",
			gs: &storage.Game{
				PlayerIDs: []string{"ABE", "BOB", "", ""},
			},
			pid:     "ABE",
			wantErr: true,
		},
		{
			name: "lead bidder",
			gs: &storage.Game{
				PlayerIDs:      []string{"ABE", "BOB", "CAL", "DON"},
				CurrentBidding: "2",
				CurrentHands:   "AH+AS+AD+AC",
			},
			pid:  "CAL",
			want: orderedBids, // all bids
		},
		{
			name: "some bidder",
			gs: &storage.Game{
				PlayerIDs:      []string{"ABE", "BOB", "CAL", "DON"},
				CurrentBidding: "1|P|P",
				CurrentHands:   "AH+AS+AD+AC",
			},
			pid:  "DON",
			want: orderedBids, // all bids
		},
		{
			name: "bidding complete",
			gs: &storage.Game{
				PlayerIDs:      []string{"ABE", "BOB", "CAL", "DON"},
				CurrentBidding: "1|P|P|P|7",
				CurrentHands:   "AH+AS+AD+AC",
			},
			pid:     "DON",
			wantErr: true,
		},
		{
			name: "unknown player",
			gs: &storage.Game{
				PlayerIDs:      []string{"ABE", "BOB", "CAL", "DON"},
				CurrentBidding: "2",
				CurrentHands:   "AH+AS+AD+AC",
			},
			pid:     "NOTINGAME", // Not in this game.
			wantErr: true,
		},
		{
			name: "dealer can take highest",
			gs: &storage.Game{
				PlayerIDs:        []string{"ABE", "BOB", "CAL", "DON"},
				CurrentBidding:   "2|C|P|P",
				CurrentDealerPos: 1,
				CurrentHands:     "AH+AS+AD+AC",
			},
			pid:  "BOB",
			want: []string{"P", "C", "CN"},
		},
		{
			name: "incorrect order",
			gs: &storage.Game{
				PlayerIDs:      []string{"ABE", "BOB", "CAL", "DON"},
				CurrentBidding: "0",
				CurrentHands:   "AH+AS+AD+AC",
			},
			pid:     "BOB", // Should be ABE's turn.
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			g, _, playerStore := buildGame(t, tc.gs)
			player := getPlayer(t, playerStore, tc.pid)

			gotBids, err := g.AvailableBids(player)

			if tc.wantErr && err == nil {
				t.Fatal("missing expected error")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tc.wantErr {
				return
			}

			var wantBids []Bid
			for _, b := range tc.want {
				wantBids = append(wantBids, buildBid(t, b))
			}
			if diff := cmp.Diff(wantBids, gotBids, compareBids); diff != "" {
				t.Errorf("AvailableBids() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestPlayerHand(t *testing.T) {
	testCases := []struct {
		name    string
		gs      *storage.Game
		pid     string
		want    []string
		wantErr bool
	}{
		{
			name: "first player",
			gs: &storage.Game{
				PlayerIDs:    []string{"ABE", "BOB", "CAL", "DON"},
				CurrentHands: "AH|KH+AD|KD+AS|KS+AC|KC",
			},
			pid:  "ABE",
			want: []string{"AH", "KH"},
		},
		{
			name: "last player",
			gs: &storage.Game{
				PlayerIDs:    []string{"ABE", "BOB", "CAL", "DON"},
				CurrentHands: "AH|KH+AD|KD+AS|KS+AC|KC",
			},
			pid:  "DON",
			want: []string{"AC", "KC"},
		},
		{
			name: "unknown player",
			gs: &storage.Game{
				PlayerIDs:    []string{"ABE", "BOB", "CAL", "DON"},
				CurrentHands: "AH|KH+AD|KD+AS|KS+AC|KC",
			},
			pid:     "NOTINGAME",
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			g, _, playerStore := buildGame(t, tc.gs)
			player := getPlayer(t, playerStore, tc.pid)

			gotHand, err := g.PlayerHand(player)

			if tc.wantErr && err == nil {
				t.Fatal("missing expected error")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tc.wantErr {
				return
			}

			wantHand := buildHand(t, tc.want)
			if diff := cmp.Diff(wantHand, gotHand, compareHands); diff != "" {
				t.Errorf("AvailableBids() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestAddPlayer(t *testing.T) {
	testCases := []struct {
		name    string
		gs      *storage.Game
		pid     string
		pos     int
		want    []string
		wantErr error
	}{
		{
			name: "first player",
			gs: &storage.Game{
				PlayerIDs: []string{"ABE", "", "", ""},
			},
			pid:  "BOB",
			pos:  1,
			want: []string{"ABE", "BOB", "", ""},
		},
		{
			name: "all players",
			gs: &storage.Game{
				PlayerIDs: []string{"ABE", "BOB", "", "DON"},
			},
			pid:  "CAL",
			pos:  2,
			want: []string{"ABE", "BOB", "CAL", "DON"},
		},
		{
			name: "position filled",
			gs: &storage.Game{
				PlayerIDs: []string{"ABE", "BOB", "", "DON"},
			},
			pid:     "CAL",
			pos:     1,
			wantErr: ErrPlayerPositionFilled,
		},
		{
			name: "duplicate player",
			gs: &storage.Game{
				PlayerIDs: []string{"ABE", "", "", ""},
			},
			pid:     "ABE",
			pos:     1,
			wantErr: ErrPlayerAlreadyAdded,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			g, _, playerStore := buildGame(t, tc.gs)
			player := getPlayer(t, playerStore, tc.pid)

			gotGame, err := g.AddPlayer(ctx, player, tc.pos)

			if tc.wantErr != nil && tc.wantErr != err {
				t.Fatalf("incorrect error got=%v want=%v", err, tc.wantErr)
			}
			if tc.wantErr == nil && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tc.wantErr != nil {
				return
			}

			var gotPlayerIDs []string
			for _, p := range gotGame.Players() {
				if p == nil {
					gotPlayerIDs = append(gotPlayerIDs, "")
					continue
				}
				gotPlayerIDs = append(gotPlayerIDs, p.ID())
			}
			if diff := cmp.Diff(tc.want, gotPlayerIDs); diff != "" {
				t.Errorf("player IDs mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestDealCards(t *testing.T) {
	pids := []string{"ABE", "BOB", "CAL", "DON"}
	testCases := []struct {
		name      string
		gs        *storage.Game
		pid       string
		want      *storage.Game
		wantState GameState
		wantErr   error
	}{
		{
			name: "not in dealing state",
			gs: &storage.Game{
				PlayerIDs: []string{"ABE"}, // Still joining
			},
			pid:     "BOB",
			wantErr: ErrNotDealing,
		},
		{
			name: "incorrect dealer order",
			gs: &storage.Game{
				PlayerIDs:        pids,
				CurrentDealerPos: 3,
			},
			pid:     "BOB",
			wantErr: ErrIncorrectDealer,
		},
		{
			name: "hard started",
			gs: &storage.Game{
				PlayerIDs:        pids,
				CurrentDealerPos: 1,
			},
			pid: "BOB",
			want: &storage.Game{
				PlayerIDs:        pids,
				Score:            "52-",
				CurrentDealerPos: 2,
				CurrentBidding:   "3|",
				CurrentTally:     "0|0|0",
				PassedCards:      "3|",
			},
			wantState: BiddingState,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			g, _, playerStore := buildGame(t, tc.gs)
			player := getPlayer(t, playerStore, tc.pid)

			gotGame, err := g.DealCards(ctx, player)

			if tc.wantErr != nil && tc.wantErr != err {
				t.Fatalf("incorrect error got=%v want=%v", err, tc.wantErr)
			}
			if tc.wantErr == nil && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tc.wantErr != nil {
				return
			}

			if got, want := gotGame.State(), tc.wantState; got != want {
				t.Errorf("State()=%s want=%s", got, want)
			}
			gotGameStorage := storageFromGame(gotGame.(*game))
			opts := []cmp.Option{ignoreDates, ignoreHands}
			if diff := cmp.Diff(tc.want, gotGameStorage, opts...); diff != "" {
				t.Errorf("game storage mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestPassCard(t *testing.T) {
	pids := []string{"ABE", "BOB", "CAL", "DON"}
	testCases := []struct {
		name      string
		gs        *storage.Game
		pid       string
		card      string
		want      *storage.Game
		wantState GameState
		wantErr   error
	}{
		{
			name: "error if rules do not allow passing",
			gs: &storage.Game{
				PlayerIDs:    pids,
				CurrentHands: "AH+AS+AD+AC",
				PassedCards:  "0|",
			},
			pid:     "ABE",
			card:    "8H",
			wantErr: ErrPassingNotAllowed,
		},
		{
			name: "not passing",
			gs: &storage.Game{
				PlayerIDs:    pids,
				CurrentHands: "AH+AS+AD+AC",
				PassedCards:  "0|8C|9C|TC|JC",
				Rules:        storage.Rules{PassCard: true},
			},
			pid:     "ABE",
			card:    "8H",
			wantErr: ErrNotPassing,
		},
		{
			name: "unavailable card",
			gs: &storage.Game{
				PlayerIDs:    pids,
				CurrentHands: "AH+AS+AD+AC",
				PassedCards:  "0|8H",
				Rules:        storage.Rules{PassCard: true},
			},
			pid:     "BOB",
			card:    "5H",
			wantErr: ErrMissingCard,
		},
		{
			name: "valid card",
			gs: &storage.Game{
				PlayerIDs:    pids,
				CurrentHands: "AH+AS|8S+8S|AD+AC|8C",
				PassedCards:  "0|8H",
				Rules:        storage.Rules{PassCard: true},
			},
			pid:  "BOB",
			card: "8S",
			want: &storage.Game{
				PlayerIDs:      pids,
				Score:          "52-",
				PassedCards:    "0|8H|8S",
				CurrentBidding: "0|",
				CurrentHands:   "AH+AS+8S|AD+AC|8C",
				CurrentTally:   "0|0|0",
				Rules:          storage.Rules{PassCard: true},
			},
			wantState: PassingState,
		},
		{
			name: "last passed card, move to bidding state",
			gs: &storage.Game{
				PlayerIDs:        pids,
				CurrentHands:     "AH+AS+AD+AC|8C",
				CurrentDealerPos: 3,
				PassedCards:      "0|8H|8S|8D",
				Rules:            storage.Rules{PassCard: true},
			},
			pid:  "DON",
			card: "8C",
			want: &storage.Game{
				PlayerIDs:        pids,
				Score:            "52-",
				CurrentDealerPos: 3,
				PassedCards:      "0|8H|8S|8D|8C",
				CurrentHands:     "AH|8D+AS|8C+8H|AD+8S|AC", // Partners get the passed cards.
				CurrentBidding:   "0|",
				CurrentTally:     "0|0|0",
				Rules:            storage.Rules{PassCard: true},
			},
			wantState: BiddingState,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			g, _, playerStore := buildGame(t, tc.gs)
			player := getPlayer(t, playerStore, tc.pid)
			card := buildCard(t, tc.card)

			gotGame, err := g.PassCard(ctx, player, card)

			if tc.wantErr != nil && tc.wantErr != err {
				t.Fatalf("incorrect error got=%v want=%v", err, tc.wantErr)
			}
			if tc.wantErr == nil && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tc.wantErr != nil {
				return
			}

			if got, want := gotGame.State(), tc.wantState; got != want {
				t.Errorf("State()=%s want=%s", got, want)
			}
			gotGameStorage := storageFromGame(gotGame.(*game))
			if diff := cmp.Diff(tc.want, gotGameStorage, ignoreDates); diff != "" {
				t.Errorf("game storage mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestPlaceBid(t *testing.T) {
	pids := []string{"ABE", "BOB", "CAL", "DON"}
	testCases := []struct {
		name      string
		gs        *storage.Game
		pid       string
		bid       string
		want      *storage.Game
		wantState GameState
		wantErr   error
	}{
		{
			name: "not bidding",
			gs: &storage.Game{
				PlayerIDs:      pids,
				CurrentBidding: "0|P|P|P|7",
				CurrentHands:   "AH+AS+AD+AC",
			},
			pid:     "ABE",
			bid:     "8",
			wantErr: ErrNotBidding,
		},
		{
			name: "unavailable bid",
			gs: &storage.Game{
				PlayerIDs:      pids,
				CurrentBidding: "0|8",
				CurrentHands:   "AH+AS+AD+AC",
			},
			pid:     "BOB",
			bid:     "7N",
			wantErr: ErrInvalidBid,
		},
		{
			name: "valid bid",
			gs: &storage.Game{
				PlayerIDs:      pids,
				CurrentBidding: "0|8",
				CurrentHands:   "AH+AS+AD+AC",
			},
			pid: "BOB",
			bid: "8N",
			want: &storage.Game{
				PlayerIDs:      pids,
				Score:          "52-",
				CurrentBidding: "0|8|8N",
				CurrentTally:   "0|0|0",
				CurrentHands:   "AH+AS+AD+AC",
				PassedCards:    "0|",
			},
			wantState: BiddingState,
		},
		{
			name: "last bid, move to call trick state",
			gs: &storage.Game{
				PlayerIDs:        pids,
				CurrentDealerPos: 3,
				CurrentBidding:   "0|8|P|P",
				CurrentHands:     "AH+AS+AD+AC",
			},
			pid: "DON",
			bid: "8",
			want: &storage.Game{
				PlayerIDs:        pids,
				Score:            "52-",
				CurrentDealerPos: 3,
				CurrentBidding:   "0|8|P|P|8",
				CurrentTally:     "0|0|0",
				CurrentHands:     "AH+AS+AD+AC",
				PassedCards:      "0|"},
			wantState: CallingState,
		},
		{
			name: "no trump bid, go straight to playing",
			gs: &storage.Game{
				PlayerIDs:      pids,
				CurrentBidding: "0|P|8N|P",
				CurrentHands:   "AH+AS+AD+AC",
			},
			pid: "DON",
			bid: "P",
			want: &storage.Game{
				PlayerIDs:      pids,
				Score:          "52-",
				CurrentBidding: "0|P|8N|P|P",
				CurrentTrick:   "1|N",
				CurrentTally:   "0|0|0",
				CurrentHands:   "AH+AS+AD+AC",
				PassedCards:    "0|",
			},
			wantState: PlayingState,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			g, _, playerStore := buildGame(t, tc.gs)
			player := getPlayer(t, playerStore, tc.pid)
			bid := buildBid(t, tc.bid)

			gotGame, err := g.PlaceBid(ctx, player, bid)

			if tc.wantErr != nil && tc.wantErr != err {
				t.Fatalf("incorrect error got=%v want=%v", err, tc.wantErr)
			}
			if tc.wantErr == nil && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tc.wantErr != nil {
				return
			}

			if got, want := gotGame.State(), tc.wantState; got != want {
				t.Errorf("State()=%s want=%s", got, want)
			}
			gotGameStorage := storageFromGame(gotGame.(*game))
			if diff := cmp.Diff(tc.want, gotGameStorage, ignoreDates); diff != "" {
				t.Errorf("game storage mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestCallTrump(t *testing.T) {
	pids := []string{"ABE", "BOB", "CAL", "DON"}
	testCases := []struct {
		name      string
		gs        *storage.Game
		pid       string
		trump     string
		want      *storage.Game
		wantState GameState
		wantErr   error
	}{
		{
			name: "bidding not complete",
			gs: &storage.Game{
				PlayerIDs:      pids,
				CurrentBidding: "0|P|P",
				CurrentHands:   "AH+AS+AD+AC",
			},
			pid:     "ABE",
			trump:   "H",
			wantErr: ErrNotCalling,
		},
		{
			name: "already playing",
			gs: &storage.Game{
				PlayerIDs:      pids,
				CurrentBidding: "0|P|P|P|7",
				CurrentTrick:   "3|H",
				CurrentHands:   "AH+AS+AD+AC",
			},
			pid:     "ABE",
			trump:   "C",
			wantErr: ErrNotCalling,
		},
		{
			name: "incorrect player calls",
			gs: &storage.Game{
				PlayerIDs:      pids,
				CurrentBidding: "0|P|P|P|7",
				CurrentHands:   "AH+AS+AD+AC",
			},
			pid:     "ABE",
			trump:   "H",
			wantErr: ErrIncorrectCaller,
		},
		{
			name: "correct player calls, game moves to playing",
			gs: &storage.Game{
				PlayerIDs:      pids,
				CurrentBidding: "0|P|P|P|7",
				CurrentHands:   "AH+AS+AD+AC",
			},
			pid:   "DON",
			trump: "S",
			want: &storage.Game{
				PlayerIDs:      pids,
				Score:          "52-",
				CurrentBidding: "0|P|P|P|7",
				CurrentTrick:   "3|S",
				CurrentTally:   "0|0|0",
				CurrentHands:   "AH+AS+AD+AC",
				PassedCards:    "0|",
			},
			wantState: PlayingState,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			g, _, playerStore := buildGame(t, tc.gs)
			player := getPlayer(t, playerStore, tc.pid)
			trump := buildSuit(t, tc.trump)

			gotGame, err := g.CallTrump(ctx, player, trump)

			if tc.wantErr != nil && tc.wantErr != err {
				t.Fatalf("incorrect error got=%v want=%v", err, tc.wantErr)
			}
			if tc.wantErr == nil && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tc.wantErr != nil {
				return
			}

			if got, want := gotGame.State(), tc.wantState; got != want {
				t.Errorf("State()=%s want=%s", got, want)
			}
			gotGameStorage := storageFromGame(gotGame.(*game))
			if diff := cmp.Diff(tc.want, gotGameStorage, ignoreDates); diff != "" {
				t.Errorf("game storage mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestPlayCard(t *testing.T) {
	pids := []string{"ABE", "BOB", "CAL", "DON"}
	// hands := "AH|KH|7D+AS|KS+AC|KC+AD|KD"
	testCases := []struct {
		name          string
		gs            *storage.Game
		pid           string
		card          string
		want          *storage.Game
		wantState     GameState
		wantEmptyHand bool
		wantErr       error
	}{
		{
			name: "bidding not complete",
			gs: &storage.Game{
				PlayerIDs:      pids,
				CurrentHands:   "AH|KH|7D+AS|KS+AC|KC+AD|KD",
				CurrentBidding: "0|P|P|P",
			},
			pid:     "ABE",
			card:    "AH",
			wantErr: ErrNotPlaying,
		},
		{
			name: "calling not complete",
			gs: &storage.Game{
				PlayerIDs:      pids,
				CurrentHands:   "AH|KH|7D+AS|KS+AC|KC+AD|KD",
				CurrentBidding: "0|P|P|P|7",
			},
			pid:     "ABE",
			card:    "AH",
			wantErr: ErrNotPlaying,
		},
		{
			name: "invalid card",
			gs: &storage.Game{
				PlayerIDs:      pids,
				CurrentHands:   "AH|KH|7D+AS|KS+AC|KC+AD|KD",
				CurrentBidding: "0|P|P|P|7",
				CurrentTrick:   "3|H",
			},
			pid:     "DON",
			card:    "5H",
			wantErr: ErrMissingCard,
		},
		{
			name: "out of order",
			gs: &storage.Game{
				PlayerIDs:      pids,
				CurrentHands:   "AH|KH|7D+AS|KS+AC|KC+AD|KD",
				CurrentBidding: "0|P|P|P|7",
				CurrentTrick:   "3|H",
			},
			pid:     "ABE",
			card:    "AH",
			wantErr: ErrIncorrectPlayOrder,
		},
		{
			name: "player must follow suit",
			gs: &storage.Game{
				PlayerIDs:      pids,
				CurrentHands:   "AH|KH|7D+AS|KS+AC|KC+KD",
				CurrentBidding: "0|P|P|P|7",
				CurrentTrick:   "3|H|AD",
			},
			pid:     "ABE",
			card:    "AH",
			wantErr: ErrNotFollowingSuit,
		},
		{
			name: "player may sluff",
			gs: &storage.Game{
				PlayerIDs:      pids,
				CurrentHands:   "AH|KH+AS|KS+AC|KC+KD",
				CurrentBidding: "0|P|P|P|7",
				CurrentTrick:   "3|H|AD|7D",
			},
			pid:  "BOB",
			card: "KS",
			want: &storage.Game{
				PlayerIDs:      pids,
				Score:          "52-",
				CurrentHands:   "AH|KH+AS+AC|KC+KD", // KS played
				CurrentBidding: "0|P|P|P|7",
				CurrentTrick:   "3|H|AD|7D|KS",
				CurrentTally:   "0|0|0",
				PassedCards:    "0|",
			},
			wantState: PlayingState,
		},
		{
			name: "last card transitions to next trick",
			gs: &storage.Game{
				PlayerIDs:      pids,
				CurrentHands:   "AH|KH+AS+AC|KC+KD",
				CurrentBidding: "0|P|P|P|7",
				CurrentTrick:   "3|H|AD|7D|KS",
			},
			pid:  "CAL",
			card: "KC",
			want: &storage.Game{
				PlayerIDs:      pids,
				Score:          "52-",
				CurrentHands:   "AH|KH+AS+AC+KD", // KC played
				CurrentBidding: "0|P|P|P|7",
				CurrentTrick:   "3|H",   // Lead-off position (3) won last trick; hearts still trump.
				CurrentTally:   "1|0|1", // One card played; team 1/3 got the point.
				LastTrick:      "3|H|AD|7D|KS|KC",
				PassedCards:    "0|",
			},
			wantState: PlayingState,
		},
		{
			name: "last card transitions to dealing",
			gs: &storage.Game{
				PlayerIDs:        pids,
				CurrentHands:     "++AC+",
				CurrentDealerPos: 3,
				CurrentBidding:   "0|P|P|P|7",
				CurrentTrick:     "3|H|AD|AH|AS",
				CurrentTally:     "7|9|0",
			},
			pid:  "CAL",
			card: "AC",
			want: &storage.Game{
				PlayerIDs:        pids,
				CurrentHands:     "", // New shuffle
				CurrentDealerPos: 3,
				CurrentBidding:   "0|P|P|P|7",
				CurrentTrick:     "", // New hand.
				CurrentTally:     "8|10|0",
				Score:            "52-10|-7**1|0|missed 7 bid", // Score added from tally.
				LastTrick:        "3|H|AD|AH|AS|AC",
				PassedCards:      "0|",
			},
			wantState:     DealingState,
			wantEmptyHand: true,
		},
		{
			name: "last card transitions to dealing if passing card rule",
			gs: &storage.Game{
				PlayerIDs:        pids,
				CurrentHands:     "++AC+",
				CurrentDealerPos: 3,
				CurrentBidding:   "0|P|P|P|7",
				CurrentTrick:     "3|H|AD|AH|AS",
				CurrentTally:     "7|9|0",
				PassedCards:      "0|7C|8C|9C|TC",
				Rules:            storage.Rules{PassCard: true},
			},
			pid:  "CAL",
			card: "AC",
			want: &storage.Game{
				PlayerIDs:        pids,
				CurrentHands:     "", // New shuffle
				CurrentDealerPos: 3,
				CurrentBidding:   "0|P|P|P|7",
				CurrentTrick:     "", // New hand.
				CurrentTally:     "8|10|0",
				Score:            "52-10|-7**1|0|missed 7 bid", // Score added from tally.
				LastTrick:        "3|H|AD|AH|AS|AC",
				PassedCards:      "0|7C|8C|9C|TC",
				Rules:            storage.Rules{PassCard: true},
			},
			wantState:     DealingState,
			wantEmptyHand: true,
		},
		{
			name: "last card transitions to game completion",
			gs: &storage.Game{
				PlayerIDs:        pids,
				CurrentHands:     "++AC+",
				CurrentDealerPos: 3,
				CurrentBidding:   "0|P|P|7|P",
				CurrentTrick:     "3|H|AD|AH|AS",
				CurrentTally:     "7|9|0",
				Score:            "52-50|0", // Score is close to completion.
			},
			pid:  "CAL",
			card: "AC",
			want: &storage.Game{
				Complete:         true,
				PlayerIDs:        pids,
				CurrentHands:     "+++",                             // Hands are empty.
				CurrentDealerPos: 3,                                 // Dealer position does not update.
				CurrentBidding:   "0|P|P|7|P",                       // Bidding does not clear.
				CurrentTrick:     "",                                // Trick resets.
				CurrentTally:     "8|10|0",                          // Tally does not clear.
				Score:            "52||0-50|0||60|0**0|1|bid out 7", // Score added from tally.
				LastTrick:        "3|H|AD|AH|AS|AC",
				PassedCards:      "0|",
			},
			wantState: CompletedState,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			g, _, playerStore := buildGame(t, tc.gs)
			player := getPlayer(t, playerStore, tc.pid)
			card := buildCard(t, tc.card)

			gotGame, err := g.PlayCard(ctx, player, card)

			if tc.wantErr != nil && tc.wantErr != err {
				t.Fatalf("incorrect error got=%v want=%v", err, tc.wantErr)
			}
			if tc.wantErr == nil && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tc.wantErr != nil {
				return
			}

			if got, want := gotGame.State(), tc.wantState; got != want {
				t.Errorf("State()=%s want=%s", got, want)
			}
			gotGameStorage := storageFromGame(gotGame.(*game))
			opts := []cmp.Option{ignoreDates}
			if tc.wantEmptyHand {
				opts = append(opts, cmpopts.IgnoreFields(storage.Game{}, "CurrentHands"))

				for i := 0; i < 4; i++ {
					hand, err := gotGame.(*game).currentHands.Hand(i)
					if err != nil {
						t.Fatal(err)
					}
					if len(hand.Cards()) != 0 {
						t.Errorf("hand %d is not 0 cards", i)
					}
				}
			}
			if diff := cmp.Diff(tc.want, gotGameStorage, opts...); diff != "" {
				t.Errorf("game storage mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func buildPlayer(t *testing.T, playerStore storage.PlayerStore, id string) Player {
	t.Helper()
	p, err := NewPlayer(context.Background(), playerStore, id, id+" NAME")
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func getPlayer(t *testing.T, playerStore storage.PlayerStore, id string) Player {
	t.Helper()
	ctx := context.Background()
	p, err := GetPlayer(ctx, playerStore, id)
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func buildGame(t *testing.T, gs *storage.Game) (Game, storage.GameStore, storage.PlayerStore) {
	t.Helper()
	ctx := context.Background()
	id := "ABC123"
	gameStore := storage.NewFakeGameStore(map[string]*storage.Game{id: gs})
	playerStore := storage.NewFakePlayerStore("ABE", "BOB", "CAL", "DON", "NOTINGAME")
	g, err := GetGame(ctx, gameStore, playerStore, id)
	if err != nil {
		t.Fatal(err)
	}
	return g, gameStore, playerStore
}
