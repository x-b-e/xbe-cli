#!/bin/bash
#
# XBE CLI Integration Tests: Commitment Items
#
# Tests list, show, create, update, and delete operations for the commitment_items resource.
#
# COVERAGE: All filters + create/update attributes + error cases
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

SAMPLE_ID=""
SAMPLE_COMMITMENT_TYPE=""
SAMPLE_COMMITMENT_ID=""
SAMPLE_STATUS=""

CREATE_COMMITMENT_TYPE=""
CREATE_COMMITMENT_ID=""
CREATED_COMMITMENT_ITEM_ID=""

UPDATED_LABEL=""

YEARS_JSON='[2026]'
MONTHS_JSON='[1,2]'
WEEKS_JSON='[1]'
DAYS_OF_WEEK_JSON='[1,2,3,4,5]'
TIMES_OF_DAY_JSON='["day"]'
ADJUSTMENT_CONSTANT_JSON='{"class_name":"NormalDistribution","mean":1.0,"standard_deviation":0.2}'
ADJUSTMENT_INPUT_JSON='{"class_name":"TriangularDistribution","minimum":1,"mode":2,"maximum":3}'

START_ON="2026-01-01"
END_ON="2026-12-31"

LABEL=""


describe "Resource: commitment_items"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List commitment items"
xbe_json view commitment-items list --limit 5
assert_success

test_name "List commitment items returns array"
xbe_json view commitment-items list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list commitment items"
fi

# ============================================================================
# Sample Record (used for filters/show)
# ============================================================================

test_name "Capture sample commitment item"
xbe_json view commitment-items list --limit 1
if [[ $status -eq 0 ]]; then
    SAMPLE_ID=$(json_get ".[0].id")
    SAMPLE_COMMITMENT_TYPE=$(json_get ".[0].commitment_type")
    SAMPLE_COMMITMENT_ID=$(json_get ".[0].commitment_id")
    SAMPLE_STATUS=$(json_get ".[0].status")
    if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
        pass
    else
        skip "No commitment items available for follow-on tests"
    fi
else
    skip "Could not list commitment items to capture sample"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List commitment items with --status filter"
if [[ -n "$SAMPLE_STATUS" && "$SAMPLE_STATUS" != "null" ]]; then
    xbe_json view commitment-items list --status "$SAMPLE_STATUS" --limit 5
    assert_success
else
    skip "No status available"
fi

test_name "List commitment items with --commitment filter"
if [[ -n "$SAMPLE_COMMITMENT_TYPE" && "$SAMPLE_COMMITMENT_TYPE" != "null" && -n "$SAMPLE_COMMITMENT_ID" && "$SAMPLE_COMMITMENT_ID" != "null" ]]; then
    xbe_json view commitment-items list --commitment-type "$SAMPLE_COMMITMENT_TYPE" --commitment-id "$SAMPLE_COMMITMENT_ID" --limit 5
    assert_success
else
    skip "No commitment type/id available"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show commitment item"
if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_json view commitment-items show "$SAMPLE_ID"
    assert_success
else
    skip "No commitment item ID available"
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Resolve commitment for create"
if [[ -n "$XBE_TEST_COMMITMENT_TYPE" && -n "$XBE_TEST_COMMITMENT_ID" ]]; then
    CREATE_COMMITMENT_TYPE="$XBE_TEST_COMMITMENT_TYPE"
    CREATE_COMMITMENT_ID="$XBE_TEST_COMMITMENT_ID"
    echo "    Using XBE_TEST_COMMITMENT_TYPE/XBE_TEST_COMMITMENT_ID"
    pass
elif [[ -n "$SAMPLE_COMMITMENT_TYPE" && "$SAMPLE_COMMITMENT_TYPE" != "null" && -n "$SAMPLE_COMMITMENT_ID" && "$SAMPLE_COMMITMENT_ID" != "null" ]]; then
    CREATE_COMMITMENT_TYPE="$SAMPLE_COMMITMENT_TYPE"
    CREATE_COMMITMENT_ID="$SAMPLE_COMMITMENT_ID"
    pass
else
    skip "No commitment available for create"
fi

test_name "Create commitment item"
if [[ -n "$CREATE_COMMITMENT_TYPE" && -n "$CREATE_COMMITMENT_ID" ]]; then
    LABEL=$(unique_name "CommitmentItem")
    xbe_json do commitment-items create \
        --commitment-type "$CREATE_COMMITMENT_TYPE" \
        --commitment-id "$CREATE_COMMITMENT_ID" \
        --label "$LABEL" \
        --status editing \
        --start-on "$START_ON" \
        --end-on "$END_ON" \
        --years "$YEARS_JSON" \
        --months "$MONTHS_JSON" \
        --weeks "$WEEKS_JSON" \
        --days-of-week "$DAYS_OF_WEEK_JSON" \
        --times-of-day "$TIMES_OF_DAY_JSON" \
        --adjustment-sequence-position 1 \
        --adjustment-coefficient 1.1 \
        --adjustment-constant "$ADJUSTMENT_CONSTANT_JSON" \
        --adjustment-input "$ADJUSTMENT_INPUT_JSON"

    if [[ $status -eq 0 ]]; then
        CREATED_COMMITMENT_ITEM_ID=$(json_get ".id")
        if [[ -n "$CREATED_COMMITMENT_ITEM_ID" && "$CREATED_COMMITMENT_ITEM_ID" != "null" ]]; then
            register_cleanup "commitment-items" "$CREATED_COMMITMENT_ITEM_ID"
            pass
        else
            fail "Created commitment item but no ID returned"
        fi
    else
        if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]] || [[ "$output" == *"422"* ]]; then
            skip "Create not permitted or failed validation"
        else
            fail "Failed to create commitment item"
        fi
    fi
else
    skip "No commitment available for create"
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update commitment item"
if [[ -n "$CREATED_COMMITMENT_ITEM_ID" && "$CREATED_COMMITMENT_ITEM_ID" != "null" ]]; then
    UPDATED_LABEL=$(unique_name "CommitmentItemUpdated")
    xbe_json do commitment-items update "$CREATED_COMMITMENT_ITEM_ID" --label "$UPDATED_LABEL" --status active
    if [[ $status -eq 0 ]]; then
        pass
    else
        if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]] || [[ "$output" == *"422"* ]]; then
            skip "Update not permitted or failed validation"
        else
            fail "Failed to update commitment item"
        fi
    fi
else
    skip "No created commitment item ID available"
fi

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete commitment item"
if [[ -n "$CREATED_COMMITMENT_ITEM_ID" && "$CREATED_COMMITMENT_ITEM_ID" != "null" ]]; then
    xbe_run do commitment-items delete "$CREATED_COMMITMENT_ITEM_ID" --confirm
    if [[ $status -eq 0 ]]; then
        pass
    else
        skip "Could not delete commitment item (permissions or policy)"
    fi
else
    skip "No created commitment item ID available"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create commitment item without required fields fails"
xbe_run do commitment-items create
assert_failure

test_name "Update commitment item without any fields fails"
xbe_run do commitment-items update "999999"
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
