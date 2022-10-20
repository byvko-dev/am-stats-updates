package workers

import (
	"github.com/byvko-dev/am-core/helpers/env"
	"github.com/byvko-dev/am-stats-updates/core/messaging"
)

func Start(cancel chan int) {
	messaging.Connect(env.MustGetString("MESSAGING_URI"))
	err := StartUpdateWorkers(cancel) // Will execute all tasks in the queue every 5 min
	if err != nil {
		panic(err)
	}
}
