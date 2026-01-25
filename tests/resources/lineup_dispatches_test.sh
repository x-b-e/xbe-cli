#!/bin/bash
#
# XBE CLI Integration Tests: Lineup Dispatches
#
# Tests list/show for lineup-dispatches resource and updates writable attributes.
#
# COVERAGE: List filters + show + update attributes
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

SAMPLE_ID=""
SAMPLE_LINEUP_ID=""

CREATED_ID=""

describe "Resource: lineup-dispatches"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List lineup dispatches"
xbe_json view lineup-dispatches list --limit 5
assert_success

test_name "List lineup dispatches returns array"
xbe_json view lineup-dispatches list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list lineup dispatches"
fi

# ============================================================================
# Sample Record (used for filters/show/update)
# ============================================================================

test_name "Capture sample lineup dispatch"
xbe_json view lineup-dispatches list --limit 1
if [[ $status -eq 0 ]]; then
    SAMPLE_ID=$(json_get ".[0].id")
    SAMPLE_LINEUP_ID=$(json_get ".[0].lineup_id")
    if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
        pass
    else
        skip "No lineup dispatches available for follow-on tests"
    fi
else
    skip "Could not list lineup dispatches to capture sample"
fi

# ============================================================================
# CREATE Tests (optional via XBE_TEST_LINEUP_ID)
# ============================================================================

test_name "Create lineup dispatch (optional)"
if [[ -n "$XBE_TEST_LINEUP_ID" ]]; then
    xbe_json do lineup-dispatches create --lineup "$XBE_TEST_LINEUP_ID" --comment "CLI test dispatch $(date +%s)"
    if [[ $status -eq 0 ]]; then
        CREATED_ID=$(json_get ".id")
        if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
            register_cleanup "lineup-dispatches" "$CREATED_ID"
            pass
        else
            fail "Created lineup dispatch but no ID returned"
        fi
    else
        fail "Failed to create lineup dispatch"
    fi
else
    skip "XBE_TEST_LINEUP_ID not set"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List lineup dispatches with --lineup filter"
if [[ -n "$SAMPLE_LINEUP_ID" && "$SAMPLE_LINEUP_ID" != "null" ]]; then
    xbe_json view lineup-dispatches list --lineup "$SAMPLE_LINEUP_ID" --limit 5
    assert_success
else
    skip "No lineup ID available"
fi

test_name "List lineup dispatches with --created-at-min filter"
xbe_json view lineup-dispatches list --created-at-min "2020-01-01T00:00:00Z" --limit 5
assert_success

test_name "List lineup dispatches with --created-at-max filter"
xbe_json view lineup-dispatches list --created-at-max "2030-01-01T00:00:00Z" --limit 5
assert_success

test_name "List lineup dispatches with --updated-at-min filter"
xbe_json view lineup-dispatches list --updated-at-min "2020-01-01T00:00:00Z" --limit 5
assert_success

test_name "List lineup dispatches with --updated-at-max filter"
xbe_json view lineup-dispatches list --updated-at-max "2030-01-01T00:00:00Z" --limit 5
assert_success

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show lineup dispatch"
if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_json view lineup-dispatches show "$SAMPLE_ID"
    assert_success
else
    skip "No lineup dispatch ID available"
fi

# ============================================================================
# UPDATE Tests - Attributes
# ============================================================================

test_name "Update lineup dispatch comment"
if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_json do lineup-dispatches update "$SAMPLE_ID" --comment "CLI update $(date +%s)"
    assert_success
else
    skip "No lineup dispatch ID available"
fi

test_name "Update lineup dispatch tender settings"
if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_json do lineup-dispatches update "$SAMPLE_ID" \
        --auto-offer-customer-tenders=false \
        --auto-offer-trucker-tenders=false \
        --auto-accept-trucker-tenders=true
    assert_success
else
    skip "No lineup dispatch ID available"
fi

# ============================================================================
# UPDATE Tests - Error Cases
# ============================================================================

test_name "Update lineup dispatch with no fields fails"
if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_json do lineup-dispatches update "$SAMPLE_ID"
    assert_failure
else
    skip "No lineup dispatch ID available"
fi

# ============================================================================
# DELETE Tests - Guard
# ============================================================================

test_name "Delete lineup dispatch requires --confirm flag"
if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_json do lineup-dispatches delete "$SAMPLE_ID"
    assert_failure
else
    skip "No lineup dispatch ID available"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
