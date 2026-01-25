#!/bin/bash
#
# XBE CLI Integration Tests: Developer Trucker Certification Classifications
#
# Tests CRUD operations for the developer_trucker_certification_classifications resource.
# These classifications define types of certifications for truckers within a developer context.
#
# COVERAGE: All create/update attributes + list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_CLASSIFICATION_ID=""
CREATED_BROKER_ID=""
CREATED_DEVELOPER_ID=""

describe "Resource: developer_trucker_certification_classifications"

# ============================================================================
# Prerequisites - Create broker and developer
# ============================================================================

test_name "Create prerequisite broker for developer trucker certification classification tests"
BROKER_NAME=$(unique_name "DTCCTestBroker")

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

test_name "Create prerequisite developer"
DEVELOPER_NAME=$(unique_name "DTCCTestDev")

xbe_json do developers create \
    --name "$DEVELOPER_NAME" \
    --broker "$CREATED_BROKER_ID"

if [[ $status -eq 0 ]]; then
    CREATED_DEVELOPER_ID=$(json_get ".id")
    if [[ -n "$CREATED_DEVELOPER_ID" && "$CREATED_DEVELOPER_ID" != "null" ]]; then
        register_cleanup "developers" "$CREATED_DEVELOPER_ID"
        pass
    else
        fail "Created developer but no ID returned"
        echo "Cannot continue without a developer"
        run_tests
    fi
else
    fail "Failed to create developer"
    echo "Cannot continue without a developer"
    run_tests
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create developer trucker certification classification with required fields"
TEST_NAME=$(unique_name "DevTruckCertClass")

xbe_json do developer-trucker-certification-classifications create \
    --name "$TEST_NAME" \
    --developer "$CREATED_DEVELOPER_ID"

if [[ $status -eq 0 ]]; then
    CREATED_CLASSIFICATION_ID=$(json_get ".id")
    if [[ -n "$CREATED_CLASSIFICATION_ID" && "$CREATED_CLASSIFICATION_ID" != "null" ]]; then
        register_cleanup "developer-trucker-certification-classifications" "$CREATED_CLASSIFICATION_ID"
        pass
    else
        fail "Created developer trucker certification classification but no ID returned"
    fi
else
    fail "Failed to create developer trucker certification classification"
fi

# Only continue if we successfully created a classification
if [[ -z "$CREATED_CLASSIFICATION_ID" || "$CREATED_CLASSIFICATION_ID" == "null" ]]; then
    echo "Cannot continue without a valid developer trucker certification classification ID"
    run_tests
fi

test_name "Create second developer trucker certification classification"
TEST_NAME2=$(unique_name "DevTruckCertClass2")
xbe_json do developer-trucker-certification-classifications create \
    --name "$TEST_NAME2" \
    --developer "$CREATED_DEVELOPER_ID"
if [[ $status -eq 0 ]]; then
    id=$(json_get ".id")
    register_cleanup "developer-trucker-certification-classifications" "$id"
    pass
else
    fail "Failed to create second developer trucker certification classification"
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update developer trucker certification classification name"
UPDATED_NAME=$(unique_name "UpdatedDTCC")
xbe_json do developer-trucker-certification-classifications update "$CREATED_CLASSIFICATION_ID" --name "$UPDATED_NAME"
assert_success

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List developer trucker certification classifications"
xbe_json view developer-trucker-certification-classifications list --limit 5
assert_success

test_name "List developer trucker certification classifications returns array"
xbe_json view developer-trucker-certification-classifications list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list developer trucker certification classifications"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List developer trucker certification classifications with --developer filter"
xbe_json view developer-trucker-certification-classifications list --developer "$CREATED_DEVELOPER_ID" --limit 10
assert_success

# ============================================================================
# LIST Tests - Pagination
# ============================================================================

test_name "List developer trucker certification classifications with --limit"
xbe_json view developer-trucker-certification-classifications list --limit 3
assert_success

test_name "List developer trucker certification classifications with --offset"
xbe_json view developer-trucker-certification-classifications list --limit 3 --offset 3
assert_success

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete developer trucker certification classification requires --confirm flag"
xbe_run do developer-trucker-certification-classifications delete "$CREATED_CLASSIFICATION_ID"
assert_failure

test_name "Delete developer trucker certification classification with --confirm"
# Create a classification specifically for deletion
TEST_DEL_NAME=$(unique_name "DeleteDTCC")
xbe_json do developer-trucker-certification-classifications create \
    --name "$TEST_DEL_NAME" \
    --developer "$CREATED_DEVELOPER_ID"
if [[ $status -eq 0 ]]; then
    DEL_ID=$(json_get ".id")
    xbe_run do developer-trucker-certification-classifications delete "$DEL_ID" --confirm
    assert_success
else
    skip "Could not create developer trucker certification classification for deletion test"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create developer trucker certification classification without name fails"
xbe_json do developer-trucker-certification-classifications create \
    --developer "$CREATED_DEVELOPER_ID"
assert_failure

test_name "Create developer trucker certification classification without developer fails"
xbe_json do developer-trucker-certification-classifications create \
    --name "NoDeveloper"
assert_failure

test_name "Update without any fields fails"
xbe_json do developer-trucker-certification-classifications update "$CREATED_CLASSIFICATION_ID"
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
