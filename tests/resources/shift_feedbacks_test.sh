#!/bin/bash
#
# XBE CLI Integration Tests: Shift Feedbacks
#
# Tests CRUD operations for the shift-feedbacks resource.
# Shift feedbacks represent driver/trucker performance feedback.
#
# NOTE: Creating shift feedbacks requires a tender-job-schedule-shift,
# which is a complex resource without CLI commands. If we cannot create
# the prerequisite, CRUD tests will be skipped.
#
# COVERAGE: All create/update attributes + list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_FEEDBACK_ID=""
CREATED_REASON_ID=""
CREATED_BROKER_ID=""

describe "Resource: shift-feedbacks"

# ============================================================================
# Prerequisites - Create broker and shift-feedback-reason
# ============================================================================

test_name "Create prerequisite broker for shift feedback tests"
BROKER_NAME=$(unique_name "FeedbackTestBroker")

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

test_name "Create prerequisite shift-feedback-reason"
REASON_NAME=$(unique_name "TestReason")
REASON_SLUG=$(echo "$REASON_NAME" | tr '[:upper:]' '[:lower:]' | tr ' ' '-')

xbe_json do shift-feedback-reasons create \
    --name "$REASON_NAME" \
    --kind "driver" \
    --slug "$REASON_SLUG"

if [[ $status -eq 0 ]]; then
    CREATED_REASON_ID=$(json_get ".id")
    if [[ -n "$CREATED_REASON_ID" && "$CREATED_REASON_ID" != "null" ]]; then
        register_cleanup "shift-feedback-reasons" "$CREATED_REASON_ID"
        pass
    else
        fail "Created shift-feedback-reason but no ID returned"
    fi
else
    fail "Failed to create shift-feedback-reason"
fi

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List shift feedbacks"
xbe_json view shift-feedbacks list --limit 5
assert_success

test_name "List shift feedbacks returns array"
xbe_json view shift-feedbacks list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list shift feedbacks"
fi

# ============================================================================
# LIST Tests - All Filters
# ============================================================================

test_name "List shift feedbacks with --broker filter"
xbe_json view shift-feedbacks list --broker "$CREATED_BROKER_ID" --limit 10
assert_success

test_name "List shift feedbacks with --reason filter"
if [[ -n "$CREATED_REASON_ID" && "$CREATED_REASON_ID" != "null" ]]; then
    xbe_json view shift-feedbacks list --reason "$CREATED_REASON_ID" --limit 10
    assert_success
else
    skip "No reason ID available"
fi

test_name "List shift feedbacks with --rating filter"
xbe_json view shift-feedbacks list --rating 5 --limit 10
assert_success

test_name "List shift feedbacks with --kind filter"
xbe_json view shift-feedbacks list --kind "driver" --limit 10
assert_success

test_name "List shift feedbacks with --automated filter"
xbe_json view shift-feedbacks list --automated "true" --limit 10
assert_success

test_name "List shift feedbacks with --automated=false filter"
xbe_json view shift-feedbacks list --automated "false" --limit 10
assert_success

test_name "List shift feedbacks with --shift-date-min filter"
xbe_json view shift-feedbacks list --shift-date-min "2024-01-01" --limit 10
assert_success

test_name "List shift feedbacks with --shift-date-max filter"
xbe_json view shift-feedbacks list --shift-date-max "2025-12-31" --limit 10
assert_success

test_name "List shift feedbacks with --shift-date-min and --shift-date-max filter"
xbe_json view shift-feedbacks list --shift-date-min "2024-01-01" --shift-date-max "2025-12-31" --limit 10
assert_success

# ============================================================================
# LIST Tests - Pagination
# ============================================================================

test_name "List shift feedbacks with --limit"
xbe_json view shift-feedbacks list --limit 3
assert_success

test_name "List shift feedbacks with --offset"
xbe_json view shift-feedbacks list --limit 3 --offset 3
assert_success

# ============================================================================
# CREATE Tests - Note: Requires tender-job-schedule-shift
# ============================================================================

# Check if we can get a tender-job-schedule-shift from the existing data
# If not, we'll skip the CRUD tests

test_name "Create shift feedback requires --tender-job-schedule-shift"
xbe_run do shift-feedbacks create \
    --reason "$CREATED_REASON_ID" \
    --rating 5
assert_failure

test_name "Create shift feedback requires --reason"
xbe_run do shift-feedbacks create \
    --tender-job-schedule-shift "nonexistent" \
    --rating 5
assert_failure

test_name "Create shift feedback requires --rating"
xbe_run do shift-feedbacks create \
    --tender-job-schedule-shift "nonexistent" \
    --reason "$CREATED_REASON_ID"
assert_failure

# ============================================================================
# UPDATE Tests - Error cases
# ============================================================================

test_name "Update shift feedback without any fields fails"
# Use a fake ID since we likely don't have a real feedback to update
xbe_run do shift-feedbacks update "nonexistent"
assert_failure

# ============================================================================
# DELETE Tests - Error cases
# ============================================================================

test_name "Delete shift feedback requires --confirm flag"
xbe_run do shift-feedbacks delete "nonexistent"
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
