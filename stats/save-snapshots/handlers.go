package savesnapshots

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"github.com/byvko-dev/am-cloud-functions/core/database"
	"github.com/byvko-dev/am-cloud-functions/core/helpers"
)

type Payload struct {
	Realm     string   `json:"realm"`
	PlayerIDs []string `json:"playerIDs"`
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
		helpers.ReplyError(w, "failed to decode body", http.StatusBadRequest)
		return
	}

	if request.Realm == "" {
		helpers.ReplyError(w, "missing realm", http.StatusBadRequest)
		return
	}

	// Get players on realm from database
	playerIDs, err := database.GetRealmAccountIDs(request.Realm)
	if err != nil {
		return
	}
	var ids []string
	for _, id := range playerIDs {
		ids = append(ids, strconv.Itoa(id))
	}

	updateErrors := savePlayerSnapshots(request.Realm, ids, true)
	if len(updateErrors) > 0 {
		badIds := make([]string, 0, len(updateErrors))
		for _, err := range updateErrors {
			badIds = append(badIds, err.AccountID)
		}
		helpers.ReplyError(w, "failed to save snapshots for: "+strings.Join(badIds, ","), http.StatusInternalServerError)
		return
	}

	// Send an HTTP response
	io.WriteString(w, "OK")
}

// SaveAccountSnapshots generates and saves snapshots for all player ids passed in
func HTTPSaveAccountSnapshots(w http.ResponseWriter, r *http.Request) {
	var request Payload
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		helpers.ReplyError(w, "failed to decode body", http.StatusBadRequest)
		return
	}

	if request.Realm == "" {
		helpers.ReplyError(w, "missing realm", http.StatusBadRequest)
		return
	}
	if len(request.PlayerIDs) == 0 {
		helpers.ReplyError(w, "missing playerIDs", http.StatusBadRequest)
		return
	}

	updateErrors := savePlayerSnapshots(request.Realm, request.PlayerIDs, true)
	if len(updateErrors) > 0 {
		badIds := make([]string, 0, len(updateErrors))
		for _, err := range updateErrors {
			badIds = append(badIds, err.AccountID)
		}
		helpers.ReplyError(w, "failed to save snapshots for: "+strings.Join(badIds, ","), http.StatusInternalServerError)
		return
	}

	// Send an HTTP response
	io.WriteString(w, "OK")
}
