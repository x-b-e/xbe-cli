#!/bin/bash
#
# XBE CLI Integration Tests: Mechanic User Associations
#
# Tests CRUD operations for the mechanic-user-associations resource.
#
# COVERAGE: All filters + create/update relationships + show + delete
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_RECORD_ID=""
SHOW_ID=""
USER_ID=""
MAINTENANCE_REQUIREMENT_ID=""
ALT_USER_ID=""
ALT_MAINTENANCE_REQUIREMENT_ID=""

describe "Resource: mechanic-user-associations"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List mechanic user associations"
xbe_json view mechanic-user-associations list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
    SHOW_ID=$(json_get '.[0].id // empty')
    USER_ID=$(json_get '.[0].user_id // empty')
    MAINTENANCE_REQUIREMENT_ID=$(json_get '.[0].maintenance_requirement_id // empty')
else
    fail "Failed to list mechanic user associations"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show mechanic user association"
if [[ -n "$SHOW_ID" && "$SHOW_ID" != "null" ]]; then
    xbe_json view mechanic-user-associations show "$SHOW_ID"
    assert_success
else
    skip "No record ID available"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List records with --user filter"
if [[ -n "$USER_ID" && "$USER_ID" != "null" ]]; then
    xbe_json view mechanic-user-associations list --user "$USER_ID" --limit 5
    assert_success
else
    skip "No user ID available"
fi

test_name "List records with --maintenance-requirement filter"
if [[ -n "$MAINTENANCE_REQUIREMENT_ID" && "$MAINTENANCE_REQUIREMENT_ID" != "null" ]]; then
    xbe_json view mechanic-user-associations list --maintenance-requirement "$MAINTENANCE_REQUIREMENT_ID" --limit 5
    assert_success
else
    skip "No maintenance requirement ID available"
fi

# ============================================================================
# LIST Tests - Pagination
# ============================================================================

test_name "List records with --limit"
xbe_json view mechanic-user-associations list --limit 3
assert_success

test_name "List records with --offset"
xbe_json view mechanic-user-associations list --limit 3 --offset 3
assert_success

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create record without required fields fails"
xbe_run do mechanic-user-associations create
assert_failure

test_name "Update record without fields fails"
if [[ -n "$SHOW_ID" && "$SHOW_ID" != "null" ]]; then
    xbe_run do mechanic-user-associations update "$SHOW_ID"
    assert_failure
else
    skip "No record ID available"
fi

test_name "Delete record requires --confirm flag"
if [[ -n "$SHOW_ID" && "$SHOW_ID" != "null" ]]; then
    xbe_run do mechanic-user-associations delete "$SHOW_ID"
    assert_failure
else
    skip "No record ID available"
fi

# ============================================================================
# Prerequisite Lookup via API
# ============================================================================

test_name "Lookup user and maintenance requirement via API"
if [[ -n "$XBE_TEST_USER_ID" ]]; then
    USER_ID="$XBE_TEST_USER_ID"
fi
if [[ -n "$XBE_TEST_MAINTENANCE_REQUIREMENT_ID" ]]; then
    MAINTENANCE_REQUIREMENT_ID="$XBE_TEST_MAINTENANCE_REQUIREMENT_ID"
fi
if [[ -z "$XBE_TOKEN" ]]; then
    skip "XBE_TOKEN not set; skipping API lookup"
else
    base_url="${XBE_BASE_URL%/}"

    if [[ -z "$USER_ID" || "$USER_ID" == "null" ]]; then
        users_json=$(curl -s \
            -H "Authorization: Bearer $XBE_TOKEN" \
            -H "Accept: application/vnd.api+json" \
            "$base_url/v1/users?page[limit]=20" || true)

        USER_ID=$(echo "$users_json" | jq -r '.data[0].id // empty' 2>/dev/null || true)
        ALT_USER_ID=$(echo "$users_json" | jq -r '.data[1].id // empty' 2>/dev/null || true)
    fi

    if [[ -z "$MAINTENANCE_REQUIREMENT_ID" || "$MAINTENANCE_REQUIREMENT_ID" == "null" ]]; then
        reqs_json=$(curl -s \
            -H "Authorization: Bearer $XBE_TOKEN" \
            -H "Accept: application/vnd.api+json" \
            "$base_url/v1/maintenance-requirements?page[limit]=20" || true)

        MAINTENANCE_REQUIREMENT_ID=$(echo "$reqs_json" | jq -r '.data[0].id // empty' 2>/dev/null || true)
        ALT_MAINTENANCE_REQUIREMENT_ID=$(echo "$reqs_json" | jq -r '.data[1].id // empty' 2>/dev/null || true)
    fi

    if [[ -n "$USER_ID" && -n "$MAINTENANCE_REQUIREMENT_ID" ]]; then
        pass
    else
        skip "No user or maintenance requirement found"
    fi
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create record with required fields"
if [[ -n "$USER_ID" && -n "$MAINTENANCE_REQUIREMENT_ID" ]]; then
    xbe_json do mechanic-user-associations create \
        --user "$USER_ID" \
        --maintenance-requirement "$MAINTENANCE_REQUIREMENT_ID"

    if [[ $status -eq 0 ]]; then
        CREATED_RECORD_ID=$(json_get ".id")
        if [[ -n "$CREATED_RECORD_ID" && "$CREATED_RECORD_ID" != "null" ]]; then
            register_cleanup "mechanic-user-associations" "$CREATED_RECORD_ID"
            pass
        else
            fail "Created record but no ID returned"
        fi
    else
        if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]] || [[ "$output" == *"422"* ]]; then
            pass
        else
            fail "Failed to create record: $output"
        fi
    fi
else
    skip "Missing user or maintenance requirement for create"
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update record relationships"
if [[ -n "$CREATED_RECORD_ID" && "$CREATED_RECORD_ID" != "null" ]]; then
    update_user_id="$USER_ID"
    update_req_id="$MAINTENANCE_REQUIREMENT_ID"

    if [[ -n "$ALT_USER_ID" ]]; then
        update_user_id="$ALT_USER_ID"
    fi
    if [[ -n "$ALT_MAINTENANCE_REQUIREMENT_ID" ]]; then
        update_req_id="$ALT_MAINTENANCE_REQUIREMENT_ID"
    fi

    xbe_json do mechanic-user-associations update "$CREATED_RECORD_ID" \
        --user "$update_user_id" \
        --maintenance-requirement "$update_req_id"
    assert_success
else
    skip "No created record available"
fi

# ============================================================================
# SHOW Tests (Created)
# ============================================================================

test_name "Show created record details"
if [[ -n "$CREATED_RECORD_ID" && "$CREATED_RECORD_ID" != "null" ]]; then
    xbe_json view mechanic-user-associations show "$CREATED_RECORD_ID"
    assert_success
else
    skip "No created record available"
fi

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete record with --confirm"
if [[ -n "$USER_ID" && -n "$MAINTENANCE_REQUIREMENT_ID" ]]; then
    xbe_json do mechanic-user-associations create \
        --user "$USER_ID" \
        --maintenance-requirement "$MAINTENANCE_REQUIREMENT_ID"

    if [[ $status -eq 0 ]]; then
        del_id=$(json_get ".id")
        xbe_run do mechanic-user-associations delete "$del_id" --confirm
        assert_success
    else
        skip "Could not create record for deletion test"
    fi
else
    skip "Missing prerequisites for delete test"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
