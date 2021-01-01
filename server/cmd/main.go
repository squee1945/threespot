package main

import (
	"log"
	"net/http"
	"os"

	"github.com/squee1945/threespot/server/pkg/web"
)

const (
	defaultPort = "8080"
)

// http.Handle("/foo", fooHandler)

// http.HandleFunc("/bar", func(w http.ResponseWriter, r *http.Request) {
// 	fmt.Fprintf(w, "Hello, %q", html.EscapeString(r.URL.Path))
// })
func main() {
	port, ok := os.LookupEnv("PORT")
	if !ok {
		port = defaultPort
	}

	http.Handle("/new", web.NewGame)
	http.Handle("/join", web.JoinGame)
	http.Handle("/bid", web.PlaceBid)
	http.Handle("/play", web.PlayCard)
	http.Handle("/game", web.GameState)

	s := &http.Server{
		Addr:    ":" + port,
		Handler: nil,
		// ReadTimeout:    10 * time.Second,
		// WriteTimeout:   10 * time.Second,
		// MaxHeaderBytes: 1 << 20,
	}
	log.Fatal(s.ListenAndServe())
}
