#!/bin/bash

# K6 Load Testing CI/CD Script
# This script is designed to run in CI/CD pipelines

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
BASE_URL="${BASE_URL:-http://localhost:8080}"
TEST_TYPE="${TEST_TYPE:-complete}"
RESULTS_DIR="results"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)

echo -e "${BLUE}=================================================${NC}"
echo -e "${BLUE}K6 Load Testing CI/CD Runner${NC}"
echo -e "${BLUE}=================================================${NC}"
echo -e "Base URL: ${GREEN}${BASE_URL}${NC}"
echo -e "Test Type: ${GREEN}${TEST_TYPE}${NC}"
echo -e "Timestamp: ${GREEN}${TIMESTAMP}${NC}"
echo ""

# Check if k6 is installed
if ! command -v k6 &> /dev/null; then
    echo -e "${RED}Error: k6 is not installed${NC}"
    echo -e "${YELLOW}Install k6: https://k6.io/docs/getting-started/installation/${NC}"
    exit 1
fi

echo -e "${GREEN}k6 version: $(k6 version)${NC}"
echo ""

# Check if service is available
echo -e "${YELLOW}Checking if service is available...${NC}"
max_retries=5
retry_count=0

while [ $retry_count -lt $max_retries ]; do
    if curl -s -f "${BASE_URL}/api/v1/orders" > /dev/null 2>&1; then
        echo -e "${GREEN}Service is available!${NC}"
        break
    else
        retry_count=$((retry_count + 1))
        if [ $retry_count -lt $max_retries ]; then
            echo -e "${YELLOW}Service not available yet, retrying in 5 seconds... (${retry_count}/${max_retries})${NC}"
            sleep 5
        else
            echo -e "${RED}Service is not available after ${max_retries} retries${NC}"
            exit 1
        fi
    fi
done

# Create results directory
mkdir -p "${RESULTS_DIR}/ci-${TIMESTAMP}"

# Select test file based on TEST_TYPE
case "${TEST_TYPE}" in
    quick)
        TEST_FILE="complete-flow.js"
        EXTRA_ARGS="--vus 1 --iterations 1"
        ;;
    create)
        TEST_FILE="create-orders.js"
        EXTRA_ARGS=""
        ;;
    complete)
        TEST_FILE="complete-flow.js"
        EXTRA_ARGS=""
        ;;
    spike)
        TEST_FILE="spike-test.js"
        EXTRA_ARGS=""
        ;;
    soak)
        TEST_FILE="soak-test.js"
        EXTRA_ARGS=""
        ;;
    *)
        echo -e "${RED}Error: Unknown test type '${TEST_TYPE}'${NC}"
        echo -e "${YELLOW}Valid types: quick, create, complete, spike, soak${NC}"
        exit 1
        ;;
esac

echo -e "${BLUE}Running ${TEST_TYPE} test...${NC}"
echo ""

# Run the test
if k6 run \
    -e BASE_URL="${BASE_URL}" \
    --out json="${RESULTS_DIR}/ci-${TIMESTAMP}/results.json" \
    --summary-export="${RESULTS_DIR}/ci-${TIMESTAMP}/summary.json" \
    ${EXTRA_ARGS} \
    "${TEST_FILE}"; then

    echo ""
    echo -e "${GREEN}=================================================${NC}"
    echo -e "${GREEN}Test completed successfully!${NC}"
    echo -e "${GREEN}=================================================${NC}"

    # Display summary if available
    if [ -f "${RESULTS_DIR}/ci-${TIMESTAMP}/summary.json" ]; then
        echo -e "\n${BLUE}Test Summary:${NC}"
        cat "${RESULTS_DIR}/ci-${TIMESTAMP}/summary.json" | python3 -m json.tool || cat "${RESULTS_DIR}/ci-${TIMESTAMP}/summary.json"
    fi

    echo -e "\n${YELLOW}Results saved to: ${RESULTS_DIR}/ci-${TIMESTAMP}/${NC}"
    exit 0
else
    echo ""
    echo -e "${RED}=================================================${NC}"
    echo -e "${RED}Test failed!${NC}"
    echo -e "${RED}=================================================${NC}"
    exit 1
fi
