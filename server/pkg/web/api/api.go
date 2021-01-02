package web

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/squee1945/threespot/server/pkg/game"
	"github.com/squee1945/threespot/server/pkg/util"
)

type errorResponse struct {
	Error string
}

func (s *Server) lookupPlayer(ctx context.Context, w http.ResponseWriter, r *http.Request) game.Player {
	playerID, err := playerID(r)
	if err != nil {
		sendServerError(w, "fetching player ID: %v", err)
		return nil
	}
	player, err := game.GetPlayer(ctx, s.PlayerStore, playerID)
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

func sendGameState(w http.ResponseWriter, id string, player game.Player) {
	game := lookupGame(w, id)
	if game == nil {
		return
	}

	if err := sendResponse(w, buildGameState(game, player)); err != nil {
		sendServerError(w, "sending response: %v", err)
	}
}

func genError(exposeMsg bool, fmt string, args ...interface{}) errorResponse {
	errorID := "[errorID:" + util.RandString(10) + "]"
	msg := fmt.Sprintf(fmt+" "+errorID, args...)
	log.Printf(msg)
	resp := errorResponse{Error: errorID}
	if exposeMsg {
		resp.Error = msg
	}
	return resp
}

func sendUserError(w http.ResponseWriter, fmt string, args ...interface{}) {
	resp := genError(true, fmt, args)
	if err := sendResponseStatus(w, resp, http.StatusBadRequest); err != nil {
		sendServerError(w, "sending response: %v", err)
	}
}

func sendServerError(w http.ResponseWriter, fmt string, args ...interface{}) {
	resp := genError(false, fmt, args)
	if err := sendResponseStatus(w, resp, http.StatusInternalServerError); err != nil {
		log.Printf("Error response failed: %v", err)
		w.WriteHeader(500)
	}
}

func sendResponse(w http.ResponseWriter, doc interface{}) error {
	return sendResponseStatus(w, doc, http.StatusOK)
}

func sendResponseStatus(w http.ResponseWriter, doc interface{}, int status) error {
	b, err := json.Marshal(doc)
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(b)
	return nil
}
