package util

import (
	"net/http"
	"time"
)

var (
	playerCookie = "pid"
	cookieTTL    = 365 * 24 * time.Hour
)

func PlayerID(r *http.Request) (string, error) {
	cookie, err := r.Cookie(playerCookie)
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}

func SetPlayerID(w http.ResponseWriter) string {
	pid := RandString(8) // TODO: add a secret hash so that people can't mess with this (Kaiser cheater!)
	cookie := http.Cookie{
		Name:    playerCookie,
		Value:   pid,
		Expires: time.Now().Add(cookieTTL),
	}
	http.SetCookie(w, &cookie)
	return pid
}

func ClearPlayerID(w http.ResponseWriter) {
	cookie := http.Cookie{
		Name:    playerCookie,
		Value:   "",
		Expires: time.Now(),
	}
	http.SetCookie(w, &cookie)
}
