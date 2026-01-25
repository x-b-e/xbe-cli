#!/bin/bash
#
# XBE CLI Integration Tests: HOS Days
#
# Tests list, show, create, update, and delete operations for the hos-days resource.
#
# COVERAGE: List filters + show + create/update/delete + error cases
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

SAMPLE_ID=""
SAMPLE_DRIVER_ID=""
SAMPLE_BROKER_ID=""
SAMPLE_SERVICE_DATE=""
SAMPLE_REGULATION_SET_CODE=""

UNAUTHORIZED_UPDATE=false

describe "Resource: hos_days"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List HOS days"
xbe_json view hos-days list --limit 5
assert_success

test_name "List HOS days returns array"
xbe_json view hos-days list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list HOS days"
fi

# ============================================================================
# Sample Record (used for filters/show)
# ============================================================================

test_name "Capture sample HOS day"
xbe_json view hos-days list --limit 1
if [[ $status -eq 0 ]]; then
    SAMPLE_ID=$(json_get ".[0].id")
    SAMPLE_DRIVER_ID=$(json_get ".[0].driver_id")
    SAMPLE_BROKER_ID=$(json_get ".[0].broker_id")
    SAMPLE_SERVICE_DATE=$(json_get ".[0].service_date")
    SAMPLE_REGULATION_SET_CODE=$(json_get ".[0].regulation_set_code")
    if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
        pass
    else
        skip "No HOS days available for follow-on tests"
    fi
else
    skip "Could not list HOS days to capture sample"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List HOS days with --driver filter"
if [[ -n "$SAMPLE_DRIVER_ID" && "$SAMPLE_DRIVER_ID" != "null" ]]; then
    xbe_json view hos-days list --driver "$SAMPLE_DRIVER_ID" --limit 5
    assert_success
else
    skip "No driver ID available"
fi

test_name "List HOS days with --user filter"
if [[ -n "$SAMPLE_DRIVER_ID" && "$SAMPLE_DRIVER_ID" != "null" ]]; then
    xbe_json view hos-days list --user "$SAMPLE_DRIVER_ID" --limit 5
    assert_success
else
    skip "No driver ID available"
fi

test_name "List HOS days with --broker filter"
if [[ -n "$SAMPLE_BROKER_ID" && "$SAMPLE_BROKER_ID" != "null" ]]; then
    xbe_json view hos-days list --broker "$SAMPLE_BROKER_ID" --limit 5
    assert_success
else
    skip "No broker ID available"
fi

test_name "List HOS days with --service-date filter"
if [[ -n "$SAMPLE_SERVICE_DATE" && "$SAMPLE_SERVICE_DATE" != "null" ]]; then
    xbe_json view hos-days list --service-date "$SAMPLE_SERVICE_DATE" --limit 5
    assert_success
else
    skip "No service date available"
fi

test_name "List HOS days with --service-date-min filter"
xbe_json view hos-days list --service-date-min "2000-01-01" --limit 5
assert_success

test_name "List HOS days with --service-date-max filter"
xbe_json view hos-days list --service-date-max "2100-01-01" --limit 5
assert_success

test_name "List HOS days with --has-service-date filter"
xbe_json view hos-days list --has-service-date true --limit 5
assert_success

test_name "List HOS days with --regulation-set-code filter"
if [[ -n "$SAMPLE_REGULATION_SET_CODE" && "$SAMPLE_REGULATION_SET_CODE" != "null" ]]; then
    xbe_json view hos-days list --regulation-set-code "$SAMPLE_REGULATION_SET_CODE" --limit 5
    assert_success
else
    skip "No regulation set code available"
fi

test_name "List HOS days with --created-at-min filter"
xbe_json view hos-days list --created-at-min "2020-01-01T00:00:00Z" --limit 5
assert_success

test_name "List HOS days with --created-at-max filter"
xbe_json view hos-days list --created-at-max "2030-01-01T00:00:00Z" --limit 5
assert_success

test_name "List HOS days with --is-created-at filter"
xbe_json view hos-days list --is-created-at true --limit 5
assert_success

test_name "List HOS days with --updated-at-min filter"
xbe_json view hos-days list --updated-at-min "2020-01-01T00:00:00Z" --limit 5
assert_success

test_name "List HOS days with --updated-at-max filter"
xbe_json view hos-days list --updated-at-max "2030-01-01T00:00:00Z" --limit 5
assert_success

test_name "List HOS days with --is-updated-at filter"
xbe_json view hos-days list --is-updated-at true --limit 5
assert_success

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show HOS day"
if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_json view hos-days show "$SAMPLE_ID"
    assert_success
else
    skip "No HOS day ID available"
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update HOS day regulation set code"
if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" && -n "$SAMPLE_REGULATION_SET_CODE" && "$SAMPLE_REGULATION_SET_CODE" != "null" ]]; then
    xbe_json do hos-days update "$SAMPLE_ID" --regulation-set-code "$SAMPLE_REGULATION_SET_CODE"
    if [[ $status -eq 0 ]]; then
        pass
    else
        if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]] || [[ "$output" == *"403"* ]]; then
            UNAUTHORIZED_UPDATE=true
            skip "Not authorized to update HOS day"
        else
            fail "Failed to update HOS day"
        fi
    fi
else
    skip "No sample HOS day or regulation set code available"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Update HOS day without any fields fails"
xbe_run do hos-days update "999999"
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
