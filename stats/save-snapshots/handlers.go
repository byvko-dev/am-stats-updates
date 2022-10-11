package savesnapshots

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
)

type Payload struct {
	Realm     string   `json:"realm"`
	PlayerIDs []string `json:"playerIDs"`
}

func replyError(w http.ResponseWriter, message string, code ...int) {
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

func init() {
	// Register an HTTP function with the Functions Framework
	functions.HTTP("SaveRealmSnapshots", HTTPSaveRealmSnapshots)
	functions.HTTP("SaveAccountSnapshots", HTTPSaveAccountSnapshots)
}

// SaveRealmSnapshots generates and saves snapshots for all players on a realm
func HTTPSaveRealmSnapshots(w http.ResponseWriter, r *http.Request) {
	// Decode the request body into a struct.
	var request Payload
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		replyError(w, "failed to decode body", http.StatusBadRequest)
		return
	}

	if request.Realm == "" {
		replyError(w, "missing realm", http.StatusBadRequest)
		return
	}

	// Get players on realm from database
	playerIDs, err := getRealmPlayers(request.Realm)
	if err != nil {
		return
	}

	updateErrors := savePlayerSnapshots(request.Realm, playerIDs, true)
	if len(updateErrors) > 0 {
		badIds := make([]string, 0, len(updateErrors))
		for _, err := range updateErrors {
			badIds = append(badIds, err.AccountID)
		}
		replyError(w, "failed to save snapshots for: "+strings.Join(badIds, ","), http.StatusInternalServerError)
		return
	}

	// Send an HTTP response
	io.WriteString(w, "OK")
}

// SaveAccountSnapshots generates and saves snapshots for all player ids passed in
func HTTPSaveAccountSnapshots(w http.ResponseWriter, r *http.Request) {
	var request Payload
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		replyError(w, "failed to decode body", http.StatusBadRequest)
		return
	}

	if request.Realm == "" {
		replyError(w, "missing realm", http.StatusBadRequest)
		return
	}
	if len(request.PlayerIDs) == 0 {
		replyError(w, "missing playerIDs", http.StatusBadRequest)
		return
	}

	updateErrors := savePlayerSnapshots(request.Realm, request.PlayerIDs, true)
	if len(updateErrors) > 0 {
		badIds := make([]string, 0, len(updateErrors))
		for _, err := range updateErrors {
			badIds = append(badIds, err.AccountID)
		}
		replyError(w, "failed to save snapshots for: "+strings.Join(badIds, ","), http.StatusInternalServerError)
		return
	}

	// Send an HTTP response
	io.WriteString(w, "OK")
}
