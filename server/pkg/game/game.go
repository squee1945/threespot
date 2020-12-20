package game

type GameState, HandState string
const (
	GameActive GameState = "active"
	GameDone GameState = "done"

	HandActive HandState = "active"
	HandDone HandState = "done"
)

type G struct {
	state GameState
	players Players
	hands []Hand
}

type Players struct {
	teamA, teamB Team
}

type Team struct {
	one, two Player
	score int
}

type Player struct {
	name string
	cards []Card
}

type Hand struct {
	state HandState
	trump Trump
	plays []Play
}

type Play struct {
	card Card
}
