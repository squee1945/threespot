package web

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/squee1945/threespot/server/pkg/game"
)

type UserError struct {
	Error string
}

func (s *Server) lookupPlayer(ctx context.Context, w http.ResponseWriter, r *http.Request) game.Player {
	playerID, err := playerID(r)
	if err != nil {
		sendServerError(w, "fetching player ID: %v", err)
		return nil
	}
	player, err := game.GetPlayer(ctx, s.playerStore, playerID)
	if err != nil {
		if err == game.ErrNotFound {
			sendUserError(w, "Player not found.")
			return nil
		}
		sendServerError(w, "looking up player: %v", err)
		return nil
	}
	return player
}

func lookupGame(w http.ResponseWriter, id string) game.Game {
	g, err := game.GetGame(id)
	if err != nil {
		if err == game.ErrNotFound {
			sendUserError(w, "Game not found.")
			return nil
		}
		sendServerError(w, "looking up game: %v", err)
		return nil
	}
	return g
}

func sendResponse(w http.ResponseWriter, doc interface{}) error {
	b, err := json.Marshal(doc)
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(b)
	return nil
}

func sendGameState(w http.ResponseWriter, id string, player game.Player) {
	game := lookupGame(w, id)
	if game == nil {
		return
	}

	if err := sendResponse(w, buildGameState(game, player)); err != nil {
		sendServerError(w, "sending response: %v", err)
	}
}

func sendUserError(w http.ResponseWriter, e string) {
	resp := UserError{Error: e}
	log.Printf("User error: %q" + e)
	if err := sendResponse(w, resp); err != nil {
		sendServerError(w, "sending response: %v", err)
	}
}

func sendServerError(w http.ResponseWriter, fmt string, args ...interface{}) {
	errorID := randString(8)
	log.Printf(fmt+"(errorID:"+errorID+")", args...)
	w.WriteHeader(500)
	w.Write([]byte("ErrorID: " + errorID))
}
