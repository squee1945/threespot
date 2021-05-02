package util

import (
	"math/rand"
	"net/http"
	"time"
)

var letters = []rune("BCDFGHJKLMNPQRSTVWXZ123456789")

// RandString generates a random string of n characters.
func RandString(n int) string {
	rand.Seed(time.Now().UnixNano())
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func Address(r *http.Request, id string) string {
	return "https://" + r.Host + "/" + id
}
