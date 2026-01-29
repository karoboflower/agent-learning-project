#!/bin/bash

# Go Agent API Test Script
# 测试所有API端点

set -e

BASE_URL="http://localhost:8080"
AGENT_ID=""
TASK_ID=""

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Print functions
print_test() {
    echo -e "${YELLOW}[TEST]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Test health check
test_health() {
    print_test "Testing health check..."
    response=$(curl -s ${BASE_URL}/health)
    echo "Response: $response"

    if echo "$response" | grep -q "ok"; then
        print_success "Health check passed"
    else
        print_error "Health check failed"
        exit 1
    fi
    echo ""
}

# Test create agent
test_create_agent() {
    print_test "Creating agent..."
    response=$(curl -s -X POST ${BASE_URL}/api/v1/agents \
        -H "Content-Type: application/json" \
        -d '{
            "name": "test-agent",
            "type": "general",
            "config": {
                "model": "gpt-4",
                "temperature": 0.7,
                "max_tokens": 2000
            }
        }')

    echo "Response: $response"

    AGENT_ID=$(echo "$response" | grep -o '"id":"[^"]*"' | cut -d'"' -f4)

    if [ -n "$AGENT_ID" ]; then
        print_success "Agent created with ID: $AGENT_ID"
    else
        print_error "Failed to create agent"
        exit 1
    fi
    echo ""
}

# Test list agents
test_list_agents() {
    print_test "Listing agents..."
    response=$(curl -s ${BASE_URL}/api/v1/agents)
    echo "Response: $response"

    if echo "$response" | grep -q "agents"; then
        print_success "Agents listed successfully"
    else
        print_error "Failed to list agents"
        exit 1
    fi
    echo ""
}

# Test get agent
test_get_agent() {
    print_test "Getting agent $AGENT_ID..."
    response=$(curl -s ${BASE_URL}/api/v1/agents/${AGENT_ID})
    echo "Response: $response"

    if echo "$response" | grep -q "$AGENT_ID"; then
        print_success "Agent retrieved successfully"
    else
        print_error "Failed to get agent"
        exit 1
    fi
    echo ""
}

# Test submit task
test_submit_task() {
    print_test "Submitting task..."
    response=$(curl -s -X POST ${BASE_URL}/api/v1/tasks \
        -H "Content-Type: application/json" \
        -d "{
            \"agent_id\": \"${AGENT_ID}\",
            \"type\": \"query\",
            \"input\": \"What is 2+2?\",
            \"priority\": 1
        }")

    echo "Response: $response"

    TASK_ID=$(echo "$response" | grep -o '"id":"[^"]*"' | cut -d'"' -f4)

    if [ -n "$TASK_ID" ]; then
        print_success "Task submitted with ID: $TASK_ID"
    else
        print_error "Failed to submit task"
        exit 1
    fi
    echo ""
}

# Test get task
test_get_task() {
    print_test "Getting task $TASK_ID..."
    response=$(curl -s ${BASE_URL}/api/v1/tasks/${TASK_ID})
    echo "Response: $response"

    if echo "$response" | grep -q "$TASK_ID"; then
        print_success "Task retrieved successfully"
    else
        print_error "Failed to get task"
        exit 1
    fi
    echo ""
}

# Test task stats
test_task_stats() {
    print_test "Getting task statistics..."
    response=$(curl -s ${BASE_URL}/api/v1/tasks/stats)
    echo "Response: $response"

    if echo "$response" | grep -q "pending_tasks"; then
        print_success "Task statistics retrieved successfully"
    else
        print_error "Failed to get task statistics"
        exit 1
    fi
    echo ""
}

# Wait for task completion
wait_for_task() {
    print_test "Waiting for task completion..."
    max_attempts=30
    attempt=0

    while [ $attempt -lt $max_attempts ]; do
        response=$(curl -s ${BASE_URL}/api/v1/tasks/${TASK_ID})
        status=$(echo "$response" | grep -o '"status":"[^"]*"' | cut -d'"' -f4)

        if [ "$status" = "completed" ] || [ "$status" = "failed" ]; then
            print_success "Task finished with status: $status"
            return 0
        fi

        echo "Task status: $status (attempt $((attempt+1))/$max_attempts)"
        sleep 2
        attempt=$((attempt+1))
    done

    print_error "Task did not complete in time"
    return 1
}

# Test get task result
test_get_task_result() {
    print_test "Getting task result..."
    response=$(curl -s ${BASE_URL}/api/v1/tasks/${TASK_ID}/result)
    echo "Response: $response"

    if echo "$response" | grep -q "output\|error"; then
        print_success "Task result retrieved successfully"
    else
        print_error "Failed to get task result"
    fi
    echo ""
}

# Test list tasks
test_list_tasks() {
    print_test "Listing tasks..."
    response=$(curl -s ${BASE_URL}/api/v1/tasks)
    echo "Response: $response"

    if echo "$response" | grep -q "tasks"; then
        print_success "Tasks listed successfully"
    else
        print_error "Failed to list tasks"
        exit 1
    fi
    echo ""
}

# Test delete agent
test_delete_agent() {
    print_test "Deleting agent $AGENT_ID..."
    response=$(curl -s -X DELETE ${BASE_URL}/api/v1/agents/${AGENT_ID})

    if [ -z "$response" ] || echo "$response" | grep -q "204"; then
        print_success "Agent deleted successfully"
    else
        print_error "Failed to delete agent"
        echo "Response: $response"
    fi
    echo ""
}

# Main test flow
main() {
    echo "======================================"
    echo "  Go Agent API Integration Tests"
    echo "======================================"
    echo ""

    test_health
    test_create_agent
    test_list_agents
    test_get_agent
    test_submit_task
    test_get_task
    test_task_stats
    test_list_tasks

    # Wait for task and get result (optional, may fail if no OpenAI key)
    if wait_for_task; then
        test_get_task_result
    fi

    test_delete_agent

    echo "======================================"
    echo -e "${GREEN}All tests passed!${NC}"
    echo "======================================"
}

# Check if server is running
if ! curl -s ${BASE_URL}/health > /dev/null; then
    print_error "Server is not running at ${BASE_URL}"
    echo "Please start the server first: make run"
    exit 1
fi

# Run tests
main
