#!/bin/bash
#
# XBE CLI Integration Tests: Key Results
#
# Tests list/show/create/update/delete operations for key-results.
#
# COVERAGE: List filters + writable attributes
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

OBJECTIVE_ID="${XBE_TEST_KEY_RESULT_OBJECTIVE_ID:-}"
OWNER_ID="${XBE_TEST_KEY_RESULT_OWNER_ID:-}"
CS_RESP_ID="${XBE_TEST_KEY_RESULT_CUSTOMER_SUCCESS_RESPONSIBLE_PERSON_ID:-}"
KEY_RESULT_ID="${XBE_TEST_KEY_RESULT_ID:-}"

SAMPLE_ID=""
CREATED_ID=""

handle_write_failure() {
    local output_text="$1"
    if [[ "$output_text" == *"Not Authorized"* ]] || [[ "$output_text" == *"not authorized"* ]] || [[ "$output_text" == *"422"* ]] || [[ "$output_text" == *"409"* ]]; then
        skip "Write blocked by server policy/validation"
        return 0
    fi
    fail "Write failed: $output_text"
    return 1
}

describe "Resource: key-results"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List key results"
xbe_json view key-results list --limit 5
assert_success

test_name "List key results returns array"
xbe_json view key-results list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list key results"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List key results with --status filter"
xbe_json view key-results list --status green --limit 5
assert_success

test_name "List key results with --is-template filter"
xbe_json view key-results list --is-template true --limit 5
assert_success

test_name "List key results with --has-customer-success-responsible-person filter"
xbe_json view key-results list --has-customer-success-responsible-person true --limit 5
assert_success

test_name "List key results with --objective filter"
if [[ -n "$OBJECTIVE_ID" && "$OBJECTIVE_ID" != "null" ]]; then
    xbe_json view key-results list --objective "$OBJECTIVE_ID" --limit 5
    assert_success
else
    skip "No objective ID available. Set XBE_TEST_KEY_RESULT_OBJECTIVE_ID to enable objective filter testing."
fi

test_name "List key results with --owner filter"
if [[ -n "$OWNER_ID" && "$OWNER_ID" != "null" ]]; then
    xbe_json view key-results list --owner "$OWNER_ID" --limit 5
    assert_success
else
    skip "No owner ID available. Set XBE_TEST_KEY_RESULT_OWNER_ID to enable owner filter testing."
fi

test_name "List key results with --customer-success-responsible-person filter"
if [[ -n "$CS_RESP_ID" && "$CS_RESP_ID" != "null" ]]; then
    xbe_json view key-results list --customer-success-responsible-person "$CS_RESP_ID" --limit 5
    assert_success
else
    skip "No customer success responsible person ID available. Set XBE_TEST_KEY_RESULT_CUSTOMER_SUCCESS_RESPONSIBLE_PERSON_ID to enable filter testing."
fi

# ============================================================================
# Sample Record (used for show)
# ============================================================================

test_name "Capture sample key result"
xbe_json view key-results list --limit 1
if [[ $status -eq 0 ]]; then
    SAMPLE_ID=$(json_get ".[0].id")
    if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
        pass
    else
        skip "No key results available for show test"
    fi
else
    skip "Could not list key results to capture sample"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show key result"
DETAIL_ID="${KEY_RESULT_ID:-$SAMPLE_ID}"
if [[ -n "$DETAIL_ID" && "$DETAIL_ID" != "null" ]]; then
    xbe_json view key-results show "$DETAIL_ID"
    if [[ $status -eq 0 ]]; then
        assert_json_has ".id"
    else
        if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]] || [[ "$output" == *"Not Found"* ]] || [[ "$output" == *"not found"* ]]; then
            skip "Show blocked by server policy or record not found"
        else
            fail "Failed to show key result: $output"
        fi
    fi
else
    skip "No key result ID available for show"
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create key result"
if [[ -n "$OBJECTIVE_ID" && "$OBJECTIVE_ID" != "null" ]]; then
    TITLE=$(unique_name "KeyResult")
    xbe_json do key-results create --title "$TITLE" --objective "$OBJECTIVE_ID"
    if [[ $status -eq 0 ]]; then
        CREATED_ID=$(json_get ".id")
        if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
            register_cleanup "key-results" "$CREATED_ID"
            pass
        else
            fail "Created key result but no ID returned"
        fi
    else
        handle_write_failure "$output"
    fi
else
    skip "No objective ID available. Set XBE_TEST_KEY_RESULT_OBJECTIVE_ID to enable create/update tests."
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

UPDATE_ID="${CREATED_ID:-$KEY_RESULT_ID}"

update_key_result() {
    local description="$1"
    shift
    test_name "$description"
    xbe_json do key-results update "$UPDATE_ID" "$@"
    if [[ $status -eq 0 ]]; then
        pass
    else
        if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]] || [[ "$output" == *"422"* ]] || [[ "$output" == *"409"* ]]; then
            skip "Update blocked by server policy/validation"
        else
            fail "Update failed: $output"
        fi
    fi
}

if [[ -n "$UPDATE_ID" && "$UPDATE_ID" != "null" ]]; then
    update_key_result "Update title" --title "Updated $(unique_name "KeyResult")"
    update_key_result "Update title summary explicit" --title-summary-explicit "Launch KPI"

    START_DATE=$(date -u +%Y-%m-%d)
    END_DATE=$(date -u -v+30d +%Y-%m-%d 2>/dev/null || date -u -d '+30 days' +%Y-%m-%d)
    update_key_result "Update start/end dates" --start-on "$START_DATE" --end-on "$END_DATE"

    update_key_result "Update completion percentage" --completion-percentage "0.35"

    if [[ -n "$OWNER_ID" && "$OWNER_ID" != "null" ]]; then
        update_key_result "Update owner" --owner "$OWNER_ID"
    else
        test_name "Update owner"
        skip "No owner ID available. Set XBE_TEST_KEY_RESULT_OWNER_ID to enable owner update testing."
    fi

    if [[ -n "$CS_RESP_ID" && "$CS_RESP_ID" != "null" ]]; then
        update_key_result "Update customer success responsible person" --customer-success-responsible-person "$CS_RESP_ID"
    else
        test_name "Update customer success responsible person"
        skip "No customer success responsible person ID available. Set XBE_TEST_KEY_RESULT_CUSTOMER_SUCCESS_RESPONSIBLE_PERSON_ID to enable update testing."
    fi
else
    test_name "Update key result"
    skip "No key result ID available for update tests"
fi

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete key result"
if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
    xbe_json do key-results delete "$CREATED_ID" --confirm
    if [[ $status -eq 0 ]]; then
        pass
    else
        if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]] || [[ "$output" == *"422"* ]] || [[ "$output" == *"409"* ]]; then
            skip "Delete blocked by server policy/validation"
        else
            fail "Failed to delete key result: $output"
        fi
    fi
else
    skip "No created key result ID available for delete test"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
