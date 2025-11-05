#!/bin/bash
set -e

echo "=========================================="
echo "Kitsune Orchestration Test"
echo "=========================================="
echo ""

echo "Starting infrastructure..."
docker-compose up -d

echo ""
echo "Waiting for services to start (10 seconds)..."
sleep 10

echo ""
echo "Checking service health..."
echo "----------------------------"

# Check Temporal
echo -n "Temporal Server: "
docker exec kitsune-temporal temporal operator cluster health 2>&1 | grep -q "SERVING" && echo "✓ Ready" || echo "✗ Not ready"

# Check PostgreSQL
echo -n "PostgreSQL: "
docker exec kitsune-postgres pg_isready -U temporal 2>&1 | grep -q "accepting connections" && echo "✓ Ready" || echo "✗ Not ready"

# Check Orchestrator
echo -n "Orchestrator Worker: "
docker logs kitsune-orchestrator 2>&1 | grep -q "orchestrator worker started" && echo "✓ Started" || echo "✗ Not started"

# Check Local Workers
echo -n "Server-1 Worker: "
docker logs kitsune-mock-server-1 2>&1 | grep -q "Local worker started" && echo "✓ Started" || echo "✗ Not started"

echo -n "Server-2 Worker: "
docker logs kitsune-mock-server-2 2>&1 | grep -q "Local worker started" && echo "✓ Started" || echo "✗ Not started"

echo -n "Server-3 Worker: "
docker logs kitsune-mock-server-3 2>&1 | grep -q "Local worker started" && echo "✓ Started" || echo "✗ Not started"

echo ""
echo "Triggering orchestration workflow..."
echo "----------------------------"

# Trigger the workflow
TIMESTAMP=$(date +%Y-%m-%d_%H:%M:%S)
WORKFLOW_ID="orchestration-test-$TIMESTAMP"

docker exec kitsune-temporal temporal workflow start \
  --task-queue execution-orchestrator \
  --type OrchestrationWorkflow \
  --input  "$(cat test-plan-orchestrator.json | jq)" \
  --workflow-id "$WORKFLOW_ID"

echo "Workflow ID: $WORKFLOW_ID"

echo ""
echo "Waiting for execution to complete (60 seconds)..."
sleep 60

echo ""
echo "Checking execution results..."
echo "----------------------------"

# Check results on each server
for i in 1 2 3; do
  echo "Server $i:"
  RESULT=$(docker exec kitsune-mock-server-$i cat /tmp/deployment-marker.txt 2>/dev/null)
  if [ $? -eq 0 ]; then
    echo "  ✓ $RESULT"
  else
    echo "  ✗ File not found"
  fi
done

echo ""
echo "Checking workflow status..."
echo "----------------------------"
docker exec kitsune-temporal temporal workflow describe --workflow-id "$WORKFLOW_ID" 2>&1 | grep -A 5 "Status" || echo "Could not get workflow status"

echo ""
echo "=========================================="
echo "Test Complete!"
echo "=========================================="
echo ""
echo "Next steps:"
echo "  • View Temporal UI:     http://localhost:8080"
echo "  • View workflow:        http://localhost:8080/namespaces/default/workflows/$WORKFLOW_ID"
echo ""
echo "View logs:"
echo "  docker logs kitsune-orchestrator"
echo "  docker logs kitsune-mock-server-1"
echo "  docker logs kitsune-mock-server-2"
echo "  docker logs kitsune-mock-server-3"
echo ""

read -p "Press Enter to continue..."