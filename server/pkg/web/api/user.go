package web

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
)

// TODO: update this to be an API (JSON docs)

type SetNameRequest struct {
	ID   string
	Name string
}

type SetNameResponse struct{}

func (s *ApiServer) SetName(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	player := s.lookupPlayer(ctx, w, r)
	if player == nil {
		return
	}

	var req SetNameRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		sendServerError(w, "decoding request: %v", err)
		return
	}

	n := strings.TrimSpace(req.Name)
	if n == "" {
		sendUserError(w, "Name is required.")
		return
	}

	_, err := player.SetName(ctx, n)
	if err != nil {
		sendServerError(w, "setting player name: %v", err)
		return
	}

	if err := sendResponse(w, SetNameResponse{}); err != nil {
		sendServerError(w, "sending response: %v", err)
	}
}
