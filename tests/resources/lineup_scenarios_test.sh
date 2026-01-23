#!/bin/bash
#
# XBE CLI Integration Tests: Lineup Scenarios
#
# Tests list/show and create/update/delete behavior for lineup-scenarios.
#
# COVERAGE: List filters + show + create attributes + update attributes + delete guard
#

source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

describe "Resource: lineup-scenarios"

SAMPLE_ID=""
SAMPLE_BROKER_ID=""
SAMPLE_CUSTOMER_ID=""
SAMPLE_DATE=""
SAMPLE_WINDOW=""
SAMPLE_GENERATOR_ID=""

CREATED_ID=""

BROKER_ID="${XBE_TEST_BROKER_ID:-}"
CUSTOMER_ID="${XBE_TEST_CUSTOMER_ID:-}"
TODAY=$(date +%Y-%m-%d)

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List lineup scenarios"
xbe_json view lineup-scenarios list --limit 5
assert_success

test_name "List lineup scenarios returns array"
xbe_json view lineup-scenarios list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list lineup scenarios"
fi

# ============================================================================
# Sample Record (used for filters/show/update)
# ============================================================================

test_name "Capture sample lineup scenario"
xbe_json view lineup-scenarios list --limit 1
if [[ $status -eq 0 ]]; then
    SAMPLE_ID=$(json_get ".[0].id")
    SAMPLE_BROKER_ID=$(json_get ".[0].broker_id")
    SAMPLE_CUSTOMER_ID=$(json_get ".[0].customer_id")
    SAMPLE_DATE=$(json_get ".[0].date")
    SAMPLE_WINDOW=$(json_get ".[0].window")
    SAMPLE_GENERATOR_ID=$(json_get ".[0].generator_id")
    if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
        pass
    else
        skip "No lineup scenarios available for follow-on tests"
    fi
else
    skip "Could not list lineup scenarios to capture sample"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List lineup scenarios with --broker filter"
if [[ -n "$SAMPLE_BROKER_ID" && "$SAMPLE_BROKER_ID" != "null" ]]; then
    xbe_json view lineup-scenarios list --broker "$SAMPLE_BROKER_ID" --limit 5
    assert_success
else
    skip "No broker ID available"
fi

test_name "List lineup scenarios with --date filter"
if [[ -n "$SAMPLE_DATE" && "$SAMPLE_DATE" != "null" ]]; then
    xbe_json view lineup-scenarios list --date "$SAMPLE_DATE" --limit 5
    assert_success
else
    skip "No date available"
fi

test_name "List lineup scenarios with --date-min filter"
xbe_json view lineup-scenarios list --date-min "2020-01-01" --limit 5
assert_success

test_name "List lineup scenarios with --date-max filter"
xbe_json view lineup-scenarios list --date-max "2030-01-01" --limit 5
assert_success

test_name "List lineup scenarios with --window filter"
if [[ -n "$SAMPLE_WINDOW" && "$SAMPLE_WINDOW" != "null" ]]; then
    xbe_json view lineup-scenarios list --window "$SAMPLE_WINDOW" --limit 5
    assert_success
else
    skip "No window available"
fi

test_name "List lineup scenarios with --generator filter"
if [[ -n "$SAMPLE_GENERATOR_ID" && "$SAMPLE_GENERATOR_ID" != "null" ]]; then
    xbe_json view lineup-scenarios list --generator "$SAMPLE_GENERATOR_ID" --limit 5
    assert_success
else
    skip "No generator ID available"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show lineup scenario"
if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_json view lineup-scenarios show "$SAMPLE_ID"
    assert_success
else
    skip "No lineup scenario ID available"
fi

# ============================================================================
# CREATE Tests - Required Fields
# ============================================================================

test_name "Fail create without broker"
xbe_run do lineup-scenarios create --date "$TODAY" --window day
assert_failure

test_name "Fail create without date"
xbe_run do lineup-scenarios create --broker "1" --window day
assert_failure

test_name "Fail create without window"
xbe_run do lineup-scenarios create --broker "1" --date "$TODAY"
assert_failure

# ============================================================================
# CREATE Tests - Optional Attributes
# ============================================================================

test_name "Create lineup scenario with optional attributes"
if [[ -n "$BROKER_ID" ]]; then
    create_args=(do lineup-scenarios create
        --broker "$BROKER_ID"
        --date "$TODAY"
        --window day
        --name "CLI Scenario $(date +%s)"
        --include-trucker-assignments-as-constraints=true
        --add-lineups-automatically=false)

    if [[ -n "$CUSTOMER_ID" ]]; then
        create_args+=(--customer "$CUSTOMER_ID")
    fi

    xbe_json "${create_args[@]}"

    if [[ $status -eq 0 ]]; then
        CREATED_ID=$(json_get ".id")
        if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
            register_cleanup "lineup-scenarios" "$CREATED_ID"
            pass
        else
            fail "Created scenario but no ID returned"
        fi
    else
        fail "Failed to create lineup scenario"
    fi
else
    skip "XBE_TEST_BROKER_ID not set"
fi

# ============================================================================
# UPDATE Tests - Attributes
# ============================================================================

test_name "Update lineup scenario name"
TARGET_ID="$SAMPLE_ID"
if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
    TARGET_ID="$CREATED_ID"
fi
if [[ -n "$TARGET_ID" && "$TARGET_ID" != "null" ]]; then
    xbe_json do lineup-scenarios update "$TARGET_ID" --name "CLI update $(date +%s)"
    assert_success
else
    skip "No lineup scenario ID available"
fi

test_name "Update lineup scenario constraints flag"
if [[ -n "$TARGET_ID" && "$TARGET_ID" != "null" ]]; then
    xbe_json do lineup-scenarios update "$TARGET_ID" --include-trucker-assignments-as-constraints=false
    assert_success
else
    skip "No lineup scenario ID available"
fi

# ============================================================================
# UPDATE Tests - Error Cases
# ============================================================================

test_name "Update lineup scenario with no fields fails"
if [[ -n "$TARGET_ID" && "$TARGET_ID" != "null" ]]; then
    xbe_json do lineup-scenarios update "$TARGET_ID"
    assert_failure
else
    skip "No lineup scenario ID available"
fi

# ============================================================================
# DELETE Tests - Guard
# ============================================================================

test_name "Delete lineup scenario requires --confirm flag"
if [[ -n "$TARGET_ID" && "$TARGET_ID" != "null" ]]; then
    xbe_json do lineup-scenarios delete "$TARGET_ID"
    assert_failure
else
    skip "No lineup scenario ID available"
fi

run_tests
