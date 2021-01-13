package web

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"

	"github.com/squee1945/threespot/server/pkg/storage"
	"github.com/squee1945/threespot/server/pkg/util"
)

const (
	templatesDir = "./pkg/web/templates/"
)

func NewServer(gameStore storage.GameStore, playerStore storage.PlayerStore) (*Server, error) {
	tmpl, err := template.ParseGlob(templatesDir + "*.tmpl")
	if err != nil {
		return nil, err
	}

	pages, err := filepath.Glob(templatesDir + "*.html")
	if err != nil {
		return nil, err
	}

	pageMap := make(map[string]*template.Template, len(pages))
	for _, p := range pages {
		overlay, err := template.Must(tmpl.Clone()).ParseFiles(p)
		if err != nil {
			return nil, err
		}
		pageMap[filepath.Base(p)] = overlay
	}
	return &Server{gameStore: gameStore, playerStore: playerStore, tmpl: tmpl, pageMap: pageMap}, nil
}

type Server struct {
	gameStore   storage.GameStore
	playerStore storage.PlayerStore
	tmpl        *template.Template
	pageMap     map[string]*template.Template
}

func (s *Server) ClearCookie(w http.ResponseWriter, r *http.Request) {
	util.ClearPlayerID(w)
	http.Redirect(w, r, "/", 302)
}

func (s *Server) render(name string, w http.ResponseWriter, args interface{}) {
	tmpl, ok := s.pageMap[name]
	if !ok {
		sendServerError(w, "template %q not found", name)
		return
	}
	if err := tmpl.ExecuteTemplate(w, name, args); err != nil {
		sendServerError(w, err.Error())
	}
}

func genError(format string, args ...interface{}) string {
	errorID := "[errorID:" + util.RandString(10) + "]"
	msg := fmt.Sprintf(format+" "+errorID, args...)
	log.Printf(msg)
	return msg
}

func sendServerError(w http.ResponseWriter, format string, args ...interface{}) {
	msg := genError(format, args)
	w.WriteHeader(500)
	w.Write([]byte(msg))
}
