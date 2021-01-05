package web

import (
	"fmt"
	"log"
	"net/http"

	"github.com/squee1945/threespot/server/pkg/storage"
	"github.com/squee1945/threespot/server/pkg/util"
)

type Server struct {
	PlayerStore storage.PlayerStore
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
