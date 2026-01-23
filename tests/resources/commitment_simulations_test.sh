#!/bin/bash
#
# XBE CLI Integration Tests: Commitment Simulations
#
# Tests list/show and create behavior for commitment-simulations.
#
# COVERAGE: List filters + show + create required attributes
#

source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

describe "Resource: commitment-simulations"

SAMPLE_ID=""
SAMPLE_COMMITMENT_ID=""
SAMPLE_COMMITMENT_TYPE=""
SAMPLE_STATUS=""

COMMITMENT_ID="${XBE_TEST_COMMITMENT_ID:-}"
COMMITMENT_TYPE="${XBE_TEST_COMMITMENT_TYPE:-}"
TODAY=$(date +%Y-%m-%d)

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List commitment simulations"
xbe_json view commitment-simulations list --limit 5
assert_success

test_name "List commitment simulations returns array"
xbe_json view commitment-simulations list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list commitment simulations"
fi

# ============================================================================
# Sample Record (used for filters/show)
# ============================================================================

test_name "Capture sample commitment simulation"
xbe_json view commitment-simulations list --limit 1
if [[ $status -eq 0 ]]; then
    SAMPLE_ID=$(json_get ".[0].id")
    SAMPLE_COMMITMENT_ID=$(json_get ".[0].commitment_id")
    SAMPLE_COMMITMENT_TYPE=$(json_get ".[0].commitment_type")
    SAMPLE_STATUS=$(json_get ".[0].status")
    if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
        pass
    else
        skip "No commitment simulations available for follow-on tests"
    fi
else
    skip "Could not list commitment simulations to capture sample"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List commitment simulations with --commitment filter"
if [[ -n "$SAMPLE_COMMITMENT_ID" && "$SAMPLE_COMMITMENT_ID" != "null" ]]; then
    xbe_json view commitment-simulations list --commitment "$SAMPLE_COMMITMENT_ID" --limit 5
    assert_success
else
    skip "No commitment ID available"
fi

test_name "List commitment simulations with --status filter"
if [[ -n "$SAMPLE_STATUS" && "$SAMPLE_STATUS" != "null" ]]; then
    xbe_json view commitment-simulations list --status "$SAMPLE_STATUS" --limit 5
    assert_success
else
    xbe_json view commitment-simulations list --status enqueued --limit 5
    assert_success
fi

test_name "List commitment simulations with --created-at-min filter"
xbe_json view commitment-simulations list --created-at-min "2020-01-01T00:00:00Z" --limit 5
assert_success

test_name "List commitment simulations with --created-at-max filter"
xbe_json view commitment-simulations list --created-at-max "2030-01-01T00:00:00Z" --limit 5
assert_success

test_name "List commitment simulations with --updated-at-min filter"
xbe_json view commitment-simulations list --updated-at-min "2020-01-01T00:00:00Z" --limit 5
assert_success

test_name "List commitment simulations with --updated-at-max filter"
xbe_json view commitment-simulations list --updated-at-max "2030-01-01T00:00:00Z" --limit 5
assert_success

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show commitment simulation"
if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_json view commitment-simulations show "$SAMPLE_ID"
    assert_success
else
    skip "No commitment simulation ID available"
fi

# ============================================================================
# CREATE Tests - Required Fields
# ============================================================================

test_name "Fail create without commitment type"
xbe_run do commitment-simulations create --commitment-id "1" --start-on "$TODAY" --end-on "$TODAY" --iteration-count 1
assert_failure

test_name "Fail create without commitment id"
xbe_run do commitment-simulations create --commitment-type commitments --start-on "$TODAY" --end-on "$TODAY" --iteration-count 1
assert_failure

test_name "Fail create without start-on"
xbe_run do commitment-simulations create --commitment-type commitments --commitment-id "1" --end-on "$TODAY" --iteration-count 1
assert_failure

test_name "Fail create without end-on"
xbe_run do commitment-simulations create --commitment-type commitments --commitment-id "1" --start-on "$TODAY" --iteration-count 1
assert_failure

test_name "Fail create without iteration-count"
xbe_run do commitment-simulations create --commitment-type commitments --commitment-id "1" --start-on "$TODAY" --end-on "$TODAY"
assert_failure

# ============================================================================
# CREATE Tests - Success Path
# ============================================================================

test_name "Create commitment simulation"
if [[ -n "$COMMITMENT_ID" && -n "$COMMITMENT_TYPE" ]]; then
    xbe_json do commitment-simulations create \
        --commitment-type "$COMMITMENT_TYPE" \
        --commitment-id "$COMMITMENT_ID" \
        --start-on "$TODAY" \
        --end-on "$TODAY" \
        --iteration-count 10

    if [[ $status -eq 0 ]]; then
        assert_json_has ".id"
    else
        fail "Failed to create commitment simulation"
    fi
else
    skip "Set XBE_TEST_COMMITMENT_ID and XBE_TEST_COMMITMENT_TYPE to run create test"
fi

run_tests
