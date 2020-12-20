package game

import (
	"github.com/squee1945/threespot/server/pkg/deck"
)

type GameState string
type TrickState string

const (
	GameActive GameState = "active"
	GameDone   GameState = "done"

	TrickActive TrickState = "active"
	TrickDone   TrickState = "done"
)

type Game interface{}

type game struct {
	state        GameState
	teamA, teamB Team
	tricks       []Trick
}

type players struct {
	teamA, teamB Team
}

type Team interface{}

type team struct {
	player1, player2 Player
	score            int
}

type Trick struct {
	state TrickState
	trump deck.Suit
	plays []Play
}

type Play struct {
	player Player
	card   deck.Card
}

func NewGame(teamA, teamB Team) (Game, error) {
	return nil, nil
}

func NewTeam(player1, player2 Player) (Team, error) {
	return nil, nil

}
