package util

import (
	"net/http"
	"time"
)

var (
	playerCookie = "pid"
	cookieTTL    = 365 * 24 * time.Hour
)

// PlayerID returns the player ID stored on the cookie.
func PlayerID(r *http.Request) (string, error) {
	cookie, err := r.Cookie(playerCookie)
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}

// SetPlayerID sets a Set-Cookie header on the response.
func SetPlayerID(w http.ResponseWriter) string {
	pid := RandString(8) // TODO: add a secret hash so that people can't mess with this (Kaiser cheater!)
	cookie := http.Cookie{
		Name:    playerCookie,
		Value:   pid,
		Expires: time.Now().Add(cookieTTL),
		Path:    "/",
	}
	http.SetCookie(w, &cookie)
	return pid
}

// ClearPlayerID clears the player cookie. There's no way to get the same one back, so this is primarily for testing.
func ClearPlayerID(w http.ResponseWriter) {
	cookie := http.Cookie{
		Name:    playerCookie,
		Value:   "",
		Expires: time.Now().Add(-24 * time.Hour),
	}
	http.SetCookie(w, &cookie)
}
