#!/bin/bash
#
# XBE CLI Integration Test Helpers
# Provides common functions for testing CLI commands against staging
#

# Fail on any error in the helper setup
set -e

# Test tracking
TESTS_RUN=0
TESTS_PASSED=0
TESTS_FAILED=0
CURRENT_RESOURCE=""
CURRENT_TEST=""

# Store output and status from last command
output=""
status=0

# Colors for output (if terminal supports it)
if [[ -t 1 ]]; then
    RED='\033[0;31m'
    GREEN='\033[0;32m'
    YELLOW='\033[1;33m'
    BLUE='\033[0;34m'
    NC='\033[0m' # No Color
else
    RED=''
    GREEN=''
    YELLOW=''
    BLUE=''
    NC=''
fi

# Get the project root directory (parent of tests/)
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
XBE_BIN="${PROJECT_ROOT}/xbe"

# Load configuration
source "$(dirname "${BASH_SOURCE[0]}")/config.sh"

# ============================================================================
# Test Lifecycle Functions
# ============================================================================

# Describe a test suite (usually called once per resource file)
describe() {
    CURRENT_RESOURCE="$1"
    echo ""
    echo -e "${BLUE}=== $1 ===${NC}"
}

# Name the current test
test_name() {
    CURRENT_TEST="$1"
    echo -e "  ${YELLOW}Testing:${NC} $1"
}

# Mark test as passed
pass() {
    TESTS_PASSED=$((TESTS_PASSED + 1))
    TESTS_RUN=$((TESTS_RUN + 1))
    echo -e "    ${GREEN}✓ PASS${NC}"
}

# Mark test as failed with message
fail() {
    local msg="${1:-Test failed}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
    TESTS_RUN=$((TESTS_RUN + 1))
    echo -e "    ${RED}✗ FAIL:${NC} $msg"
    if [[ -n "$output" ]]; then
        echo "    Output: ${output:0:200}"
    fi
}

# Skip a test with reason
skip() {
    local reason="${1:-Skipped}"
    echo -e "    ${YELLOW}⊘ SKIP:${NC} $reason"
}

# Print test results summary
print_summary() {
    echo ""
    echo "============================================"
    echo -e "Results for ${BLUE}${CURRENT_RESOURCE}${NC}:"
    echo -e "  ${GREEN}Passed:${NC} $TESTS_PASSED"
    echo -e "  ${RED}Failed:${NC} $TESTS_FAILED"
    echo -e "  Total:  $TESTS_RUN"
    echo "============================================"

    if [[ $TESTS_FAILED -gt 0 ]]; then
        return 1
    fi
    return 0
}

# Run all tests and exit with appropriate code
run_tests() {
    print_summary
    exit $?
}

# ============================================================================
# Command Execution
# ============================================================================

# Run a command and capture output and status
# Usage: run ./xbe do users create --name "Test" --json
run() {
    set +e
    output=$("$@" 2>&1)
    status=$?
    set -e
}

# Run xbe command with common flags
# Usage: xbe_run do users create --name "Test"
xbe_run() {
    if [[ -n "$XBE_TOKEN" ]]; then
        run "$XBE_BIN" --base-url "$XBE_BASE_URL" --token "$XBE_TOKEN" "$@"
    else
        # Using stored auth - don't pass empty token
        run "$XBE_BIN" --base-url "$XBE_BASE_URL" "$@"
    fi
}

# Run xbe command with --json flag
# Usage: xbe_json do users create --name "Test"
xbe_json() {
    xbe_run "$@" --json
}

# ============================================================================
# Assertions
# ============================================================================

# Assert command exited with success (status 0)
assert_success() {
    if [[ $status -eq 0 ]]; then
        pass
    else
        fail "Expected success (exit 0), got exit $status"
    fi
}

# Assert command failed (non-zero status)
assert_failure() {
    if [[ $status -ne 0 ]]; then
        pass
    else
        fail "Expected failure, but command succeeded"
    fi
}

# Assert output contains a substring
# Usage: assert_output_contains "Created user"
assert_output_contains() {
    local expected="$1"
    if [[ "$output" == *"$expected"* ]]; then
        pass
    else
        fail "Output does not contain '$expected'"
    fi
}

# Assert output does NOT contain a substring
# Usage: assert_output_not_contains "error"
assert_output_not_contains() {
    local unexpected="$1"
    if [[ "$output" != *"$unexpected"* ]]; then
        pass
    else
        fail "Output unexpectedly contains '$unexpected'"
    fi
}

# Assert JSON output has a specific key (using jq)
# Usage: assert_json_has ".id"
assert_json_has() {
    local jq_path="$1"
    if echo "$output" | jq -e "$jq_path" > /dev/null 2>&1; then
        pass
    else
        fail "JSON output missing key: $jq_path"
    fi
}

# Assert JSON output key equals expected value
# Usage: assert_json_equals ".name" "Test User"
assert_json_equals() {
    local jq_path="$1"
    local expected="$2"
    local actual
    actual=$(echo "$output" | jq -r "$jq_path" 2>/dev/null)
    if [[ "$actual" == "$expected" ]]; then
        pass
    else
        fail "Expected $jq_path = '$expected', got '$actual'"
    fi
}

# Assert JSON output is an array
assert_json_is_array() {
    if echo "$output" | jq -e 'type == "array"' > /dev/null 2>&1; then
        pass
    else
        fail "JSON output is not an array"
    fi
}

# Assert JSON array is not empty
assert_json_array_not_empty() {
    local length
    length=$(echo "$output" | jq 'length' 2>/dev/null)
    if [[ "$length" -gt 0 ]]; then
        pass
    else
        fail "JSON array is empty"
    fi
}

# Assert JSON boolean equals expected value
# Usage: assert_json_bool ".deleted" "true"
assert_json_bool() {
    local jq_path="$1"
    local expected="$2"
    local actual
    actual=$(echo "$output" | jq -r "$jq_path" 2>/dev/null)
    if [[ "$actual" == "$expected" ]]; then
        pass
    else
        fail "Expected $jq_path = $expected, got $actual"
    fi
}

# ============================================================================
# JSON Helpers
# ============================================================================

# Extract a value from JSON output
# Usage: id=$(json_get ".id")
json_get() {
    echo "$output" | jq -r "$1" 2>/dev/null
}

# ============================================================================
# Test Data Helpers
# ============================================================================

# Generate a unique test suffix for resource names
# Usage: suffix=$(unique_suffix)
unique_suffix() {
    echo "test_$(date +%s)_${RANDOM}"
}

# Generate a unique email address for testing
# Usage: email=$(unique_email)
unique_email() {
    echo "test-$(date +%s)-${RANDOM}@xbe-cli-test.example.com"
}

# Generate a unique name for testing
# Usage: name=$(unique_name "User")
unique_name() {
    local prefix="${1:-Test}"
    echo "CLI-Test-${prefix}-$(date +%s)-${RANDOM}"
}

# Generate a unique mobile number for testing
# Usage: mobile=$(unique_mobile)
# Note: Uses a real area code (815) to pass phone validation
unique_mobile() {
    # Generate a valid US phone number format that passes validation
    # Using area code 815 (Illinois) with random subscriber number
    local rand4=$(printf "%04d" $((RANDOM % 10000)))
    echo "+1815347${rand4}"
}

# ============================================================================
# Cleanup Helpers
# ============================================================================

# Store IDs for cleanup at the end
declare -a CLEANUP_IDS
declare -a CLEANUP_TYPES

# Register a resource for cleanup
# Usage: register_cleanup "users" "$user_id"
register_cleanup() {
    local resource_type="$1"
    local resource_id="$2"
    CLEANUP_TYPES+=("$resource_type")
    CLEANUP_IDS+=("$resource_id")
}

# Run cleanup for all registered resources (in reverse order)
run_cleanup() {
    echo ""
    echo -e "${YELLOW}Cleaning up test resources...${NC}"

    local i
    for ((i=${#CLEANUP_IDS[@]}-1; i>=0; i--)); do
        local resource_type="${CLEANUP_TYPES[$i]}"
        local resource_id="${CLEANUP_IDS[$i]}"

        if [[ -n "$resource_id" && "$resource_id" != "null" ]]; then
            echo "  Deleting $resource_type $resource_id..."
            xbe_run do "$resource_type" delete "$resource_id" --confirm 2>/dev/null || true
        fi
    done

    echo -e "${GREEN}Cleanup complete.${NC}"
}

# Trap to ensure cleanup runs on exit
trap run_cleanup EXIT

# ============================================================================
# Pre-requisite Checks
# ============================================================================

# Check that jq is installed
check_jq() {
    if ! command -v jq &> /dev/null; then
        echo -e "${RED}Error: jq is required but not installed.${NC}"
        echo "Install with: brew install jq"
        exit 1
    fi
}

# Check that the xbe binary exists
check_xbe_binary() {
    if [[ ! -x "$XBE_BIN" ]]; then
        echo -e "${RED}Error: xbe binary not found at $XBE_BIN${NC}"
        echo "Run 'make build' first."
        exit 1
    fi
}

# Run all pre-requisite checks
init_tests() {
    check_jq
    check_xbe_binary
    check_config
}

# Initialize on source
init_tests
