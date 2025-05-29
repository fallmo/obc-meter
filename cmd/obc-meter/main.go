package main

import (
	"github.com/fallmo/obc-meter/cmd/obc-meter/api"
	"github.com/fallmo/obc-meter/cmd/obc-meter/utils"
)

func main() {
	utils.StartupTasks()
	// k8s.StartMeteringObjectBuckets()
	api.StartServer()
}
