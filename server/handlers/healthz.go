package handlers

import (
	"encoding/json"
	"net/http"
	"sync/atomic"
)

type serverStatus string
type serverHealthz struct {
	Status serverStatus `json:"status"`
}

const serverStatusUp serverStatus = "UP"
const serverStatusDown serverStatus = "DOWN"

func Healthz(w http.ResponseWriter, _ *http.Request) {
	sendMessage(w, http.StatusOK, serverStatusUp)
}

func Readyz(isReady *atomic.Value) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		if isReady == nil || !isReady.Load().(bool) {
			sendMessage(w, http.StatusServiceUnavailable, serverStatusDown)
			return
		}
		sendMessage(w, http.StatusOK, serverStatusUp)
	}
}

func sendMessage(w http.ResponseWriter, statusCode int, serverStatus serverStatus) {
	w.WriteHeader(statusCode)
	serverHealth := serverHealthz{
		Status: serverStatus,
	}
	b, _ := json.MarshalIndent(serverHealth, "", "  ")
	w.Write(b)
}
