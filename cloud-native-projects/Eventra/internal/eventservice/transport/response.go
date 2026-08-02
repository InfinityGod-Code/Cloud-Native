package transport

import (
	"encoding/json"
	"net/http"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

func WriteJSONResponse(w http.ResponseWriter, httpStatusCode int, resp interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatusCode)
	json.NewEncoder(w).Encode(resp)
}

func WriteErrorResponse(w http.ResponseWriter, httpStatusCode int, msg string) {
	WriteJSONResponse(w, httpStatusCode, ErrorResponse{Error: msg})
}
