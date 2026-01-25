#!/bin/bash
#
# XBE CLI Integration Tests: Public Praise Culture Values
#
# Tests CRUD operations for the public-praise-culture-values resource.
# Public praise culture values link public praises to culture values.
#
# COVERAGE: create/update/delete + list filters + show
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_BROKER_ID=""
GIVER_USER_ID=""
RECEIVER_USER_ID=""
GIVER_MEMBERSHIP_ID=""
RECEIVER_MEMBERSHIP_ID=""
CREATED_PRAISE_ID=""
CREATED_CULTURE_VALUE_ID=""
CREATED_PPCV_ID=""

describe "Resource: public-praise-culture-values"

# ============================================================================
# Prerequisites - Create broker, users, memberships, public praise, culture value
# ============================================================================

test_name "Create prerequisite broker for public praise culture value tests"
BROKER_NAME=$(unique_name "PraiseCultureBroker")

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

test_name "Create giver user for public praise culture value tests"
GIVER_EMAIL=$(unique_email)
GIVER_NAME=$(unique_name "PraiseCultureGiver")

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

test_name "Create receiver user for public praise culture value tests"
RECEIVER_EMAIL=$(unique_email)
RECEIVER_NAME=$(unique_name "PraiseCultureReceiver")

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

test_name "Create public praise for culture value link"
xbe_json do public-praises create \
    --description "Great teamwork on the build!" \
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
        echo "Cannot continue without a public praise"
        run_tests
    fi
else
    fail "Failed to create public praise"
    echo "Cannot continue without a public praise"
    run_tests
fi

test_name "Create culture value for culture value link"
CULTURE_VALUE_NAME=$(unique_name "PraiseCultureValue")

xbe_json do culture-values create \
    --name "$CULTURE_VALUE_NAME" \
    --organization "Broker|$CREATED_BROKER_ID"

if [[ $status -eq 0 ]]; then
    CREATED_CULTURE_VALUE_ID=$(json_get ".id")
    if [[ -n "$CREATED_CULTURE_VALUE_ID" && "$CREATED_CULTURE_VALUE_ID" != "null" ]]; then
        register_cleanup "culture-values" "$CREATED_CULTURE_VALUE_ID"
        pass
    else
        fail "Created culture value but no ID returned"
        echo "Cannot continue without a culture value"
        run_tests
    fi
else
    fail "Failed to create culture value"
    echo "Cannot continue without a culture value"
    run_tests
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create public praise culture value without required fields fails"
xbe_run do public-praise-culture-values create
assert_failure

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create public praise culture value"
xbe_json do public-praise-culture-values create \
    --public-praise "$CREATED_PRAISE_ID" \
    --culture-value "$CREATED_CULTURE_VALUE_ID"

if [[ $status -eq 0 ]]; then
    CREATED_PPCV_ID=$(json_get ".id")
    if [[ -n "$CREATED_PPCV_ID" && "$CREATED_PPCV_ID" != "null" ]]; then
        register_cleanup "public-praise-culture-values" "$CREATED_PPCV_ID"
        pass
    else
        fail "Created public praise culture value but no ID returned"
    fi
else
    fail "Failed to create public praise culture value"
fi

# Only continue if we successfully created a public praise culture value
if [[ -z "$CREATED_PPCV_ID" || "$CREATED_PPCV_ID" == "null" ]]; then
    echo "Cannot continue without a valid public praise culture value ID"
    run_tests
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update public praise culture value relationships"
xbe_json do public-praise-culture-values update "$CREATED_PPCV_ID" \
    --public-praise "$CREATED_PRAISE_ID" \
    --culture-value "$CREATED_CULTURE_VALUE_ID"
assert_success

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show public praise culture value"
xbe_json view public-praise-culture-values show "$CREATED_PPCV_ID"
assert_success

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List public praise culture values"
xbe_json view public-praise-culture-values list --limit 5
assert_success

test_name "List public praise culture values returns array"
xbe_json view public-praise-culture-values list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list public praise culture values"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List public praise culture values with --public-praise filter"
xbe_json view public-praise-culture-values list --public-praise "$CREATED_PRAISE_ID" --limit 5
assert_success

test_name "List public praise culture values with --culture-value filter"
xbe_json view public-praise-culture-values list --culture-value "$CREATED_CULTURE_VALUE_ID" --limit 5
assert_success

test_name "List public praise culture values with --created-at-min filter"
xbe_json view public-praise-culture-values list --created-at-min "2020-01-01T00:00:00Z" --limit 5
assert_success

test_name "List public praise culture values with --created-at-max filter"
xbe_json view public-praise-culture-values list --created-at-max "2030-01-01T00:00:00Z" --limit 5
assert_success

test_name "List public praise culture values with --updated-at-min filter"
xbe_json view public-praise-culture-values list --updated-at-min "2020-01-01T00:00:00Z" --limit 5
assert_success

test_name "List public praise culture values with --updated-at-max filter"
xbe_json view public-praise-culture-values list --updated-at-max "2030-01-01T00:00:00Z" --limit 5
assert_success

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete public praise culture value"
if [[ -n "$CREATED_PPCV_ID" && "$CREATED_PPCV_ID" != "null" ]]; then
    xbe_run do public-praise-culture-values delete "$CREATED_PPCV_ID" --confirm
    assert_success
else
    skip "No public praise culture value created for deletion"
fi

run_tests
