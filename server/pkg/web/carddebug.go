package web

import (
	"net/http"
)

func (s *Server) CardDebug(w http.ResponseWriter, r *http.Request) {
	// ctx := appengine.NewContext(r)
	s.render("card-debug.html", w, nil)
}
