#!/bin/bash
#
# XBE CLI Integration Tests: Public Praises
#
# Tests CRUD operations for the public-praises resource.
# Public praises represent employee recognition.
#
# NOTE: This test requires creating prerequisite resources: broker and user
#
# COVERAGE: All create/update attributes + list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_PRAISE_ID=""
CREATED_BROKER_ID=""
GIVER_USER_ID=""
RECEIVER_USER_ID=""

describe "Resource: public-praises"

# ============================================================================
# Prerequisites - Create broker and user
# ============================================================================

test_name "Create prerequisite broker for public praise tests"
BROKER_NAME=$(unique_name "PraiseTestBroker")

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

test_name "Create giver user for public praise tests"
GIVER_EMAIL=$(unique_email)
GIVER_NAME=$(unique_name "PraiseGiver")

xbe_json do users create \
    --email "$GIVER_EMAIL" \
    --name "$GIVER_NAME"

if [[ $status -eq 0 ]]; then
    GIVER_USER_ID=$(json_get ".id")
    if [[ -n "$GIVER_USER_ID" && "$GIVER_USER_ID" != "null" ]]; then
        pass
    else
        fail "Created user but no ID returned"
        echo "Cannot continue without a giver user"
        run_tests
    fi
else
    fail "Failed to create giver user"
    echo "Cannot continue without a giver user"
    run_tests
fi

test_name "Create receiver user for public praise tests"
RECEIVER_EMAIL=$(unique_email)
RECEIVER_NAME=$(unique_name "PraiseReceiver")

xbe_json do users create \
    --email "$RECEIVER_EMAIL" \
    --name "$RECEIVER_NAME"

if [[ $status -eq 0 ]]; then
    RECEIVER_USER_ID=$(json_get ".id")
    if [[ -n "$RECEIVER_USER_ID" && "$RECEIVER_USER_ID" != "null" ]]; then
        pass
    else
        fail "Created user but no ID returned"
        echo "Cannot continue without a receiver user"
        run_tests
    fi
else
    fail "Failed to create receiver user"
    echo "Cannot continue without a receiver user"
    run_tests
fi

test_name "Create membership for giver user to broker"
xbe_json do memberships create \
    --user "$GIVER_USER_ID" \
    --organization "Broker|$CREATED_BROKER_ID"

if [[ $status -eq 0 ]]; then
    GIVER_MEMBERSHIP_ID=$(json_get ".id")
    if [[ -n "$GIVER_MEMBERSHIP_ID" && "$GIVER_MEMBERSHIP_ID" != "null" ]]; then
        register_cleanup "memberships" "$GIVER_MEMBERSHIP_ID"
        pass
    else
        fail "Created membership but no ID returned"
        echo "Cannot continue without a giver membership"
        run_tests
    fi
else
    fail "Failed to create giver membership"
    echo "Cannot continue without a giver membership"
    run_tests
fi

test_name "Create membership for receiver user to broker"
xbe_json do memberships create \
    --user "$RECEIVER_USER_ID" \
    --organization "Broker|$CREATED_BROKER_ID"

if [[ $status -eq 0 ]]; then
    RECEIVER_MEMBERSHIP_ID=$(json_get ".id")
    if [[ -n "$RECEIVER_MEMBERSHIP_ID" && "$RECEIVER_MEMBERSHIP_ID" != "null" ]]; then
        register_cleanup "memberships" "$RECEIVER_MEMBERSHIP_ID"
        pass
    else
        fail "Created membership but no ID returned"
        echo "Cannot continue without a receiver membership"
        run_tests
    fi
else
    fail "Failed to create receiver membership"
    echo "Cannot continue without a receiver membership"
    run_tests
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create public praise with required fields"

xbe_json do public-praises create \
    --description "Great job on the project!" \
    --given-by "$GIVER_USER_ID" \
    --received-by "$RECEIVER_USER_ID" \
    --organization-type "brokers" \
    --organization-id "$CREATED_BROKER_ID"

if [[ $status -eq 0 ]]; then
    CREATED_PRAISE_ID=$(json_get ".id")
    if [[ -n "$CREATED_PRAISE_ID" && "$CREATED_PRAISE_ID" != "null" ]]; then
        register_cleanup "public-praises" "$CREATED_PRAISE_ID"
        pass
    else
        fail "Created public praise but no ID returned"
    fi
else
    fail "Failed to create public praise"
fi

# Only continue if we successfully created a public praise
if [[ -z "$CREATED_PRAISE_ID" || "$CREATED_PRAISE_ID" == "null" ]]; then
    echo "Cannot continue without a valid public praise ID"
    run_tests
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update public praise --description"
xbe_json do public-praises update "$CREATED_PRAISE_ID" --description "Updated praise description"
assert_success

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List public praises"
xbe_json view public-praises list --limit 5
assert_success

test_name "List public praises returns array"
xbe_json view public-praises list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list public praises"
fi

# ============================================================================
# LIST Tests - All Filters
# ============================================================================

test_name "List public praises with --broker filter"
xbe_json view public-praises list --broker "$CREATED_BROKER_ID" --limit 10
assert_success

test_name "List public praises with --given-by filter"
xbe_json view public-praises list --given-by "$GIVER_USER_ID" --limit 10
assert_success

# ============================================================================
# LIST Tests - Pagination
# ============================================================================

test_name "List public praises with --limit"
xbe_json view public-praises list --limit 3
assert_success

test_name "List public praises with --offset"
xbe_json view public-praises list --limit 3 --offset 3
assert_success

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete public praise requires --confirm flag"
xbe_run do public-praises delete "$CREATED_PRAISE_ID"
assert_failure

test_name "Delete public praise with --confirm"
# Create a praise specifically for deletion
xbe_json do public-praises create \
    --description "Praise to delete" \
    --given-by "$GIVER_USER_ID" \
    --received-by "$RECEIVER_USER_ID" \
    --organization-type "brokers" \
    --organization-id "$CREATED_BROKER_ID"

if [[ $status -eq 0 ]]; then
    DEL_PRAISE_ID=$(json_get ".id")
    if [[ -n "$DEL_PRAISE_ID" && "$DEL_PRAISE_ID" != "null" ]]; then
        xbe_run do public-praises delete "$DEL_PRAISE_ID" --confirm
        assert_success
    else
        skip "Could not create public praise for deletion test"
    fi
else
    skip "Could not create public praise for deletion test"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create public praise without --description fails"
xbe_json do public-praises create \
    --given-by "$GIVER_USER_ID" \
    --received-by "$RECEIVER_USER_ID" \
    --organization-type "brokers" \
    --organization-id "$CREATED_BROKER_ID"
assert_failure

test_name "Create public praise without --given-by fails"
xbe_json do public-praises create \
    --description "Missing given-by" \
    --received-by "$RECEIVER_USER_ID" \
    --organization-type "brokers" \
    --organization-id "$CREATED_BROKER_ID"
assert_failure

test_name "Create public praise without --received-by fails"
xbe_json do public-praises create \
    --description "Missing received-by" \
    --given-by "$GIVER_USER_ID" \
    --organization-type "brokers" \
    --organization-id "$CREATED_BROKER_ID"
assert_failure

test_name "Create public praise without --organization-type fails"
xbe_json do public-praises create \
    --description "Missing org type" \
    --given-by "$GIVER_USER_ID" \
    --received-by "$RECEIVER_USER_ID" \
    --organization-id "$CREATED_BROKER_ID"
assert_failure

test_name "Create public praise without --organization-id fails"
xbe_json do public-praises create \
    --description "Missing org id" \
    --given-by "$GIVER_USER_ID" \
    --received-by "$RECEIVER_USER_ID" \
    --organization-type "brokers"
assert_failure

test_name "Update public praise without any fields fails"
xbe_run do public-praises update "$CREATED_PRAISE_ID"
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
