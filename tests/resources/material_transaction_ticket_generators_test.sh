#!/bin/bash
#
# XBE CLI Integration Tests: Material Transaction Ticket Generators
#
# Tests CRUD operations for the material-transaction-ticket-generators resource.
#
# COVERAGE: All create/update attributes + list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_GENERATOR_ID=""
CREATED_BROKER_ID=""
CURRENT_USER_ID=""
CURRENT_USER_IS_ADMIN=""
CREATE_SUPPORTED="true"

describe "Resource: material-transaction-ticket-generators"

# ============================================================================
# Prerequisites - Current user and broker
# ============================================================================

test_name "Fetch current authenticated user"
xbe_json auth whoami

if [[ $status -eq 0 ]]; then
    CURRENT_USER_ID=$(json_get ".id")
    CURRENT_USER_IS_ADMIN=$(json_get ".is_admin")
    if [[ -n "$CURRENT_USER_ID" && "$CURRENT_USER_ID" != "null" ]]; then
        pass
    else
        fail "Could not determine current user ID"
        run_tests
    fi
else
    fail "Failed to fetch authenticated user"
    run_tests
fi

test_name "Create prerequisite broker for ticket generator tests"
if [[ "$CURRENT_USER_IS_ADMIN" == "true" ]]; then
    BROKER_NAME=$(unique_name "MTXTicketGenBroker")
    xbe_json do brokers create --name "$BROKER_NAME"

    if [[ $status -eq 0 ]]; then
        CREATED_BROKER_ID=$(json_get ".id")
        if [[ -n "$CREATED_BROKER_ID" && "$CREATED_BROKER_ID" != "null" ]]; then
            register_cleanup "brokers" "$CREATED_BROKER_ID"
            pass
        else
            fail "Created broker but no ID returned"
            run_tests
        fi
    else
        if [[ -n "$XBE_TEST_BROKER_ID" ]]; then
            CREATED_BROKER_ID="$XBE_TEST_BROKER_ID"
            echo "    Using XBE_TEST_BROKER_ID: $CREATED_BROKER_ID"
            pass
        else
            fail "Failed to create broker and XBE_TEST_BROKER_ID not set"
            run_tests
        fi
    fi
else
    if [[ -n "$XBE_TEST_BROKER_ID" ]]; then
        CREATED_BROKER_ID="$XBE_TEST_BROKER_ID"
        echo "    Using XBE_TEST_BROKER_ID: $CREATED_BROKER_ID"
        pass
    else
        fail "Non-admin user requires XBE_TEST_BROKER_ID to proceed"
        run_tests
    fi
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create material transaction ticket generator"
FORMAT_RULE="MTX-$(unique_suffix)"
xbe_json do material-transaction-ticket-generators create \
    --format-rule "$FORMAT_RULE" \
    --organization-type "brokers" \
    --organization-id "$CREATED_BROKER_ID"

if [[ $status -eq 0 ]]; then
    CREATED_GENERATOR_ID=$(json_get ".id")
    if [[ -n "$CREATED_GENERATOR_ID" && "$CREATED_GENERATOR_ID" != "null" ]]; then
        register_cleanup "material-transaction-ticket-generators" "$CREATED_GENERATOR_ID"
        pass
    else
        fail "Created ticket generator but no ID returned"
    fi
else
    if [[ "$output" == *"Not Authorized"* ]] || \
       [[ "$output" == *"not authorized"* ]] || \
       [[ "$output" == *"403"* ]] || \
       [[ "$output" == *"Forbidden"* ]]; then
        CREATE_SUPPORTED="false"
        skip "Not authorized to create ticket generators"
    else
        fail "Failed to create ticket generator"
    fi
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update material transaction ticket generator --format-rule"
if [[ "$CREATE_SUPPORTED" == "true" && -n "$CREATED_GENERATOR_ID" && "$CREATED_GENERATOR_ID" != "null" ]]; then
    UPDATED_RULE="MTX-UPDATED-$(unique_suffix)"
    xbe_json do material-transaction-ticket-generators update "$CREATED_GENERATOR_ID" --format-rule "$UPDATED_RULE"
    assert_success
else
    skip "Create not supported; update skipped"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show material transaction ticket generator"
if [[ -n "$CREATED_GENERATOR_ID" && "$CREATED_GENERATOR_ID" != "null" ]]; then
    xbe_json view material-transaction-ticket-generators show "$CREATED_GENERATOR_ID"
    assert_success
else
    skip "No ticket generator ID available"
fi

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List material transaction ticket generators"
xbe_json view material-transaction-ticket-generators list --limit 5
if [[ $status -eq 0 ]]; then
    pass
else
    if [[ "$output" == *"Not Authorized"* ]] || \
       [[ "$output" == *"not authorized"* ]] || \
       [[ "$output" == *"403"* ]] || \
       [[ "$output" == *"Forbidden"* ]]; then
        skip "Listing requires additional access"
    else
        fail "Failed to list ticket generators"
    fi
fi

test_name "List material transaction ticket generators returns array"
xbe_json view material-transaction-ticket-generators list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list ticket generators"
fi

# ============================================================================
# LIST Tests - All Filters
# ============================================================================

if [[ -n "$CREATED_BROKER_ID" && "$CREATED_BROKER_ID" != "null" ]]; then
    test_name "List ticket generators with --broker filter"
    xbe_json view material-transaction-ticket-generators list --broker "$CREATED_BROKER_ID" --limit 10
    assert_success

    test_name "List ticket generators with --organization filter"
    xbe_json view material-transaction-ticket-generators list --organization "Broker|$CREATED_BROKER_ID" --limit 10
    assert_success

    test_name "List ticket generators with --organization-type filter"
    xbe_json view material-transaction-ticket-generators list --organization-type "Broker" --limit 10
    assert_success

    test_name "List ticket generators with --organization-id filter"
    xbe_json view material-transaction-ticket-generators list --organization-type "Broker" --organization-id "$CREATED_BROKER_ID" --limit 10
    assert_success
else
    skip "Broker ID unavailable; organization filters skipped"
fi

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete material transaction ticket generator"
if [[ "$CREATE_SUPPORTED" == "true" && -n "$CREATED_GENERATOR_ID" && "$CREATED_GENERATOR_ID" != "null" ]]; then
    xbe_json do material-transaction-ticket-generators delete "$CREATED_GENERATOR_ID" --confirm
    assert_success
else
    skip "Create not supported; delete skipped"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create ticket generator without required flags fails"
xbe_run do material-transaction-ticket-generators create
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
