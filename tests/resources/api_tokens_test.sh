#!/bin/bash
#
# XBE CLI Integration Tests: API Tokens
#
# Tests create, update, and list filters for the api-tokens resource.
#
# COVERAGE: All create/update attributes + list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

# Override cleanup trap since API tokens cannot be deleted
trap - EXIT
trap cleanup_api_tokens EXIT

CREATED_USER_ID=""
CREATED_TOKEN_IDS=()

cleanup_api_tokens() {
    if [[ ${#CREATED_TOKEN_IDS[@]} -gt 0 ]]; then
        echo ""
        echo -e "${YELLOW}Revoking test API tokens...${NC}"
        local now
        now=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
        for token_id in "${CREATED_TOKEN_IDS[@]}"; do
            if [[ -n "$token_id" && "$token_id" != "null" ]]; then
                xbe_run do api-tokens update "$token_id" --revoked-at "$now" >/dev/null 2>&1 || true
            fi
        done
        echo -e "${GREEN}Cleanup complete.${NC}"
    fi
}

describe "Resource: api_tokens"

# ============================================================================
# Prerequisites - Create user
# ============================================================================

test_name "Create prerequisite user"
USER_NAME=$(unique_name "ApiTokenUser")
USER_EMAIL=$(unique_email)

xbe_json do users create \
    --name "$USER_NAME" \
    --email "$USER_EMAIL"

if [[ $status -eq 0 ]]; then
    CREATED_USER_ID=$(json_get ".id")
    if [[ -n "$CREATED_USER_ID" && "$CREATED_USER_ID" != "null" ]]; then
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

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create API token with required fields"

xbe_json do api-tokens create --user "$CREATED_USER_ID"

if [[ $status -eq 0 ]]; then
    TOKEN_ID_REQUIRED=$(json_get ".id")
    TOKEN_VALUE=$(json_get ".token")
    if [[ -n "$TOKEN_ID_REQUIRED" && "$TOKEN_ID_REQUIRED" != "null" && -n "$TOKEN_VALUE" && "$TOKEN_VALUE" != "null" ]]; then
        CREATED_TOKEN_IDS+=("$TOKEN_ID_REQUIRED")
        pass
    else
        fail "Created API token but missing ID or token value"
    fi
else
    fail "Failed to create API token"
fi

# Only continue if we successfully created a token
if [[ -z "$TOKEN_ID_REQUIRED" || "$TOKEN_ID_REQUIRED" == "null" ]]; then
    echo "Cannot continue without a valid API token ID"
    run_tests
fi

test_name "Create API token with name"
TOKEN_NAME=$(unique_name "ApiToken")

xbe_json do api-tokens create --user "$CREATED_USER_ID" --name "$TOKEN_NAME"

if [[ $status -eq 0 ]]; then
    TOKEN_ID_NAME=$(json_get ".id")
    if [[ -n "$TOKEN_ID_NAME" && "$TOKEN_ID_NAME" != "null" ]]; then
        CREATED_TOKEN_IDS+=("$TOKEN_ID_NAME")
        assert_json_equals ".name" "$TOKEN_NAME"
    else
        fail "Created API token but no ID returned"
    fi
else
    fail "Failed to create API token with name"
fi

test_name "Create API token with expires-at"
EXPIRES_AT="2030-01-01T00:00:00Z"

xbe_json do api-tokens create --user "$CREATED_USER_ID" --expires-at "$EXPIRES_AT"

if [[ $status -eq 0 ]]; then
    TOKEN_ID_EXPIRES=$(json_get ".id")
    if [[ -n "$TOKEN_ID_EXPIRES" && "$TOKEN_ID_EXPIRES" != "null" ]]; then
        CREATED_TOKEN_IDS+=("$TOKEN_ID_EXPIRES")
        EXPIRES_AT_PREFIX="${EXPIRES_AT%Z}"
        if echo "$output" | jq -e --arg ts "$EXPIRES_AT_PREFIX" '.expires_at | startswith($ts)' >/dev/null; then
            pass
        else
            fail "expires_at did not match expected prefix"
        fi
    else
        fail "Created API token but no ID returned"
    fi
else
    fail "Failed to create API token with expires-at"
fi

# ============================================================================
# LIST/SHOW Tests
# ============================================================================

test_name "List API tokens filtered by user"

xbe_json view api-tokens list --user "$CREATED_USER_ID"

if [[ $status -eq 0 ]]; then
    if echo "$output" | jq -e --arg id "$TOKEN_ID_REQUIRED" 'any(.[]; .id == $id)' >/dev/null; then
        pass
    else
        fail "Filtered list did not include expected API token ID"
    fi
else
    fail "Failed to list API tokens with user filter"
fi

test_name "Show API token details"

xbe_json view api-tokens show "$TOKEN_ID_REQUIRED"

if [[ $status -eq 0 ]]; then
    assert_json_equals ".id" "$TOKEN_ID_REQUIRED"
else
    fail "Failed to show API token"
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Revoke API token"
REVOKED_AT=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

xbe_json do api-tokens update "$TOKEN_ID_REQUIRED" --revoked-at "$REVOKED_AT"

if [[ $status -eq 0 ]]; then
    REVOKED_AT_PREFIX="${REVOKED_AT%Z}"
    if echo "$output" | jq -e --arg ts "$REVOKED_AT_PREFIX" '.revoked_at | startswith($ts)' >/dev/null; then
        pass
    else
        fail "revoked_at did not match expected prefix"
    fi
else
    fail "Failed to revoke API token"
fi

run_tests
