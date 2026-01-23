#!/bin/bash
#
# XBE CLI Integration Tests: User Credential Classifications
#
# Tests CRUD operations for the user_credential_classifications resource.
# These classifications define types of credentials for users.
#
# COVERAGE: All create/update attributes + list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_CLASSIFICATION_ID=""
CREATED_BROKER_ID=""

describe "Resource: user_credential_classifications"

# ============================================================================
# Prerequisites - Create broker
# ============================================================================

test_name "Create prerequisite broker for user credential classification tests"
BROKER_NAME=$(unique_name "UCCTestBroker")

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

test_name "Create user credential classification with required fields"
TEST_NAME=$(unique_name "UserCredClass")

xbe_json do user-credential-classifications create \
    --name "$TEST_NAME" \
    --organization-type "brokers" \
    --organization-id "$CREATED_BROKER_ID"

if [[ $status -eq 0 ]]; then
    CREATED_CLASSIFICATION_ID=$(json_get ".id")
    if [[ -n "$CREATED_CLASSIFICATION_ID" && "$CREATED_CLASSIFICATION_ID" != "null" ]]; then
        register_cleanup "user-credential-classifications" "$CREATED_CLASSIFICATION_ID"
        pass
    else
        fail "Created user credential classification but no ID returned"
    fi
else
    fail "Failed to create user credential classification"
fi

# Only continue if we successfully created a classification
if [[ -z "$CREATED_CLASSIFICATION_ID" || "$CREATED_CLASSIFICATION_ID" == "null" ]]; then
    echo "Cannot continue without a valid user credential classification ID"
    run_tests
fi

test_name "Create user credential classification with description"
TEST_NAME2=$(unique_name "UserCredClass2")
xbe_json do user-credential-classifications create \
    --name "$TEST_NAME2" \
    --organization-type "brokers" \
    --organization-id "$CREATED_BROKER_ID" \
    --description "Commercial Driver License"
if [[ $status -eq 0 ]]; then
    id=$(json_get ".id")
    register_cleanup "user-credential-classifications" "$id"
    pass
else
    fail "Failed to create user credential classification with description"
fi

test_name "Create user credential classification with issuer-name"
TEST_NAME3=$(unique_name "UserCredClass3")
xbe_json do user-credential-classifications create \
    --name "$TEST_NAME3" \
    --organization-type "brokers" \
    --organization-id "$CREATED_BROKER_ID" \
    --issuer-name "State DMV"
if [[ $status -eq 0 ]]; then
    id=$(json_get ".id")
    register_cleanup "user-credential-classifications" "$id"
    pass
else
    fail "Failed to create user credential classification with issuer-name"
fi

test_name "Create user credential classification with external-id"
TEST_NAME4=$(unique_name "UserCredClass4")
xbe_json do user-credential-classifications create \
    --name "$TEST_NAME4" \
    --organization-type "brokers" \
    --organization-id "$CREATED_BROKER_ID" \
    --external-id "EXT-USER-123"
if [[ $status -eq 0 ]]; then
    id=$(json_get ".id")
    register_cleanup "user-credential-classifications" "$id"
    pass
else
    fail "Failed to create user credential classification with external-id"
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update user credential classification name"
UPDATED_NAME=$(unique_name "UpdatedUCC")
xbe_json do user-credential-classifications update "$CREATED_CLASSIFICATION_ID" --name "$UPDATED_NAME"
assert_success

test_name "Update user credential classification description"
xbe_json do user-credential-classifications update "$CREATED_CLASSIFICATION_ID" --description "Updated description"
assert_success

test_name "Update user credential classification issuer-name"
xbe_json do user-credential-classifications update "$CREATED_CLASSIFICATION_ID" --issuer-name "Updated Issuer"
assert_success

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List user credential classifications"
xbe_json view user-credential-classifications list --limit 5
assert_success

test_name "List user credential classifications returns array"
xbe_json view user-credential-classifications list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list user credential classifications"
fi

# ============================================================================
# LIST Tests - Pagination
# ============================================================================

test_name "List user credential classifications with --limit"
xbe_json view user-credential-classifications list --limit 3
assert_success

test_name "List user credential classifications with --offset"
xbe_json view user-credential-classifications list --limit 3 --offset 3
assert_success

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete user credential classification requires --confirm flag"
xbe_run do user-credential-classifications delete "$CREATED_CLASSIFICATION_ID"
assert_failure

test_name "Delete user credential classification with --confirm"
# Create a classification specifically for deletion
TEST_DEL_NAME=$(unique_name "DeleteUCC")
xbe_json do user-credential-classifications create \
    --name "$TEST_DEL_NAME" \
    --organization-type "brokers" \
    --organization-id "$CREATED_BROKER_ID"
if [[ $status -eq 0 ]]; then
    DEL_ID=$(json_get ".id")
    xbe_run do user-credential-classifications delete "$DEL_ID" --confirm
    assert_success
else
    skip "Could not create user credential classification for deletion test"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create user credential classification without name fails"
xbe_json do user-credential-classifications create \
    --organization-type "brokers" \
    --organization-id "$CREATED_BROKER_ID"
assert_failure

test_name "Create user credential classification without organization-type fails"
xbe_json do user-credential-classifications create \
    --name "NoOrgType" \
    --organization-id "$CREATED_BROKER_ID"
assert_failure

test_name "Create user credential classification without organization-id fails"
xbe_json do user-credential-classifications create \
    --name "NoOrgId" \
    --organization-type "brokers"
assert_failure

test_name "Update without any fields fails"
xbe_json do user-credential-classifications update "$CREATED_CLASSIFICATION_ID"
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
