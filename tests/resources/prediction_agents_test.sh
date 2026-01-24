#!/bin/bash
#
# XBE CLI Integration Tests: Prediction Agents
#
# Tests list/show/create/update/delete operations for prediction-agents.
#
# COVERAGE: All list filters + writable attributes
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

PREDICTION_SUBJECT_ID="${XBE_TEST_PREDICTION_AGENT_SUBJECT_ID:-}"
if [[ -z "$PREDICTION_SUBJECT_ID" ]]; then
    PREDICTION_SUBJECT_ID="${XBE_TEST_PREDICTION_SUBJECT_ID:-}"
fi

CREATED_BY_ID="${XBE_TEST_USER_ID:-}"

SAMPLE_ID=""
SAMPLE_PREDICTION_SUBJECT_ID=""
SAMPLE_CREATED_BY_ID=""
CREATED_ID=""

CUSTOM_INSTRUCTIONS="Focus on recent performance trends."

update_blocked_message() {
    local msg="$1"
    if [[ "$msg" == *"Not Authorized"* ]] || [[ "$msg" == *"not authorized"* ]] || [[ "$msg" == *"insufficient predictions"* ]] || [[ "$msg" == *"already has a prediction agent"* ]] || [[ "$msg" == *"already exists"* ]] || [[ "$msg" == *"request is already in progress"* ]]; then
        return 0
    fi
    return 1
}

describe "Resource: prediction-agents"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List prediction agents"
xbe_json view prediction-agents list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
    SAMPLE_ID=$(json_get ".[0].id")
    SAMPLE_PREDICTION_SUBJECT_ID=$(json_get ".[0].prediction_subject_id")
    SAMPLE_CREATED_BY_ID=$(json_get ".[0].created_by_id")
else
    fail "Failed to list prediction agents"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show prediction agent"
if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_json view prediction-agents show "$SAMPLE_ID"
    if [[ $status -eq 0 ]]; then
        assert_json_has ".id"
    else
        if update_blocked_message "$output" || [[ "$output" == *"Not Found"* ]] || [[ "$output" == *"not found"* ]]; then
            skip "Show blocked by server policy or record not found"
        else
            fail "Failed to show prediction agent: $output"
        fi
    fi
else
    skip "No prediction agent ID available for show"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List prediction agents with --prediction-subject filter"
FILTER_PREDICTION_SUBJECT_ID="${SAMPLE_PREDICTION_SUBJECT_ID:-$PREDICTION_SUBJECT_ID}"
if [[ -n "$FILTER_PREDICTION_SUBJECT_ID" && "$FILTER_PREDICTION_SUBJECT_ID" != "null" ]]; then
    xbe_json view prediction-agents list --prediction-subject "$FILTER_PREDICTION_SUBJECT_ID" --limit 5
    assert_success
else
    skip "No prediction subject ID available"
fi

if [[ -z "$CREATED_BY_ID" ]]; then
    CREATED_BY_ID="$SAMPLE_CREATED_BY_ID"
fi

if [[ -z "$CREATED_BY_ID" || "$CREATED_BY_ID" == "null" ]]; then
    xbe_json auth whoami
    if [[ $status -eq 0 ]]; then
        CREATED_BY_ID=$(json_get ".id")
    fi
fi

test_name "List prediction agents with --created-by filter"
if [[ -n "$CREATED_BY_ID" && "$CREATED_BY_ID" != "null" ]]; then
    xbe_json view prediction-agents list --created-by "$CREATED_BY_ID" --limit 5
    assert_success
else
    skip "No created-by ID available (set XBE_TEST_USER_ID)"
fi

test_name "List prediction agents with --has-prediction true"
xbe_json view prediction-agents list --has-prediction true --limit 5
assert_success

test_name "List prediction agents with --has-prediction false"
xbe_json view prediction-agents list --has-prediction false --limit 5
assert_success

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create prediction agent requires prediction subject"
xbe_run do prediction-agents create
assert_failure

test_name "Create prediction agent"
if [[ -n "$PREDICTION_SUBJECT_ID" && "$PREDICTION_SUBJECT_ID" != "null" ]]; then
    xbe_json do prediction-agents create --prediction-subject "$PREDICTION_SUBJECT_ID"
    if [[ $status -eq 0 ]]; then
        CREATED_ID=$(json_get ".id")
        if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
            register_cleanup "prediction-agents" "$CREATED_ID"
            pass
        else
            fail "Created prediction agent but no ID returned"
        fi
    else
        if update_blocked_message "$output" || [[ "$output" == *"Not Found"* ]] || [[ "$output" == *"not found"* ]]; then
            skip "Create blocked by server policy or invalid prediction subject"
        else
            fail "Failed to create prediction agent: $output"
        fi
    fi
else
    skip "No prediction subject ID available (set XBE_TEST_PREDICTION_AGENT_SUBJECT_ID or XBE_TEST_PREDICTION_SUBJECT_ID)"
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update prediction agent custom instructions"
if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
    xbe_json do prediction-agents update "$CREATED_ID" --custom-instructions "$CUSTOM_INSTRUCTIONS"
    if [[ $status -eq 0 ]]; then
        assert_success
    else
        if update_blocked_message "$output"; then
            skip "Update blocked by server policy or validation"
        else
            fail "Failed to update prediction agent: $output"
        fi
    fi
else
    skip "No created prediction agent available for update"
fi

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete prediction agent requires --confirm"
if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
    xbe_run do prediction-agents delete "$CREATED_ID"
    assert_failure
else
    skip "No created prediction agent available for delete"
fi

test_name "Delete prediction agent"
if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
    xbe_run do prediction-agents delete "$CREATED_ID" --confirm
    if [[ $status -eq 0 ]]; then
        assert_success
    else
        if update_blocked_message "$output"; then
            skip "Delete blocked by server policy"
        else
            fail "Failed to delete prediction agent: $output"
        fi
    fi
else
    skip "No created prediction agent available for delete"
fi

run_tests
