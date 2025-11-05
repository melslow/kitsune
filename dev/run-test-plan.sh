#!/bin/bash
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Usage function
usage() {
    echo "Usage: $0 <test-plan.json> [options]"
    echo ""
    echo "Options:"
    echo "  --skip-startup    Skip docker-compose startup (assumes services are running)"
    echo "  --wait-time <sec> Wait time after workflow trigger (default: 60)"
    echo "  --no-logs         Don't show logs after test"
    echo "  --help            Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0 test-plan-yum-upgrade-success.json"
    echo "  $0 test-plan-yum-upgrade-rollback.json --wait-time 120"
    echo "  $0 test-plan-orchestrator.json --skip-startup"
    exit 1
}

# Check if help is requested
if [ "$1" == "--help" ] || [ "$1" == "-h" ]; then
    usage
fi

# Check arguments
if [ $# -lt 1 ]; then
    echo -e "${RED}Error: Missing test plan file${NC}"
    echo ""
    usage
fi

TEST_PLAN_FILE="$1"
shift

# Parse optional arguments
SKIP_STARTUP=false
WAIT_TIME=60
SHOW_LOGS=true

while [[ $# -gt 0 ]]; do
    case $1 in
        --skip-startup)
            SKIP_STARTUP=true
            shift
            ;;
        --wait-time)
            WAIT_TIME="$2"
            shift 2
            ;;
        --no-logs)
            SHOW_LOGS=false
            shift
            ;;
        *)
            echo -e "${RED}Unknown option: $1${NC}"
            usage
            ;;
    esac
done

# Validate test plan file exists
if [ ! -f "$TEST_PLAN_FILE" ]; then
    echo -e "${RED}Error: Test plan file not found: $TEST_PLAN_FILE${NC}"
    exit 1
fi

# Validate JSON
if ! jq empty "$TEST_PLAN_FILE" 2>/dev/null; then
    echo -e "${RED}Error: Invalid JSON in test plan file${NC}"
    exit 1
fi

echo -e "${BLUE}==========================================${NC}"
echo -e "${BLUE}Kitsune Test Plan Orchestrator${NC}"
echo -e "${BLUE}==========================================${NC}"
echo ""
echo -e "${GREEN}Test Plan:${NC} $TEST_PLAN_FILE"
echo ""

# Extract test plan details
SERVERS=$(jq -r '.servers | join(", ")' "$TEST_PLAN_FILE")
STEP_COUNT=$(jq '.steps | length' "$TEST_PLAN_FILE")
STRATEGY=$(jq -r '.rolloutStrategy.type' "$TEST_PLAN_FILE")

echo -e "${YELLOW}Configuration:${NC}"
echo -e "  Servers: $SERVERS"
echo -e "  Steps: $STEP_COUNT"
echo -e "  Strategy: $STRATEGY"
echo ""

# Start infrastructure if needed
if [ "$SKIP_STARTUP" = false ]; then
    echo -e "${BLUE}Starting infrastructure...${NC}"
    docker-compose up -d
    
    echo ""
    echo -e "${YELLOW}Waiting for services to start (10 seconds)...${NC}"
    sleep 10
else
    echo -e "${YELLOW}Skipping startup (using existing services)${NC}"
fi

echo ""
echo -e "${BLUE}Checking service health...${NC}"
echo "----------------------------"

# Check Temporal
echo -n "Temporal Server: "
if docker exec kitsune-temporal temporal operator cluster health 2>&1 | grep -q "SERVING"; then
    echo -e "${GREEN}✓ Ready${NC}"
else
    echo -e "${RED}✗ Not ready${NC}"
fi

# Check PostgreSQL
echo -n "PostgreSQL: "
if docker exec kitsune-postgres pg_isready -U temporal 2>&1 | grep -q "accepting connections"; then
    echo -e "${GREEN}✓ Ready${NC}"
else
    echo -e "${RED}✗ Not ready${NC}"
fi

# Check Orchestrator
echo -n "Orchestrator Worker: "
if docker logs kitsune-orchestrator 2>&1 | grep -q "orchestrator worker started"; then
    echo -e "${GREEN}✓ Started${NC}"
else
    echo -e "${RED}✗ Not started${NC}"
fi

# Check Local Workers (dynamically from test plan)
for SERVER in $(jq -r '.servers[]' "$TEST_PLAN_FILE"); do
    echo -n "$SERVER Worker: "
    if docker logs "kitsune-mock-$SERVER" 2>&1 | grep -q "Local worker started"; then
        echo -e "${GREEN}✓ Started${NC}"
    else
        echo -e "${RED}✗ Not started${NC}"
    fi
done

echo ""
echo -e "${BLUE}Triggering orchestration workflow...${NC}"
echo "----------------------------"

# Generate unique workflow ID
TIMESTAMP=$(date +%Y-%m-%d_%H-%M-%S)
TEST_NAME=$(basename "$TEST_PLAN_FILE" .json)
WORKFLOW_ID="$TEST_NAME-$TIMESTAMP"

# Trigger the workflow
docker exec kitsune-temporal temporal workflow start \
  --task-queue execution-orchestrator \
  --type OrchestrationWorkflow \
  --input "$(cat "$TEST_PLAN_FILE" | jq -c)" \
  --workflow-id "$WORKFLOW_ID"

echo -e "${GREEN}Workflow ID:${NC} $WORKFLOW_ID"

echo ""
echo -e "${YELLOW}Waiting for execution to complete ($WAIT_TIME seconds)...${NC}"
sleep "$WAIT_TIME"

echo ""
echo -e "${BLUE}Checking workflow status...${NC}"
echo "----------------------------"

# Get workflow status
WORKFLOW_STATUS=$(docker exec kitsune-temporal temporal workflow describe --workflow-id "$WORKFLOW_ID" 2>&1)

if echo "$WORKFLOW_STATUS" | grep -q "Status.*Completed"; then
    echo -e "${GREEN}✓ Workflow Completed${NC}"
elif echo "$WORKFLOW_STATUS" | grep -q "Status.*Failed"; then
    echo -e "${RED}✗ Workflow Failed${NC}"
elif echo "$WORKFLOW_STATUS" | grep -q "Status.*Running"; then
    echo -e "${YELLOW}⚠ Workflow Still Running${NC}"
else
    echo -e "${YELLOW}⚠ Unknown Status${NC}"
fi

echo ""
echo "$WORKFLOW_STATUS" | grep -A 10 "Status" || echo "Could not get workflow details"

# Show logs if requested
if [ "$SHOW_LOGS" = true ]; then
    echo ""
    echo -e "${BLUE}Recent Orchestrator Logs:${NC}"
    echo "----------------------------"
    docker logs --tail 50 kitsune-orchestrator 2>&1 | tail -20
    
    echo ""
    echo -e "${BLUE}Recent Worker Logs:${NC}"
    echo "----------------------------"
    for SERVER in $(jq -r '.servers[]' "$TEST_PLAN_FILE"); do
        echo -e "${YELLOW}$SERVER:${NC}"
        docker logs --tail 30 "kitsune-mock-$SERVER" 2>&1 | tail -10
        echo ""
    done
fi

echo ""
echo -e "${BLUE}==========================================${NC}"
echo -e "${GREEN}Test Complete!${NC}"
echo -e "${BLUE}==========================================${NC}"
echo ""
echo -e "${YELLOW}Next steps:${NC}"
echo "  • View Temporal UI:     http://localhost:8080"
echo "  • View workflow:        http://localhost:8080/namespaces/default/workflows/$WORKFLOW_ID"
echo ""
echo -e "${YELLOW}View full logs:${NC}"
echo "  docker logs kitsune-orchestrator"
for SERVER in $(jq -r '.servers[]' "$TEST_PLAN_FILE"); do
    echo "  docker logs kitsune-mock-$SERVER"
done
echo ""
echo -e "${YELLOW}Stop services:${NC}"
echo "  docker-compose down"
echo ""
