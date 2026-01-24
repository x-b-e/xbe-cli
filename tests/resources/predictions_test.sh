#!/bin/bash
#
# XBE CLI Integration Tests: Predictions
#
# Tests CRUD operations for the predictions resource.
# Requires a prediction subject that allows creating predictions.
#
# COVERAGE: All create/update attributes + list filters
#

source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

PREDICTION_SUBJECT_ID="${XBE_TEST_PREDICTION_PREDICTION_SUBJECT_ID:-${XBE_TEST_PREDICTION_SUBJECT_ID:-}}"
PREDICTED_BY_ID="${XBE_TEST_PREDICTION_PREDICTED_BY_ID:-}"
PREDICTION_AGENT_ID="${XBE_TEST_PREDICTION_AGENT_ID:-}"
EXISTING_PREDICTION_ID="${XBE_TEST_PREDICTION_ID:-}"

STATUS_CREATE="${XBE_TEST_PREDICTION_STATUS_CREATE:-draft}"
STATUS_UPDATE="${XBE_TEST_PREDICTION_STATUS_UPDATE:-submitted}"
STATUS_FILTER_VALUE="${XBE_TEST_PREDICTION_STATUS_FILTER:-submitted}"

DEFAULT_DISTRIBUTION='{"class_name":"TriangularDistribution","minimum":100,"mode":120,"maximum":140}'
DEFAULT_UPDATE_DISTRIBUTION='{"class_name":"TriangularDistribution","minimum":110,"mode":130,"maximum":150}'
PROBABILITY_DISTRIBUTION="${XBE_TEST_PREDICTION_PROBABILITY_DISTRIBUTION:-$DEFAULT_DISTRIBUTION}"
UPDATE_PROBABILITY_DISTRIBUTION="${XBE_TEST_PREDICTION_PROBABILITY_DISTRIBUTION_UPDATE:-$DEFAULT_UPDATE_DISTRIBUTION}"

CREATED_ID=""

describe "Resource: predictions"

# ==========================================================================
# CREATE Tests
# ==========================================================================

if [[ -n "$PREDICTION_SUBJECT_ID" ]]; then
    test_name "Create prediction"
    create_args=(do predictions create --prediction-subject "$PREDICTION_SUBJECT_ID" --status "$STATUS_CREATE" --probability-distribution "$PROBABILITY_DISTRIBUTION")
    if [[ -n "$PREDICTED_BY_ID" ]]; then
        create_args+=(--predicted-by "$PREDICTED_BY_ID")
    fi
    xbe_json "${create_args[@]}"

    if [[ $status -eq 0 ]]; then
        CREATED_ID=$(json_get ".id")
        if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
            register_cleanup "predictions" "$CREATED_ID"
            pass
        else
            fail "Created prediction but no ID returned"
        fi
    else
        fail "Failed to create prediction"
    fi
else
    test_name "Skip create tests (missing prediction subject)"
    skip "Set XBE_TEST_PREDICTION_PREDICTION_SUBJECT_ID to run create tests"
fi

# ==========================================================================
# SHOW Tests
# ==========================================================================

if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
    test_name "Show prediction by ID"
    xbe_json view predictions show "$CREATED_ID"
    assert_success
elif [[ -n "$EXISTING_PREDICTION_ID" ]]; then
    test_name "Show prediction by existing ID"
    xbe_json view predictions show "$EXISTING_PREDICTION_ID"
    assert_success
else
    test_name "Skip show tests (missing prediction ID)"
    skip "Set XBE_TEST_PREDICTION_ID or enable create tests"
fi

# ==========================================================================
# LIST Tests - Basic
# ==========================================================================

test_name "List predictions"
xbe_json view predictions list --limit 5
assert_success

test_name "List predictions returns array"
xbe_json view predictions list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list predictions"
fi

# ==========================================================================
# LIST Tests - Filters
# ==========================================================================

if [[ -n "$PREDICTION_SUBJECT_ID" ]]; then
    test_name "List predictions with --prediction-subject filter"
    xbe_json view predictions list --prediction-subject "$PREDICTION_SUBJECT_ID" --limit 10
    assert_success
else
    test_name "Skip --prediction-subject filter (missing prediction subject ID)"
    skip "Set XBE_TEST_PREDICTION_PREDICTION_SUBJECT_ID to test prediction subject filter"
fi

if [[ -n "$PREDICTED_BY_ID" ]]; then
    test_name "List predictions with --predicted-by filter"
    xbe_json view predictions list --predicted-by "$PREDICTED_BY_ID" --limit 10
    assert_success
else
    test_name "Skip --predicted-by filter (missing predicted-by ID)"
    skip "Set XBE_TEST_PREDICTION_PREDICTED_BY_ID to test predicted-by filter"
fi

if [[ -n "$PREDICTION_AGENT_ID" ]]; then
    test_name "List predictions with --prediction-agent filter"
    xbe_json view predictions list --prediction-agent "$PREDICTION_AGENT_ID" --limit 10
    assert_success
else
    test_name "Skip --prediction-agent filter (missing prediction agent ID)"
    skip "Set XBE_TEST_PREDICTION_AGENT_ID to test prediction agent filter"
fi

test_name "List predictions with --status filter"
xbe_json view predictions list --status "$STATUS_FILTER_VALUE" --limit 10
assert_success

# ==========================================================================
# LIST Tests - Pagination
# ==========================================================================

test_name "List predictions with --limit"
xbe_json view predictions list --limit 3
assert_success

test_name "List predictions with --offset"
xbe_json view predictions list --limit 3 --offset 3
assert_success

# ==========================================================================
# UPDATE Tests
# ==========================================================================

if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
    test_name "Update prediction status"
    xbe_json do predictions update "$CREATED_ID" --status "$STATUS_UPDATE"
    assert_success

    test_name "Update prediction probability distribution"
    xbe_json do predictions update "$CREATED_ID" --probability-distribution "$UPDATE_PROBABILITY_DISTRIBUTION"
    assert_success
else
    test_name "Skip update tests (missing created prediction)"
    skip "Enable create tests to run update"
fi

# ==========================================================================
# DELETE Tests
# ==========================================================================

if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
    test_name "Delete prediction requires --confirm flag"
    xbe_run do predictions delete "$CREATED_ID"
    assert_failure

    test_name "Delete prediction with --confirm"
    xbe_run do predictions delete "$CREATED_ID" --confirm
    assert_success
else
    test_name "Skip delete tests (missing created prediction)"
    skip "Enable create tests to run delete"
fi

# ==========================================================================
# Error Cases
# ==========================================================================

test_name "Create prediction without prediction subject fails"
xbe_run do predictions create --status draft
assert_failure

test_name "Update prediction without fields fails"
xbe_run do predictions update 1
assert_failure

test_name "Delete prediction without --confirm fails"
xbe_run do predictions delete 1
assert_failure

run_tests
