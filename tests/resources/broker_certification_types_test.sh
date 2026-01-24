#!/bin/bash
#
# XBE CLI Integration Tests: Broker Certification Types
#
# Tests CRUD operations for the broker_certification_types resource.
#
# COVERAGE: All create/update attributes + list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_BROKER_ID=""
CREATED_CERTIFICATION_TYPE_ID=""
CREATED_BROKER_CERT_TYPE_ID=""
SECOND_BROKER_ID=""
SECOND_CERTIFICATION_TYPE_ID=""

describe "Resource: broker_certification_types"

# ============================================================================
# Prerequisites - Create broker and certification type
# ============================================================================

test_name "Create prerequisite broker for broker certification type tests"
BROKER_NAME=$(unique_name "BCTBroker")

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

test_name "Create prerequisite certification type for broker certification type tests"
CERT_NAME=$(unique_name "BCTCert")

xbe_json do certification-types create \
    --name "$CERT_NAME" \
    --can-apply-to "Trucker" \
    --broker "$CREATED_BROKER_ID"

if [[ $status -eq 0 ]]; then
    CREATED_CERTIFICATION_TYPE_ID=$(json_get ".id")
    if [[ -n "$CREATED_CERTIFICATION_TYPE_ID" && "$CREATED_CERTIFICATION_TYPE_ID" != "null" ]]; then
        register_cleanup "certification-types" "$CREATED_CERTIFICATION_TYPE_ID"
        pass
    else
        fail "Created certification type but no ID returned"
        echo "Cannot continue without a certification type"
        run_tests
    fi
else
    fail "Failed to create certification type"
    echo "Cannot continue without a certification type"
    run_tests
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create broker certification type with required fields"
xbe_json do broker-certification-types create \
    --broker "$CREATED_BROKER_ID" \
    --certification-type "$CREATED_CERTIFICATION_TYPE_ID"

if [[ $status -eq 0 ]]; then
    CREATED_BROKER_CERT_TYPE_ID=$(json_get ".id")
    if [[ -n "$CREATED_BROKER_CERT_TYPE_ID" && "$CREATED_BROKER_CERT_TYPE_ID" != "null" ]]; then
        register_cleanup "broker-certification-types" "$CREATED_BROKER_CERT_TYPE_ID"
        pass
    else
        fail "Created broker certification type but no ID returned"
    fi
else
    fail "Failed to create broker certification type"
fi

if [[ -z "$CREATED_BROKER_CERT_TYPE_ID" || "$CREATED_BROKER_CERT_TYPE_ID" == "null" ]]; then
    echo "Cannot continue without a valid broker certification type ID"
    run_tests
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show broker certification type"
xbe_json view broker-certification-types show "$CREATED_BROKER_CERT_TYPE_ID"
assert_success

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Create secondary broker for update"
SECOND_BROKER_NAME=$(unique_name "BCTBroker2")

xbe_json do brokers create --name "$SECOND_BROKER_NAME"

if [[ $status -eq 0 ]]; then
    SECOND_BROKER_ID=$(json_get ".id")
    if [[ -n "$SECOND_BROKER_ID" && "$SECOND_BROKER_ID" != "null" ]]; then
        register_cleanup "brokers" "$SECOND_BROKER_ID"
        pass
    else
        fail "Created second broker but no ID returned"
        echo "Cannot continue without a second broker"
        run_tests
    fi
else
    fail "Failed to create second broker"
    echo "Cannot continue without a second broker"
    run_tests
fi

test_name "Create certification type for second broker"
SECOND_CERT_NAME=$(unique_name "BCTCert2")

xbe_json do certification-types create \
    --name "$SECOND_CERT_NAME" \
    --can-apply-to "Trucker" \
    --broker "$SECOND_BROKER_ID"

if [[ $status -eq 0 ]]; then
    SECOND_CERTIFICATION_TYPE_ID=$(json_get ".id")
    if [[ -n "$SECOND_CERTIFICATION_TYPE_ID" && "$SECOND_CERTIFICATION_TYPE_ID" != "null" ]]; then
        register_cleanup "certification-types" "$SECOND_CERTIFICATION_TYPE_ID"
        pass
    else
        fail "Created second certification type but no ID returned"
        echo "Cannot continue without a second certification type"
        run_tests
    fi
else
    fail "Failed to create second certification type"
    echo "Cannot continue without a second certification type"
    run_tests
fi

test_name "Update broker certification type broker and certification type"
xbe_json do broker-certification-types update "$CREATED_BROKER_CERT_TYPE_ID" \
    --broker "$SECOND_BROKER_ID" \
    --certification-type "$SECOND_CERTIFICATION_TYPE_ID"
assert_success

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List broker certification types"
xbe_json view broker-certification-types list --limit 5
assert_success

test_name "List broker certification types returns array"
xbe_json view broker-certification-types list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list broker certification types"
fi

# ============================================================================
# LIST Tests - All Filters
# ============================================================================

test_name "List broker certification types with --created-at-min filter"
xbe_json view broker-certification-types list --created-at-min "2020-01-01T00:00:00Z" --limit 5
assert_success

test_name "List broker certification types with --created-at-max filter"
xbe_json view broker-certification-types list --created-at-max "2030-01-01T00:00:00Z" --limit 5
assert_success

test_name "List broker certification types with --is-created-at filter"
xbe_json view broker-certification-types list --is-created-at true --limit 5
assert_success

test_name "List broker certification types with --updated-at-min filter"
xbe_json view broker-certification-types list --updated-at-min "2020-01-01T00:00:00Z" --limit 5
assert_success

test_name "List broker certification types with --updated-at-max filter"
xbe_json view broker-certification-types list --updated-at-max "2030-01-01T00:00:00Z" --limit 5
assert_success

test_name "List broker certification types with --is-updated-at filter"
xbe_json view broker-certification-types list --is-updated-at true --limit 5
assert_success

# ============================================================================
# LIST Tests - Pagination
# ============================================================================

test_name "List broker certification types with --limit"
xbe_json view broker-certification-types list --limit 3
assert_success

test_name "List broker certification types with --offset"
xbe_json view broker-certification-types list --limit 3 --offset 3
assert_success

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete broker certification type requires --confirm flag"
xbe_run do broker-certification-types delete "$CREATED_BROKER_CERT_TYPE_ID"
assert_failure

test_name "Delete broker certification type with --confirm"
xbe_run do broker-certification-types delete "$CREATED_BROKER_CERT_TYPE_ID" --confirm
assert_success

run_tests
