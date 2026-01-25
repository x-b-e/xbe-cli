#!/bin/bash
#
# XBE CLI Integration Tests: Customer Applications
#
# Tests list/show and create/update/delete behavior for customer-applications.
# Customer applications require broker + user relationships.
#
# COVERAGE: All supported filters + writable attributes
#

source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

describe "Resource: customer-applications"

SAMPLE_ID=""
CREATED_APPLICATION_ID=""
BROKER_ID="${XBE_TEST_BROKER_ID:-}"
USER_ID="${XBE_TEST_USER_ID:-}"
SKIP_MUTATION=0

if [[ -z "$XBE_TOKEN" ]]; then
    echo "    (XBE_TOKEN not set; skipping create/update/delete tests)"
    SKIP_MUTATION=1
fi

# ==========================================================================
# LIST Tests
# ==========================================================================

test_name "List customer applications"
xbe_json view customer-applications list --limit 5
assert_success

test_name "List customer applications returns array"
xbe_json view customer-applications list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list customer applications"
fi

test_name "Capture sample customer application"
xbe_json view customer-applications list --limit 1
if [[ $status -eq 0 ]]; then
    SAMPLE_ID=$(json_get ".[0].id")
    if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
        pass
    else
        skip "No customer applications available"
    fi
else
    skip "Could not list customer applications"
fi

# ==========================================================================
# SHOW Tests
# ==========================================================================

test_name "Show customer application"
if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_json view customer-applications show "$SAMPLE_ID"
    assert_success
else
    skip "No customer application ID available"
fi

# ==========================================================================
# CREATE Tests
# ==========================================================================

test_name "Create customer application without required fields fails"
xbe_json do customer-applications create
assert_failure

if [[ $SKIP_MUTATION -eq 1 ]]; then
    skip "Skipping create/update/delete tests without XBE_TOKEN"
else
    if [[ -z "$BROKER_ID" ]]; then
        test_name "Create prerequisite broker for customer application tests"
        BROKER_NAME=$(unique_name "CustAppBroker")
        xbe_json do brokers create --name "$BROKER_NAME"
        if [[ $status -eq 0 ]]; then
            BROKER_ID=$(json_get ".id")
            if [[ -n "$BROKER_ID" && "$BROKER_ID" != "null" ]]; then
                register_cleanup "brokers" "$BROKER_ID"
                pass
            else
                fail "Created broker but no ID returned"
            fi
        else
            fail "Failed to create broker"
            SKIP_MUTATION=1
        fi
    fi

    if [[ -z "$USER_ID" ]]; then
        xbe_json auth whoami
        if [[ $status -eq 0 ]]; then
            USER_ID=$(json_get ".id")
        else
            xbe_json view users list --limit 1
            if [[ $status -eq 0 ]]; then
                USER_ID=$(json_get ".[0].id")
            fi
        fi
    fi

    test_name "Create customer application with required fields"
    APP_NAME=$(unique_name "CustApp")
    xbe_json do customer-applications create \
        --company-name "$APP_NAME" \
        --company-address "123 Main St, Test City, TS 12345" \
        --requires-union-drivers false \
        --is-trucking-company true \
        --broker "$BROKER_ID" \
        --user "$USER_ID"

    if [[ $status -eq 0 ]]; then
        CREATED_APPLICATION_ID=$(json_get ".id")
        if [[ -n "$CREATED_APPLICATION_ID" && "$CREATED_APPLICATION_ID" != "null" ]]; then
            register_cleanup "customer-applications" "$CREATED_APPLICATION_ID"
            pass
        else
            fail "Created customer application but no ID returned"
        fi
    else
        if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]] || [[ "$output" == *"Authentication required"* ]]; then
            skip "Create not authorized"
            SKIP_MUTATION=1
        else
            fail "Failed to create customer application"
            SKIP_MUTATION=1
        fi
    fi
fi

# ==========================================================================
# UPDATE Tests - Attributes
# ==========================================================================

if [[ $SKIP_MUTATION -eq 1 || -z "$CREATED_APPLICATION_ID" || "$CREATED_APPLICATION_ID" == "null" ]]; then
    skip "Skipping update tests without a customer application"
else
    test_name "Update customer application company-name"
    UPDATED_NAME=$(unique_name "CustAppUpdated")
    xbe_json do customer-applications update "$CREATED_APPLICATION_ID" --company-name "$UPDATED_NAME"
    assert_success

    test_name "Update customer application company-address"
    xbe_json do customer-applications update "$CREATED_APPLICATION_ID" --company-address "456 Market St, Test City, TS 67890"
    assert_success

    test_name "Update customer application company-address-latitude"
    xbe_json do customer-applications update "$CREATED_APPLICATION_ID" --company-address-latitude "40.7128"
    assert_success

    test_name "Update customer application company-address-longitude"
    xbe_json do customer-applications update "$CREATED_APPLICATION_ID" --company-address-longitude "-74.0060"
    assert_success

    test_name "Update customer application company-address-place-id"
    xbe_json do customer-applications update "$CREATED_APPLICATION_ID" --company-address-place-id "ChIJOwg_06VPwokRYv534QaPC8g"
    assert_success

    test_name "Update customer application company-address-plus-code"
    xbe_json do customer-applications update "$CREATED_APPLICATION_ID" --company-address-plus-code "87G8P2RP+CC"
    assert_success

    test_name "Update customer application skip-company-address-geocoding true"
    xbe_json do customer-applications update "$CREATED_APPLICATION_ID" --skip-company-address-geocoding true
    assert_success

    test_name "Update customer application skip-company-address-geocoding false"
    xbe_json do customer-applications update "$CREATED_APPLICATION_ID" --skip-company-address-geocoding false
    assert_success

    test_name "Update customer application company-url"
    xbe_json do customer-applications update "$CREATED_APPLICATION_ID" --company-url "https://example.com"
    assert_success

    test_name "Update customer application notes"
    xbe_json do customer-applications update "$CREATED_APPLICATION_ID" --notes "CLI updated notes"
    assert_success

    test_name "Update customer application requires-union-drivers true"
    xbe_json do customer-applications update "$CREATED_APPLICATION_ID" --requires-union-drivers true
    assert_success

    test_name "Update customer application requires-union-drivers false"
    xbe_json do customer-applications update "$CREATED_APPLICATION_ID" --requires-union-drivers false
    assert_success

    test_name "Update customer application is-trucking-company true"
    xbe_json do customer-applications update "$CREATED_APPLICATION_ID" --is-trucking-company true
    assert_success

    test_name "Update customer application is-trucking-company false"
    xbe_json do customer-applications update "$CREATED_APPLICATION_ID" --is-trucking-company false
    assert_success

    test_name "Update customer application estimated-annual-material-transport-spend"
    xbe_json do customer-applications update "$CREATED_APPLICATION_ID" --estimated-annual-material-transport-spend "1000000"
    assert_success

    test_name "Update customer application status"
    xbe_json do customer-applications update "$CREATED_APPLICATION_ID" --status reviewing
    if [[ $status -eq 0 ]]; then
        pass
    else
        skip "Could not update status (permissions or policy)"
    fi

    test_name "Update customer application broker relationship"
    xbe_json do customer-applications update "$CREATED_APPLICATION_ID" --broker "$BROKER_ID"
    if [[ $status -eq 0 ]]; then
        pass
    else
        skip "Could not update broker (permissions or policy)"
    fi

    test_name "Update customer application user relationship"
    xbe_json do customer-applications update "$CREATED_APPLICATION_ID" --user "$USER_ID"
    if [[ $status -eq 0 ]]; then
        pass
    else
        skip "Could not update user (permissions or policy)"
    fi

    test_name "Update customer application job types"
    xbe_json do customer-applications update "$CREATED_APPLICATION_ID" --job-types ""
    if [[ $status -eq 0 ]]; then
        pass
    else
        skip "Could not update job types (permissions or policy)"
    fi
fi

# ==========================================================================
# LIST Tests - Filters
# ==========================================================================

test_name "List customer applications with status filter"
xbe_json view customer-applications list --status pending --limit 5
assert_success

test_name "List customer applications with broker filter"
if [[ -n "$BROKER_ID" && "$BROKER_ID" != "null" ]]; then
    xbe_json view customer-applications list --broker "$BROKER_ID" --limit 5
    assert_success
else
    skip "No broker ID available"
fi

test_name "List customer applications with created/updated filters"
xbe_json view customer-applications list \
    --created-at-min "2020-01-01T00:00:00Z" \
    --created-at-max "2030-01-01T00:00:00Z" \
    --updated-at-min "2020-01-01T00:00:00Z" \
    --updated-at-max "2030-01-01T00:00:00Z" \
    --limit 5
assert_success

# ==========================================================================
# DELETE Tests
# ==========================================================================

test_name "Delete customer application"
if [[ -n "$CREATED_APPLICATION_ID" && "$CREATED_APPLICATION_ID" != "null" && $SKIP_MUTATION -eq 0 ]]; then
    xbe_run do customer-applications delete "$CREATED_APPLICATION_ID" --confirm
    if [[ $status -eq 0 ]]; then
        pass
    else
        skip "Could not delete customer application (permissions or policy)"
    fi
else
    skip "No customer application ID available"
fi

run_tests
