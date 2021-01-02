package web

import (
	"net/http"
	"time"

	"github.com/squee1945/threespot/server/pkg/storage"
	"github.com/squee1945/threespot/server/pkg/util"
)

var (
	playerCookie = "pid"
	cookieTTL    = 365 * 24 * time.Hour
)

type Server struct {
	PlayerStore storage.PlayerStore
}

func playerID(r *http.Request) (string, error) {
	cookie, err := r.Cookie(playerCookie)
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}

func setPlayerID(w http.ResponseWriter) string {
	pid := util.RandString(8) // TODO: add a secret hash so that people can't mess with this (Kaiser cheater!)
	cookie := http.Cookie{
		Name:    playerCookie,
		Value:   pid,
		Expires: time.Now().Add(cookieTTL),
	}
	http.SetCookie(w, &cookie)
	return pid
}
