#!/bin/bash
#
# XBE CLI Integration Tests: Tractor Trailer Credential Classifications
#
# Tests CRUD operations for the tractor_trailer_credential_classifications resource.
# These classifications define types of credentials for tractors and trailers.
#
# NOTE: The organization type for tractor/trailer credential classifications must be
# "truckers" (not brokers). This test creates a broker and trucker as prerequisites.
#
# COVERAGE: All create/update attributes + list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_CLASSIFICATION_ID=""
CREATED_BROKER_ID=""
CREATED_TRUCKER_ID=""

describe "Resource: tractor_trailer_credential_classifications"

# ============================================================================
# Prerequisites - Create broker and trucker
# ============================================================================

test_name "Create prerequisite broker for tractor trailer credential classification tests"
BROKER_NAME=$(unique_name "TTCCTestBroker")

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

test_name "Create prerequisite trucker"
TRUCKER_NAME=$(unique_name "TTCCTestTrucker")
TRUCKER_ADDRESS="350 N Orleans St, Chicago, IL 60654"

xbe_json do truckers create \
    --name "$TRUCKER_NAME" \
    --broker "$CREATED_BROKER_ID" \
    --company-address "$TRUCKER_ADDRESS"

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

test_name "Create tractor trailer credential classification with required fields"
TEST_NAME=$(unique_name "TTCredClass")

xbe_json do tractor-trailer-credential-classifications create \
    --name "$TEST_NAME" \
    --organization-type "truckers" \
    --organization-id "$CREATED_TRUCKER_ID"

if [[ $status -eq 0 ]]; then
    CREATED_CLASSIFICATION_ID=$(json_get ".id")
    if [[ -n "$CREATED_CLASSIFICATION_ID" && "$CREATED_CLASSIFICATION_ID" != "null" ]]; then
        register_cleanup "tractor-trailer-credential-classifications" "$CREATED_CLASSIFICATION_ID"
        pass
    else
        fail "Created tractor trailer credential classification but no ID returned"
    fi
else
    fail "Failed to create tractor trailer credential classification"
fi

# Only continue if we successfully created a classification
if [[ -z "$CREATED_CLASSIFICATION_ID" || "$CREATED_CLASSIFICATION_ID" == "null" ]]; then
    echo "Cannot continue without a valid tractor trailer credential classification ID"
    run_tests
fi

test_name "Create tractor trailer credential classification with description"
TEST_NAME2=$(unique_name "TTCredClass2")
xbe_json do tractor-trailer-credential-classifications create \
    --name "$TEST_NAME2" \
    --organization-type "truckers" \
    --organization-id "$CREATED_TRUCKER_ID" \
    --description "Insurance credential for tractors"
if [[ $status -eq 0 ]]; then
    id=$(json_get ".id")
    register_cleanup "tractor-trailer-credential-classifications" "$id"
    pass
else
    fail "Failed to create tractor trailer credential classification with description"
fi

test_name "Create tractor trailer credential classification with issuer-name"
TEST_NAME3=$(unique_name "TTCredClass3")
xbe_json do tractor-trailer-credential-classifications create \
    --name "$TEST_NAME3" \
    --organization-type "truckers" \
    --organization-id "$CREATED_TRUCKER_ID" \
    --issuer-name "State DMV"
if [[ $status -eq 0 ]]; then
    id=$(json_get ".id")
    register_cleanup "tractor-trailer-credential-classifications" "$id"
    pass
else
    fail "Failed to create tractor trailer credential classification with issuer-name"
fi

test_name "Create tractor trailer credential classification with external-id"
TEST_NAME4=$(unique_name "TTCredClass4")
xbe_json do tractor-trailer-credential-classifications create \
    --name "$TEST_NAME4" \
    --organization-type "truckers" \
    --organization-id "$CREATED_TRUCKER_ID" \
    --external-id "EXT-123"
if [[ $status -eq 0 ]]; then
    id=$(json_get ".id")
    register_cleanup "tractor-trailer-credential-classifications" "$id"
    pass
else
    fail "Failed to create tractor trailer credential classification with external-id"
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update tractor trailer credential classification name"
UPDATED_NAME=$(unique_name "UpdatedTTCC")
xbe_json do tractor-trailer-credential-classifications update "$CREATED_CLASSIFICATION_ID" --name "$UPDATED_NAME"
assert_success

test_name "Update tractor trailer credential classification description"
xbe_json do tractor-trailer-credential-classifications update "$CREATED_CLASSIFICATION_ID" --description "Updated description"
assert_success

test_name "Update tractor trailer credential classification issuer-name"
xbe_json do tractor-trailer-credential-classifications update "$CREATED_CLASSIFICATION_ID" --issuer-name "Updated Issuer"
assert_success

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List tractor trailer credential classifications"
xbe_json view tractor-trailer-credential-classifications list --limit 5
assert_success

test_name "List tractor trailer credential classifications returns array"
xbe_json view tractor-trailer-credential-classifications list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list tractor trailer credential classifications"
fi

# ============================================================================
# LIST Tests - Pagination
# ============================================================================

test_name "List tractor trailer credential classifications with --limit"
xbe_json view tractor-trailer-credential-classifications list --limit 3
assert_success

test_name "List tractor trailer credential classifications with --offset"
xbe_json view tractor-trailer-credential-classifications list --limit 3 --offset 3
assert_success

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete tractor trailer credential classification requires --confirm flag"
xbe_run do tractor-trailer-credential-classifications delete "$CREATED_CLASSIFICATION_ID"
assert_failure

test_name "Delete tractor trailer credential classification with --confirm"
# Create a classification specifically for deletion
TEST_DEL_NAME=$(unique_name "DeleteTTCC")
xbe_json do tractor-trailer-credential-classifications create \
    --name "$TEST_DEL_NAME" \
    --organization-type "truckers" \
    --organization-id "$CREATED_TRUCKER_ID"
if [[ $status -eq 0 ]]; then
    DEL_ID=$(json_get ".id")
    xbe_run do tractor-trailer-credential-classifications delete "$DEL_ID" --confirm
    assert_success
else
    skip "Could not create tractor trailer credential classification for deletion test"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create tractor trailer credential classification without name fails"
xbe_json do tractor-trailer-credential-classifications create \
    --organization-type "truckers" \
    --organization-id "$CREATED_TRUCKER_ID"
assert_failure

test_name "Create tractor trailer credential classification without organization-type fails"
xbe_json do tractor-trailer-credential-classifications create \
    --name "NoOrgType" \
    --organization-id "$CREATED_TRUCKER_ID"
assert_failure

test_name "Create tractor trailer credential classification without organization-id fails"
xbe_json do tractor-trailer-credential-classifications create \
    --name "NoOrgId" \
    --organization-type "truckers"
assert_failure

test_name "Update without any fields fails"
xbe_json do tractor-trailer-credential-classifications update "$CREATED_CLASSIFICATION_ID"
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
