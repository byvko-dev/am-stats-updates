package accounts

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/byvko-dev/am-core/logs"
	"github.com/byvko-dev/am-stats-updates/core/database"
	"github.com/byvko-dev/am-stats-updates/core/helpers"
	"github.com/byvko-dev/am-types/stats/v3"
	"github.com/byvko-dev/am-types/wargaming/v2/accounts"
	"github.com/byvko-dev/am-types/wargaming/v2/clans"
	wg "github.com/cufee/am-wg-proxy-next/client"
)

func updateAccounts(realm string, playerIDs []string) ([]helpers.UpdateResult, []string) {
	opts := wg.ClientOptons{
		Debug: true,
	}

	client := wg.NewClient(os.Getenv("WG_PROXY_HOST"), time.Second*60, opts)
	defer client.Close()
	logs.Debug("Requesting %d accounts", len(playerIDs))
	accountData, err := client.BulkGetAccountsByID(playerIDs, realm)
	if err != nil {
		logs.Error("Failed to get accounts: %s", err.Error())
		var result = make([]helpers.UpdateResult, len(playerIDs))
		for i, id := range playerIDs {
			result[i] = helpers.UpdateResult{AccountID: id, Error: fmt.Sprintf("failed to get accounts: %s", err.Error()), WillRetry: true}
		}
		return result, playerIDs
	}
	if len(accountData) == 0 {
		logs.Error("No accounts returned")
		var result = make([]helpers.UpdateResult, len(playerIDs))
		for i, id := range playerIDs {
			result[i] = helpers.UpdateResult{AccountID: id, Error: "no accounts returned", WillRetry: true}
		}
		return result, playerIDs
	}

	logs.Debug("Requesting %d account clans", len(playerIDs))
	clansData, err := client.BulkGetAccountsClans(playerIDs, realm)
	if err != nil && !strings.Contains(err.Error(), "SOURCE_NOT_AVAILABLE") { // Ignore SOURCE_NOT_AVAILABLE error, it's not critical
		logs.Error("Failed to get clans: %s", err.Error())
		var result = make([]helpers.UpdateResult, len(playerIDs))
		for i, id := range playerIDs {
			result[i] = helpers.UpdateResult{AccountID: id, Error: fmt.Sprintf("failed to get clans: %s", err.Error()), WillRetry: true}
		}
		return result, playerIDs

	}

	// Save all snapshots in goroutines
	var wg sync.WaitGroup
	var retry = make(chan string, len(accountData))
	var result = make(chan helpers.UpdateResult, len(accountData))
	for _, id := range playerIDs {
		wg.Add(1)
		account := accountData[id]
		clan := clansData[id]

		go func(account accounts.CompleteProfile, clan clans.MemberProfile, id string) {
			defer wg.Done()
			if account.AccountID == 0 || id == "" {
				result <- helpers.UpdateResult{AccountID: id, Error: "account not found"}

				// This is a rare case where an account was deleted from WG servers
				idInt, err := strconv.Atoi(id)
				if err == nil {
					database.DeleteAccount(idInt)
				}
				return
			}

			var accountInfo stats.AccountInfo
			accountInfo.AccountID = int(account.AccountID)
			accountInfo.Nickname = account.Nickname
			accountInfo.Realm = strings.ToUpper(realm)

			var clanInfo stats.AccountClan
			if clan.ClanID != 0 {
				clanInfo.ID = int(clan.ClanID)
				clanInfo.Name = clan.Clan.Name
				clanInfo.Tag = clan.Clan.Tag
				clanInfo.Role = clan.Role
				clanInfo.JoinedAt = int(clan.JoinedAt)
				accountInfo.Clan = clanInfo
			}

			err := database.UpsertAccountInfo(accountInfo)
			if err != nil {
				retry <- id
				result <- helpers.UpdateResult{AccountID: id, Error: fmt.Sprintf("failed to update account info: %s", err.Error()), WillRetry: true}
				return
			}

			result <- helpers.UpdateResult{AccountID: id, Success: true}
		}(account, clan, id)

	}
	wg.Wait()
	close(result)
	close(retry)

	var results []helpers.UpdateResult
	for r := range result {
		results = append(results, r)
	}
	var retryIDs []string
	for id := range retry {
		retryIDs = append(retryIDs, id)
	}
	return results, retryIDs
}
