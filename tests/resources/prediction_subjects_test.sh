#!/bin/bash
#
# XBE CLI Integration Tests: Prediction Subjects
#
# Tests list, show, create, update, delete operations for prediction-subjects.
#
# COVERAGE: All list filters + create/update attributes + delete confirmation
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_BROKER_ID=""
CREATED_SUBJECT_ID=""
CREATED_BUSINESS_UNIT_ID=""
CREATED_BY_ID=""
LIST_SUPPORTED="false"

PREDICTIONS_DUE_AT=""
ACTUAL_DUE_AT=""
UPDATED_NAME=""
REFERENCE_NUMBER=""
ACTUAL_VALUE=""


describe "Resource: prediction-subjects"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List prediction subjects"
xbe_json view prediction-subjects list --limit 5
if [[ $status -eq 0 ]]; then
    LIST_SUPPORTED="true"
    pass
else
    if [[ "$output" == *"404"* ]]; then
        skip "Prediction subjects list endpoint not available"
    else
        fail "Expected success (exit 0), got exit $status"
    fi
fi

test_name "List prediction subjects returns array"
if [[ "$LIST_SUPPORTED" == "true" ]]; then
    xbe_json view prediction-subjects list --limit 5
    if [[ $status -eq 0 ]]; then
        assert_json_is_array
    else
        fail "Failed to list prediction subjects"
    fi
else
    skip "Prediction subjects list endpoint not available"
fi

# ============================================================================
# Prerequisites - Create broker
# ============================================================================

test_name "Create prerequisite broker"
BROKER_NAME=$(unique_name "PredictionSubjectBroker")

xbe_json do brokers create --name "$BROKER_NAME"

if [[ $status -eq 0 ]]; then
    CREATED_BROKER_ID=$(json_get ".id")
    if [[ -n "$CREATED_BROKER_ID" && "$CREATED_BROKER_ID" != "null" ]]; then
        register_cleanup "brokers" "$CREATED_BROKER_ID"
        pass
    else
        fail "Created broker but no ID returned"
        echo "Cannot continue without a broker"
        run_tests
    fi
else
    if [[ -n "$XBE_TEST_BROKER_ID" ]]; then
        CREATED_BROKER_ID="$XBE_TEST_BROKER_ID"
        echo "    Using XBE_TEST_BROKER_ID: $CREATED_BROKER_ID"
        pass
    else
        fail "Failed to create broker and XBE_TEST_BROKER_ID not set"
        echo "Cannot continue without a broker"
        run_tests
    fi
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create prediction subject without required flags fails"
xbe_run do prediction-subjects create --name "Missing Parent"
assert_failure

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create prediction subject"
PREDICTIONS_DUE_AT="2026-02-01"
ACTUAL_DUE_AT="2026-02-10"
REFERENCE_NUMBER="ref-$(date +%s | tail -c 6)"
ACTUAL_VALUE="150.25"

xbe_json do prediction-subjects create \
    --name "$(unique_name "PredictionSubject")" \
    --parent-type brokers \
    --parent-id "$CREATED_BROKER_ID" \
    --broker "$CREATED_BROKER_ID" \
    --status active \
    --kind lowest_losing_bid \
    --predictions-due-at "$PREDICTIONS_DUE_AT" \
    --actual-due-at "$ACTUAL_DUE_AT"

if [[ $status -eq 0 ]]; then
    CREATED_SUBJECT_ID=$(json_get ".id")
    if [[ -n "$CREATED_SUBJECT_ID" && "$CREATED_SUBJECT_ID" != "null" ]]; then
        register_cleanup "prediction-subjects" "$CREATED_SUBJECT_ID"
        pass
    else
        fail "Created prediction subject but no ID returned"
        echo "Cannot continue without a prediction subject"
        run_tests
    fi
else
    fail "Failed to create prediction subject"
    run_tests
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show prediction subject"
if [[ -n "$CREATED_SUBJECT_ID" && "$CREATED_SUBJECT_ID" != "null" ]]; then
    xbe_json view prediction-subjects show "$CREATED_SUBJECT_ID"
    if [[ $status -eq 0 ]]; then
        CREATED_BY_ID=$(json_get ".created_by_id")
        pass
    else
        fail "Failed to show prediction subject"
    fi
else
    skip "No prediction subject ID available"
fi

# ============================================================================
# UPDATE Tests - Attributes
# ============================================================================

test_name "Update prediction subject name"
UPDATED_NAME=$(unique_name "PredictionSubjectUpdated")
xbe_json do prediction-subjects update "$CREATED_SUBJECT_ID" --name "$UPDATED_NAME"
assert_success

test_name "Update prediction subject description"
xbe_json do prediction-subjects update "$CREATED_SUBJECT_ID" --description "Updated description"
assert_success

test_name "Update prediction subject status to abandoned"
xbe_json do prediction-subjects update "$CREATED_SUBJECT_ID" --status abandoned
assert_success

test_name "Update prediction subject status back to active"
xbe_json do prediction-subjects update "$CREATED_SUBJECT_ID" --status active
assert_success

test_name "Update prediction subject kind"
xbe_json do prediction-subjects update "$CREATED_SUBJECT_ID" --kind lowest_losing_bid
assert_success

test_name "Update predictions due at"
PREDICTIONS_DUE_AT="2026-02-05"
xbe_json do prediction-subjects update "$CREATED_SUBJECT_ID" --predictions-due-at "$PREDICTIONS_DUE_AT"
assert_success

test_name "Update actual due at"
ACTUAL_DUE_AT="2026-02-20"
xbe_json do prediction-subjects update "$CREATED_SUBJECT_ID" --actual-due-at "$ACTUAL_DUE_AT"
assert_success

test_name "Update domain min"
xbe_json do prediction-subjects update "$CREATED_SUBJECT_ID" --domain-min 100
assert_success

test_name "Update domain max"
xbe_json do prediction-subjects update "$CREATED_SUBJECT_ID" --domain-max 200
assert_success

test_name "Update additional attributes"
ADDITIONAL_ATTRS=$(printf '{"reference_number":"%s","source":"cli"}' "$REFERENCE_NUMBER")
xbe_json do prediction-subjects update "$CREATED_SUBJECT_ID" --additional-attributes "$ADDITIONAL_ATTRS"
assert_success

test_name "Update reference number"
xbe_json do prediction-subjects update "$CREATED_SUBJECT_ID" --reference-number "$REFERENCE_NUMBER"
assert_success

test_name "Update actual value"
xbe_json do prediction-subjects update "$CREATED_SUBJECT_ID" --actual "$ACTUAL_VALUE"
assert_success

# ============================================================================
# UPDATE Tests - Relationships
# ============================================================================

test_name "Create business unit for relationship update"
BUSINESS_UNIT_NAME=$(unique_name "PredictionSubjectBU")
xbe_json do business-units create --name "$BUSINESS_UNIT_NAME" --broker "$CREATED_BROKER_ID"
if [[ $status -eq 0 ]]; then
    CREATED_BUSINESS_UNIT_ID=$(json_get ".id")
    if [[ -n "$CREATED_BUSINESS_UNIT_ID" && "$CREATED_BUSINESS_UNIT_ID" != "null" ]]; then
        register_cleanup "business-units" "$CREATED_BUSINESS_UNIT_ID"
        pass
    else
        fail "Created business unit but no ID returned"
    fi
else
    skip "Failed to create business unit for relationship update"
fi

test_name "Update business unit relationship"
if [[ -n "$CREATED_BUSINESS_UNIT_ID" && "$CREATED_BUSINESS_UNIT_ID" != "null" ]]; then
    xbe_json do prediction-subjects update "$CREATED_SUBJECT_ID" --business-unit "$CREATED_BUSINESS_UNIT_ID"
    assert_success
else
    skip "No business unit available for relationship update"
fi

# Prediction consensus and parent updates require existing prediction/project IDs.

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List prediction subjects with --name filter"
xbe_json view prediction-subjects list --name "$UPDATED_NAME" --limit 10
assert_success

test_name "List prediction subjects with --status filter"
xbe_json view prediction-subjects list --status active --limit 10
assert_success

test_name "List prediction subjects with --broker filter"
xbe_json view prediction-subjects list --broker "$CREATED_BROKER_ID" --limit 10
assert_success

test_name "List prediction subjects with --parent filter"
xbe_json view prediction-subjects list --parent "Broker|$CREATED_BROKER_ID" --limit 10
assert_success

test_name "List prediction subjects with --business-unit filter"
if [[ -n "$CREATED_BUSINESS_UNIT_ID" && "$CREATED_BUSINESS_UNIT_ID" != "null" ]]; then
    xbe_json view prediction-subjects list --business-unit "$CREATED_BUSINESS_UNIT_ID" --limit 10
    assert_success
else
    skip "No business unit available for filter test"
fi

test_name "List prediction subjects with --created-by filter"
if [[ -n "$CREATED_BY_ID" && "$CREATED_BY_ID" != "null" ]]; then
    xbe_json view prediction-subjects list --created-by "$CREATED_BY_ID" --limit 10
    assert_success
else
    skip "No created-by ID available for filter test"
fi

test_name "List prediction subjects with --predictions-due-at filter"
xbe_json view prediction-subjects list --predictions-due-at "$PREDICTIONS_DUE_AT" --limit 10
assert_success

test_name "List prediction subjects with --predictions-due-at-min filter"
xbe_json view prediction-subjects list --predictions-due-at-min "2026-01-01" --limit 10
assert_success

test_name "List prediction subjects with --predictions-due-at-max filter"
xbe_json view prediction-subjects list --predictions-due-at-max "2026-12-31" --limit 10
assert_success

test_name "List prediction subjects with --has-predictions-due-at filter"
xbe_json view prediction-subjects list --has-predictions-due-at true --limit 10
assert_success

test_name "List prediction subjects with --actual-due-at filter"
xbe_json view prediction-subjects list --actual-due-at "$ACTUAL_DUE_AT" --limit 10
assert_success

test_name "List prediction subjects with --actual-due-at-min filter"
xbe_json view prediction-subjects list --actual-due-at-min "2026-01-01" --limit 10
assert_success

test_name "List prediction subjects with --actual-due-at-max filter"
xbe_json view prediction-subjects list --actual-due-at-max "2026-12-31" --limit 10
assert_success

test_name "List prediction subjects with --has-actual-due-at filter"
xbe_json view prediction-subjects list --has-actual-due-at true --limit 10
assert_success

test_name "List prediction subjects with --actual filter"
xbe_json view prediction-subjects list --actual "$ACTUAL_VALUE" --limit 10
assert_success

test_name "List prediction subjects with --actual-min filter"
xbe_json view prediction-subjects list --actual-min 100 --limit 10
assert_success

test_name "List prediction subjects with --actual-max filter"
xbe_json view prediction-subjects list --actual-max 300 --limit 10
assert_success

test_name "List prediction subjects with --reference-number filter"
xbe_json view prediction-subjects list --reference-number "$REFERENCE_NUMBER" --limit 10
assert_success

test_name "List prediction subjects with --tagged-with filter"
xbe_json view prediction-subjects list --tagged-with "cli-test" --limit 10
assert_success

test_name "List prediction subjects with --tagged-with-any filter"
xbe_json view prediction-subjects list --tagged-with-any "cli-test,cli-test-2" --limit 10
assert_success

test_name "List prediction subjects with --tagged-with-all filter"
xbe_json view prediction-subjects list --tagged-with-all "cli-test,cli-test-2" --limit 10
assert_success

test_name "List prediction subjects with --in-tag-category filter"
xbe_json view prediction-subjects list --in-tag-category "test-category" --limit 10
assert_success

test_name "List prediction subjects with --bidder filter"
xbe_json view prediction-subjects list --bidder "1" --limit 10
assert_success

test_name "List prediction subjects with --lowest-bid-amount-min filter"
xbe_json view prediction-subjects list --lowest-bid-amount-min "100" --limit 10
assert_success

test_name "List prediction subjects with --lowest-bid-amount-max filter"
xbe_json view prediction-subjects list --lowest-bid-amount-max "1000" --limit 10
assert_success

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete prediction subject requires --confirm flag"
xbe_run do prediction-subjects delete "$CREATED_SUBJECT_ID"
assert_failure

test_name "Delete prediction subject with --confirm"
# Create a prediction subject specifically for deletion
DEL_NAME=$(unique_name "PredictionSubjectDelete")
xbe_json do prediction-subjects create \
    --name "$DEL_NAME" \
    --parent-type brokers \
    --parent-id "$CREATED_BROKER_ID" \
    --status active
if [[ $status -eq 0 ]]; then
    DEL_ID=$(json_get ".id")
    xbe_run do prediction-subjects delete "$DEL_ID" --confirm
    assert_success
else
    skip "Could not create prediction subject for deletion test"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
