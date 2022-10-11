package main

import (
	"log"
	"os"

	"github.com/GoogleCloudPlatform/functions-framework-go/funcframework"
	_ "github.com/byvko-dev/am-cloud-functions/stats/save-snapshots"
	_ "github.com/byvko-dev/am-cloud-functions/stats/update-accounts"
	// savesnapshots "github.com/byvko-dev/am-cloud-functions/stats/cache/save-snapshots"
)

func init() {
	os.Setenv("FUNCTION_TARGET", "UpdateSomePlayers")
}

func main() {

	// Start the Functions Framework HTTP server
	if err := funcframework.Start("9093"); err != nil {
		log.Fatalf("funcframework.Start: %v", err)
	}
}
