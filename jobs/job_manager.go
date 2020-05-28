package jobs

import (
	"github.com/chris-sg/bst_api/db"
	"github.com/chris-sg/bst_api/ddr"
	"github.com/chris-sg/bst_api/eagate/user"
	"github.com/chris-sg/bst_api/utilities"
	bst_models "github.com/chris-sg/bst_server_models"
	"github.com/golang/glog"
	"time"
)

func StartJobs() {
	go RunJobs()
}

func RunJobs() {
	counter := 0
	for range time.Tick(time.Hour) {
		counter++
		glog.Infof("Updating ea states %d", counter)

		// Update ea state for all assumed logged in users
		user.RunUpdatesOnAllEaUsers()

		profilesToUpdate, errs := db.GetApiDb().RetrieveUpdateableProfiles()
		if utilities.PrintErrors("failed to retrieve updatable profiles", errs) {
			continue
		}

		// Run ddr updates
		ddrUpdateCount := 0
		ddrFailedCount := 0
		for _, profile := range profilesToUpdate {
			func() {
				if !profile.DdrAutoUpdate {
					return
				}
				usernames, errs := db.GetUserDb().RetrieveUsernamesByWebId(profile.User)
				if len(usernames) == 0 {
					return
				}
				u, exists, errs := db.GetUserDb().RetrieveUserByUserId(usernames[0])
				if utilities.PrintErrors("failed to retrieve user", errs) || !exists {
					ddrFailedCount++
					return
				}
				client, err := user.CreateClientForUser(u)
				defer client.UpdateCookie()
				if !err.Equals(bst_models.ErrorOK) {
					ddrFailedCount++
					glog.Warning(err)
					return
				}
				err = ddr.UpdatePlayerProfile(u, client)
				if !err.Equals(bst_models.ErrorOK) {
					ddrFailedCount++
					glog.Warning(err)
					return
				}
				ddrUpdateCount++
			}()
		}
		glog.Infof("successfully updated %d/%d ddr profiles (%d failed)", ddrUpdateCount, len(profilesToUpdate), ddrFailedCount)
	}
}