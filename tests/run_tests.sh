#!/bin/bash
#
# XBE CLI Integration Test Runner
#
# Runs all integration tests against staging server.
#
# Usage:
#   ./tests/run_tests.sh                    # Run all tests sequentially
#   ./tests/run_tests.sh -p                 # Run all tests in parallel (8 jobs)
#   ./tests/run_tests.sh -p -j 4            # Run all tests in parallel (4 jobs)
#   ./tests/run_tests.sh users              # Run only users tests
#   ./tests/run_tests.sh -p users brokers   # Run users and brokers in parallel
#

set -e

# Get the directory where this script lives
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Default settings
PARALLEL=false
JOBS=8

# Parse flags
while [[ $# -gt 0 ]]; do
    case "$1" in
        -p|--parallel)
            PARALLEL=true
            shift
            ;;
        -j|--jobs)
            JOBS="$2"
            shift 2
            ;;
        -*)
            echo -e "${RED}Unknown option: $1${NC}"
            echo "Usage: $0 [-p|--parallel] [-j|--jobs N] [test_names...]"
            exit 1
            ;;
        *)
            break
            ;;
    esac
done

echo ""
echo -e "${BLUE}╔════════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║          XBE CLI Integration Test Suite                ║${NC}"
echo -e "${BLUE}╚════════════════════════════════════════════════════════╝${NC}"
echo ""

# Load configuration
source "$SCRIPT_DIR/lib/config.sh"

# Check configuration
check_config

# Build fresh binary
echo -e "${YELLOW}Building xbe binary...${NC}"
cd "$PROJECT_ROOT"
make build
echo -e "${GREEN}Build complete.${NC}"
echo ""

# Track overall results
TOTAL_PASSED=0
TOTAL_FAILED=0
TOTAL_TESTS=0
declare -a FAILED_SUITES

# Function to run a single test file (sequential mode)
run_test_file() {
    local test_file="$1"
    local test_name
    test_name=$(basename "$test_file" _test.sh)

    echo ""
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${BLUE}Running: $test_name${NC}"
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"

    if bash "$test_file"; then
        echo -e "${GREEN}✓ $test_name: PASSED${NC}"
        return 0
    else
        echo -e "${RED}✗ $test_name: FAILED${NC}"
        FAILED_SUITES+=("$test_name")
        return 1
    fi
}

# Determine which tests to run
if [[ $# -gt 0 ]]; then
    # Run specific tests
    TEST_FILES=()
    for arg in "$@"; do
        test_file="$SCRIPT_DIR/resources/${arg}_test.sh"
        if [[ -f "$test_file" ]]; then
            TEST_FILES+=("$test_file")
        else
            echo -e "${RED}Error: Test file not found: $test_file${NC}"
            exit 1
        fi
    done
else
    # Run all tests
    TEST_FILES=("$SCRIPT_DIR"/resources/*_test.sh)
fi

# Check that we have tests to run
if [[ ${#TEST_FILES[@]} -eq 0 ]]; then
    echo -e "${YELLOW}No test files found.${NC}"
    exit 0
fi

TOTAL_TESTS=${#TEST_FILES[@]}

if [[ "$PARALLEL" == true ]]; then
    # Parallel execution mode
    echo -e "${YELLOW}Running ${TOTAL_TESTS} test suite(s) in parallel (${JOBS} jobs)...${NC}"
    echo ""

    # Create temp directory for results
    RESULTS_DIR=$(mktemp -d)
    trap "rm -rf $RESULTS_DIR" EXIT

    # Export variables needed by subprocesses
    export RESULTS_DIR
    export XBE_BASE_URL
    export XBE_TOKEN

    # Function to run a single test and record result (for parallel mode)
    run_single_test() {
        local test_file="$1"
        local test_name
        test_name=$(basename "$test_file" _test.sh)
        local output_file="$RESULTS_DIR/${test_name}.log"
        local status_file="$RESULTS_DIR/${test_name}.status"

        # Run the test, capturing output
        if bash "$test_file" > "$output_file" 2>&1; then
            echo "0" > "$status_file"
        else
            echo "1" > "$status_file"
        fi
    }
    export -f run_single_test

    # Run tests in parallel using xargs
    printf '%s\n' "${TEST_FILES[@]}" | xargs -P "$JOBS" -I {} bash -c 'run_single_test "$@"' _ {}

    # Collect results
    for test_file in "${TEST_FILES[@]}"; do
        test_name=$(basename "$test_file" _test.sh)
        status_file="$RESULTS_DIR/${test_name}.status"
        output_file="$RESULTS_DIR/${test_name}.log"

        if [[ -f "$status_file" ]]; then
            status=$(cat "$status_file")
            if [[ "$status" == "0" ]]; then
                echo -e "${GREEN}✓ ${test_name}: PASSED${NC}"
                ((TOTAL_PASSED++))
            else
                echo -e "${RED}✗ ${test_name}: FAILED${NC}"
                ((TOTAL_FAILED++))
                FAILED_SUITES+=("$test_name")
            fi
        else
            echo -e "${RED}✗ ${test_name}: NO RESULT${NC}"
            ((TOTAL_FAILED++))
            FAILED_SUITES+=("$test_name")
        fi
    done

    # Show failed test output
    if [[ ${#FAILED_SUITES[@]} -gt 0 ]]; then
        echo ""
        echo -e "${RED}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
        echo -e "${RED}Failed Test Output:${NC}"
        echo -e "${RED}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
        for suite in "${FAILED_SUITES[@]}"; do
            output_file="$RESULTS_DIR/${suite}.log"
            if [[ -f "$output_file" ]]; then
                echo ""
                echo -e "${YELLOW}--- $suite ---${NC}"
                tail -50 "$output_file"
            fi
        done
    fi
else
    # Sequential execution mode
    echo -e "${YELLOW}Running ${TOTAL_TESTS} test suite(s) sequentially...${NC}"

    # Run each test file
    for test_file in "${TEST_FILES[@]}"; do
        if [[ -f "$test_file" ]]; then
            if run_test_file "$test_file"; then
                ((TOTAL_PASSED++))
            else
                ((TOTAL_FAILED++))
            fi
        fi
    done
fi

# Print summary
echo ""
echo ""
echo -e "${BLUE}╔════════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║                    Final Summary                       ║${NC}"
echo -e "${BLUE}╚════════════════════════════════════════════════════════╝${NC}"
echo ""
echo -e "  Test Suites Run:    $TOTAL_TESTS"
echo -e "  ${GREEN}Passed:${NC}             $TOTAL_PASSED"
echo -e "  ${RED}Failed:${NC}             $TOTAL_FAILED"

if [[ ${#FAILED_SUITES[@]} -gt 0 ]]; then
    echo ""
    echo -e "${RED}Failed test suites:${NC}"
    for suite in "${FAILED_SUITES[@]}"; do
        echo -e "  - $suite"
    done
fi

echo ""

# Exit with appropriate code
if [[ $TOTAL_FAILED -gt 0 ]]; then
    echo -e "${RED}Some tests failed!${NC}"
    exit 1
else
    echo -e "${GREEN}All tests passed!${NC}"
    exit 0
fi
