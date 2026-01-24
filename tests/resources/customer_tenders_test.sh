#!/bin/bash
#
# XBE CLI Integration Tests: Customer Tenders
#
# Tests CRUD operations for the customer-tenders resource.
#
# COVERAGE: Create + update + delete + list filters + show
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

BROKER_ID="${XBE_TEST_BROKER_ID:-}"
JOB_ID="${XBE_TEST_CUSTOMER_TENDER_JOB_ID:-}"
JOB_SITE_ID="${XBE_TEST_CUSTOMER_TENDER_JOB_SITE_ID:-}"
CUSTOMER_ID="${XBE_TEST_CUSTOMER_TENDER_CUSTOMER_ID:-}"
JOB_NUMBER=""
BROKER_USER_ID=""
CUSTOMER_USER_ID=""
TENDER_ID=""
CERT_TYPE_ID=""
CERT_REQ_ID=""
EXTERNAL_ID_TYPE_ID=""
EXTERNAL_ID_ID=""
EXTERNAL_ID_VALUE=""

describe "Resource: customer-tenders"

# ============================================================================
# Broker selection
# ============================================================================

test_name "Find broker for customer tenders"
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
# Job discovery
# ============================================================================

test_name "Find job for customer tenders"
if [[ -n "$JOB_ID" ]]; then
    echo "    Using XBE_TEST_CUSTOMER_TENDER_JOB_ID: $JOB_ID"
    pass

    if [[ -z "$CUSTOMER_ID" && -n "$XBE_TOKEN" ]]; then
        base_url="${XBE_BASE_URL%/}"
        job_json=$(curl -s \
            -H "Authorization: Bearer $XBE_TOKEN" \
            -H "Accept: application/vnd.api+json" \
            "$base_url/v1/jobs/$JOB_ID?fields[jobs]=job-number,job-site,customer" || true)

        CUSTOMER_ID=$(echo "$job_json" | jq -r '.data.relationships.customer.data.id // empty' 2>/dev/null)
        JOB_NUMBER=$(echo "$job_json" | jq -r '.data.attributes["job-number"] // empty' 2>/dev/null)
        if [[ -z "$JOB_SITE_ID" ]]; then
            JOB_SITE_ID=$(echo "$job_json" | jq -r '.data.relationships["job-site"].data.id // empty' 2>/dev/null)
        fi
    fi
else
    if [[ -z "$XBE_TOKEN" ]]; then
        skip "XBE_TOKEN not set; set XBE_TEST_CUSTOMER_TENDER_JOB_ID to run create/update tests"
    else
        base_url="${XBE_BASE_URL%/}"
        jobs_json=$(curl -s \
            -H "Authorization: Bearer $XBE_TOKEN" \
            -H "Accept: application/vnd.api+json" \
            "$base_url/v1/jobs?filter[broker]=$BROKER_ID&page[limit]=50&fields[jobs]=tenderable,job-number,job-site,customer" || true)

        JOB_ID=$(echo "$jobs_json" | jq -r '.data[] | select(.attributes.tenderable==true) | .id' 2>/dev/null | head -n 1)
        if [[ -z "$JOB_ID" ]]; then
            JOB_ID=$(echo "$jobs_json" | jq -r '.data[0].id // empty' 2>/dev/null | head -n 1)
        fi
        JOB_NUMBER=$(echo "$jobs_json" | jq -r --arg id "$JOB_ID" '.data[] | select(.id==$id) | .attributes["job-number"] // empty' 2>/dev/null | head -n 1)
        if [[ -z "$JOB_SITE_ID" ]]; then
            JOB_SITE_ID=$(echo "$jobs_json" | jq -r --arg id "$JOB_ID" '.data[] | select(.id==$id) | .relationships["job-site"].data.id // empty' 2>/dev/null | head -n 1)
        fi
        if [[ -z "$CUSTOMER_ID" ]]; then
            CUSTOMER_ID=$(echo "$jobs_json" | jq -r --arg id "$JOB_ID" '.data[] | select(.id==$id) | .relationships.customer.data.id // empty' 2>/dev/null | head -n 1)
        fi

        if [[ -n "$JOB_ID" ]]; then
            echo "    Using job ID: $JOB_ID"
            pass
        else
            skip "No job available for broker (set XBE_TEST_CUSTOMER_TENDER_JOB_ID)"
        fi
    fi
fi

if [[ -z "$JOB_ID" ]]; then
    echo "Cannot continue without a job ID"
    run_tests
fi

if [[ -z "$CUSTOMER_ID" ]]; then
    fail "No customer available for job (set XBE_TEST_CUSTOMER_TENDER_CUSTOMER_ID)"
    run_tests
fi

# ============================================================================
# Contact lookup
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

test_name "Find customer user for contact updates"
xbe_json view memberships list --organization "Customer|$CUSTOMER_ID" --limit 1
if [[ $status -eq 0 ]]; then
    CUSTOMER_USER_ID=$(json_get ".[0].user_id")
    if [[ -n "$CUSTOMER_USER_ID" && "$CUSTOMER_USER_ID" != "null" ]]; then
        echo "    Using customer user ID: $CUSTOMER_USER_ID"
        pass
    else
        skip "No customer membership user available"
    fi
else
    skip "Failed to list customer memberships"
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create customer tender with required fields"
xbe_json do customer-tenders create --job "$JOB_ID" --customer "$CUSTOMER_ID" --broker "$BROKER_ID" --note "Test tender"
if [[ $status -eq 0 ]]; then
    TENDER_ID=$(json_get ".id")
    if [[ -n "$TENDER_ID" && "$TENDER_ID" != "null" ]]; then
        register_cleanup "customer-tenders" "$TENDER_ID"
        pass
    else
        fail "Created customer tender but no ID returned"
        run_tests
    fi
else
    fail "Failed to create customer tender: $output"
    run_tests
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show customer tender"
xbe_json view customer-tenders show "$TENDER_ID"
assert_success

# ============================================================================
# UPDATE Tests - Attributes
# ============================================================================

test_name "Update customer tender note"
xbe_json do customer-tenders update "$TENDER_ID" --note "Updated note"
assert_success

test_name "Update customer tender expires-at"
xbe_json do customer-tenders update "$TENDER_ID" --expires-at "$(date -u +%Y-%m-%dT%H:%M:%SZ)"
assert_success

test_name "Update customer tender payment terms"
xbe_json do customer-tenders update "$TENDER_ID" --payment-terms 30
assert_success

test_name "Update customer tender payment terms and conditions"
xbe_json do customer-tenders update "$TENDER_ID" --payment-terms-and-conditions "Net 30"
assert_success

test_name "Update customer tender restrict-to-customer-truckers true"
xbe_json do customer-tenders update "$TENDER_ID" --restrict-to-customer-truckers true
assert_success

test_name "Update customer tender restrict-to-customer-truckers false"
xbe_json do customer-tenders update "$TENDER_ID" --restrict-to-customer-truckers false
assert_success

test_name "Update customer tender maximum travel minutes"
xbe_json do customer-tenders update "$TENDER_ID" --maximum-travel-minutes 45
assert_success

test_name "Update customer tender billable travel minutes per travel mile"
xbe_json do customer-tenders update "$TENDER_ID" --billable-travel-minutes-per-travel-mile 2
assert_success

test_name "Update customer tender displays-trips true"
xbe_json do customer-tenders update "$TENDER_ID" --displays-trips true
assert_success

test_name "Update customer tender displays-trips false"
xbe_json do customer-tenders update "$TENDER_ID" --displays-trips false
assert_success

test_name "Update customer tender shift rejection permitted true"
xbe_json do customer-tenders update "$TENDER_ID" --is-trucker-shift-rejection-permitted true
assert_success

test_name "Update customer tender shift rejection permitted false"
xbe_json do customer-tenders update "$TENDER_ID" --is-trucker-shift-rejection-permitted false
assert_success

# ============================================================================
# UPDATE Tests - Relationships
# ============================================================================

test_name "Update customer tender job (same ID)"
xbe_json do customer-tenders update "$TENDER_ID" --job "$JOB_ID"
assert_success

test_name "Update customer tender customer (same ID)"
xbe_json do customer-tenders update "$TENDER_ID" --customer "$CUSTOMER_ID"
assert_success

test_name "Update customer tender broker (same ID)"
xbe_json do customer-tenders update "$TENDER_ID" --broker "$BROKER_ID"
assert_success

if [[ -n "$BROKER_USER_ID" ]]; then
    test_name "Update customer tender seller contacts"
    xbe_json do customer-tenders update "$TENDER_ID" \
        --seller-operations-contact "$BROKER_USER_ID" \
        --seller-financial-contact "$BROKER_USER_ID"
    assert_success
else
    skip "No broker user available for seller contact updates"
fi

if [[ -n "$CUSTOMER_USER_ID" ]]; then
    test_name "Update customer tender buyer contacts"
    xbe_json do customer-tenders update "$TENDER_ID" \
        --buyer-operations-contact "$CUSTOMER_USER_ID" \
        --buyer-financial-contact "$CUSTOMER_USER_ID"
    assert_success
else
    skip "No customer user available for buyer contact updates"
fi

# ============================================================================
# UPDATE Tests - Certification Requirements
# ============================================================================

test_name "Create certification type for customer tender"
CERT_TYPE_NAME=$(unique_name "CustomerTenderCert")
xbe_json do certification-types create \
    --name "$CERT_TYPE_NAME" \
    --can-apply-to Customer \
    --can-be-requirement-of CustomerTender \
    --broker "$BROKER_ID"

if [[ $status -eq 0 ]]; then
    CERT_TYPE_ID=$(json_get ".id")
    if [[ -n "$CERT_TYPE_ID" && "$CERT_TYPE_ID" != "null" ]]; then
        register_cleanup "certification-types" "$CERT_TYPE_ID"
        pass
    else
        skip "Created certification type but no ID returned"
    fi
else
    skip "Unable to create certification type"
fi

if [[ -n "$CERT_TYPE_ID" ]]; then
    test_name "Create certification requirement for customer tender"
    xbe_json do certification-requirements create \
        --certification-type "$CERT_TYPE_ID" \
        --required-by-type customer-tenders \
        --required-by-id "$TENDER_ID"

    if [[ $status -eq 0 ]]; then
        CERT_REQ_ID=$(json_get ".id")
        if [[ -n "$CERT_REQ_ID" && "$CERT_REQ_ID" != "null" ]]; then
            register_cleanup "certification-requirements" "$CERT_REQ_ID"
            pass
        else
            skip "Created certification requirement but no ID returned"
        fi
    else
        skip "Unable to create certification requirement"
    fi
fi

if [[ -n "$CERT_REQ_ID" ]]; then
    test_name "Update customer tender certification requirements"
    xbe_json do customer-tenders update "$TENDER_ID" --certification-requirements "$CERT_REQ_ID"
    assert_success
else
    skip "No certification requirement available for update"
fi

# ============================================================================
# External identification setup
# ============================================================================

test_name "Create external identification type for customer tender"
EXT_TYPE_NAME=$(unique_name "CustomerTenderExt")
xbe_json do external-identification-types create \
    --name "$EXT_TYPE_NAME" \
    --can-apply-to CustomerTender

if [[ $status -eq 0 ]]; then
    EXTERNAL_ID_TYPE_ID=$(json_get ".id")
    if [[ -n "$EXTERNAL_ID_TYPE_ID" && "$EXTERNAL_ID_TYPE_ID" != "null" ]]; then
        register_cleanup "external-identification-types" "$EXTERNAL_ID_TYPE_ID"
        pass
    else
        skip "Created external identification type but no ID returned"
    fi
else
    skip "Unable to create external identification type"
fi

if [[ -n "$EXTERNAL_ID_TYPE_ID" ]]; then
    test_name "Create external identification for customer tender"
    EXTERNAL_ID_VALUE="CT-${TENDER_ID}-${RANDOM}"
    xbe_json do external-identifications create \
        --external-identification-type "$EXTERNAL_ID_TYPE_ID" \
        --identifies-type customer-tenders \
        --identifies-id "$TENDER_ID" \
        --value "$EXTERNAL_ID_VALUE"

    if [[ $status -eq 0 ]]; then
        EXTERNAL_ID_ID=$(json_get ".id")
        if [[ -n "$EXTERNAL_ID_ID" && "$EXTERNAL_ID_ID" != "null" ]]; then
            register_cleanup "external-identifications" "$EXTERNAL_ID_ID"
            pass
        else
            skip "Created external identification but no ID returned"
        fi
    else
        skip "Unable to create external identification"
    fi
else
    skip "No external identification type available"
fi

# ============================================================================
# LIST Tests
# ============================================================================

run_filter() {
    local name="$1"
    shift
    test_name "$name"
    xbe_json view customer-tenders list "$@" --limit 5
    assert_success
}

run_filter "List customer tenders"
run_filter "Filter by buyer" --buyer "$CUSTOMER_ID"
run_filter "Filter by seller" --seller "$BROKER_ID"
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

if [[ -n "$EXTERNAL_ID_VALUE" ]]; then
    run_filter "Filter by external identification value" --external-identification-value "$EXTERNAL_ID_VALUE"
else
    run_filter "Filter by external identification value" --external-identification-value "CT-TEST"
fi

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete customer tender"
xbe_run do customer-tenders delete "$TENDER_ID" --confirm
assert_success

run_tests
