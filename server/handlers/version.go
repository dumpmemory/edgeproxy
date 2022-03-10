package handlers

import (
	"edgeproxy/version"
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"net/http"
)

func VersionHandler(w http.ResponseWriter, r *http.Request) {
	body, err := json.Marshal(version.GetVersion())
	if err != nil {
		log.Errorf("Could not encode info data: %v", err)
		http.Error(w, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(body)
}
