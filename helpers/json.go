package helpers

import (
	"encoding/json"
	"net/http"
	http2 "vodpackager/api"
)

func WriteJSONError(w http.ResponseWriter, statusCode int, msg string) {
	response := http2.HttpResponse[any]{
		Status: "error",
		Data:   nil,
		Error:  msg,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(response)
}
