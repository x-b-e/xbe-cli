#!/bin/bash
#
# XBE CLI Integration Tests: User Languages
#
# Tests CRUD operations for the user-languages resource.
# User languages associate users with preferred languages and default settings.
#
# COVERAGE: All create/update attributes + list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_USER_LANGUAGE_ID=""
CREATED_USER_ID=""
LANGUAGE_ID=""

describe "Resource: user-languages"

# ============================================================================
# Prerequisites - Create user and locate language
# ============================================================================

test_name "Create prerequisite user"
USER_NAME=$(unique_name "UserLanguage")
USER_EMAIL=$(unique_email)

xbe_json do users create \
    --name "$USER_NAME" \
    --email "$USER_EMAIL"

if [[ $status -eq 0 ]]; then
    CREATED_USER_ID=$(json_get ".id")
    if [[ -n "$CREATED_USER_ID" && "$CREATED_USER_ID" != "null" ]]; then
        # Note: Users cannot be deleted via API, so we don't register cleanup
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

test_name "Locate language ID"
xbe_json view languages list --code "en"
if [[ $status -eq 0 ]]; then
    LANGUAGE_ID=$(json_get ".[0].id")
fi

if [[ -z "$LANGUAGE_ID" || "$LANGUAGE_ID" == "null" ]]; then
    xbe_json view languages list
    if [[ $status -eq 0 ]]; then
        LANGUAGE_ID=$(json_get ".[0].id")
    fi
fi

if [[ -n "$LANGUAGE_ID" && "$LANGUAGE_ID" != "null" ]]; then
    pass
else
    fail "Failed to locate a language ID"
    run_tests
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create user language with required fields"

xbe_json do user-languages create \
    --user "$CREATED_USER_ID" \
    --language "$LANGUAGE_ID" \
    --is-default true

if [[ $status -eq 0 ]]; then
    CREATED_USER_LANGUAGE_ID=$(json_get ".id")
    if [[ -n "$CREATED_USER_LANGUAGE_ID" && "$CREATED_USER_LANGUAGE_ID" != "null" ]]; then
        register_cleanup "user-languages" "$CREATED_USER_LANGUAGE_ID"
        pass
    else
        fail "Created user language but no ID returned"
    fi
else
    fail "Failed to create user language"
fi

if [[ -z "$CREATED_USER_LANGUAGE_ID" || "$CREATED_USER_LANGUAGE_ID" == "null" ]]; then
    echo "Cannot continue without a valid user language ID"
    run_tests
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show user language"
xbe_json view user-languages show "$CREATED_USER_LANGUAGE_ID"
assert_success

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List user languages"
xbe_json view user-languages list
assert_success

test_name "List user languages with --user"
xbe_json view user-languages list --user "$CREATED_USER_ID"
assert_success

test_name "List user languages with --language"
xbe_json view user-languages list --language "$LANGUAGE_ID"
assert_success

test_name "List user languages with --is-default"
xbe_json view user-languages list --is-default true
assert_success

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update user language is-default"
xbe_json do user-languages update "$CREATED_USER_LANGUAGE_ID" --is-default false
assert_success

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete user language"
xbe_json do user-languages delete "$CREATED_USER_LANGUAGE_ID" --confirm
assert_success

# ============================================================================
# Summary
# ============================================================================

run_tests
