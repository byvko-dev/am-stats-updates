package main

import (
	"log"

	"github.com/GoogleCloudPlatform/functions-framework-go/funcframework"
	_ "github.com/byvko-dev/am-cloud-functions/stats/save-snapshots"
	// savesnapshots "github.com/byvko-dev/am-cloud-functions/stats/cache/save-snapshots"
)

func main() {
	// Start the Functions Framework HTTP server
	if err := funcframework.Start("9093"); err != nil {
		log.Fatalf("funcframework.Start: %v", err)
	}
}
