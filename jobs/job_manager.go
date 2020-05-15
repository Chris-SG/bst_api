package jobs

import (
	"github.com/chris-sg/bst_api/eagate/user"
	"github.com/golang/glog"
	"time"
)

func StartJobs() {
	go UpdateEaState()
}

func UpdateEaState() {
	counter := 0
	for range time.Tick(time.Hour) {
		glog.Infof("Updating ea states %d", counter)
		user.RunUpdatesOnAllEaUsers()
	}
}