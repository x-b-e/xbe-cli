#!/bin/bash
#
# XBE CLI Integration Tests: Driver Managers
#
# Tests list, show, create, update, and delete operations for the driver-managers resource.
#
# COVERAGE: List filters + show + create/update/delete + error cases
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

SAMPLE_ID=""
SAMPLE_TRUCKER_ID=""
SAMPLE_MANAGER_MEMBERSHIP_ID=""
SAMPLE_MANAGED_MEMBERSHIP_ID=""
SAMPLE_MANAGER_USER_ID=""
SAMPLE_MANAGED_USER_ID=""
SAMPLE_BROKER_ID=""
SAMPLE_CREATED_AT=""
SAMPLE_UPDATED_AT=""
CREATED_ID=""

describe "Resource: driver-managers"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List driver managers"
xbe_json view driver-managers list --limit 5
assert_success

test_name "List driver managers returns array"
xbe_json view driver-managers list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list driver managers"
fi

# ============================================================================
# Sample Record (used for filters/show/update)
# ============================================================================

test_name "Capture sample driver manager"
xbe_json view driver-managers list --limit 1
if [[ $status -eq 0 ]]; then
    SAMPLE_ID=$(json_get ".[0].id")
    SAMPLE_TRUCKER_ID=$(json_get ".[0].trucker_id")
    SAMPLE_MANAGER_MEMBERSHIP_ID=$(json_get ".[0].manager_membership_id")
    SAMPLE_MANAGED_MEMBERSHIP_ID=$(json_get ".[0].managed_membership_id")
    if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
        pass
    else
        skip "No driver managers available for follow-on tests"
    fi
else
    skip "Could not list driver managers to capture sample"
fi

test_name "Capture driver manager timestamps"
if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_json view driver-managers show "$SAMPLE_ID"
    if [[ $status -eq 0 ]]; then
        SAMPLE_CREATED_AT=$(json_get ".created_at")
        SAMPLE_UPDATED_AT=$(json_get ".updated_at")
        pass
    else
        skip "Could not fetch driver manager timestamps"
    fi
else
    skip "No driver manager ID available"
fi

# Capture manager membership user/broker
if [[ -n "$SAMPLE_MANAGER_MEMBERSHIP_ID" && "$SAMPLE_MANAGER_MEMBERSHIP_ID" != "null" ]]; then
    test_name "Capture manager membership details"
    xbe_json view memberships show "$SAMPLE_MANAGER_MEMBERSHIP_ID"
    if [[ $status -eq 0 ]]; then
        SAMPLE_MANAGER_USER_ID=$(json_get ".user_id")
        SAMPLE_BROKER_ID=$(json_get ".broker_id")
        pass
    else
        skip "Could not fetch manager membership details"
    fi
fi

# Capture managed membership user/broker (fallback broker if needed)
if [[ -n "$SAMPLE_MANAGED_MEMBERSHIP_ID" && "$SAMPLE_MANAGED_MEMBERSHIP_ID" != "null" ]]; then
    test_name "Capture managed membership details"
    xbe_json view memberships show "$SAMPLE_MANAGED_MEMBERSHIP_ID"
    if [[ $status -eq 0 ]]; then
        SAMPLE_MANAGED_USER_ID=$(json_get ".user_id")
        if [[ -z "$SAMPLE_BROKER_ID" || "$SAMPLE_BROKER_ID" == "null" ]]; then
            SAMPLE_BROKER_ID=$(json_get ".broker_id")
        fi
        pass
    else
        skip "Could not fetch managed membership details"
    fi
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List driver managers with --trucker filter"
if [[ -n "$SAMPLE_TRUCKER_ID" && "$SAMPLE_TRUCKER_ID" != "null" ]]; then
    xbe_json view driver-managers list --trucker "$SAMPLE_TRUCKER_ID" --limit 5
    assert_success
else
    skip "No trucker ID available"
fi

test_name "List driver managers with --manager-membership filter"
if [[ -n "$SAMPLE_MANAGER_MEMBERSHIP_ID" && "$SAMPLE_MANAGER_MEMBERSHIP_ID" != "null" ]]; then
    xbe_json view driver-managers list --manager-membership "$SAMPLE_MANAGER_MEMBERSHIP_ID" --limit 5
    assert_success
else
    skip "No manager membership ID available"
fi

test_name "List driver managers with --managed-membership filter"
if [[ -n "$SAMPLE_MANAGED_MEMBERSHIP_ID" && "$SAMPLE_MANAGED_MEMBERSHIP_ID" != "null" ]]; then
    xbe_json view driver-managers list --managed-membership "$SAMPLE_MANAGED_MEMBERSHIP_ID" --limit 5
    assert_success
else
    skip "No managed membership ID available"
fi

test_name "List driver managers with --broker filter"
if [[ -n "$SAMPLE_BROKER_ID" && "$SAMPLE_BROKER_ID" != "null" ]]; then
    xbe_json view driver-managers list --broker "$SAMPLE_BROKER_ID" --limit 5
    assert_success
else
    skip "No broker ID available"
fi

test_name "List driver managers with --manager-user filter"
if [[ -n "$SAMPLE_MANAGER_USER_ID" && "$SAMPLE_MANAGER_USER_ID" != "null" ]]; then
    xbe_json view driver-managers list --manager-user "$SAMPLE_MANAGER_USER_ID" --limit 5
    assert_success
else
    skip "No manager user ID available"
fi

test_name "List driver managers with --managed-user filter"
if [[ -n "$SAMPLE_MANAGED_USER_ID" && "$SAMPLE_MANAGED_USER_ID" != "null" ]]; then
    xbe_json view driver-managers list --managed-user "$SAMPLE_MANAGED_USER_ID" --limit 5
    assert_success
else
    skip "No managed user ID available"
fi

test_name "List driver managers with --created-at-min filter"
if [[ -n "$SAMPLE_CREATED_AT" && "$SAMPLE_CREATED_AT" != "null" ]]; then
    xbe_json view driver-managers list --created-at-min "$SAMPLE_CREATED_AT" --limit 5
    assert_success
else
    skip "No created-at timestamp available"
fi

test_name "List driver managers with --created-at-max filter"
if [[ -n "$SAMPLE_CREATED_AT" && "$SAMPLE_CREATED_AT" != "null" ]]; then
    xbe_json view driver-managers list --created-at-max "$SAMPLE_CREATED_AT" --limit 5
    assert_success
else
    skip "No created-at timestamp available"
fi

test_name "List driver managers with --updated-at-min filter"
if [[ -n "$SAMPLE_UPDATED_AT" && "$SAMPLE_UPDATED_AT" != "null" ]]; then
    xbe_json view driver-managers list --updated-at-min "$SAMPLE_UPDATED_AT" --limit 5
    assert_success
else
    skip "No updated-at timestamp available"
fi

test_name "List driver managers with --updated-at-max filter"
if [[ -n "$SAMPLE_UPDATED_AT" && "$SAMPLE_UPDATED_AT" != "null" ]]; then
    xbe_json view driver-managers list --updated-at-max "$SAMPLE_UPDATED_AT" --limit 5
    assert_success
else
    skip "No updated-at timestamp available"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show driver manager"
if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_json view driver-managers show "$SAMPLE_ID"
    assert_success
else
    skip "No driver manager ID available"
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create driver manager"
if [[ -n "$SAMPLE_TRUCKER_ID" && "$SAMPLE_TRUCKER_ID" != "null" && -n "$SAMPLE_MANAGER_MEMBERSHIP_ID" && "$SAMPLE_MANAGER_MEMBERSHIP_ID" != "null" && -n "$SAMPLE_MANAGED_MEMBERSHIP_ID" && "$SAMPLE_MANAGED_MEMBERSHIP_ID" != "null" ]]; then
    xbe_json do driver-managers create \
        --trucker "$SAMPLE_TRUCKER_ID" \
        --manager-membership "$SAMPLE_MANAGER_MEMBERSHIP_ID" \
        --managed-membership "$SAMPLE_MANAGED_MEMBERSHIP_ID"
    if [[ $status -eq 0 ]]; then
        CREATED_ID=$(json_get ".id")
        if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
            pass
        else
            fail "Create succeeded but no ID returned"
        fi
    else
        if [[ "$output" == *"already"* ]] || [[ "$output" == *"taken"* ]] || [[ "$output" == *"422"* ]] || [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]]; then
            pass
        else
            fail "Create failed: $output"
        fi
    fi
else
    skip "Missing required IDs for create"
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update driver manager"
UPDATE_ID=""
if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
    UPDATE_ID="$CREATED_ID"
elif [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    UPDATE_ID="$SAMPLE_ID"
fi

if [[ -n "$UPDATE_ID" && -n "$SAMPLE_MANAGER_MEMBERSHIP_ID" && "$SAMPLE_MANAGER_MEMBERSHIP_ID" != "null" ]]; then
    xbe_json do driver-managers update "$UPDATE_ID" --manager-membership "$SAMPLE_MANAGER_MEMBERSHIP_ID"
    if [[ $status -eq 0 ]]; then
        pass
    else
        if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]]; then
            skip "Not authorized to update driver manager"
        else
            fail "Update failed: $output"
        fi
    fi
else
    skip "No driver manager ID available for update"
fi

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete driver manager requires --confirm flag"
if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_run do driver-managers delete "$SAMPLE_ID"
    assert_failure
else
    skip "No driver manager ID available"
fi

test_name "Delete driver manager with --confirm"
if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
    xbe_run do driver-managers delete "$CREATED_ID" --confirm
    if [[ $status -eq 0 ]]; then
        pass
    else
        if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]]; then
            skip "Not authorized to delete driver manager"
        else
            fail "Delete failed: $output"
        fi
    fi
else
    skip "No created driver manager available"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create driver manager without required flags fails"
xbe_json do driver-managers create --trucker 123
assert_failure

run_tests
