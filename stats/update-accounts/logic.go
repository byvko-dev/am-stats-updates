package updateplayers

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/byvko-dev/am-cloud-functions/core/database"
	"github.com/byvko-dev/am-cloud-functions/core/helpers"
	"github.com/byvko-dev/am-types/stats/v3"
	"github.com/byvko-dev/am-types/wargaming/v1/accounts"
	"github.com/byvko-dev/am-types/wargaming/v1/clans"
	wg "github.com/cufee/am-wg-proxy-next/client"
)

func updateAccounts(realm string, playerIDs []string) ([]helpers.UpdateResult, []string) {
	client := wg.NewClient(os.Getenv("WG_PROXY_HOST"), time.Second*30)
	accountData, err := client.BulkGetAccountsByID(playerIDs, realm)
	if err != nil {
		var result = make([]helpers.UpdateResult, len(playerIDs))
		for i, id := range playerIDs {
			result[i] = helpers.UpdateResult{AccountID: id, Error: fmt.Sprintf("failed to get accounts: %s", err.Error()), WillRetry: true}
		}
		return result, playerIDs
	}

	clansData, err := client.BulkGetAccountsClans(playerIDs, realm)
	if err != nil {
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
				// This should never happen but just in case
				result <- helpers.UpdateResult{AccountID: id, Error: "account not found"}
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
