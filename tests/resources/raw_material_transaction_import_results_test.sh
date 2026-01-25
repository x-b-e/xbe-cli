#!/bin/bash
#
# XBE CLI Integration Tests: Raw Material Transaction Import Results
#
# Tests list and show operations for the raw-material-transaction-import-results resource.
#
# COVERAGE: List filters + show
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

SAMPLE_ID=""
SOURCE_TYPE=""
SOURCE_ID=""
BROKER_ID=""
BATCH_ID=""
LOCATION_ID=""
HAS_ERRORS=""
SKIP_ID_FILTERS=0

describe "Resource: raw-material-transaction-import-results"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List raw material transaction import results"
xbe_json view raw-material-transaction-import-results list --limit 5
assert_success

test_name "List raw material transaction import results returns array"
xbe_json view raw-material-transaction-import-results list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list raw material transaction import results"
fi

# ============================================================================
# Sample Data
# ============================================================================

test_name "Find sample raw material transaction import result"
xbe_json view raw-material-transaction-import-results list --limit 1
if [[ $status -eq 0 ]]; then
    SAMPLE_ID=$(json_get ".[0].id")
    SOURCE_TYPE=$(json_get ".[0].source_type")
    SOURCE_ID=$(json_get ".[0].source_id")
    BROKER_ID=$(json_get ".[0].broker_id")
    BATCH_ID=$(json_get ".[0].batch_id")
    LOCATION_ID=$(json_get ".[0].location_id")
    HAS_ERRORS=$(json_get ".[0].has_errors")
    if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
        pass
    else
        SKIP_ID_FILTERS=1
        skip "No raw material transaction import results available"
    fi
else
    SKIP_ID_FILTERS=1
    fail "Failed to list raw material transaction import results"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List import results with --source-type/--source-id filter"
if [[ $SKIP_ID_FILTERS -eq 0 && -n "$SOURCE_TYPE" && "$SOURCE_TYPE" != "null" && -n "$SOURCE_ID" && "$SOURCE_ID" != "null" ]]; then
    xbe_json view raw-material-transaction-import-results list --source-type "$SOURCE_TYPE" --source-id "$SOURCE_ID" --limit 5
    assert_success
else
    skip "No source type/id available"
fi

test_name "List import results with --source filter"
if [[ $SKIP_ID_FILTERS -eq 0 && -n "$SOURCE_TYPE" && "$SOURCE_TYPE" != "null" && -n "$SOURCE_ID" && "$SOURCE_ID" != "null" ]]; then
    xbe_json view raw-material-transaction-import-results list --source "${SOURCE_TYPE}|${SOURCE_ID}" --limit 5
    assert_success
else
    skip "No source type/id available"
fi

test_name "List import results with --broker filter"
if [[ -n "$BROKER_ID" && "$BROKER_ID" != "null" ]]; then
    xbe_json view raw-material-transaction-import-results list --broker "$BROKER_ID" --limit 5
    assert_success
else
    skip "No broker ID available"
fi

test_name "List import results with --batch-id filter"
if [[ -n "$BATCH_ID" && "$BATCH_ID" != "null" ]]; then
    xbe_json view raw-material-transaction-import-results list --batch-id "$BATCH_ID" --limit 5
    assert_success
else
    skip "No batch ID available"
fi

test_name "List import results with --location-id filter"
if [[ -n "$LOCATION_ID" && "$LOCATION_ID" != "null" ]]; then
    xbe_json view raw-material-transaction-import-results list --location-id "$LOCATION_ID" --limit 5
    assert_success
else
    skip "No location ID available"
fi

test_name "List import results with --has-errors filter"
if [[ -n "$HAS_ERRORS" && "$HAS_ERRORS" != "null" ]]; then
    xbe_json view raw-material-transaction-import-results list --has-errors "$HAS_ERRORS" --limit 5
    assert_success
else
    xbe_json view raw-material-transaction-import-results list --has-errors true --limit 5
    assert_success
fi

test_name "List import results with --earliest-created-transaction-at-min filter"
xbe_json view raw-material-transaction-import-results list --earliest-created-transaction-at-min "2020-01-01T00:00:00Z" --limit 5
assert_success

test_name "List import results with --earliest-created-transaction-at-max filter"
xbe_json view raw-material-transaction-import-results list --earliest-created-transaction-at-max "2030-01-01T00:00:00Z" --limit 5
assert_success

test_name "List import results with --is-earliest-created-transaction-at filter"
xbe_json view raw-material-transaction-import-results list --is-earliest-created-transaction-at true --limit 5
assert_success

test_name "List import results with --latest-created-transaction-at-min filter"
xbe_json view raw-material-transaction-import-results list --latest-created-transaction-at-min "2020-01-01T00:00:00Z" --limit 5
assert_success

test_name "List import results with --latest-created-transaction-at-max filter"
xbe_json view raw-material-transaction-import-results list --latest-created-transaction-at-max "2030-01-01T00:00:00Z" --limit 5
assert_success

test_name "List import results with --is-latest-created-transaction-at filter"
xbe_json view raw-material-transaction-import-results list --is-latest-created-transaction-at true --limit 5
assert_success

test_name "List import results with --disconnected-at-min filter"
xbe_json view raw-material-transaction-import-results list --disconnected-at-min "2020-01-01T00:00:00Z" --limit 5
assert_success

test_name "List import results with --disconnected-at-max filter"
xbe_json view raw-material-transaction-import-results list --disconnected-at-max "2030-01-01T00:00:00Z" --limit 5
assert_success

test_name "List import results with --is-disconnected-at filter"
xbe_json view raw-material-transaction-import-results list --is-disconnected-at true --limit 5
assert_success

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show raw material transaction import result"
if [[ $SKIP_ID_FILTERS -eq 0 && -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_json view raw-material-transaction-import-results show "$SAMPLE_ID"
    assert_success
else
    skip "No import result ID available"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
