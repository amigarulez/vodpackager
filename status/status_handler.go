package status

import "net/http"

func StatusHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func StatusExHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	// TODO build version
	// TODO library versions
	w.Write([]byte("OK"))
}
