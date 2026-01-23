#!/bin/bash
#
# XBE CLI Integration Tests: Invoice Generations
#
# Tests list/show/create operations for invoice-generations.
#
# COVERAGE: List filters + writable attributes
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

SAMPLE_ID=""
SAMPLE_ORG_TYPE=""
SAMPLE_ORG_ID=""
SAMPLE_TIME_CARD_ID=""
SAMPLE_COMPLETED_AT=""
SAMPLE_CREATED_AT=""
SAMPLE_UPDATED_AT=""
SAMPLE_IS_COMPLETED=""

CREATED_ID=""

TIME_CARD_ID="${XBE_TEST_TIME_CARD_ID:-}"
ORG_TYPE=""
ORG_ID=""

normalize_org_type() {
    case "${1:-}" in
        brokers|broker|Broker) echo "Broker" ;;
        customers|customer|Customer) echo "Customer" ;;
        truckers|trucker|Trucker) echo "Trucker" ;;
        *) echo "$1" ;;
    esac
}

pick_org_type_id_from_time_card() {
    local tc_id="$1"
    if [[ -z "$tc_id" || "$tc_id" == "null" ]]; then
        return
    fi
    xbe_json view time-cards show "$tc_id"
    if [[ $status -ne 0 ]]; then
        return
    fi
    local broker_id
    local customer_id
    local trucker_id
    broker_id=$(json_get ".broker_id")
    customer_id=$(json_get ".customer_id")
    trucker_id=$(json_get ".trucker_id")
    if [[ -n "$broker_id" && "$broker_id" != "null" ]]; then
        ORG_TYPE="brokers"
        ORG_ID="$broker_id"
        return
    fi
    if [[ -n "$customer_id" && "$customer_id" != "null" ]]; then
        ORG_TYPE="customers"
        ORG_ID="$customer_id"
        return
    fi
    if [[ -n "$trucker_id" && "$trucker_id" != "null" ]]; then
        ORG_TYPE="truckers"
        ORG_ID="$trucker_id"
        return
    fi
}

describe "Resource: invoice-generations"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List invoice generations"
xbe_json view invoice-generations list --limit 5
assert_success

test_name "List invoice generations returns array"
xbe_json view invoice-generations list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list invoice generations"
fi

# ============================================================================
# Sample Record (used for filters/show)
# ============================================================================

test_name "Capture sample invoice generation"
xbe_json view invoice-generations list --limit 1
if [[ $status -eq 0 ]]; then
    SAMPLE_ID=$(json_get ".[0].id")
    SAMPLE_ORG_TYPE=$(json_get ".[0].organization_type")
    SAMPLE_ORG_ID=$(json_get ".[0].organization_id")
    SAMPLE_TIME_CARD_ID=$(json_get ".[0].time_card_ids[0]")
    SAMPLE_COMPLETED_AT=$(json_get ".[0].completed_at")
    SAMPLE_CREATED_AT=$(json_get ".[0].created_at")
    SAMPLE_UPDATED_AT=$(json_get ".[0].updated_at")
    SAMPLE_IS_COMPLETED=$(json_get ".[0].is_completed")
    if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
        pass
    else
        skip "No invoice generations available for follow-on tests"
    fi
else
    skip "Could not list invoice generations to capture sample"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List invoice generations with --organization filter"
if [[ -n "$SAMPLE_ORG_ID" && "$SAMPLE_ORG_ID" != "null" ]]; then
    FILTER_ORG_TYPE=$(normalize_org_type "$SAMPLE_ORG_TYPE")
    if [[ -n "$FILTER_ORG_TYPE" && "$FILTER_ORG_TYPE" != "null" ]]; then
        xbe_json view invoice-generations list --organization "${FILTER_ORG_TYPE}|${SAMPLE_ORG_ID}" --limit 5
        assert_success
    else
        skip "No organization type available"
    fi
else
    skip "No organization ID available"
fi

test_name "List invoice generations with --organization-id filter"
if [[ -n "$SAMPLE_ORG_ID" && "$SAMPLE_ORG_ID" != "null" ]]; then
    FILTER_ORG_TYPE=$(normalize_org_type "$SAMPLE_ORG_TYPE")
    if [[ -n "$FILTER_ORG_TYPE" && "$FILTER_ORG_TYPE" != "null" ]]; then
        xbe_json view invoice-generations list --organization-type "$FILTER_ORG_TYPE" --organization-id "$SAMPLE_ORG_ID" --limit 5
        assert_success
    else
        skip "No organization type available"
    fi
else
    skip "No organization ID available"
fi

test_name "List invoice generations with --organization-type filter"
FILTER_ORG_TYPE=$(normalize_org_type "$SAMPLE_ORG_TYPE")
if [[ -n "$FILTER_ORG_TYPE" && "$FILTER_ORG_TYPE" != "null" ]]; then
    xbe_json view invoice-generations list --organization-type "$FILTER_ORG_TYPE" --limit 5
    assert_success
else
    skip "No organization type available"
fi

test_name "List invoice generations with --not-organization-type filter"
xbe_json view invoice-generations list --not-organization-type "Customer" --limit 5
assert_success

test_name "List invoice generations with --time-cards filter"
FILTER_TIME_CARD_ID="$TIME_CARD_ID"
if [[ -z "$FILTER_TIME_CARD_ID" || "$FILTER_TIME_CARD_ID" == "null" ]]; then
    FILTER_TIME_CARD_ID="$SAMPLE_TIME_CARD_ID"
fi
if [[ -n "$FILTER_TIME_CARD_ID" && "$FILTER_TIME_CARD_ID" != "null" ]]; then
    xbe_json view invoice-generations list --time-cards "$FILTER_TIME_CARD_ID" --limit 5
    assert_success
else
    skip "No time card ID available"
fi

COMPLETED_AT_FILTER="$SAMPLE_COMPLETED_AT"
if [[ -z "$COMPLETED_AT_FILTER" || "$COMPLETED_AT_FILTER" == "null" ]]; then
    COMPLETED_AT_FILTER=$(date -u +%Y-%m-%dT%H:%M:%SZ)
fi

test_name "List invoice generations with --completed-at-min filter"
xbe_json view invoice-generations list --completed-at-min "$COMPLETED_AT_FILTER" --limit 5
assert_success

test_name "List invoice generations with --completed-at-max filter"
xbe_json view invoice-generations list --completed-at-max "$COMPLETED_AT_FILTER" --limit 5
assert_success

test_name "List invoice generations with --is-completed-at=true filter"
xbe_json view invoice-generations list --is-completed-at true --limit 5
assert_success

test_name "List invoice generations with --is-completed-at=false filter"
xbe_json view invoice-generations list --is-completed-at false --limit 5
assert_success

test_name "List invoice generations with --is-completed=true filter"
xbe_json view invoice-generations list --is-completed true --limit 5
assert_success

test_name "List invoice generations with --is-completed=false filter"
xbe_json view invoice-generations list --is-completed false --limit 5
assert_success

CREATED_AT_FILTER="$SAMPLE_CREATED_AT"
if [[ -z "$CREATED_AT_FILTER" || "$CREATED_AT_FILTER" == "null" ]]; then
    CREATED_AT_FILTER=$(date -u +%Y-%m-%dT%H:%M:%SZ)
fi

test_name "List invoice generations with --created-at-min filter"
xbe_json view invoice-generations list --created-at-min "$CREATED_AT_FILTER" --limit 5
assert_success

test_name "List invoice generations with --created-at-max filter"
xbe_json view invoice-generations list --created-at-max "$CREATED_AT_FILTER" --limit 5
assert_success

test_name "List invoice generations with --is-created-at=true filter"
xbe_json view invoice-generations list --is-created-at true --limit 5
assert_success

test_name "List invoice generations with --is-created-at=false filter"
xbe_json view invoice-generations list --is-created-at false --limit 5
assert_success

UPDATED_AT_FILTER="$SAMPLE_UPDATED_AT"
if [[ -z "$UPDATED_AT_FILTER" || "$UPDATED_AT_FILTER" == "null" ]]; then
    UPDATED_AT_FILTER=$(date -u +%Y-%m-%dT%H:%M:%SZ)
fi

test_name "List invoice generations with --updated-at-min filter"
xbe_json view invoice-generations list --updated-at-min "$UPDATED_AT_FILTER" --limit 5
assert_success

test_name "List invoice generations with --updated-at-max filter"
xbe_json view invoice-generations list --updated-at-max "$UPDATED_AT_FILTER" --limit 5
assert_success

test_name "List invoice generations with --is-updated-at=true filter"
xbe_json view invoice-generations list --is-updated-at true --limit 5
assert_success

test_name "List invoice generations with --is-updated-at=false filter"
xbe_json view invoice-generations list --is-updated-at false --limit 5
assert_success

# ============================================================================
# LIST Tests - Pagination
# ============================================================================

test_name "List invoice generations with --limit"
xbe_json view invoice-generations list --limit 2
assert_success

test_name "List invoice generations with --offset"
xbe_json view invoice-generations list --limit 2 --offset 2
assert_success

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show invoice generation"
SHOW_ID="$SAMPLE_ID"
if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
    SHOW_ID="$CREATED_ID"
fi
if [[ -n "$SHOW_ID" && "$SHOW_ID" != "null" ]]; then
    xbe_json view invoice-generations show "$SHOW_ID"
    assert_success
else
    skip "No invoice generation ID available"
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create invoice generation requires organization-type"
xbe_run do invoice-generations create --organization-id 123
assert_failure

test_name "Create invoice generation requires organization-id"
xbe_run do invoice-generations create --organization-type brokers
assert_failure

if [[ -z "$ORG_ID" ]]; then
    if [[ -n "${XBE_TEST_TIME_CARD_BROKER_ID:-}" ]]; then
        ORG_TYPE="brokers"
        ORG_ID="$XBE_TEST_TIME_CARD_BROKER_ID"
    elif [[ -n "${XBE_TEST_TIME_CARD_CUSTOMER_ID:-}" ]]; then
        ORG_TYPE="customers"
        ORG_ID="$XBE_TEST_TIME_CARD_CUSTOMER_ID"
    elif [[ -n "${XBE_TEST_TIME_CARD_TRUCKER_ID:-}" ]]; then
        ORG_TYPE="truckers"
        ORG_ID="$XBE_TEST_TIME_CARD_TRUCKER_ID"
    fi
fi

if [[ -z "$TIME_CARD_ID" || "$TIME_CARD_ID" == "null" ]]; then
    TIME_CARD_ID="$SAMPLE_TIME_CARD_ID"
fi

if [[ -z "$TIME_CARD_ID" || "$TIME_CARD_ID" == "null" ]]; then
    xbe_json view time-cards list --limit 1
    if [[ $status -eq 0 ]]; then
        TIME_CARD_ID=$(json_get ".[0].id")
    fi
fi

if [[ -z "$ORG_ID" && -n "$TIME_CARD_ID" && "$TIME_CARD_ID" != "null" ]]; then
    pick_org_type_id_from_time_card "$TIME_CARD_ID"
fi

test_name "Create invoice generation"
if [[ -n "$ORG_ID" && "$ORG_ID" != "null" && -n "$ORG_TYPE" && -n "$TIME_CARD_ID" && "$TIME_CARD_ID" != "null" ]]; then
    NOTE=$(unique_name "InvoiceGeneration")
    xbe_json do invoice-generations create \
        --organization-type "$ORG_TYPE" \
        --organization-id "$ORG_ID" \
        --time-card-ids "$TIME_CARD_ID" \
        --note "$NOTE"
    if [[ $status -eq 0 ]]; then
        CREATED_ID=$(json_get ".id")
        if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
            pass
        else
            fail "Created invoice generation but no ID returned"
        fi
    else
        if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]] || [[ "$output" == *"422"* ]] || [[ "$output" == *"409"* ]] || [[ "$output" == *"overlaps"* ]] || [[ "$output" == *"must all relate"* ]]; then
            skip "Create blocked by server policy/validation"
        else
            fail "Failed to create invoice generation: $output"
        fi
    fi
else
    skip "Missing organization or time card IDs (set XBE_TEST_TIME_CARD_ID + XBE_TEST_TIME_CARD_BROKER_ID/CUSTOMER_ID/TRUCKER_ID)"
fi

run_tests
