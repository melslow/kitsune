#!/bin/bash
set -e

echo "Starting worker..."
cd ..
go run cmd/local-worker/main.go &
WORKER_PID=$!

sleep 3

echo "Reading test plan..."
STEPS=$(cat dev/test-plan.json | jq -c '.steps')

echo "Triggering execution workflow..."
docker exec kitsune-temporal temporal workflow start \
  --task-queue dev-local \
  --type ServerExecutionWorkflow \
  --input "dev-local"
  --input "$steps"
  --workflow-id test-full-$(date +%s)

sleep 10

echo ""
echo "Checking test file..."
if [ -f /tmp/kitsune-test.txt ]; then
  echo "✓ File created successfully:"
  cat /tmp/kitsune-test.txt
  rm /tmp/kitsune-test.txt
else
  echo "✗ File not found"
fi

kill $WORKER_PID

echo ""
echo "Done! Check Temporal UI at http://localhost:8080"