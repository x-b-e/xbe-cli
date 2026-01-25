#!/bin/bash
#
# XBE CLI Integration Tests: Lineup Scenario Generators
#
# Tests list/show and create/delete behavior for lineup-scenario-generators.
#
# COVERAGE: List filters + show + create attributes + delete guard
#

source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

describe "Resource: lineup-scenario-generators"

SAMPLE_ID=""
SAMPLE_BROKER_ID=""
SAMPLE_CUSTOMER_ID=""
SAMPLE_DATE=""
SAMPLE_WINDOW=""

CREATED_ID=""

BROKER_ID="${XBE_TEST_BROKER_ID:-}"
CUSTOMER_ID="${XBE_TEST_CUSTOMER_ID:-}"
TODAY=$(date +%Y-%m-%d)

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List lineup scenario generators"
xbe_json view lineup-scenario-generators list --limit 5
assert_success

test_name "List lineup scenario generators returns array"
xbe_json view lineup-scenario-generators list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list lineup scenario generators"
fi

# ============================================================================
# Sample Record (used for filters/show)
# ============================================================================

test_name "Capture sample lineup scenario generator"
xbe_json view lineup-scenario-generators list --limit 1
if [[ $status -eq 0 ]]; then
    SAMPLE_ID=$(json_get ".[0].id")
    SAMPLE_BROKER_ID=$(json_get ".[0].broker_id")
    SAMPLE_CUSTOMER_ID=$(json_get ".[0].customer_id")
    SAMPLE_DATE=$(json_get ".[0].date")
    SAMPLE_WINDOW=$(json_get ".[0].window")
    if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
        pass
    else
        skip "No lineup scenario generators available for follow-on tests"
    fi
else
    skip "Could not list lineup scenario generators to capture sample"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List lineup scenario generators with --broker filter"
if [[ -n "$SAMPLE_BROKER_ID" && "$SAMPLE_BROKER_ID" != "null" ]]; then
    xbe_json view lineup-scenario-generators list --broker "$SAMPLE_BROKER_ID" --limit 5
    assert_success
else
    skip "No broker ID available"
fi

test_name "List lineup scenario generators with --date filter"
if [[ -n "$SAMPLE_DATE" && "$SAMPLE_DATE" != "null" ]]; then
    xbe_json view lineup-scenario-generators list --date "$SAMPLE_DATE" --limit 5
    assert_success
else
    skip "No date available"
fi

test_name "List lineup scenario generators with --date-min filter"
xbe_json view lineup-scenario-generators list --date-min "2020-01-01" --limit 5
assert_success

test_name "List lineup scenario generators with --date-max filter"
xbe_json view lineup-scenario-generators list --date-max "2030-01-01" --limit 5
assert_success

test_name "List lineup scenario generators with --window filter"
if [[ -n "$SAMPLE_WINDOW" && "$SAMPLE_WINDOW" != "null" ]]; then
    xbe_json view lineup-scenario-generators list --window "$SAMPLE_WINDOW" --limit 5
    assert_success
else
    skip "No window available"
fi

test_name "List lineup scenario generators with --completed-at-min filter"
xbe_json view lineup-scenario-generators list --completed-at-min "2020-01-01T00:00:00Z" --limit 5
assert_success

test_name "List lineup scenario generators with --completed-at-max filter"
xbe_json view lineup-scenario-generators list --completed-at-max "2030-01-01T00:00:00Z" --limit 5
assert_success

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show lineup scenario generator"
if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_json view lineup-scenario-generators show "$SAMPLE_ID"
    assert_success
else
    skip "No lineup scenario generator ID available"
fi

# ============================================================================
# CREATE Tests - Required Fields
# ============================================================================

test_name "Fail create without broker"
xbe_run do lineup-scenario-generators create --date "$TODAY" --window day
assert_failure

test_name "Fail create without date"
xbe_run do lineup-scenario-generators create --broker "1" --window day
assert_failure

test_name "Fail create without window"
xbe_run do lineup-scenario-generators create --broker "1" --date "$TODAY"
assert_failure

# ============================================================================
# CREATE Tests - Optional Attributes
# ============================================================================

test_name "Create lineup scenario generator with optional attributes"
if [[ -n "$BROKER_ID" ]]; then
    create_args=(do lineup-scenario-generators create
        --broker "$BROKER_ID"
        --date "$TODAY"
        --window day
        --include-trucker-assignments-as-constraints=true
        --trucker-assignment-limits-lookback-window-days 3
        --skip-minimum-assignment-count=true
        --skip-create-lineup-scenario-solution=true
        --use-most-recent-lineup-scenario-constraints=false)

    if [[ -n "$CUSTOMER_ID" ]]; then
        create_args+=(--customer "$CUSTOMER_ID")
    fi

    xbe_json "${create_args[@]}"

    if [[ $status -eq 0 ]]; then
        CREATED_ID=$(json_get ".id")
        if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
            register_cleanup "lineup-scenario-generators" "$CREATED_ID"
            pass
        else
            fail "Created generator but no ID returned"
        fi
    else
        fail "Failed to create lineup scenario generator"
    fi
else
    skip "XBE_TEST_BROKER_ID not set"
fi

# ============================================================================
# DELETE Tests - Guard
# ============================================================================

test_name "Delete lineup scenario generator requires --confirm flag"
if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
    xbe_run do lineup-scenario-generators delete "$CREATED_ID"
    assert_failure
else
    skip "No generator ID available"
fi

run_tests
