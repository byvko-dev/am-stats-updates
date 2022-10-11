package updateplayers

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
	functions.HTTP("UpdateRealmPlayers", HTTPUpdateRealmPlayers)
	functions.HTTP("UpdateSomePlayers", HTTPUpdateSomePlayers)
}

// HTTPUpdateRealmPlayers updates all players on a realm
func HTTPUpdateRealmPlayers(w http.ResponseWriter, r *http.Request) {
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

	ids, err := database.GetRealmAccountIDs(request.Realm)
	if err != nil {
		helpers.ReplyError(w, "failed to get realm accounts", http.StatusInternalServerError)
		return
	}

	idsStr := make([]string, 0, len(ids))
	for _, id := range ids {
		idsStr = append(idsStr, strconv.Itoa(id))
	}

	updateErrors := updateAccounts(request.Realm, idsStr)
	if len(updateErrors) > 0 {
		badIds := make([]string, 0, len(updateErrors))
		for _, err := range updateErrors {
			badIds = append(badIds, err.AccountID)
		}
		helpers.ReplyError(w, "failed to update accounts: "+strings.Join(badIds, ","), http.StatusInternalServerError)
		return
	}

	// Send an HTTP response
	io.WriteString(w, "OK")
}

// UpdateSomePlayers updates all player ids passed in
func HTTPUpdateSomePlayers(w http.ResponseWriter, r *http.Request) {
	var request Payload
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		helpers.ReplyError(w, "failed to decode body", http.StatusBadRequest)
		return
	}
	if len(request.PlayerIDs) == 0 {
		helpers.ReplyError(w, "missing playerIDs", http.StatusBadRequest)
		return
	}

	updateErrors := updateAccounts(request.Realm, request.PlayerIDs)
	if len(updateErrors) > 0 {
		badIds := make([]string, 0, len(updateErrors))
		for _, err := range updateErrors {
			badIds = append(badIds, err.AccountID)
		}
		helpers.ReplyError(w, "failed to update accounts: "+strings.Join(badIds, ","), http.StatusInternalServerError)
		return
	}

	// Send an HTTP response
	io.WriteString(w, "OK")
}
