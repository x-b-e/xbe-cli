#!/bin/bash
#
# XBE CLI Integration Tests: Material Supplier Memberships
#
# Tests CRUD operations for the material-supplier-memberships resource.
# Material supplier memberships link users to material suppliers.
#
# NOTE: This test requires creating prerequisite resources: broker, material supplier, user
#
# COVERAGE: All create/update attributes + list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_MEMBERSHIP_ID=""
CREATED_BROKER_ID=""
CREATED_MATERIAL_SUPPLIER_ID=""
CREATED_PROJECT_OFFICE_ID=""
CREATED_USER_ID=""

describe "Resource: material-supplier-memberships"

# ==========================================================================
# Prerequisites - Create broker, material supplier, project office, and user
# ==========================================================================

test_name "Create prerequisite broker for material supplier membership tests"
BROKER_NAME=$(unique_name "MSMBroker")

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

test_name "Create prerequisite material supplier"
SUPPLIER_NAME=$(unique_name "MaterialSupplier")

xbe_json do material-suppliers create --name "$SUPPLIER_NAME" --broker "$CREATED_BROKER_ID"

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
    fail "Failed to create material supplier"
    echo "Cannot continue without a material supplier"
    run_tests
fi

test_name "Create prerequisite project office"
PROJECT_OFFICE_NAME=$(unique_name "MSMOffice")

xbe_json do project-offices create --name "$PROJECT_OFFICE_NAME" --broker "$CREATED_BROKER_ID"

if [[ $status -eq 0 ]]; then
    CREATED_PROJECT_OFFICE_ID=$(json_get ".id")
    if [[ -n "$CREATED_PROJECT_OFFICE_ID" && "$CREATED_PROJECT_OFFICE_ID" != "null" ]]; then
        register_cleanup "project-offices" "$CREATED_PROJECT_OFFICE_ID"
        pass
    else
        fail "Created project office but no ID returned"
    fi
else
    xbe_json view project-offices list --broker "$CREATED_BROKER_ID" --limit 1
    if [[ $status -eq 0 ]]; then
        CREATED_PROJECT_OFFICE_ID=$(json_get ".[0].id")
        if [[ -n "$CREATED_PROJECT_OFFICE_ID" && "$CREATED_PROJECT_OFFICE_ID" != "null" ]]; then
            echo "    Using existing project office: $CREATED_PROJECT_OFFICE_ID"
            pass
        else
            skip "Failed to create or locate project office"
        fi
    else
        skip "Failed to create or locate project office"
    fi
fi

if [[ "$CREATED_PROJECT_OFFICE_ID" == "null" ]]; then
    CREATED_PROJECT_OFFICE_ID=""
fi

test_name "Create prerequisite user"
TEST_EMAIL=$(unique_email)

xbe_json do users create \
    --name "Material Supplier Member" \
    --email "$TEST_EMAIL"

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

# ==========================================================================
# CREATE Tests
# ==========================================================================

test_name "Create material supplier membership with required fields"

xbe_json do material-supplier-memberships create \
    --user "$CREATED_USER_ID" \
    --material-supplier "$CREATED_MATERIAL_SUPPLIER_ID"

if [[ $status -eq 0 ]]; then
    CREATED_MEMBERSHIP_ID=$(json_get ".id")
    if [[ -n "$CREATED_MEMBERSHIP_ID" && "$CREATED_MEMBERSHIP_ID" != "null" ]]; then
        register_cleanup "material-supplier-memberships" "$CREATED_MEMBERSHIP_ID"
        pass
    else
        fail "Created membership but no ID returned"
    fi
else
    fail "Failed to create material supplier membership"
fi

# Only continue if we successfully created a membership
if [[ -z "$CREATED_MEMBERSHIP_ID" || "$CREATED_MEMBERSHIP_ID" == "null" ]]; then
    echo "Cannot continue without a valid membership ID"
    run_tests
fi

test_name "Create material supplier membership with role and admin"
TEST_EMAIL2=$(unique_email)
xbe_json do users create --name "MSM Manager User" --email "$TEST_EMAIL2"
if [[ $status -eq 0 ]]; then
    MANAGER_USER_ID=$(json_get ".id")
    xbe_json do material-supplier-memberships create \
        --user "$MANAGER_USER_ID" \
        --material-supplier "$CREATED_MATERIAL_SUPPLIER_ID" \
        --kind "manager" \
        --title "Operations Manager" \
        --is-admin "true"
    if [[ $status -eq 0 ]]; then
        id=$(json_get ".id")
        register_cleanup "material-supplier-memberships" "$id"
        pass
    else
        fail "Failed to create material supplier membership with role and admin"
    fi
else
    skip "Could not create user for manager membership test"
fi

# ==========================================================================
# SHOW Tests
# ==========================================================================

test_name "Show material supplier membership"
xbe_json view material-supplier-memberships show "$CREATED_MEMBERSHIP_ID"
assert_success

# ==========================================================================
# UPDATE Tests
# ==========================================================================

test_name "Update material supplier membership core fields"
xbe_json do material-supplier-memberships update "$CREATED_MEMBERSHIP_ID" \
    --kind "manager" \
    --is-admin "true" \
    --title "Plant Manager" \
    --color-hex "#FF5500" \
    --external-employee-id "MSM-EXT-123" \
    --explicit-sort-order "12"
assert_success

test_name "Update material supplier membership timing"
xbe_json do material-supplier-memberships update "$CREATED_MEMBERSHIP_ID" \
    --start-at "2025-01-01T08:00:00Z" \
    --end-at "2025-02-01T17:00:00Z" \
    --drives-shift-type "day"
assert_success

if [[ -n "$CREATED_PROJECT_OFFICE_ID" ]]; then
    test_name "Update material supplier membership project office"
    xbe_json do material-supplier-memberships update "$CREATED_MEMBERSHIP_ID" \
        --project-office "$CREATED_PROJECT_OFFICE_ID"
    assert_success
else
    test_name "Update material supplier membership project office"
    skip "Project office not available"
fi

test_name "Update material supplier membership permissions and notifications"
xbe_json do material-supplier-memberships update "$CREATED_MEMBERSHIP_ID" \
    --can-see-rates-as-manager "false" \
    --can-validate-profit-improvements "false" \
    --is-geofence-violation-team-member "true" \
    --is-unapproved-time-card-subscriber "true" \
    --is-default-job-production-plan-subscriber "false" \
    --enable-recap-notifications "true" \
    --enable-inventory-capacity-notifications "true"
assert_success

# ==========================================================================
# LIST Tests - Basic
# ==========================================================================

test_name "List material supplier memberships"
xbe_json view material-supplier-memberships list --limit 5
assert_success

test_name "List material supplier memberships returns array"
xbe_json view material-supplier-memberships list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list material supplier memberships"
fi

# ==========================================================================
# LIST Tests - Filters
# ==========================================================================

test_name "List material supplier memberships with --material-supplier filter"
xbe_json view material-supplier-memberships list --material-supplier "$CREATED_MATERIAL_SUPPLIER_ID" --limit 10
assert_success

test_name "List material supplier memberships with --broker filter"
xbe_json view material-supplier-memberships list --broker "$CREATED_BROKER_ID" --limit 10
assert_success

test_name "List material supplier memberships with --user filter"
xbe_json view material-supplier-memberships list --user "$CREATED_USER_ID" --limit 10
assert_success

if [[ -n "$CREATED_PROJECT_OFFICE_ID" ]]; then
    test_name "List material supplier memberships with --project-office filter"
    xbe_json view material-supplier-memberships list --project-office "$CREATED_PROJECT_OFFICE_ID" --limit 10
    assert_success
else
    test_name "List material supplier memberships with --project-office filter"
    skip "Project office not available"
fi

test_name "List material supplier memberships with --kind filter"
xbe_json view material-supplier-memberships list --kind "manager" --limit 10
assert_success

test_name "List material supplier memberships with --q filter"
xbe_json view material-supplier-memberships list --q "Material Supplier" --limit 10
assert_success

test_name "List material supplier memberships with --drives-shift-type filter"
xbe_json view material-supplier-memberships list --drives-shift-type "day" --limit 10
assert_success

test_name "List material supplier memberships with --external-employee-id filter"
xbe_json view material-supplier-memberships list --external-employee-id "MSM-EXT-123" --limit 10
assert_success

test_name "List material supplier memberships with --is-rate-editor filter"
xbe_json view material-supplier-memberships list --is-rate-editor "false" --limit 10
assert_success

test_name "List material supplier memberships with --is-time-card-auditor filter"
xbe_json view material-supplier-memberships list --is-time-card-auditor "false" --limit 10
assert_success

test_name "List material supplier memberships with --is-equipment-rental-team-member filter"
xbe_json view material-supplier-memberships list --is-equipment-rental-team-member "false" --limit 10
assert_success

test_name "List material supplier memberships with --is-geofence-violation-team-member filter"
xbe_json view material-supplier-memberships list --is-geofence-violation-team-member "true" --limit 10
assert_success

test_name "List material supplier memberships with --is-unapproved-time-card-subscriber filter"
xbe_json view material-supplier-memberships list --is-unapproved-time-card-subscriber "true" --limit 10
assert_success

test_name "List material supplier memberships with --is-default-job-production-plan-subscriber filter"
xbe_json view material-supplier-memberships list --is-default-job-production-plan-subscriber "false" --limit 10
assert_success

# ==========================================================================
# LIST Tests - Pagination
# ==========================================================================

test_name "List material supplier memberships with --limit"
xbe_json view material-supplier-memberships list --limit 3
assert_success

test_name "List material supplier memberships with --offset"
xbe_json view material-supplier-memberships list --limit 3 --offset 3
assert_success

# ==========================================================================
# DELETE Tests
# ==========================================================================

test_name "Delete material supplier membership requires --confirm flag"
xbe_run do material-supplier-memberships delete "$CREATED_MEMBERSHIP_ID"
assert_failure

test_name "Delete material supplier membership with --confirm"
TEST_DEL_EMAIL=$(unique_email)
xbe_json do users create --name "MSM Delete User" --email "$TEST_DEL_EMAIL"
if [[ $status -eq 0 ]]; then
    DEL_USER_ID=$(json_get ".id")
    xbe_json do material-supplier-memberships create \
        --user "$DEL_USER_ID" \
        --material-supplier "$CREATED_MATERIAL_SUPPLIER_ID"
    if [[ $status -eq 0 ]]; then
        DEL_ID=$(json_get ".id")
        xbe_run do material-supplier-memberships delete "$DEL_ID" --confirm
        assert_success
    else
        skip "Could not create membership for deletion test"
    fi
else
    skip "Could not create user for deletion test"
fi

# ==========================================================================
# Error Cases
# ==========================================================================

test_name "Create material supplier membership without user fails"
xbe_json do material-supplier-memberships create --material-supplier "$CREATED_MATERIAL_SUPPLIER_ID"
assert_failure

test_name "Create material supplier membership without material supplier fails"
xbe_json do material-supplier-memberships create --user "$CREATED_USER_ID"
assert_failure

test_name "Update without any fields fails"
xbe_json do material-supplier-memberships update "$CREATED_MEMBERSHIP_ID"
assert_failure

# ==========================================================================
# Summary
# ==========================================================================

run_tests
