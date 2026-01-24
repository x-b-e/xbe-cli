#!/bin/bash
#
# XBE CLI Integration Tests: Developer Certified Weighers
#
# Tests list, show, create, update, and delete operations for developer-certified-weighers.
#
# COVERAGE: All list filters + writable attributes
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_BROKER_ID=""
CREATED_DEVELOPER_ID=""
CREATED_MATERIAL_SUPPLIER_ID=""
CREATED_USER_ID=""
CREATED_MEMBERSHIP_ID=""
CREATED_DEVELOPER_CERTIFIED_WEIGHER_ID=""

CERT_NUMBER=""
UPDATED_CERT_NUMBER=""

describe "Resource: developer-certified-weighers"

# ============================================================================
# Prerequisites - Create broker, developer, material supplier, user, membership
# ============================================================================

test_name "Create broker for developer certified weigher tests"
BROKER_NAME=$(unique_name "DCWBroker")

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

test_name "Create developer for developer certified weigher tests"
DEVELOPER_NAME=$(unique_name "DCWDeveloper")

xbe_json do developers create --name "$DEVELOPER_NAME" --broker "$CREATED_BROKER_ID"

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

test_name "Create material supplier for developer certified weigher tests"
MATERIAL_SUPPLIER_NAME=$(unique_name "DCWSupplier")

xbe_json do material-suppliers create --name "$MATERIAL_SUPPLIER_NAME" --broker "$CREATED_BROKER_ID"

if [[ $status -eq 0 ]]; then
    CREATED_MATERIAL_SUPPLIER_ID=$(json_get ".id")
    if [[ -n "$CREATED_MATERIAL_SUPPLIER_ID" && "$CREATED_MATERIAL_SUPPLIER_ID" != "null" ]]; then
        register_cleanup "material-suppliers" "$CREATED_MATERIAL_SUPPLIER_ID"
        pass
    else
        fail "Created material supplier but no ID returned"
        echo "Cannot continue without a material supplier"
        run_tests
    fi
else
    if [[ -n "$XBE_TEST_MATERIAL_SUPPLIER_ID" ]]; then
        CREATED_MATERIAL_SUPPLIER_ID="$XBE_TEST_MATERIAL_SUPPLIER_ID"
        echo "    Using XBE_TEST_MATERIAL_SUPPLIER_ID: $CREATED_MATERIAL_SUPPLIER_ID"
        pass
    else
        fail "Failed to create material supplier and XBE_TEST_MATERIAL_SUPPLIER_ID not set"
        echo "Cannot continue without a material supplier"
        run_tests
    fi
fi

test_name "Create user for developer certified weigher tests"
USER_NAME=$(unique_name "DCWUser")
USER_EMAIL=$(unique_email)

xbe_json do users create --name "$USER_NAME" --email "$USER_EMAIL"

if [[ $status -eq 0 ]]; then
    CREATED_USER_ID=$(json_get ".id")
    if [[ -n "$CREATED_USER_ID" && "$CREATED_USER_ID" != "null" ]]; then
        register_cleanup "users" "$CREATED_USER_ID"
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

test_name "Create material supplier membership for developer certified weigher tests"
if [[ -n "$CREATED_USER_ID" && -n "$CREATED_MATERIAL_SUPPLIER_ID" ]]; then
    xbe_json do memberships create --user "$CREATED_USER_ID" --organization "MaterialSupplier|$CREATED_MATERIAL_SUPPLIER_ID"
    if [[ $status -eq 0 ]]; then
        CREATED_MEMBERSHIP_ID=$(json_get ".id")
        if [[ -n "$CREATED_MEMBERSHIP_ID" && "$CREATED_MEMBERSHIP_ID" != "null" ]]; then
            register_cleanup "memberships" "$CREATED_MEMBERSHIP_ID"
            pass
        else
            fail "Created membership but no ID returned"
            echo "Cannot continue without a membership"
            run_tests
        fi
    else
        fail "Failed to create membership"
        echo "Cannot continue without a membership"
        run_tests
    fi
else
    fail "Missing user or material supplier ID for membership"
    echo "Cannot continue without a membership"
    run_tests
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create developer certified weigher"
CERT_NUMBER="CW-$(date +%s)-${RANDOM}"

xbe_json do developer-certified-weighers create \
    --developer "$CREATED_DEVELOPER_ID" \
    --user "$CREATED_USER_ID" \
    --number "$CERT_NUMBER" \
    --is-active true

if [[ $status -eq 0 ]]; then
    CREATED_DEVELOPER_CERTIFIED_WEIGHER_ID=$(json_get ".id")
    if [[ -n "$CREATED_DEVELOPER_CERTIFIED_WEIGHER_ID" && "$CREATED_DEVELOPER_CERTIFIED_WEIGHER_ID" != "null" ]]; then
        register_cleanup "developer-certified-weighers" "$CREATED_DEVELOPER_CERTIFIED_WEIGHER_ID"
        pass
    else
        fail "Created developer certified weigher but no ID returned"
    fi
else
    fail "Failed to create developer certified weigher"
fi

if [[ -z "$CREATED_DEVELOPER_CERTIFIED_WEIGHER_ID" || "$CREATED_DEVELOPER_CERTIFIED_WEIGHER_ID" == "null" ]]; then
    echo "Cannot continue without a valid developer certified weigher ID"
    run_tests
fi

# ============================================================================
# LIST Tests
# ============================================================================

test_name "List developer certified weighers"
xbe_json view developer-certified-weighers list --limit 50
assert_json_is_array

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show developer certified weigher"
xbe_json view developer-certified-weighers show "$CREATED_DEVELOPER_CERTIFIED_WEIGHER_ID"
assert_success

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "Filter by developer"
xbe_json view developer-certified-weighers list --developer "$CREATED_DEVELOPER_ID" --limit 5
assert_success

test_name "Filter by user"
xbe_json view developer-certified-weighers list --user "$CREATED_USER_ID" --limit 5
assert_success

test_name "Filter by active status (true)"
xbe_json view developer-certified-weighers list --is-active true --limit 5
assert_success

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update developer certified weigher"
UPDATED_CERT_NUMBER="CW-UPDATED-$(date +%s)-${RANDOM}"

xbe_json do developer-certified-weighers update "$CREATED_DEVELOPER_CERTIFIED_WEIGHER_ID" \
    --number "$UPDATED_CERT_NUMBER" \
    --is-active false

assert_success

# ============================================================================
# LIST Tests - Filters (post-update)
# ============================================================================

test_name "Filter by active status (false)"
xbe_json view developer-certified-weighers list --is-active false --limit 5
assert_success

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete developer certified weigher"
xbe_run do developer-certified-weighers delete "$CREATED_DEVELOPER_CERTIFIED_WEIGHER_ID" --confirm
assert_success

run_tests
