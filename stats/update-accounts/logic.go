package updateplayers

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/byvko-dev/am-cloud-functions/core/database"
	"github.com/byvko-dev/am-cloud-functions/core/helpers"
	"github.com/byvko-dev/am-types/stats/v3"
	"github.com/byvko-dev/am-types/wargaming/v1/accounts"
	"github.com/byvko-dev/am-types/wargaming/v1/clans"
	wg "github.com/cufee/am-wg-proxy-next/client"
)

func updateAccounts(realm string, playerIDs []string) []helpers.FailedUpdate {
	client := wg.NewClient(os.Getenv("WG_PROXY_HOST"), time.Second*30)
	accountData, err := client.BulkGetAccountsByID(playerIDs, realm)
	if err != nil {
		return []helpers.FailedUpdate{{AccountID: "all", Err: fmt.Errorf("failed to get accounts: %w", err)}}
	}

	clansData, err := client.BulkGetAccountsClans(playerIDs, realm)
	if err != nil {
		return []helpers.FailedUpdate{{AccountID: "all", Err: fmt.Errorf("failed to get clans: %w", err)}}
	}

	// Save all snapshots in goroutines
	var wg sync.WaitGroup
	var failed = make(chan helpers.FailedUpdate, len(accountData))
	for _, id := range playerIDs {
		wg.Add(1)
		account := accountData[id]
		clan := clansData[id]

		go func(account accounts.CompleteProfile, clan clans.MemberProfile, id string) {
			defer wg.Done()
			if account.AccountID == 0 || id == "" {
				// This should never happen but just in case
				failed <- helpers.FailedUpdate{AccountID: id, Err: fmt.Errorf("account not found")}
				return
			}

			var accountInfo stats.AccountInfo
			accountInfo.AccountID = int(account.AccountID)
			accountInfo.Nickname = account.Nickname
			accountInfo.Realm = realm

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
				failed <- helpers.FailedUpdate{AccountID: id, Err: fmt.Errorf("failed to update account info: %w", err)}
				return
			}
		}(account, clan, id)

	}
	wg.Wait()
	close(failed)

	var failedUpdates []helpers.FailedUpdate
	for failedUpdate := range failed {
		failedUpdates = append(failedUpdates, failedUpdate)
	}
	return failedUpdates
}
