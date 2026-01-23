#!/bin/bash
#
# XBE CLI Integration Tests: Job Production Plan Broadcast Messages
#
# Tests create/update and view operations for job production plan broadcast messages.
#
# COVERAGE: Create attributes (message, summary, user-ids), update (is-hidden),
#           list filters + show.
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_BROKER_ID=""
CREATED_CUSTOMER_ID=""
CREATED_JPP_ID=""
CREATED_MESSAGE_ID=""

describe "Resource: job-production-plan-broadcast-messages"

# ============================================================================
# Prerequisites - Create broker, customer, and job production plan
# ============================================================================

test_name "Create prerequisite broker for broadcast message tests"
BROKER_NAME=$(unique_name "JPPBroadcastBroker")

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

test_name "Create prerequisite customer"
CUSTOMER_NAME=$(unique_name "JPPBroadcastCustomer")

xbe_json do customers create \
    --name "$CUSTOMER_NAME" \
    --broker "$CREATED_BROKER_ID" \
    --can-manage-crew-requirements true

if [[ $status -eq 0 ]]; then
    CREATED_CUSTOMER_ID=$(json_get ".id")
    if [[ -n "$CREATED_CUSTOMER_ID" && "$CREATED_CUSTOMER_ID" != "null" ]]; then
        register_cleanup "customers" "$CREATED_CUSTOMER_ID"
        pass
    else
        fail "Created customer but no ID returned"
        echo "Cannot continue without a customer"
        run_tests
    fi
else
    fail "Failed to create customer"
    echo "Cannot continue without a customer"
    run_tests
fi

test_name "Create job production plan"
JPP_NAME=$(unique_name "JPPBroadcast")
TODAY=$(date +%Y-%m-%d)

xbe_json do job-production-plans create \
    --job-name "$JPP_NAME" \
    --start-on "$TODAY" \
    --start-time "07:00" \
    --customer "$CREATED_CUSTOMER_ID"

if [[ $status -eq 0 ]]; then
    CREATED_JPP_ID=$(json_get ".id")
    if [[ -n "$CREATED_JPP_ID" && "$CREATED_JPP_ID" != "null" ]]; then
        pass
    else
        fail "Created job production plan but no ID returned"
        echo "Cannot continue without a job production plan"
        run_tests
    fi
else
    fail "Failed to create job production plan"
    echo "Cannot continue without a job production plan"
    run_tests
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create broadcast message with required fields"
xbe_json do job-production-plan-broadcast-messages create \
    --job-production-plan "$CREATED_JPP_ID" \
    --message "Test broadcast message"

if [[ $status -eq 0 ]]; then
    CREATED_MESSAGE_ID=$(json_get ".id")
    if [[ -n "$CREATED_MESSAGE_ID" && "$CREATED_MESSAGE_ID" != "null" ]]; then
        pass
    else
        fail "Created broadcast message but no ID returned"
    fi
else
    fail "Failed to create broadcast message"
fi

# Only continue if we successfully created a broadcast message
if [[ -z "$CREATED_MESSAGE_ID" || "$CREATED_MESSAGE_ID" == "null" ]]; then
    echo "Cannot continue without a valid broadcast message ID"
    run_tests
fi

test_name "Create broadcast message with summary"
xbe_json do job-production-plan-broadcast-messages create \
    --job-production-plan "$CREATED_JPP_ID" \
    --message "Message with summary" \
    --summary "Summary only"
assert_success

test_name "Create broadcast message without --message fails"
xbe_json do job-production-plan-broadcast-messages create \
    --job-production-plan "$CREATED_JPP_ID"
assert_failure

test_name "Create broadcast message with --user-ids fails when not default recipients"
xbe_json do job-production-plan-broadcast-messages create \
    --job-production-plan "$CREATED_JPP_ID" \
    --message "Recipients test" \
    --user-ids 999999
assert_failure

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Hide broadcast message"
xbe_json do job-production-plan-broadcast-messages update "$CREATED_MESSAGE_ID" --is-hidden
assert_success

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show broadcast message"
xbe_json view job-production-plan-broadcast-messages show "$CREATED_MESSAGE_ID"
assert_success

CREATED_BY_ID=$(json_get ".created_by_id")

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List broadcast messages"
xbe_json view job-production-plan-broadcast-messages list --limit 5
assert_success

test_name "List broadcast messages returns array"
xbe_json view job-production-plan-broadcast-messages list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list broadcast messages"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List broadcast messages with --job-production-plan filter"
xbe_json view job-production-plan-broadcast-messages list --job-production-plan "$CREATED_JPP_ID" --limit 10
assert_success

test_name "List broadcast messages with --is-hidden filter"
xbe_json view job-production-plan-broadcast-messages list --is-hidden true --limit 10
assert_success

test_name "List broadcast messages with --created-by filter"
if [[ -n "$CREATED_BY_ID" && "$CREATED_BY_ID" != "null" ]]; then
    xbe_json view job-production-plan-broadcast-messages list --created-by "$CREATED_BY_ID" --limit 10
    assert_success
else
    skip "No created-by ID available"
fi

# ============================================================================
# LIST Tests - Pagination
# ============================================================================

test_name "List broadcast messages with --limit"
xbe_json view job-production-plan-broadcast-messages list --limit 3
assert_success

test_name "List broadcast messages with --offset"
xbe_json view job-production-plan-broadcast-messages list --limit 3 --offset 3
assert_success

# ============================================================================
# UPDATE Tests - Unhide
# ============================================================================

test_name "Unhide broadcast message"
xbe_json do job-production-plan-broadcast-messages update "$CREATED_MESSAGE_ID" --no-is-hidden
assert_success

run_tests
