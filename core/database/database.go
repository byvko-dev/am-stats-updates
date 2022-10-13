package database

import (
	"time"

	"github.com/byvko-dev/am-core/helpers/env"
	"github.com/byvko-dev/am-core/logs"
	"github.com/byvko-dev/am-core/mongodb/driver"
)

func init() {
	// Initialize the mongodb connection
	mongoUri := env.MustGet("MONGO_URI")[0].(string)
	databaseName := env.MustGet("MONGO_DATABASE")[0].(string)
	err := driver.InitGlobalConnection(mongoUri, databaseName, time.Minute*1)
	if err != nil {
		panic(err)
	}
	logs.Info("MongoDB connection initialized")
}
