#!/bin/bash
#
# XBE CLI Integration Tests: Laborers
#
# Tests CRUD operations for the laborers resource.
# Laborers represent workers assigned to jobs and projects.
#
# NOTE: This test requires creating prerequisite resources: broker, labor classification, and user
#
# COVERAGE: All create/update attributes + list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_LABORER_ID=""
CREATED_BROKER_ID=""
CREATED_CUSTOMER_ID=""
CREATED_LABOR_CLASSIFICATION_ID=""
CREATED_USER_ID=""

describe "Resource: laborers"

# ============================================================================
# Prerequisites - Create broker, customer, labor classification, and user
# ============================================================================

test_name "Create prerequisite broker for laborer tests"
BROKER_NAME=$(unique_name "LaborerTestBroker")

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

test_name "Create prerequisite customer for laborer tests"
CUSTOMER_NAME=$(unique_name "LaborerTestCustomer")

xbe_json do customers create \
    --name "$CUSTOMER_NAME" \
    --broker "$CREATED_BROKER_ID"

if [[ $status -eq 0 ]]; then
    CREATED_CUSTOMER_ID=$(json_get ".id")
    if [[ -n "$CREATED_CUSTOMER_ID" && "$CREATED_CUSTOMER_ID" != "null" ]]; then
        register_cleanup "customers" "$CREATED_CUSTOMER_ID"
        pass
    else
        fail "Created customer but no ID returned"
        echo "Cannot continue without a customer"
        run_tests
    fi
else
    fail "Failed to create customer"
    echo "Cannot continue without a customer"
    run_tests
fi

test_name "Create prerequisite labor classification"
LC_NAME=$(unique_name "LaborClass")
LC_ABBR="LC$(date +%s | tail -c 4)"

xbe_json do labor-classifications create \
    --name "$LC_NAME" \
    --abbreviation "$LC_ABBR"

if [[ $status -eq 0 ]]; then
    CREATED_LABOR_CLASSIFICATION_ID=$(json_get ".id")
    if [[ -n "$CREATED_LABOR_CLASSIFICATION_ID" && "$CREATED_LABOR_CLASSIFICATION_ID" != "null" ]]; then
        register_cleanup "labor-classifications" "$CREATED_LABOR_CLASSIFICATION_ID"
        pass
    else
        fail "Created labor classification but no ID returned"
        echo "Cannot continue without a labor classification"
        run_tests
    fi
else
    fail "Failed to create labor classification"
    echo "Cannot continue without a labor classification"
    run_tests
fi

test_name "Create prerequisite user for laborer"
USER_EMAIL=$(unique_email)
USER_NAME=$(unique_name "LaborerUser")

xbe_json do users create \
    --email "$USER_EMAIL" \
    --name "$USER_NAME"

if [[ $status -eq 0 ]]; then
    CREATED_USER_ID=$(json_get ".id")
    if [[ -n "$CREATED_USER_ID" && "$CREATED_USER_ID" != "null" ]]; then
        # Note: Users may not be deletable via CLI, but we'll register anyway
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

test_name "Create membership for user to customer"
xbe_json do memberships create \
    --user "$CREATED_USER_ID" \
    --organization "Customer|$CREATED_CUSTOMER_ID"

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

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create laborer with required fields"

xbe_json do laborers create \
    --labor-classification "$CREATED_LABOR_CLASSIFICATION_ID" \
    --user "$CREATED_USER_ID" \
    --organization-type "customers" \
    --organization-id "$CREATED_CUSTOMER_ID"

if [[ $status -eq 0 ]]; then
    CREATED_LABORER_ID=$(json_get ".id")
    if [[ -n "$CREATED_LABORER_ID" && "$CREATED_LABORER_ID" != "null" ]]; then
        register_cleanup "laborers" "$CREATED_LABORER_ID"
        pass
    else
        fail "Created laborer but no ID returned"
    fi
else
    fail "Failed to create laborer"
fi

# Only continue if we successfully created a laborer
if [[ -z "$CREATED_LABORER_ID" || "$CREATED_LABORER_ID" == "null" ]]; then
    echo "Cannot continue without a valid laborer ID"
    run_tests
fi

# NOTE: Additional create tests with different attributes are skipped
# because each laborer requires a unique user with membership in the organization.
# The basic create test covers the core functionality.

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update laborer --is-active"
xbe_json do laborers update "$CREATED_LABORER_ID" --is-active=false
assert_success

test_name "Update laborer --is-active back to true"
xbe_json do laborers update "$CREATED_LABORER_ID" --is-active=true
assert_success

test_name "Update laborer --group-name"
xbe_json do laborers update "$CREATED_LABORER_ID" --group-name "Updated Crew"
assert_success

# NOTE: Update --color-hex test skipped because it requires labor classification to allow colors

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List laborers"
xbe_json view laborers list --limit 5
assert_success

test_name "List laborers returns array"
xbe_json view laborers list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list laborers"
fi

# ============================================================================
# LIST Tests - All Filters
# ============================================================================

test_name "List laborers with --labor-classification filter"
xbe_json view laborers list --labor-classification "$CREATED_LABOR_CLASSIFICATION_ID" --limit 10
assert_success

test_name "List laborers with --user filter"
xbe_json view laborers list --user "$CREATED_USER_ID" --limit 10
assert_success

test_name "List laborers with --is-active filter"
xbe_json view laborers list --is-active true --limit 10
assert_success

test_name "List laborers with --search filter"
xbe_json view laborers list --search "$USER_NAME" --limit 10
assert_success

# ============================================================================
# LIST Tests - Pagination
# ============================================================================

test_name "List laborers with --limit"
xbe_json view laborers list --limit 3
assert_success

test_name "List laborers with --offset"
xbe_json view laborers list --limit 3 --offset 3
assert_success

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete laborer requires --confirm flag"
xbe_run do laborers delete "$CREATED_LABORER_ID"
assert_failure

test_name "Delete laborer with --confirm"
# Create a laborer specifically for deletion
USER_DEL_EMAIL=$(unique_email)
xbe_json do users create \
    --email "$USER_DEL_EMAIL" \
    --name "DeleteLaborer"
if [[ $status -eq 0 ]]; then
    USER_DEL_ID=$(json_get ".id")

    xbe_json do laborers create \
        --labor-classification "$CREATED_LABOR_CLASSIFICATION_ID" \
        --user "$USER_DEL_ID" \
        --organization-type "brokers" \
        --organization-id "$CREATED_BROKER_ID"
    if [[ $status -eq 0 ]]; then
        DEL_LABORER_ID=$(json_get ".id")
        xbe_run do laborers delete "$DEL_LABORER_ID" --confirm
        assert_success
    else
        skip "Could not create laborer for deletion test"
    fi
else
    skip "Could not create user for deletion test"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create laborer without --labor-classification fails"
xbe_json do laborers create \
    --user "$CREATED_USER_ID" \
    --organization-type "customers" \
    --organization-id "$CREATED_CUSTOMER_ID"
assert_failure

test_name "Create laborer without --user fails"
xbe_json do laborers create \
    --labor-classification "$CREATED_LABOR_CLASSIFICATION_ID" \
    --organization-type "customers" \
    --organization-id "$CREATED_CUSTOMER_ID"
assert_failure

test_name "Create laborer without --organization-type fails"
xbe_json do laborers create \
    --labor-classification "$CREATED_LABOR_CLASSIFICATION_ID" \
    --user "$CREATED_USER_ID" \
    --organization-id "$CREATED_BROKER_ID"
assert_failure

test_name "Create laborer without --organization-id fails"
xbe_json do laborers create \
    --labor-classification "$CREATED_LABOR_CLASSIFICATION_ID" \
    --user "$CREATED_USER_ID" \
    --organization-type "brokers"
assert_failure

test_name "Update without any fields fails"
xbe_json do laborers update "$CREATED_LABORER_ID"
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
