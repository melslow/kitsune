package main

import (
	"log"
	"os"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	"github.com/melslow/kitsune/pkg/activities"
	"github.com/melslow/kitsune/pkg/activities/handlers"
	"github.com/melslow/kitsune/pkg/workflows"
)

func main() {
	serverID := os.Getenv("SERVER_ID")
	if serverID == "" {
		serverID = "dev-local"
	}

	temporalAddress := os.Getenv("TEMPORAL_ADDRESS")
	if temporalAddress == "" {
		temporalAddress = "localhost:7233"
	}
	c, err := client.Dial(client.Options{
		HostPort: temporalAddress,
	})
	if err != nil {
		log.Fatalln("Unable to create Temporal client:", err)
	}
	defer c.Close()

	// Create step handler registry
	registry := activities.NewStepHandlerRegistry()
	registry.Register("echo", &handlers.EchoHandler{})
	registry.Register("script", &handlers.ScriptHandler{})
	registry.Register("sleep", &handlers.SleepHandler{})
	registry.Register("file_write", &handlers.FileWriteHandler{})

	// Create worker listening on server-specific task queue
	w := worker.New(c, serverID, worker.Options{})

	//Register Workflow
	w.RegisterWorkflow(workflows.ServerExecutionWorkflow)

	// Register activities
	stepActivities := activities.NewStepActivities(serverID, registry)
	w.RegisterActivity(stepActivities)

	log.Printf("Local worker started for server: %s with %d registered handlers", serverID, 4)

	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("Unable to start worker:", err)
	}
}
