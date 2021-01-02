package web

import (
	"math/rand"
	"net/http"
	"time"

	"github.com/squee1945/threespot/server/pkg/storage"
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
	pid := randString(8)
	cookie := http.Cookie{
		Name:    playerCookie,
		Value:   pid,
		Expires: time.Now().Add(cookieTTL),
	}
	http.SetCookie(w, &cookie)
	return pid
}

var letters = []rune("BCDFGHJKLMNPQRSTVWXZ123456789")

func randString(n int) string {
	rand.Seed(time.Now().UnixNano())
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
