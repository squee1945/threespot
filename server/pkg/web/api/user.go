package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/squee1945/threespot/server/pkg/game"
	"github.com/squee1945/threespot/server/pkg/util"
	"google.golang.org/appengine"
)

type UpdateUserRequest struct {
	Name string
}

type UpdateUserResponse struct {
	Name string
}

func (s *ApiServer) UpdateUser(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	if r.Method != "POST" {
		sendUserError(w, "Invalid method")
		return
	}

	playerID, err := util.PlayerID(r)
	if err != nil {
		sendServerError(w, "looking up player ID in cookie: %v", err)
		return
	}

	var req UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendServerError(w, "decoding request: %v", err)
		return
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		sendUserError(w, "Name is required.")
		return
	}

	player, err := game.GetPlayer(ctx, s.playerStore, playerID)
	if err != nil {
		if err == game.ErrNotFound {
			player, err = game.NewPlayer(ctx, s.playerStore, playerID, name)
			if err != nil {
				sendServerError(w, "creating player: %v", err)
				return
			}
		} else {
			sendServerError(w, "searching for player: %v", err)
			return
		}
	}

	player, err = player.SetName(ctx, name)
	if err != nil {
		sendServerError(w, "setting player name: %v", err)
		return
	}

	if err := sendResponse(w, UpdateUserResponse{Name: player.Name()}); err != nil {
		sendServerError(w, "sending response: %v", err)
	}
}
