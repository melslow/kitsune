package main

import (
	"log"
	"os"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	"github.com/melslow/kitsune/pkg/workflows"
)

func main() {
	temporalAddress := os.Getenv("TEMPORAL_ADDRESS")
	if temporalAddress == "" {
		temporalAddress = "localhost:7233"
	}

	// Connect to Temporal
	c, err := client.Dial(client.Options{
		HostPort: temporalAddress,
	})
	if err != nil {
		log.Fatalln("Unable to create Temporal client:", err)
	}
	defer c.Close()

	// Create worker listening on orchestrator task queue
	w := worker.New(c, "execution-orchestrator", worker.Options{})

	// Register ONLY orchestration workflow
	w.RegisterWorkflow(workflows.OrchestrationWorkflow)

	log.Printf("Central orchestrator worker started on queue: execution-orchestrator")

	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("Unable to start worker:", err)
	}
}
