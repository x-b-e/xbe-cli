#!/bin/bash
#
# XBE CLI Integration Tests: External Identifications
#
# Tests CRUD operations for the external-identifications resource.
# External identifications link external ID values (e.g., license numbers, tax IDs)
# to entities like truckers, brokers, and material sites.
#
# NOTE: This test requires creating prerequisite resources: broker, external identification type, and trucker
#
# COVERAGE: All create/update attributes + list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_EXT_ID=""
CREATED_BROKER_ID=""
CREATED_EXT_ID_TYPE_ID=""
CREATED_TRUCKER_ID=""

describe "Resource: external-identifications"

# ============================================================================
# Prerequisites - Create broker, external identification type, and trucker
# ============================================================================

test_name "Create prerequisite broker for external identification tests"
BROKER_NAME=$(unique_name "ExtIdTestBroker")

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

test_name "Create prerequisite external identification type"
EXT_ID_TYPE_NAME=$(unique_name "ExtIdType")

xbe_json do external-identification-types create \
    --name "$EXT_ID_TYPE_NAME" \
    --can-apply-to "Trucker"

if [[ $status -eq 0 ]]; then
    CREATED_EXT_ID_TYPE_ID=$(json_get ".id")
    if [[ -n "$CREATED_EXT_ID_TYPE_ID" && "$CREATED_EXT_ID_TYPE_ID" != "null" ]]; then
        register_cleanup "external-identification-types" "$CREATED_EXT_ID_TYPE_ID"
        pass
    else
        fail "Created external identification type but no ID returned"
        echo "Cannot continue without an external identification type"
        run_tests
    fi
else
    fail "Failed to create external identification type"
    echo "Cannot continue without an external identification type"
    run_tests
fi

test_name "Create prerequisite trucker"
TRUCKER_NAME=$(unique_name "ExtIdTestTrucker")

xbe_json do truckers create \
    --name "$TRUCKER_NAME" \
    --broker "$CREATED_BROKER_ID" \
    --company-address "100 Test Lane, Chicago, IL 60601"

if [[ $status -eq 0 ]]; then
    CREATED_TRUCKER_ID=$(json_get ".id")
    if [[ -n "$CREATED_TRUCKER_ID" && "$CREATED_TRUCKER_ID" != "null" ]]; then
        register_cleanup "truckers" "$CREATED_TRUCKER_ID"
        pass
    else
        fail "Created trucker but no ID returned"
        echo "Cannot continue without a trucker"
        run_tests
    fi
else
    fail "Failed to create trucker"
    echo "Cannot continue without a trucker"
    run_tests
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create external identification with required fields"

EXT_ID_VALUE=$(unique_name "EXTID")

xbe_json do external-identifications create \
    --external-identification-type "$CREATED_EXT_ID_TYPE_ID" \
    --identifies-type "truckers" \
    --identifies-id "$CREATED_TRUCKER_ID" \
    --value "$EXT_ID_VALUE"

if [[ $status -eq 0 ]]; then
    CREATED_EXT_ID=$(json_get ".id")
    if [[ -n "$CREATED_EXT_ID" && "$CREATED_EXT_ID" != "null" ]]; then
        register_cleanup "external-identifications" "$CREATED_EXT_ID"
        pass
    else
        fail "Created external identification but no ID returned"
    fi
else
    fail "Failed to create external identification"
fi

# Only continue if we successfully created an external identification
if [[ -z "$CREATED_EXT_ID" || "$CREATED_EXT_ID" == "null" ]]; then
    echo "Cannot continue without a valid external identification ID"
    run_tests
fi

test_name "Create external identification with --skip-value-validation"

# Create a second trucker for this test
TRUCKER2_NAME=$(unique_name "ExtIdTestTrucker2")
xbe_json do truckers create \
    --name "$TRUCKER2_NAME" \
    --broker "$CREATED_BROKER_ID" \
    --company-address "200 Test Lane, Chicago, IL 60602"

if [[ $status -eq 0 ]]; then
    TRUCKER2_ID=$(json_get ".id")
    register_cleanup "truckers" "$TRUCKER2_ID"

    EXT_ID_VALUE2=$(unique_name "SKIP")
    xbe_json do external-identifications create \
        --external-identification-type "$CREATED_EXT_ID_TYPE_ID" \
        --identifies-type "truckers" \
        --identifies-id "$TRUCKER2_ID" \
        --value "$EXT_ID_VALUE2" \
        --skip-value-validation

    if [[ $status -eq 0 ]]; then
        id=$(json_get ".id")
        register_cleanup "external-identifications" "$id"
        pass
    else
        fail "Failed to create external identification with --skip-value-validation"
    fi
else
    skip "Could not create trucker for skip-value-validation test"
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update external identification --value"
NEW_VALUE=$(unique_name "UPDATED")
xbe_json do external-identifications update "$CREATED_EXT_ID" --value "$NEW_VALUE"
assert_success

test_name "Update external identification --skip-value-validation"
xbe_json do external-identifications update "$CREATED_EXT_ID" --skip-value-validation
assert_success

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List external identifications"
xbe_json view external-identifications list --limit 5
assert_success

test_name "List external identifications returns array"
xbe_json view external-identifications list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list external identifications"
fi

# ============================================================================
# LIST Tests - Pagination
# ============================================================================

test_name "List external identifications with --limit"
xbe_json view external-identifications list --limit 3
assert_success

test_name "List external identifications with --offset"
xbe_json view external-identifications list --limit 3 --offset 3
assert_success

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete external identification requires --confirm flag"
xbe_run do external-identifications delete "$CREATED_EXT_ID"
assert_failure

test_name "Delete external identification with --confirm"
# Create an external identification specifically for deletion
TRUCKER3_NAME=$(unique_name "ExtIdDelTrucker")
xbe_json do truckers create \
    --name "$TRUCKER3_NAME" \
    --broker "$CREATED_BROKER_ID" \
    --company-address "300 Test Lane, Chicago, IL 60603"
if [[ $status -eq 0 ]]; then
    TRUCKER3_ID=$(json_get ".id")
    register_cleanup "truckers" "$TRUCKER3_ID"

    DEL_VALUE=$(unique_name "DEL")
    xbe_json do external-identifications create \
        --external-identification-type "$CREATED_EXT_ID_TYPE_ID" \
        --identifies-type "truckers" \
        --identifies-id "$TRUCKER3_ID" \
        --value "$DEL_VALUE"
    if [[ $status -eq 0 ]]; then
        DEL_EXT_ID=$(json_get ".id")
        xbe_run do external-identifications delete "$DEL_EXT_ID" --confirm
        assert_success
    else
        skip "Could not create external identification for deletion test"
    fi
else
    skip "Could not create trucker for deletion test"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create external identification without --external-identification-type fails"
xbe_json do external-identifications create --identifies-type "truckers" --identifies-id "$CREATED_TRUCKER_ID" --value "TEST"
assert_failure

test_name "Create external identification without --identifies-type fails"
xbe_json do external-identifications create --external-identification-type "$CREATED_EXT_ID_TYPE_ID" --identifies-id "$CREATED_TRUCKER_ID" --value "TEST"
assert_failure

test_name "Create external identification without --identifies-id fails"
xbe_json do external-identifications create --external-identification-type "$CREATED_EXT_ID_TYPE_ID" --identifies-type "truckers" --value "TEST"
assert_failure

test_name "Create external identification without --value fails"
xbe_json do external-identifications create --external-identification-type "$CREATED_EXT_ID_TYPE_ID" --identifies-type "truckers" --identifies-id "$CREATED_TRUCKER_ID"
assert_failure

test_name "Update without any fields fails"
xbe_json do external-identifications update "$CREATED_EXT_ID"
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
