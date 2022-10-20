package helpers

import (
	"encoding/json"
	"net/http"
)

func ReplyError(w http.ResponseWriter, message string, code ...int) {
	payload := make(map[string]interface{})
	payload["error"] = message

	// Send an HTTP response
	w.Header().Set("Content-Type", "application/json")
	if len(code) > 0 {
		w.WriteHeader(code[0])
	} else {
		w.WriteHeader(http.StatusInternalServerError)
	}
	json.NewEncoder(w).Encode(payload)
}
