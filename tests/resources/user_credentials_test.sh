#!/bin/bash
#
# XBE CLI Integration Tests: User Credentials
#
# Tests CRUD operations for the user_credentials resource.
# User credentials track credentials (licenses, certifications) held by users.
#
# COVERAGE: All create/update attributes + list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_CREDENTIAL_ID=""
CREATED_BROKER_ID=""
CREATED_USER_ID=""
CREATED_CLASSIFICATION_ID=""

describe "Resource: user_credentials"

# ============================================================================
# Prerequisites - Create broker, user, and classification
# ============================================================================

test_name "Create prerequisite broker for user credentials tests"
BROKER_NAME=$(unique_name "UCTestBroker")

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

test_name "Create prerequisite user"
USER_NAME=$(unique_name "UCTestUser")
USER_EMAIL=$(unique_email)

xbe_json do users create \
    --name "$USER_NAME" \
    --email "$USER_EMAIL"

if [[ $status -eq 0 ]]; then
    CREATED_USER_ID=$(json_get ".id")
    if [[ -n "$CREATED_USER_ID" && "$CREATED_USER_ID" != "null" ]]; then
        # Note: Users cannot be deleted via API, so we don't register for cleanup
        pass
    else
        fail "Created user but no ID returned"
        echo "Cannot continue without a user"
        run_tests
    fi
else
    fail "Failed to create user"
    echo "Cannot continue without a user"
    run_tests
fi

test_name "Create prerequisite user credential classification"
CLASS_NAME=$(unique_name "UCTestClass")

xbe_json do user-credential-classifications create \
    --name "$CLASS_NAME" \
    --organization-type "brokers" \
    --organization-id "$CREATED_BROKER_ID"

if [[ $status -eq 0 ]]; then
    CREATED_CLASSIFICATION_ID=$(json_get ".id")
    if [[ -n "$CREATED_CLASSIFICATION_ID" && "$CREATED_CLASSIFICATION_ID" != "null" ]]; then
        register_cleanup "user-credential-classifications" "$CREATED_CLASSIFICATION_ID"
        pass
    else
        fail "Created classification but no ID returned"
        echo "Cannot continue without a classification"
        run_tests
    fi
else
    fail "Failed to create user credential classification"
    echo "Cannot continue without a classification"
    run_tests
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create user credential with required fields"

xbe_json do user-credentials create \
    --user "$CREATED_USER_ID" \
    --user-credential-classification "$CREATED_CLASSIFICATION_ID"

if [[ $status -eq 0 ]]; then
    CREATED_CREDENTIAL_ID=$(json_get ".id")
    if [[ -n "$CREATED_CREDENTIAL_ID" && "$CREATED_CREDENTIAL_ID" != "null" ]]; then
        register_cleanup "user-credentials" "$CREATED_CREDENTIAL_ID"
        pass
    else
        fail "Created user credential but no ID returned"
    fi
else
    fail "Failed to create user credential"
fi

# Only continue if we successfully created a credential
if [[ -z "$CREATED_CREDENTIAL_ID" || "$CREATED_CREDENTIAL_ID" == "null" ]]; then
    echo "Cannot continue without a valid user credential ID"
    run_tests
fi

# NOTE: Each create test uses a unique classification because credentials
# for the same user+classification cannot have overlapping date ranges

test_name "Create user credential with issued-on"
# Create a unique classification to avoid overlap
CLASS_NAME2=$(unique_name "UCTestClass2")
xbe_json do user-credential-classifications create \
    --name "$CLASS_NAME2" \
    --organization-type "brokers" \
    --organization-id "$CREATED_BROKER_ID"
if [[ $status -eq 0 ]]; then
    CLASS_ID2=$(json_get ".id")
    register_cleanup "user-credential-classifications" "$CLASS_ID2"
    xbe_json do user-credentials create \
        --user "$CREATED_USER_ID" \
        --user-credential-classification "$CLASS_ID2" \
        --issued-on "2024-01-15"
    if [[ $status -eq 0 ]]; then
        id=$(json_get ".id")
        register_cleanup "user-credentials" "$id"
        pass
    else
        fail "Failed to create user credential with issued-on"
    fi
else
    fail "Failed to create classification for issued-on test"
fi

test_name "Create user credential with expires-on"
# Create a unique classification to avoid overlap
CLASS_NAME3=$(unique_name "UCTestClass3")
xbe_json do user-credential-classifications create \
    --name "$CLASS_NAME3" \
    --organization-type "brokers" \
    --organization-id "$CREATED_BROKER_ID"
if [[ $status -eq 0 ]]; then
    CLASS_ID3=$(json_get ".id")
    register_cleanup "user-credential-classifications" "$CLASS_ID3"
    xbe_json do user-credentials create \
        --user "$CREATED_USER_ID" \
        --user-credential-classification "$CLASS_ID3" \
        --expires-on "2025-01-15"
    if [[ $status -eq 0 ]]; then
        id=$(json_get ".id")
        register_cleanup "user-credentials" "$id"
        pass
    else
        fail "Failed to create user credential with expires-on"
    fi
else
    fail "Failed to create classification for expires-on test"
fi

test_name "Create user credential with both dates"
# Create a unique classification to avoid overlap
CLASS_NAME4=$(unique_name "UCTestClass4")
xbe_json do user-credential-classifications create \
    --name "$CLASS_NAME4" \
    --organization-type "brokers" \
    --organization-id "$CREATED_BROKER_ID"
if [[ $status -eq 0 ]]; then
    CLASS_ID4=$(json_get ".id")
    register_cleanup "user-credential-classifications" "$CLASS_ID4"
    xbe_json do user-credentials create \
        --user "$CREATED_USER_ID" \
        --user-credential-classification "$CLASS_ID4" \
        --issued-on "2024-01-01" \
        --expires-on "2025-12-31"
    if [[ $status -eq 0 ]]; then
        id=$(json_get ".id")
        register_cleanup "user-credentials" "$id"
        pass
    else
        fail "Failed to create user credential with both dates"
    fi
else
    fail "Failed to create classification for both dates test"
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update user credential issued-on"
xbe_json do user-credentials update "$CREATED_CREDENTIAL_ID" --issued-on "2024-02-01"
assert_success

test_name "Update user credential expires-on"
xbe_json do user-credentials update "$CREATED_CREDENTIAL_ID" --expires-on "2026-02-01"
assert_success

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List user credentials"
xbe_json view user-credentials list --limit 5
assert_success

test_name "List user credentials returns array"
xbe_json view user-credentials list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list user credentials"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List user credentials with --user filter"
xbe_json view user-credentials list --user "$CREATED_USER_ID" --limit 10
assert_success

# ============================================================================
# LIST Tests - Pagination
# ============================================================================

test_name "List user credentials with --limit"
xbe_json view user-credentials list --limit 3
assert_success

test_name "List user credentials with --offset"
xbe_json view user-credentials list --limit 3 --offset 3
assert_success

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete user credential requires --confirm flag"
xbe_run do user-credentials delete "$CREATED_CREDENTIAL_ID"
assert_failure

test_name "Delete user credential with --confirm"
# Create a credential specifically for deletion (using unique classification)
CLASS_NAME_DEL=$(unique_name "UCTestClassDel")
xbe_json do user-credential-classifications create \
    --name "$CLASS_NAME_DEL" \
    --organization-type "brokers" \
    --organization-id "$CREATED_BROKER_ID"
if [[ $status -eq 0 ]]; then
    CLASS_ID_DEL=$(json_get ".id")
    register_cleanup "user-credential-classifications" "$CLASS_ID_DEL"
    xbe_json do user-credentials create \
        --user "$CREATED_USER_ID" \
        --user-credential-classification "$CLASS_ID_DEL"
    if [[ $status -eq 0 ]]; then
        DEL_ID=$(json_get ".id")
        xbe_run do user-credentials delete "$DEL_ID" --confirm
        assert_success
    else
        skip "Could not create user credential for deletion test"
    fi
else
    skip "Could not create classification for deletion test"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create user credential without user fails"
xbe_json do user-credentials create \
    --user-credential-classification "$CREATED_CLASSIFICATION_ID"
assert_failure

test_name "Create user credential without classification fails"
xbe_json do user-credentials create \
    --user "$CREATED_USER_ID"
assert_failure

test_name "Update without any fields fails"
xbe_json do user-credentials update "$CREATED_CREDENTIAL_ID"
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
