#!/bin/bash
#
# XBE CLI Integration Tests: Prediction Subject Gaps
#
# Tests CRUD operations for the prediction-subject-gaps resource.
# Requires a prediction subject that satisfies gap prerequisites.
#
# COVERAGE: All create/update attributes + list filters
#

source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

PREDICTION_SUBJECT_ID="${XBE_TEST_PREDICTION_SUBJECT_GAP_PREDICTION_SUBJECT_ID:-${XBE_TEST_PREDICTION_SUBJECT_ID:-}}"
GAP_TYPE="${XBE_TEST_PREDICTION_SUBJECT_GAP_TYPE:-}"
EXISTING_GAP_ID="${XBE_TEST_PREDICTION_SUBJECT_GAP_ID:-}"

CREATED_ID=""

STATUS_FILTER_VALUE="${XBE_TEST_PREDICTION_SUBJECT_GAP_STATUS_FILTER:-pending}"
GAP_TYPE_FILTER_VALUE="${XBE_TEST_PREDICTION_SUBJECT_GAP_GAP_TYPE_FILTER:-actual_vs_consensus}"


describe "Resource: prediction-subject-gaps"

# ==========================================================================
# CREATE Tests
# ==========================================================================

if [[ -n "$PREDICTION_SUBJECT_ID" && -n "$GAP_TYPE" ]]; then
    test_name "Create prediction subject gap"
    xbe_json do prediction-subject-gaps create \
        --prediction-subject "$PREDICTION_SUBJECT_ID" \
        --gap-type "$GAP_TYPE"

    if [[ $status -eq 0 ]]; then
        CREATED_ID=$(json_get ".id")
        if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
            register_cleanup "prediction-subject-gaps" "$CREATED_ID"
            pass
        else
            fail "Created prediction subject gap but no ID returned"
        fi
    else
        fail "Failed to create prediction subject gap"
    fi
else
    test_name "Skip create tests (missing prediction subject gap prerequisites)"
    skip "Set XBE_TEST_PREDICTION_SUBJECT_GAP_PREDICTION_SUBJECT_ID and XBE_TEST_PREDICTION_SUBJECT_GAP_TYPE to run create tests"
fi

# ==========================================================================
# SHOW Tests
# ==========================================================================

if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
    test_name "Show prediction subject gap by ID"
    xbe_json view prediction-subject-gaps show "$CREATED_ID"
    assert_success
elif [[ -n "$EXISTING_GAP_ID" ]]; then
    test_name "Show prediction subject gap by existing ID"
    xbe_json view prediction-subject-gaps show "$EXISTING_GAP_ID"
    assert_success
else
    test_name "Skip show tests (missing prediction subject gap ID)"
    skip "Set XBE_TEST_PREDICTION_SUBJECT_GAP_ID or enable create tests"
fi

# ==========================================================================
# LIST Tests - Basic
# ==========================================================================

test_name "List prediction subject gaps"
xbe_json view prediction-subject-gaps list --limit 5
assert_success

test_name "List prediction subject gaps returns array"
xbe_json view prediction-subject-gaps list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list prediction subject gaps"
fi

# ==========================================================================
# LIST Tests - Filters
# ==========================================================================

if [[ -n "$PREDICTION_SUBJECT_ID" ]]; then
    test_name "List prediction subject gaps with --prediction-subject filter"
    xbe_json view prediction-subject-gaps list --prediction-subject "$PREDICTION_SUBJECT_ID" --limit 10
    assert_success
else
    test_name "Skip --prediction-subject filter (missing prediction subject ID)"
    skip "Set XBE_TEST_PREDICTION_SUBJECT_GAP_PREDICTION_SUBJECT_ID to test prediction subject filter"
fi

test_name "List prediction subject gaps with --status filter"
xbe_json view prediction-subject-gaps list --status "$STATUS_FILTER_VALUE" --limit 10
assert_success

test_name "List prediction subject gaps with --gap-type filter"
xbe_json view prediction-subject-gaps list --gap-type "$GAP_TYPE_FILTER_VALUE" --limit 10
assert_success

# ==========================================================================
# LIST Tests - Pagination
# ==========================================================================

test_name "List prediction subject gaps with --limit"
xbe_json view prediction-subject-gaps list --limit 3
assert_success

test_name "List prediction subject gaps with --offset"
xbe_json view prediction-subject-gaps list --limit 3 --offset 3
assert_success

# ==========================================================================
# UPDATE Tests
# ==========================================================================

if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
    test_name "Update prediction subject gap status"
    xbe_json do prediction-subject-gaps update "$CREATED_ID" --status approved
    assert_success

    test_name "Update prediction subject gap type"
    xbe_json do prediction-subject-gaps update "$CREATED_ID" --gap-type "$GAP_TYPE"
    assert_success
else
    test_name "Skip update tests (missing created gap)"
    skip "Enable create tests to run update"
fi

# ==========================================================================
# DELETE Tests
# ==========================================================================

if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
    test_name "Delete prediction subject gap requires --confirm flag"
    xbe_run do prediction-subject-gaps delete "$CREATED_ID"
    assert_failure

    test_name "Delete prediction subject gap with --confirm"
    xbe_run do prediction-subject-gaps delete "$CREATED_ID" --confirm
    assert_success
else
    test_name "Skip delete tests (missing created gap)"
    skip "Enable create tests to run delete"
fi

# ==========================================================================
# Error Cases
# ==========================================================================

test_name "Create prediction subject gap without prediction subject fails"
xbe_run do prediction-subject-gaps create --gap-type "actual_vs_consensus"
assert_failure

test_name "Create prediction subject gap without gap type fails"
xbe_run do prediction-subject-gaps create --prediction-subject "1"
assert_failure

test_name "Update prediction subject gap without fields fails"
xbe_run do prediction-subject-gaps update 1
assert_failure

test_name "Delete prediction subject gap without --confirm fails"
xbe_run do prediction-subject-gaps delete 1
assert_failure

run_tests
