#!/bin/bash
#
# XBE CLI Integration Tests: Developers
#
# Tests CRUD operations for the developers resource.
# Developers are companies that develop projects.
#
# COVERAGE: All create/update attributes + list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_DEVELOPER_ID=""
CREATED_BROKER_ID=""

describe "Resource: developers"

# ============================================================================
# Prerequisites - Create a broker for developer tests
# ============================================================================

test_name "Create prerequisite broker for developer tests"
BROKER_NAME=$(unique_name "DevTestBroker")

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
# CREATE Tests
# ============================================================================

test_name "Create developer with required fields"
TEST_NAME=$(unique_name "Developer")

xbe_json do developers create \
    --name "$TEST_NAME" \
    --broker "$CREATED_BROKER_ID"

if [[ $status -eq 0 ]]; then
    CREATED_DEVELOPER_ID=$(json_get ".id")
    if [[ -n "$CREATED_DEVELOPER_ID" && "$CREATED_DEVELOPER_ID" != "null" ]]; then
        register_cleanup "developers" "$CREATED_DEVELOPER_ID"
        pass
    else
        fail "Created developer but no ID returned"
    fi
else
    fail "Failed to create developer"
fi

# Only continue if we successfully created a developer
if [[ -z "$CREATED_DEVELOPER_ID" || "$CREATED_DEVELOPER_ID" == "null" ]]; then
    echo "Cannot continue without a valid developer ID"
    run_tests
fi

test_name "Create developer with weigher-seal-label"
TEST_NAME2=$(unique_name "Developer2")
TEST_SEAL="SEAL$(date +%s | tail -c 5)"
xbe_json do developers create \
    --name "$TEST_NAME2" \
    --broker "$CREATED_BROKER_ID" \
    --weigher-seal-label "$TEST_SEAL"
if [[ $status -eq 0 ]]; then
    id=$(json_get ".id")
    register_cleanup "developers" "$id"
    pass
else
    fail "Failed to create developer with weigher-seal-label"
fi

test_name "Create developer with is-prevailing-wage-explicit"
TEST_NAME3=$(unique_name "Developer3")
xbe_json do developers create \
    --name "$TEST_NAME3" \
    --broker "$CREATED_BROKER_ID" \
    --is-prevailing-wage-explicit
if [[ $status -eq 0 ]]; then
    id=$(json_get ".id")
    register_cleanup "developers" "$id"
    pass
else
    fail "Failed to create developer with is-prevailing-wage-explicit"
fi

test_name "Create developer with is-certification-required-explicit"
TEST_NAME4=$(unique_name "Developer4")
xbe_json do developers create \
    --name "$TEST_NAME4" \
    --broker "$CREATED_BROKER_ID" \
    --is-certification-required-explicit
if [[ $status -eq 0 ]]; then
    id=$(json_get ".id")
    register_cleanup "developers" "$id"
    pass
else
    fail "Failed to create developer with is-certification-required-explicit"
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update developer name"
UPDATED_NAME=$(unique_name "UpdatedDev")
xbe_json do developers update "$CREATED_DEVELOPER_ID" --name "$UPDATED_NAME"
assert_success

test_name "Update developer weigher-seal-label"
UPDATED_SEAL="UPD$(date +%s | tail -c 5)"
xbe_json do developers update "$CREATED_DEVELOPER_ID" --weigher-seal-label "$UPDATED_SEAL"
assert_success

test_name "Update developer is-prevailing-wage-explicit to true"
xbe_json do developers update "$CREATED_DEVELOPER_ID" --is-prevailing-wage-explicit
assert_success

test_name "Update developer is-certification-required-explicit to true"
xbe_json do developers update "$CREATED_DEVELOPER_ID" --is-certification-required-explicit
assert_success

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List developers"
xbe_json view developers list --limit 5
assert_success

test_name "List developers returns array"
xbe_json view developers list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list developers"
fi

# ============================================================================
# LIST Tests - All Filters
# ============================================================================

test_name "List developers with --name filter"
xbe_json view developers list --name "$UPDATED_NAME" --limit 10
assert_success

test_name "List developers with --exact-name filter"
xbe_json view developers list --exact-name "$UPDATED_NAME" --limit 10
assert_success

test_name "List developers with --broker filter"
xbe_json view developers list --broker "$CREATED_BROKER_ID" --limit 10
assert_success

# ============================================================================
# LIST Tests - Pagination
# ============================================================================

test_name "List developers with --limit"
xbe_json view developers list --limit 3
assert_success

test_name "List developers with --offset"
xbe_json view developers list --limit 3 --offset 3
assert_success

test_name "List developers with pagination (limit + offset)"
xbe_json view developers list --limit 5 --offset 10
assert_success

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete developer requires --confirm flag"
xbe_json do developers delete "$CREATED_DEVELOPER_ID"
assert_failure

test_name "Delete developer with --confirm"
# Create a developer specifically for deletion
TEST_DEL_NAME=$(unique_name "DeleteMe")
xbe_json do developers create \
    --name "$TEST_DEL_NAME" \
    --broker "$CREATED_BROKER_ID"
if [[ $status -eq 0 ]]; then
    DEL_ID=$(json_get ".id")
    xbe_json do developers delete "$DEL_ID" --confirm
    assert_success
else
    skip "Could not create developer for deletion test"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create developer without name fails"
xbe_json do developers create --broker "$CREATED_BROKER_ID"
assert_failure

test_name "Create developer without broker fails"
xbe_json do developers create --name "Test Developer"
assert_failure

test_name "Update without any fields fails"
xbe_json do developers update "$CREATED_DEVELOPER_ID"
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
