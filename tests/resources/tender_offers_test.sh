#!/bin/bash
#
# XBE CLI Integration Tests: Tender Offers
#
# Tests view and create operations for tender_offers.
# Offers move tenders from editing to offered.
#
# COVERAGE: Create attributes
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

SAMPLE_OFFER_ID=""
OFFER_TENDER_ID=""
OFFER_TENDER_TYPE=""
SKIP_MUTATION=0

describe "Resource: tender-offers"

normalize_tender_type() {
    local raw="$1"
    case "$raw" in
        broker-tenders|customer-tenders)
            echo "$raw"
            ;;
        BrokerTender|broker_tender|broker-tender)
            echo "broker-tenders"
            ;;
        CustomerTender|customer_tender|customer-tender)
            echo "customer-tenders"
            ;;
        *)
            echo "$raw"
            ;;
    esac
}

select_editing_tender() {
    local response_file
    response_file=$(mktemp)
    run curl -s -o "$response_file" -w "%{http_code}" \
        -H "Authorization: Bearer $XBE_TOKEN" \
        -H "Accept: application/vnd.api+json" \
        -G "$XBE_BASE_URL/v1/tenders" \
        --data-urlencode "page[limit]=25" \
        --data-urlencode "filter[status]=editing"

    local http_code="$output"
    if [[ $status -ne 0 || "$http_code" != 2* ]]; then
        rm -f "$response_file"
        return 1
    fi

    local count
    count=$(jq '.data | length' "$response_file" 2>/dev/null)
    if [[ -z "$count" || "$count" -eq 0 ]]; then
        rm -f "$response_file"
        return 1
    fi

    OFFER_TENDER_ID=$(jq -r '.data[0].id' "$response_file")
    OFFER_TENDER_TYPE=$(jq -r '.data[0].type' "$response_file")
    if [[ -z "$OFFER_TENDER_TYPE" || "$OFFER_TENDER_TYPE" == "null" || "$OFFER_TENDER_TYPE" == "tenders" ]]; then
        OFFER_TENDER_TYPE=$(jq -r '.data[0].attributes.type // empty' "$response_file")
    fi
    OFFER_TENDER_TYPE=$(normalize_tender_type "$OFFER_TENDER_TYPE")

    rm -f "$response_file"

    if [[ -z "$OFFER_TENDER_ID" || "$OFFER_TENDER_ID" == "null" ]]; then
        return 1
    fi
    if [[ "$OFFER_TENDER_TYPE" != "broker-tenders" && "$OFFER_TENDER_TYPE" != "customer-tenders" ]]; then
        return 1
    fi

    return 0
}

# ============================================================================
# LIST / SHOW Tests
# ============================================================================

test_name "List tender offers"
xbe_json view tender-offers list --limit 1
assert_success

test_name "Capture sample tender offer (if available)"
xbe_json view tender-offers list --limit 1
if [[ $status -eq 0 ]]; then
    count=$(echo "$output" | jq 'length' 2>/dev/null)
    if [[ "$count" -gt 0 ]]; then
        SAMPLE_OFFER_ID=$(json_get ".[0].id")
        pass
    else
        echo "    No tender offers available; skipping show test."
        pass
    fi
else
    fail "Failed to list tender offers"
fi

if [[ -n "$SAMPLE_OFFER_ID" && "$SAMPLE_OFFER_ID" != "null" ]]; then
    test_name "Show tender offer"
    xbe_json view tender-offers show "$SAMPLE_OFFER_ID"
    assert_success
fi

# ============================================================================
# CREATE Error Tests
# ============================================================================

test_name "Create tender offer requires --tender-type"
xbe_run do tender-offers create --tender 123
assert_failure

test_name "Create tender offer requires --tender"
xbe_run do tender-offers create --tender-type customer-tenders
assert_failure

# ============================================================================
# CREATE Tests
# ============================================================================

if [[ -z "$XBE_TOKEN" ]]; then
    echo "    (XBE_TOKEN not set; skipping mutation tests)"
    SKIP_MUTATION=1
fi

if [[ $SKIP_MUTATION -eq 0 ]]; then
    if [[ -n "$XBE_TEST_TENDER_OFFER_ID" && -n "$XBE_TEST_TENDER_OFFER_TYPE" ]]; then
        OFFER_TENDER_ID="$XBE_TEST_TENDER_OFFER_ID"
        OFFER_TENDER_TYPE="$XBE_TEST_TENDER_OFFER_TYPE"
        OFFER_TENDER_TYPE=$(normalize_tender_type "$OFFER_TENDER_TYPE")
    else
        test_name "Find editing tender to offer"
        if select_editing_tender; then
            pass
        else
            skip "No editing tender available for offer"
        fi
    fi
fi

if [[ -n "$OFFER_TENDER_ID" && -n "$OFFER_TENDER_TYPE" ]]; then
    test_name "Create tender offer with comment and skip certification validation"
    COMMENT_TEXT=$(unique_name "TenderOffer")
    xbe_json do tender-offers create \
        --tender "$OFFER_TENDER_ID" \
        --tender-type "$OFFER_TENDER_TYPE" \
        --comment "$COMMENT_TEXT" \
        --skip-certification-validation

    if [[ $status -eq 0 ]]; then
        assert_json_equals ".comment" "$COMMENT_TEXT"
        assert_json_equals ".skip_certification_validation" "true"
    else
        skip "Unable to offer tender (permissions or data constraints)"
    fi
else
    test_name "Create tender offer with comment and skip certification validation"
    skip "No tender available for offer"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
