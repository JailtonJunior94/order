.PHONY: help install check-k6 check-service test-create test-mark-paid test-complete test-spike test-soak test-all clean results view-results

# Variables
BASE_URL ?= http://localhost:8001
K6_OUT_DIR = results
TIMESTAMP := $(shell date +%Y%m%d_%H%M%S)

# Colors for output
GREEN  := $(shell tput -Txterm setaf 2)
YELLOW := $(shell tput -Txterm setaf 3)
BLUE   := $(shell tput -Txterm setaf 4)
RED    := $(shell tput -Txterm setaf 1)
RESET  := $(shell tput -Txterm sgr0)

help: ## Show this help message
	@echo '${BLUE}Available targets:${RESET}'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  ${GREEN}%-20s${RESET} %s\n", $$1, $$2}' $(MAKEFILE_LIST)

install: ## Install k6
	@echo "${YELLOW}Installing k6...${RESET}"
	@if [ "$$(uname)" = "Darwin" ]; then \
		brew install k6; \
	elif [ "$$(uname)" = "Linux" ]; then \
		echo "${YELLOW}Please visit: https://k6.io/docs/getting-started/installation/${RESET}"; \
	else \
		echo "${RED}Unsupported OS. Please install k6 manually.${RESET}"; \
		exit 1; \
	fi
	@echo "${GREEN}k6 installed successfully!${RESET}"

check-k6: ## Check if k6 is installed
	@command -v k6 >/dev/null 2>&1 || { \
		echo "${RED}k6 is not installed. Run 'make install' first.${RESET}"; \
		exit 1; \
	}
	@echo "${GREEN}k6 is installed: $$(k6 version)${RESET}"

check-service: ## Check if the service is running
	@echo "${YELLOW}Checking if service is running at ${BASE_URL}...${RESET}"
	@curl -s -o /dev/null -w "%{http_code}" ${BASE_URL}/api/v1/orders > /dev/null 2>&1 || { \
		echo "${RED}Service is not responding at ${BASE_URL}${RESET}"; \
		echo "${YELLOW}Please start the service first${RESET}"; \
		exit 1; \
	}
	@echo "${GREEN}Service is running!${RESET}"

test-create: check-k6 check-service ## Run create orders load test
	@echo "${BLUE}Running create orders test...${RESET}"
	@mkdir -p $(K6_OUT_DIR)
	k6 run -e BASE_URL=$(BASE_URL) \
		--out json=$(K6_OUT_DIR)/create-orders-$(TIMESTAMP).json \
		create-orders.js
	@echo "${GREEN}Test completed! Results saved to $(K6_OUT_DIR)/create-orders-$(TIMESTAMP).json${RESET}"

test-mark-paid: check-k6 check-service ## Run mark as paid load test
	@echo "${BLUE}Running mark as paid test...${RESET}"
	@mkdir -p $(K6_OUT_DIR)
	k6 run -e BASE_URL=$(BASE_URL) \
		--out json=$(K6_OUT_DIR)/mark-as-paid-$(TIMESTAMP).json \
		mark-as-paid.js
	@echo "${GREEN}Test completed! Results saved to $(K6_OUT_DIR)/mark-as-paid-$(TIMESTAMP).json${RESET}"

test-complete: check-k6 check-service ## Run complete flow load test (recommended)
	@echo "${BLUE}Running complete flow test...${RESET}"
	@mkdir -p $(K6_OUT_DIR)
	k6 run -e BASE_URL=$(BASE_URL) \
		--out json=$(K6_OUT_DIR)/complete-flow-$(TIMESTAMP).json \
		complete-flow.js
	@echo "${GREEN}Test completed! Results saved to $(K6_OUT_DIR)/complete-flow-$(TIMESTAMP).json${RESET}"

test-spike: check-k6 check-service ## Run spike test
	@echo "${BLUE}Running spike test...${RESET}"
	@mkdir -p $(K6_OUT_DIR)
	k6 run -e BASE_URL=$(BASE_URL) \
		--out json=$(K6_OUT_DIR)/spike-test-$(TIMESTAMP).json \
		spike-test.js
	@echo "${GREEN}Test completed! Results saved to $(K6_OUT_DIR)/spike-test-$(TIMESTAMP).json${RESET}"

test-soak: check-k6 check-service ## Run soak test (30 minutes)
	@echo "${BLUE}Running soak test (this will take ~30 minutes)...${RESET}"
	@mkdir -p $(K6_OUT_DIR)
	k6 run -e BASE_URL=$(BASE_URL) \
		--out json=$(K6_OUT_DIR)/soak-test-$(TIMESTAMP).json \
		soak-test.js
	@echo "${GREEN}Test completed! Results saved to $(K6_OUT_DIR)/soak-test-$(TIMESTAMP).json${RESET}"

test-all: check-k6 check-service ## Run all tests sequentially (except soak)
	@echo "${BLUE}Running all tests...${RESET}"
	@$(MAKE) test-create
	@sleep 10
	@$(MAKE) test-complete
	@sleep 10
	@$(MAKE) test-spike
	@echo "${GREEN}All tests completed!${RESET}"

test-complete-html: check-k6 check-service ## Run complete flow test with HTML report
	@echo "${BLUE}Running complete flow test with HTML report...${RESET}"
	@mkdir -p $(K6_OUT_DIR)
	k6 run -e BASE_URL=$(BASE_URL) \
		--out json=$(K6_OUT_DIR)/complete-flow-$(TIMESTAMP).json \
		--summary-export=$(K6_OUT_DIR)/summary-$(TIMESTAMP).json \
		complete-flow.js
	@echo "${GREEN}Test completed!${RESET}"
	@echo "${YELLOW}Results: $(K6_OUT_DIR)/complete-flow-$(TIMESTAMP).json${RESET}"
	@echo "${YELLOW}Summary: $(K6_OUT_DIR)/summary-$(TIMESTAMP).json${RESET}"

test-quick: check-k6 check-service ## Run quick smoke test (1 VU, 1 iteration)
	@echo "${BLUE}Running quick smoke test...${RESET}"
	k6 run -e BASE_URL=$(BASE_URL) \
		--vus 1 \
		--iterations 1 \
		complete-flow.js
	@echo "${GREEN}Quick test completed!${RESET}"

test-custom: check-k6 check-service ## Run custom test with VUS and DURATION env vars
	@echo "${BLUE}Running custom test...${RESET}"
	@if [ -z "$(VUS)" ] || [ -z "$(DURATION)" ]; then \
		echo "${RED}Please provide VUS and DURATION${RESET}"; \
		echo "${YELLOW}Example: make test-custom VUS=50 DURATION=5m${RESET}"; \
		exit 1; \
	fi
	@mkdir -p $(K6_OUT_DIR)
	k6 run -e BASE_URL=$(BASE_URL) \
		--vus $(VUS) \
		--duration $(DURATION) \
		--out json=$(K6_OUT_DIR)/custom-$(TIMESTAMP).json \
		complete-flow.js
	@echo "${GREEN}Custom test completed!${RESET}"

clean: ## Clean test results
	@echo "${YELLOW}Cleaning test results...${RESET}"
	@rm -rf $(K6_OUT_DIR)
	@echo "${GREEN}Results cleaned!${RESET}"

results: ## Show latest test results directory
	@echo "${BLUE}Latest test results:${RESET}"
	@ls -lt $(K6_OUT_DIR) 2>/dev/null | head -n 10 || echo "${YELLOW}No results found${RESET}"

view-results: ## View the most recent test result summary
	@echo "${BLUE}Most recent test summary:${RESET}"
	@latest=$$(ls -t $(K6_OUT_DIR)/summary-*.json 2>/dev/null | head -n 1); \
	if [ -n "$$latest" ]; then \
		cat "$$latest" | python3 -m json.tool; \
	else \
		echo "${YELLOW}No summary files found${RESET}"; \
	fi

watch-service: ## Watch service logs (assumes docker-compose)
	@echo "${BLUE}Watching service logs...${RESET}"
	docker-compose logs -f || echo "${YELLOW}docker-compose not found${RESET}"

benchmark: check-k6 check-service ## Run benchmark suite (create, complete, spike)
	@echo "${BLUE}==================================================${RESET}"
	@echo "${BLUE}Starting Benchmark Suite${RESET}"
	@echo "${BLUE}==================================================${RESET}"
	@mkdir -p $(K6_OUT_DIR)/benchmark-$(TIMESTAMP)
	@echo "\n${BLUE}[1/3] Create Orders Test${RESET}"
	@k6 run -e BASE_URL=$(BASE_URL) \
		--out json=$(K6_OUT_DIR)/benchmark-$(TIMESTAMP)/create-orders.json \
		create-orders.js
	@sleep 10
	@echo "\n${BLUE}[2/3] Complete Flow Test${RESET}"
	@k6 run -e BASE_URL=$(BASE_URL) \
		--out json=$(K6_OUT_DIR)/benchmark-$(TIMESTAMP)/complete-flow.json \
		complete-flow.js
	@sleep 10
	@echo "\n${BLUE}[3/3] Spike Test${RESET}"
	@k6 run -e BASE_URL=$(BASE_URL) \
		--out json=$(K6_OUT_DIR)/benchmark-$(TIMESTAMP)/spike-test.json \
		spike-test.js
	@echo "\n${GREEN}==================================================${RESET}"
	@echo "${GREEN}Benchmark Suite Completed!${RESET}"
	@echo "${GREEN}Results saved to: $(K6_OUT_DIR)/benchmark-$(TIMESTAMP)/${RESET}"
	@echo "${GREEN}==================================================${RESET}"

info: ## Show configuration information
	@echo "${BLUE}K6 Load Test Configuration${RESET}"
	@echo "  Base URL: ${YELLOW}$(BASE_URL)${RESET}"
	@echo "  Output Directory: ${YELLOW}$(K6_OUT_DIR)${RESET}"
	@echo "  k6 Version: ${YELLOW}$$(k6 version 2>/dev/null || echo 'not installed')${RESET}"
	@echo ""
	@echo "${BLUE}Available Tests:${RESET}"
	@echo "  ${GREEN}test-quick${RESET}     - Quick smoke test (1 VU)"
	@echo "  ${GREEN}test-create${RESET}    - Create orders load test"
	@echo "  ${GREEN}test-complete${RESET}  - Complete flow load test ${YELLOW}(recommended)${RESET}"
	@echo "  ${GREEN}test-spike${RESET}     - Spike test (sudden load)"
	@echo "  ${GREEN}test-soak${RESET}      - Soak test (30 min duration)"
	@echo "  ${GREEN}benchmark${RESET}      - Run full benchmark suite"

.DEFAULT_GOAL := help
