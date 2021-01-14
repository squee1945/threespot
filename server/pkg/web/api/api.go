package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/squee1945/threespot/server/pkg/game"
	"github.com/squee1945/threespot/server/pkg/storage"
	"github.com/squee1945/threespot/server/pkg/util"
)

func NewServer(gameStore storage.GameStore, playerStore storage.PlayerStore, cache storage.Cache) *ApiServer {
	return &ApiServer{
		playerStore: playerStore,
		gameStore:   gameStore,
		cache:       cache,
	}
}

type ApiServer struct {
	playerStore storage.PlayerStore
	gameStore   storage.GameStore
	cache       storage.Cache
}

type errorResponse struct {
	Error string
}

func (s *ApiServer) lookupPlayer(ctx context.Context, w http.ResponseWriter, r *http.Request) game.Player {
	playerID, err := util.PlayerID(r)
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

func (s *ApiServer) lookupGame(ctx context.Context, w http.ResponseWriter, id string) game.Game {
	g, err := game.GetGame(ctx, s.gameStore, s.playerStore, id)
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

func (s *ApiServer) sendGameState(ctx context.Context, w http.ResponseWriter, id string, g game.Game, player game.Player) {
	s.setGameStateVersion(ctx, id, g.Version())
	// sendGameState(ctx, w, id, g, player)
	// }

	// func sendGameState(ctx context.Context, w http.ResponseWriter, id string, g game.Game, player game.Player) {
	state, err := BuildGameState(g, player)
	if err != nil {
		sendServerError(w, "building state: %v", err)
		return
	}
	log.Printf("Sending %#v\n", state)
	w.Header().Set("Etag", state.Version)
	if err := sendResponse(w, state); err != nil {
		sendServerError(w, "sending response: %v", err)
	}
}

func genError(exposeMsg bool, format string, args ...interface{}) errorResponse {
	errorID := "[errorID:" + util.RandString(10) + "]"
	msg := fmt.Sprintf(format+" "+errorID, args...)
	log.Printf(msg)
	resp := errorResponse{Error: errorID}
	if exposeMsg {
		resp.Error = msg
	}
	return resp
}

func sendUserError(w http.ResponseWriter, format string, args ...interface{}) {
	resp := genError(true, format, args...)
	if err := sendResponseStatus(w, resp, http.StatusBadRequest); err != nil {
		sendServerError(w, "sending response: %v", err)
	}
}

func sendServerError(w http.ResponseWriter, format string, args ...interface{}) {
	resp := genError(false, format, args...)
	if err := sendResponseStatus(w, resp, http.StatusInternalServerError); err != nil {
		log.Printf("Error response failed: %v", err)
		w.WriteHeader(500)
	}
}

func sendResponse(w http.ResponseWriter, doc interface{}) error {
	return sendResponseStatus(w, doc, http.StatusOK)
}

func sendResponseStatus(w http.ResponseWriter, doc interface{}, status int) error {
	b, err := json.Marshal(doc)
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(b)
	return nil
}

func (s *ApiServer) setGameStateVersion(ctx context.Context, id, version string) {
	if id == "" || version == "" {
		log.Printf("ID %q and version %q must not be empty string. Skipping cache.", id, version)
	}
	key := id + "-version"
	if err := s.cache.Set(ctx, key, version, 10*time.Minute); err != nil {
		log.Printf("Failed to write cache. Suppressing error: %v", err)
	}
}

func (s *ApiServer) getGameStateVersion(ctx context.Context, id string) string {
	if id == "" {
		log.Printf("ID must not be empty string. Skipping cache.")
	}
	key := id + "-version"
	version, err := s.cache.Get(ctx, key)
	if err != nil {
		if err == storage.ErrCacheMiss {
			return ""
		}
		log.Printf("Failed to read cache. Suppressing error: %v", err)
		return ""
	}
	return version
}
