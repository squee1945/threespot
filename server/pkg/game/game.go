package game

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/squee1945/threespot/server/pkg/deck"
	"github.com/squee1945/threespot/server/pkg/storage"
)

type Game interface {
	ID() string
	Version() string
	State() GameState
	Players() []Player
	PlayerHand(player Player) (Hand, error)
	Score() Score
	PlayerPos(player Player) (int, error)
	DealerPos() int
	PosToPlay() (int, error)
	Tally() Tally

	CurrentBidding() BiddingRound
	CurrentTrick() Trick
	LastTrick() Trick
	AvailableBids(Player) ([]Bid, error)

	AddPlayer(ctx context.Context, player Player, pos int) (Game, error)
	PlaceBid(ctx context.Context, player Player, bid Bid) (Game, error)
	CallTrump(ctx context.Context, player Player, trump deck.Suit) (Game, error)
	PlayCard(ctx context.Context, player Player, card deck.Card) (Game, error)
}

var (
	ErrNotFound             = errors.New("Not found")
	ErrInvalidPosition      = errors.New("Invalid position")
	ErrPlayerPositionFilled = errors.New("Player position is already filled")
	ErrPlayerAlreadyAdded   = errors.New("Player is already added")
	ErrIncorrectBidOrder    = errors.New("Bidding out of order")
	ErrIncorrectCaller      = errors.New("Player cannot call trump")
	ErrIncorrectPlayOrder   = errors.New("Playing out of order")
	ErrInvalidBid           = errors.New("Invalid bid")
	ErrNotBidding           = errors.New("Not currently bidding")
	ErrNotCalling           = errors.New("Not currently calling trump")
	ErrNotPlaying           = errors.New("Not currently playing cards")
	ErrMissingCard          = errors.New("Player does not have this card")
	ErrNotFollowingSuit     = errors.New("Must follow lead suit")
)

type GameState string

func (gs GameState) String() string {
	return string(gs)
}

var (
	JoiningState   GameState = "JOINING"
	BiddingState   GameState = "BIDDING"
	CallingState   GameState = "CALLING" // Trump
	PlayingState   GameState = "PLAYING"
	CompletedState GameState = "COMPLETED"
)

type game struct {
	gameStore   storage.GameStore
	playerStore storage.PlayerStore

	id       string
	players  []Player // Position 0/2 are a team, 1/3 are a team; organizer is position 0.
	created  time.Time
	updated  time.Time
	complete bool

	score Score // The score of the game.

	currentDealerPos int          // The position of the current dealer.
	currentBidding   BiddingRound // The bids for the current hand.
	currentHands     Hands        // Cards held by each player, parallel with the PlayerIDs above.
	currentTrick     Trick        // Cards played for current trick.
	lastTrick        Trick        // Last trick played.
	currentTally     Tally        // The running tally for the current hand.
}

var _ Game = (*game)(nil) // Ensure interface is implemented.

// NewGame creates a new game, storing it in the GameStore.
func NewGame(ctx context.Context, gameStore storage.GameStore, playerStore storage.PlayerStore, id string, organizer Player) (Game, error) {
	gs, err := gameStore.Create(ctx, id, organizer.ID())
	if err != nil {
		return nil, err
	}
	g, err := gameFromStorage(ctx, gameStore, playerStore, id, gs)
	if err != nil {
		return nil, err
	}
	return g, nil
}

// GetGame fetches the game from the GameStore, returning ErrNotFound if not found.
func GetGame(ctx context.Context, gameStore storage.GameStore, playerStore storage.PlayerStore, id string) (Game, error) {
	gs, err := gameStore.Get(ctx, id)
	if err != nil {
		if err == storage.ErrNotFound {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("fetching game from storage: %v", err)
	}
	g, err := gameFromStorage(ctx, gameStore, playerStore, id, gs)
	if err != nil {
		return nil, err
	}
	return g, nil
}

func GetCurrentGames(ctx context.Context, gameStore storage.GameStore, playerStore storage.PlayerStore, playerID string, count int) ([]Game, error) {
	gss, err := gameStore.GetCurrentGames(ctx, playerID, count)
	if err != nil {
		return nil, err
	}
	var games []Game
	for _, gs := range gss {
		g, err := gameFromStorage(ctx, gameStore, playerStore, gs.Key.StringID(), gs)
		if err != nil {
			return nil, err
		}
		games = append(games, g)
	}
	return games, nil
}

func (g *game) ID() string {
	return g.id
}

func (g *game) Version() string {
	return strconv.FormatInt(g.updated.UnixNano(), 10)
}

func (g *game) State() GameState {
	if g.complete {
		return CompletedState
	}
	if g.playerCount() < 4 {
		return JoiningState
	}
	if !g.currentBidding.IsDone() {
		return BiddingState
	}
	if g.currentTrick == nil {
		return CallingState
	}
	return PlayingState
}

func (g *game) Players() []Player {
	return g.players
}

func (g *game) PlayerPos(player Player) (int, error) {
	for pos, p := range g.Players() {
		if p.ID() == player.ID() {
			return pos, nil
		}
	}
	return -1, errors.New("unknown player")
}

func (g *game) CurrentBidding() BiddingRound {
	return g.currentBidding
}

func (g *game) CurrentTrick() Trick {
	return g.currentTrick
}

func (g *game) LastTrick() Trick {
	return g.lastTrick
}

func (g *game) DealerPos() int {
	return g.currentDealerPos
}

func (g *game) PosToPlay() (int, error) {
	switch g.State() {
	case BiddingState:
		return g.currentBidding.CurrentTurnPos()
	case CallingState:
		_, pos, err := g.currentBidding.WinningBidAndPos()
		if err != nil {
			return 0, err
		}
		return pos, nil
	case PlayingState:
		pos, err := g.currentTrick.CurrentTurnPos()
		if err != nil {
			return 0, err
		}
		return pos, nil
	}
	return -1, nil
}

func (g *game) Tally() Tally {
	return g.currentTally
}

func (g *game) Score() Score {
	return g.score
}

func (g *game) AvailableBids(player Player) ([]Bid, error) {
	if g.State() != BiddingState {
		return nil, ErrNotBidding
	}
	if g.currentBidding.IsDone() {
		return nil, ErrNotBidding
	}
	pos, err := g.PlayerPos(player)
	if err != nil {
		return nil, err
	}
	// Is it this player's turn to bid?
	currentTurnPos, err := g.currentBidding.CurrentTurnPos()
	if err != nil {
		return nil, err
	}
	if pos != currentTurnPos {
		return nil, ErrIncorrectBidOrder
	}
	return nextBidValues(g.currentBidding.Bids(), pos == g.currentDealerPos), nil
}

func (g *game) PlayerHand(player Player) (Hand, error) {
	pos, err := g.PlayerPos(player)
	if err != nil {
		return nil, err
	}
	return g.currentHands.Hand(pos)
}

func (g *game) AddPlayer(ctx context.Context, player Player, pos int) (Game, error) {
	gs, err := g.gameStore.AddPlayer(ctx, g.id, player.ID(), pos)
	if err != nil {
		if err == storage.ErrPlayerPositionFilled {
			return nil, ErrPlayerPositionFilled
		}
		if err == storage.ErrPlayerAlreadyAdded {
			return nil, ErrPlayerAlreadyAdded
		}
		return nil, fmt.Errorf("adding player: %v", err)
	}
	newG, err := gameFromStorage(ctx, g.gameStore, g.playerStore, g.id, gs)
	if err != nil {
		return nil, err
	}

	if newG.playerCount() == 4 {
		newG.currentDealerPos = rand.Int() % 4 // Assign a random dealer.
		if err := newG.startHand(); err != nil {
			return nil, fmt.Errorf("starting hand: %v", err)
		}
		newG, err = newG.save(ctx)
		if err != nil {
			return nil, fmt.Errorf("saving game: %v", err)
		}
	}
	return newG, nil
}

func (g *game) PlaceBid(ctx context.Context, player Player, bid Bid) (Game, error) {
	// Are we in bidding state?
	if g.State() != BiddingState {
		return nil, ErrNotBidding
	}

	// Is bid in available bids?
	found := false
	available, err := g.AvailableBids(player)
	if err != nil {
		return nil, err
	}
	for _, ab := range available {
		if ab.IsEqualTo(bid) {
			found = true
			break
		}
	}
	if !found {
		return nil, ErrInvalidBid
	}

	pos, err := g.PlayerPos(player)
	if err != nil {
		return nil, err
	}
	if err := g.currentBidding.placeBid(pos, bid); err != nil {
		return nil, err
	}

	// If we have all the bids, start playing.
	if g.currentBidding.IsDone() {
		bid, pos, err := g.currentBidding.WinningBidAndPos()
		if err != nil {
			return nil, err
		}
		// If no-trump, we skip past trump selection.
		if bid.IsNoTrump() {
			if err := g.startTrick(deck.NoTrump, pos); err != nil {
				return nil, err
			}
		}
	}

	return g.save(ctx)
}

func (g *game) CallTrump(ctx context.Context, player Player, trump deck.Suit) (Game, error) {
	// Are we in calling state?
	if g.State() != CallingState {
		return nil, ErrNotCalling
	}

	// Is this the right player to set trump?
	_, winningPos, err := g.currentBidding.WinningBidAndPos()
	if err != nil {
		return nil, err
	}
	pos, err := g.PlayerPos(player)
	if err != nil {
		return nil, err
	}
	if pos != winningPos {
		return nil, ErrIncorrectCaller
	}

	if err := g.startTrick(trump, winningPos); err != nil {
		return nil, err
	}

	return g.save(ctx)
}

func (g *game) PlayCard(ctx context.Context, player Player, card deck.Card) (Game, error) {
	// Are we in playing state?
	if g.State() != PlayingState {
		return nil, ErrNotPlaying
	}

	pos, err := g.PlayerPos(player)
	if err != nil {
		return nil, err
	}

	playerHand, err := g.currentHands.Hand(pos)
	if err != nil {
		return nil, err
	}

	// Does the player have the card to play?
	if !playerHand.Contains(card) {
		return nil, ErrMissingCard
	}

	// Is the card a valid card to play (i.e., does it follow suit)?
	if g.currentTrick.NumPlayed() > 0 {
		leadSuit, err := g.currentTrick.LeadSuit()
		if err != nil {
			return nil, err
		}
		if card.Suit() != leadSuit && playerHand.ContainsSuit(leadSuit, card) {
			return nil, ErrNotFollowingSuit
		}

	}

	// Remove the card from the player hand.
	if err := playerHand.removeCard(card); err != nil {
		return nil, err
	}

	// Add the card to the current trick
	if err := g.currentTrick.playCard(pos, card); err != nil {
		return nil, err
	}

	// If the last card, compute the results
	if g.currentTrick.IsDone() {

		// Who won the trick?
		winningPos, err := g.currentTrick.WinningPos()
		if err != nil {
			return nil, err
		}

		// Add the trick to the tally.
		if g.currentTally == nil {
			g.currentTally = NewTally()
		}
		if err := g.currentTally.addTrick(g.currentTrick); err != nil {
			return nil, err
		}

		// Store the last trick for player reference.
		if g.currentTrick != nil {
			lastTrick, err := NewTrickFromEncoded(g.currentTrick.Encoded())
			if err != nil {
				return nil, err
			}
			g.lastTrick = lastTrick
		}

		// If all cards are played, update the score.
		someHand, err := g.currentHands.Hand(0)
		if err != nil {
			return nil, err
		}
		if someHand.IsEmpty() {
			if g.score == nil {
				g.score = NewScore()
			}
			if err := g.score.addTally(g.currentBidding, g.currentTally); err != nil {
				return nil, err
			}

			// If the score is a winning (note: bid out, etc.), complete the game.
			// TODO: need better stuff here: consider bid-out, stealing the 5, etc.
			if g.score.CurrentScore()[0] > g.score.ToWin() || g.score.CurrentScore()[1] > g.score.ToWin() {
				g.complete = true
			} else {
				// Else deal a new hand and go to bidding round.
				if err := g.startHand(); err != nil {
					return nil, err
				}
			}
		} else {
			if err := g.startTrick(g.currentTrick.Trump(), winningPos); err != nil {
				return nil, err
			}
		}
	}

	return g.save(ctx)
}

func (g *game) startHand() error {
	deck, err := deck.NewDeck()
	if err != nil {
		return err
	}
	deck.Shuffle()

	g.currentHands, err = NewHands(deck.Deal())
	if err != nil {
		return err
	}
	g.currentDealerPos = (g.currentDealerPos + 1) % 4

	leadBidder := (g.currentDealerPos + 1) % 4 // to the left of the dealer
	biddingRound, err := NewBiddingRound(leadBidder)
	if err != nil {
		return err
	}
	g.currentBidding = biddingRound
	g.currentTrick = nil
	g.currentTally = NewTally()
	return nil
}

func (g *game) startTrick(trump deck.Suit, leadPos int) error {
	trick, err := NewTrick(trump, leadPos)
	if err != nil {
		return err
	}
	g.currentTrick = trick
	return nil
}

func (g *game) playerCount() int {
	c := 0
	for _, p := range g.players {
		if p != nil {
			c++
		}
	}
	return c
}

func (g *game) save(ctx context.Context) (*game, error) {
	g.updated = time.Now().UTC()
	gs := storageFromGame(g)
	if err := g.gameStore.Set(ctx, g.id, gs); err != nil {
		return nil, fmt.Errorf("saving game: %v", err)
	}
	return g, nil
}

func storageFromGame(g *game) *storage.Game {
	var playerIDs []string
	for _, player := range g.players {
		if player == nil {
			playerIDs = append(playerIDs, "")
			continue
		}
		playerIDs = append(playerIDs, player.ID())
	}

	var currentTrick, lastTrick string
	if g.currentTrick != nil {
		currentTrick = g.currentTrick.Encoded()
	}
	if g.lastTrick != nil {
		lastTrick = g.lastTrick.Encoded()
	}

	return &storage.Game{
		PlayerIDs:        playerIDs,
		Created:          g.created,
		Updated:          g.updated,
		Complete:         g.complete,
		Score:            g.score.Encoded(),
		CurrentDealerPos: g.currentDealerPos,
		CurrentBidding:   g.currentBidding.Encoded(),
		CurrentHands:     g.currentHands.Encoded(),
		CurrentTrick:     currentTrick,
		LastTrick:        lastTrick,
		CurrentTally:     g.currentTally.Encoded(),
	}
}

func gameFromStorage(ctx context.Context, gameStore storage.GameStore, playerStore storage.PlayerStore, id string, gs *storage.Game) (*game, error) {
	if gs == nil {
		return nil, errors.New("nil game")
	}
	// Multi-get the players
	pss, err := playerStore.GetMulti(ctx, gs.PlayerIDs)
	if err != nil {
		return nil, fmt.Errorf("fetching player info from storage: %v", err)
	}
	players := make([]Player, 4)
	for i, ps := range pss {
		if ps == nil {
			continue
		}
		player, err := playerFromStorage(playerStore, gs.PlayerIDs[i], ps)
		if err != nil {
			return nil, err
		}
		players[i] = player
	}

	bidding, err := NewBiddingRoundFromEncoded(gs.CurrentBidding)
	if err != nil {
		return nil, err
	}

	hands, err := NewHandsFromEncoded(gs.CurrentHands)
	if err != nil {
		return nil, err
	}

	var trick Trick
	if gs.CurrentTrick != "" {
		trick, err = NewTrickFromEncoded(gs.CurrentTrick)
		if err != nil {
			return nil, err
		}
	}

	var lastTrick Trick
	if gs.LastTrick != "" {
		lastTrick, err = NewTrickFromEncoded(gs.LastTrick)
		if err != nil {
			return nil, err
		}
	}

	score, err := NewScoreFromEncoded(gs.Score)
	if err != nil {
		return nil, err
	}

	tally, err := NewTallyFromEncoded(gs.CurrentTally)
	if err != nil {
		return nil, err
	}

	g := &game{
		gameStore:        gameStore,
		playerStore:      playerStore,
		id:               id,
		created:          gs.Created,
		updated:          gs.Updated,
		players:          players,
		complete:         gs.Complete,
		score:            score,
		currentDealerPos: gs.CurrentDealerPos,
		currentBidding:   bidding,
		currentHands:     hands,
		currentTrick:     trick,
		lastTrick:        lastTrick,
		currentTally:     tally,
	}
	return g, nil
}
