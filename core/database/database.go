package database

import (
	"github.com/byvko-dev/am-core/helpers/env"
	"github.com/byvko-dev/am-core/logs"
	"github.com/byvko-dev/am-core/mongodb/driver"
)

func init() {
	// Initialize the mongodb connection
	mongoUri := env.MustGet("MONGO_URI")[0].(string)
	databaseName := env.MustGet("MONGO_DATABASE")[0].(string)
	err := driver.InitGlobalConnetion(mongoUri, databaseName)
	if err != nil {
		panic(err)
	}
	logs.Info("MongoDB connection initialized")
}
