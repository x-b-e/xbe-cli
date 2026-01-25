#!/bin/bash
#
# XBE CLI Integration Tests: Change Requests
#
# Tests CRUD operations for the change-requests resource.
#
# COVERAGE: All create/update attributes + list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_CHANGE_REQUEST_ID=""
CREATED_BROKER_ID=""
CREATED_MEMBERSHIP_ID=""
CURRENT_USER_ID=""
HAS_BROKER_MEMBERSHIP=0

REQUESTS_CREATE='[{"field":"status","from":"draft","to":"approved"}]'
REQUESTS_UPDATE='[{"field":"status","from":"approved","to":"rejected"}]'

describe "Resource: change-requests"

# ============================================================================
# Prerequisites
# ============================================================================

test_name "Resolve current user for change request tests"
xbe_json auth whoami
if [[ $status -eq 0 ]]; then
    CURRENT_USER_ID=$(json_get ".id")
    if [[ -n "$CURRENT_USER_ID" && "$CURRENT_USER_ID" != "null" ]]; then
        pass
    else
        skip "No user ID returned from auth whoami"
    fi
else
    skip "Unable to resolve current user"
fi

test_name "Resolve broker for change request tests"
if [[ -n "$XBE_TEST_BROKER_ID" ]]; then
    CREATED_BROKER_ID="$XBE_TEST_BROKER_ID"
    echo "    Using XBE_TEST_BROKER_ID: $CREATED_BROKER_ID"
    pass
else
    BROKER_NAME=$(unique_name "ChangeRequestBroker")
    xbe_json do brokers create --name "$BROKER_NAME"
    if [[ $status -eq 0 ]]; then
        CREATED_BROKER_ID=$(json_get ".id")
        if [[ -n "$CREATED_BROKER_ID" && "$CREATED_BROKER_ID" != "null" ]]; then
            register_cleanup "brokers" "$CREATED_BROKER_ID"
            pass
        else
            fail "Created broker but no ID returned"
        fi
    else
        skip "Failed to create broker"
    fi
fi

if [[ -n "$CURRENT_USER_ID" && -n "$CREATED_BROKER_ID" ]]; then
    test_name "Ensure current user membership for broker"
    xbe_json view memberships list --user "$CURRENT_USER_ID" --limit 50
    if [[ $status -eq 0 ]]; then
        EXISTING_MEMBERSHIP_ID=$(echo "$output" | jq -r --arg broker_id "$CREATED_BROKER_ID" '.[] | select((.organization_type=="Broker" or .organization_type=="brokers") and .organization_id==$broker_id) | .id' | head -n 1)
        if [[ -n "$EXISTING_MEMBERSHIP_ID" && "$EXISTING_MEMBERSHIP_ID" != "null" ]]; then
            HAS_BROKER_MEMBERSHIP=1
            pass
        else
            xbe_json do memberships create --user "$CURRENT_USER_ID" --organization "Broker|$CREATED_BROKER_ID"
            if [[ $status -eq 0 ]]; then
                CREATED_MEMBERSHIP_ID=$(json_get ".id")
                if [[ -n "$CREATED_MEMBERSHIP_ID" && "$CREATED_MEMBERSHIP_ID" != "null" ]]; then
                    register_cleanup "memberships" "$CREATED_MEMBERSHIP_ID"
                    HAS_BROKER_MEMBERSHIP=1
                    pass
                else
                    skip "Created membership but no ID returned"
                fi
            else
                skip "Failed to create membership for current user"
            fi
        fi
    else
        skip "Failed to list memberships for current user"
    fi
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create change request"
if [[ -n "$CREATED_BROKER_ID" ]]; then
    xbe_json do change-requests create \
        --requests "$REQUESTS_CREATE" \
        --organization-type Broker \
        --organization-id "$CREATED_BROKER_ID"
    if [[ $status -ne 0 ]]; then
        xbe_json do change-requests create --requests "$REQUESTS_CREATE"
    fi
else
    xbe_json do change-requests create --requests "$REQUESTS_CREATE"
fi

if [[ $status -eq 0 ]]; then
    CREATED_CHANGE_REQUEST_ID=$(json_get ".id")
    if [[ -n "$CREATED_CHANGE_REQUEST_ID" && "$CREATED_CHANGE_REQUEST_ID" != "null" ]]; then
        register_cleanup "change-requests" "$CREATED_CHANGE_REQUEST_ID"
        pass
    else
        fail "Created change request but no ID returned"
    fi
else
    fail "Failed to create change request"
fi

# Only continue if we successfully created a change request
if [[ -z "$CREATED_CHANGE_REQUEST_ID" || "$CREATED_CHANGE_REQUEST_ID" == "null" ]]; then
    echo "Cannot continue without a valid change request ID"
    run_tests
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update change request --requests"
xbe_json do change-requests update "$CREATED_CHANGE_REQUEST_ID" --requests "$REQUESTS_UPDATE"
assert_success

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show change request"
xbe_json view change-requests show "$CREATED_CHANGE_REQUEST_ID"
assert_success

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List change requests"
xbe_json view change-requests list --limit 5
assert_success

test_name "List change requests returns array"
xbe_json view change-requests list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list change requests"
fi

# ============================================================================
# LIST Tests - All Filters
# ============================================================================

test_name "List change requests with --organization"
if [[ -n "$CREATED_BROKER_ID" ]]; then
    xbe_json view change-requests list --organization "Broker|$CREATED_BROKER_ID" --limit 5
    assert_success
else
    skip "No broker available for --organization filter"
fi

test_name "List change requests with --organization-type and --organization-id"
if [[ -n "$CREATED_BROKER_ID" ]]; then
    xbe_json view change-requests list --organization-type Broker --organization-id "$CREATED_BROKER_ID" --limit 5
    assert_success
else
    skip "No broker available for organization-type/id filters"
fi

test_name "List change requests with --not-organization-type"
xbe_json view change-requests list --not-organization-type Trucker --limit 5
assert_success

test_name "List change requests with --broker"
if [[ -n "$CREATED_BROKER_ID" ]]; then
    xbe_json view change-requests list --broker "$CREATED_BROKER_ID" --limit 5
    assert_success
else
    skip "No broker available for --broker filter"
fi

test_name "List change requests with --created-by"
if [[ -n "$CURRENT_USER_ID" && "$CURRENT_USER_ID" != "null" ]]; then
    xbe_json view change-requests list --created-by "$CURRENT_USER_ID" --limit 5
    assert_success
else
    skip "No current user ID available for --created-by filter"
fi

run_tests
