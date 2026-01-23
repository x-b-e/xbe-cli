#!/bin/bash
#
# XBE CLI Integration Tests: Culture Values
#
# Tests CRUD operations for the culture_values resource.
# Culture values define organizational values used for public praise and recognition.
#
# COVERAGE: All create/update attributes + list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_CULTURE_VALUE_ID=""
CREATED_BROKER_ID=""

describe "Resource: culture_values"

# ============================================================================
# Prerequisites - Create broker for organization
# ============================================================================

test_name "Create prerequisite broker for culture value tests"
BROKER_NAME=$(unique_name "CVTestBroker")

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

test_name "Create culture value with required fields"
TEST_NAME=$(unique_name "CultureVal")

xbe_json do culture-values create \
    --name "$TEST_NAME" \
    --organization "Broker|$CREATED_BROKER_ID"

if [[ $status -eq 0 ]]; then
    CREATED_CULTURE_VALUE_ID=$(json_get ".id")
    if [[ -n "$CREATED_CULTURE_VALUE_ID" && "$CREATED_CULTURE_VALUE_ID" != "null" ]]; then
        register_cleanup "culture-values" "$CREATED_CULTURE_VALUE_ID"
        pass
    else
        fail "Created culture value but no ID returned"
    fi
else
    fail "Failed to create culture value"
fi

# Only continue if we successfully created a culture value
if [[ -z "$CREATED_CULTURE_VALUE_ID" || "$CREATED_CULTURE_VALUE_ID" == "null" ]]; then
    echo "Cannot continue without a valid culture value ID"
    run_tests
fi

test_name "Create culture value with description"
TEST_NAME2=$(unique_name "CultureVal2")
xbe_json do culture-values create \
    --name "$TEST_NAME2" \
    --description "A value that defines excellence" \
    --organization "Broker|$CREATED_BROKER_ID"
if [[ $status -eq 0 ]]; then
    id=$(json_get ".id")
    register_cleanup "culture-values" "$id"
    pass
else
    fail "Failed to create culture value with description"
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update culture value name"
UPDATED_NAME=$(unique_name "UpdatedCV")
xbe_json do culture-values update "$CREATED_CULTURE_VALUE_ID" --name "$UPDATED_NAME"
assert_success

test_name "Update culture value description"
xbe_json do culture-values update "$CREATED_CULTURE_VALUE_ID" --description "Updated description"
assert_success

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List culture values"
xbe_json view culture-values list --limit 5
assert_success

test_name "List culture values returns array"
xbe_json view culture-values list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list culture values"
fi

# ============================================================================
# LIST Tests - Pagination
# ============================================================================

test_name "List culture values with --limit"
xbe_json view culture-values list --limit 3
assert_success

test_name "List culture values with --offset"
xbe_json view culture-values list --limit 3 --offset 3
assert_success

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete culture value requires --confirm flag"
xbe_run do culture-values delete "$CREATED_CULTURE_VALUE_ID"
assert_failure

test_name "Delete culture value with --confirm"
# Create a culture value specifically for deletion
TEST_DEL_NAME=$(unique_name "DeleteCV")
xbe_json do culture-values create \
    --name "$TEST_DEL_NAME" \
    --organization "Broker|$CREATED_BROKER_ID"
if [[ $status -eq 0 ]]; then
    DEL_ID=$(json_get ".id")
    xbe_run do culture-values delete "$DEL_ID" --confirm
    assert_success
else
    skip "Could not create culture value for deletion test"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create culture value without name fails"
xbe_json do culture-values create --organization "Broker|$CREATED_BROKER_ID"
assert_failure

test_name "Create culture value without organization fails"
xbe_json do culture-values create --name "Test Value"
assert_failure

test_name "Update without any fields fails"
xbe_json do culture-values update "$CREATED_CULTURE_VALUE_ID"
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
