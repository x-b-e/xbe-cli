#!/bin/bash
#
# XBE CLI Integration Tests: Broker Tenders
#
# Tests CRUD operations for the broker-tenders resource.
#
# COVERAGE: Create + update + delete + list filters + show
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

BROKER_ID="${XBE_TEST_BROKER_ID:-}"
TRUCKER_ID=""
JOB_ID="${XBE_TEST_BROKER_TENDER_JOB_ID:-}"
JOB_SITE_ID="${XBE_TEST_BROKER_TENDER_JOB_SITE_ID:-}"
JOB_NUMBER=""
BROKER_USER_ID=""
TRUCKER_MEMBERSHIP_ID=""
TENDER_ID=""

describe "Resource: broker-tenders"

# ============================================================================
# Broker selection
# ============================================================================

test_name "Find broker for broker tenders"
if [[ -n "$BROKER_ID" ]]; then
    echo "    Using XBE_TEST_BROKER_ID: $BROKER_ID"
    pass
else
    xbe_json view brokers list --limit 1
    if [[ $status -eq 0 ]]; then
        BROKER_ID=$(json_get ".[0].id")
        if [[ -n "$BROKER_ID" && "$BROKER_ID" != "null" ]]; then
            echo "    Using broker ID: $BROKER_ID"
            pass
        else
            fail "No broker available for tests"
            run_tests
        fi
    else
        fail "Failed to list brokers"
        run_tests
    fi
fi

# ============================================================================
# Trucker setup
# ============================================================================

test_name "Create trucker for broker tenders"
TRUCKER_NAME=$(unique_name "BrokerTenderTrucker")
TRUCKER_ADDRESS="100 Tender Lane, Haul City, HC 55555"

xbe_json do truckers create \
    --name "$TRUCKER_NAME" \
    --broker "$BROKER_ID" \
    --company-address "$TRUCKER_ADDRESS"

if [[ $status -eq 0 ]]; then
    TRUCKER_ID=$(json_get ".id")
    if [[ -n "$TRUCKER_ID" && "$TRUCKER_ID" != "null" ]]; then
        register_cleanup "truckers" "$TRUCKER_ID"
        pass
    else
        fail "Created trucker but no ID returned"
        run_tests
    fi
else
    if [[ -n "$XBE_TEST_TRUCKER_ID" ]]; then
        TRUCKER_ID="$XBE_TEST_TRUCKER_ID"
        echo "    Using XBE_TEST_TRUCKER_ID: $TRUCKER_ID"
        pass
    else
        fail "Failed to create trucker and XBE_TEST_TRUCKER_ID not set"
        run_tests
    fi
fi

# ============================================================================
# Job discovery
# ============================================================================

test_name "Find job for broker tenders"
if [[ -n "$JOB_ID" ]]; then
    echo "    Using XBE_TEST_BROKER_TENDER_JOB_ID: $JOB_ID"
    pass
else
    if [[ -z "$XBE_TOKEN" ]]; then
        skip "XBE_TOKEN not set; set XBE_TEST_BROKER_TENDER_JOB_ID to run create/update tests"
    else
        base_url="${XBE_BASE_URL%/}"
        jobs_json=$(curl -s \
            -H "Authorization: Bearer $XBE_TOKEN" \
            -H "Accept: application/vnd.api+json" \
            "$base_url/v1/jobs?filter[broker]=$BROKER_ID&page[limit]=50&fields[jobs]=tenderable,job-number,job-site" || true)

        JOB_ID=$(echo "$jobs_json" | jq -r '.data[] | select(.attributes.tenderable==true) | .id' 2>/dev/null | head -n 1)
        if [[ -z "$JOB_ID" ]]; then
            JOB_ID=$(echo "$jobs_json" | jq -r '.data[0].id // empty' 2>/dev/null | head -n 1)
        fi
        JOB_NUMBER=$(echo "$jobs_json" | jq -r --arg id "$JOB_ID" '.data[] | select(.id==$id) | .attributes["job-number"] // empty' 2>/dev/null | head -n 1)
        if [[ -z "$JOB_SITE_ID" ]]; then
            JOB_SITE_ID=$(echo "$jobs_json" | jq -r --arg id "$JOB_ID" '.data[] | select(.id==$id) | .relationships["job-site"].data.id // empty' 2>/dev/null | head -n 1)
        fi

        if [[ -n "$JOB_ID" ]]; then
            echo "    Using job ID: $JOB_ID"
            pass
        else
            skip "No job available for broker (set XBE_TEST_BROKER_TENDER_JOB_ID)"
        fi
    fi
fi

if [[ -z "$JOB_ID" ]]; then
    echo "Cannot continue without a job ID"
    run_tests
fi

# ============================================================================
# Buyer contact lookup
# ============================================================================

test_name "Find broker user for contact updates"
xbe_json view broker-memberships list --broker "$BROKER_ID" --limit 1
if [[ $status -eq 0 ]]; then
    BROKER_USER_ID=$(json_get ".[0].user_id")
    if [[ -n "$BROKER_USER_ID" && "$BROKER_USER_ID" != "null" ]]; then
        echo "    Using broker user ID: $BROKER_USER_ID"
        pass
    else
        skip "No broker membership user available"
    fi
else
    skip "Failed to list broker memberships"
fi

if [[ -n "$BROKER_USER_ID" ]]; then
    test_name "Create trucker membership for broker user"
    xbe_json do memberships create \
        --user "$BROKER_USER_ID" \
        --organization "Trucker|$TRUCKER_ID" \
        --kind manager
    if [[ $status -eq 0 ]]; then
        TRUCKER_MEMBERSHIP_ID=$(json_get ".id")
        if [[ -n "$TRUCKER_MEMBERSHIP_ID" && "$TRUCKER_MEMBERSHIP_ID" != "null" ]]; then
            register_cleanup "memberships" "$TRUCKER_MEMBERSHIP_ID"
        fi
        pass
    else
        skip "Unable to create trucker membership for contact tests"
    fi
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create broker tender with required fields"
xbe_json do broker-tenders create --job "$JOB_ID" --broker "$BROKER_ID" --trucker "$TRUCKER_ID" --note "Test tender"
if [[ $status -eq 0 ]]; then
    TENDER_ID=$(json_get ".id")
    if [[ -n "$TENDER_ID" && "$TENDER_ID" != "null" ]]; then
        register_cleanup "broker-tenders" "$TENDER_ID"
        pass
    else
        fail "Created broker tender but no ID returned"
        run_tests
    fi
else
    fail "Failed to create broker tender: $output"
    run_tests
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show broker tender"
xbe_json view broker-tenders show "$TENDER_ID"
assert_success

# ============================================================================
# UPDATE Tests - Attributes
# ============================================================================

test_name "Update broker tender note"
xbe_json do broker-tenders update "$TENDER_ID" --note "Updated note"
assert_success

test_name "Update broker tender expires-at"
xbe_json do broker-tenders update "$TENDER_ID" --expires-at "$(date -u +%Y-%m-%dT%H:%M:%SZ)"
assert_success

test_name "Update broker tender payment terms"
xbe_json do broker-tenders update "$TENDER_ID" --payment-terms 30
assert_success

test_name "Update broker tender payment terms and conditions"
xbe_json do broker-tenders update "$TENDER_ID" --payment-terms-and-conditions "Net 30"
assert_success

test_name "Update broker tender restrict-to-customer-truckers true"
xbe_json do broker-tenders update "$TENDER_ID" --restrict-to-customer-truckers true
assert_success

test_name "Update broker tender restrict-to-customer-truckers false"
xbe_json do broker-tenders update "$TENDER_ID" --restrict-to-customer-truckers false
assert_success

test_name "Update broker tender maximum travel minutes"
xbe_json do broker-tenders update "$TENDER_ID" --maximum-travel-minutes 45
assert_success

test_name "Update broker tender billable travel minutes per travel mile"
xbe_json do broker-tenders update "$TENDER_ID" --billable-travel-minutes-per-travel-mile 2
assert_success

test_name "Update broker tender displays-trips true"
xbe_json do broker-tenders update "$TENDER_ID" --displays-trips true
assert_success

test_name "Update broker tender displays-trips false"
xbe_json do broker-tenders update "$TENDER_ID" --displays-trips false
assert_success

test_name "Update broker tender shift rejection permitted true"
xbe_json do broker-tenders update "$TENDER_ID" --is-trucker-shift-rejection-permitted true
assert_success

test_name "Update broker tender shift rejection permitted false"
xbe_json do broker-tenders update "$TENDER_ID" --is-trucker-shift-rejection-permitted false
assert_success

# ============================================================================
# UPDATE Tests - Relationships
# ============================================================================

test_name "Update broker tender job (same ID)"
xbe_json do broker-tenders update "$TENDER_ID" --job "$JOB_ID"
assert_success

test_name "Update broker tender broker (same ID)"
xbe_json do broker-tenders update "$TENDER_ID" --broker "$BROKER_ID"
assert_success

test_name "Update broker tender trucker (same ID)"
xbe_json do broker-tenders update "$TENDER_ID" --trucker "$TRUCKER_ID"
assert_success

if [[ -n "$BROKER_USER_ID" ]]; then
    test_name "Update broker tender buyer contacts"
    xbe_json do broker-tenders update "$TENDER_ID" \
        --buyer-operations-contact "$BROKER_USER_ID" \
        --buyer-financial-contact "$BROKER_USER_ID"
    assert_success

    if [[ -n "$TRUCKER_MEMBERSHIP_ID" ]]; then
        test_name "Update broker tender seller contacts"
        xbe_json do broker-tenders update "$TENDER_ID" \
            --seller-operations-contact "$BROKER_USER_ID" \
            --seller-financial-contact "$BROKER_USER_ID"
        assert_success
    else
        skip "No trucker membership for seller contact updates"
    fi
else
    skip "No broker user available for contact updates"
fi

# ============================================================================
# LIST Tests
# ============================================================================

run_filter() {
    local name="$1"
    shift
    test_name "$name"
    xbe_json view broker-tenders list "$@" --limit 5
    assert_success
}

run_filter "List broker tenders" 
run_filter "Filter by buyer" --buyer "$BROKER_ID"
run_filter "Filter by seller" --seller "$TRUCKER_ID"
run_filter "Filter by broker" --broker "$BROKER_ID"
run_filter "Filter by job" --job "$JOB_ID"
run_filter "Filter by status" --status editing

if [[ -n "$JOB_SITE_ID" ]]; then
    run_filter "Filter by job site" --job-site "$JOB_SITE_ID"
else
    test_name "Filter by job site"
    skip "No job site ID available"
fi

if [[ -n "$JOB_NUMBER" ]]; then
    run_filter "Filter by job number" --job-number "$JOB_NUMBER"
else
    run_filter "Filter by job number" --job-number "JOB"
fi

run_filter "Filter by start-at-min" --start-at-min "2024-01-01T00:00:00Z"
run_filter "Filter by start-at-max" --start-at-max "2030-01-01T00:00:00Z"
run_filter "Filter by end-at-max" --end-at-max "2030-01-01T00:00:00Z"
run_filter "Filter by with-alive-shifts" --with-alive-shifts true
run_filter "Filter by has-flexible-shifts" --has-flexible-shifts false
run_filter "Filter by job production plan name or number" --job-production-plan-name-or-number-like "Plan"
run_filter "Filter by business unit" --business-unit 1
run_filter "Filter by trailer classification or equivalent" --job-production-plan-trailer-classification-or-equivalent 1
run_filter "Filter by job production plan material sites" --job-production-plan-material-sites 1
run_filter "Filter by job production plan material types" --job-production-plan-material-types 1
run_filter "Filter by created-at-min" --created-at-min "2024-01-01T00:00:00Z"
run_filter "Filter by created-at-max" --created-at-max "2030-01-01T00:00:00Z"
run_filter "Filter by updated-at-min" --updated-at-min "2024-01-01T00:00:00Z"
run_filter "Filter by updated-at-max" --updated-at-max "2030-01-01T00:00:00Z"

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete broker tender"
xbe_run do broker-tenders delete "$TENDER_ID" --confirm
assert_success

run_tests
