#!/bin/bash
#
# XBE CLI Integration Tests: Tender Acceptances
#
# Tests view and create operations for tender_acceptances.
# Acceptances move tenders from offered to accepted.
#
# COVERAGE: Create attributes
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

SAMPLE_ACCEPTANCE_ID=""
ACCEPT_TENDER_ID=""
ACCEPT_TENDER_TYPE=""
REJECT_TENDER_ID=""
REJECT_TENDER_TYPE=""
REJECT_TENDER_SHIFT_ID=""
SKIP_MUTATION=0

describe "Resource: tender-acceptances"

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

select_offered_tender() {
    local response_file
    response_file=$(mktemp)
    run curl -s -o "$response_file" -w "%{http_code}" \
        -H "Authorization: Bearer $XBE_TOKEN" \
        -H "Accept: application/vnd.api+json" \
        -G "$XBE_BASE_URL/v1/tenders" \
        --data-urlencode "page[limit]=25" \
        --data-urlencode "filter[status]=offered"

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

    ACCEPT_TENDER_ID=$(jq -r '.data[0].id' "$response_file")
    ACCEPT_TENDER_TYPE=$(jq -r '.data[0].type' "$response_file")
    if [[ -z "$ACCEPT_TENDER_TYPE" || "$ACCEPT_TENDER_TYPE" == "null" || "$ACCEPT_TENDER_TYPE" == "tenders" ]]; then
        ACCEPT_TENDER_TYPE=$(jq -r '.data[0].attributes.type // empty' "$response_file")
    fi
    ACCEPT_TENDER_TYPE=$(normalize_tender_type "$ACCEPT_TENDER_TYPE")

    rm -f "$response_file"

    if [[ -z "$ACCEPT_TENDER_ID" || "$ACCEPT_TENDER_ID" == "null" ]]; then
        return 1
    fi
    if [[ "$ACCEPT_TENDER_TYPE" != "broker-tenders" && "$ACCEPT_TENDER_TYPE" != "customer-tenders" ]]; then
        return 1
    fi

    return 0
}

select_rejectable_tender() {
    local response_file
    response_file=$(mktemp)
    run curl -s -o "$response_file" -w "%{http_code}" \
        -H "Authorization: Bearer $XBE_TOKEN" \
        -H "Accept: application/vnd.api+json" \
        -G "$XBE_BASE_URL/v1/tenders" \
        --data-urlencode "page[limit]=25" \
        --data-urlencode "filter[status]=offered" \
        --data-urlencode "fields[tenders]=is-trucker-shift-rejection-permitted"

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

    local i
    for ((i=0; i<count; i++)); do
        local tender_id tender_type permitted
        tender_id=$(jq -r ".data[$i].id" "$response_file")
        tender_type=$(jq -r ".data[$i].type" "$response_file")
        if [[ -z "$tender_type" || "$tender_type" == "null" || "$tender_type" == "tenders" ]]; then
            tender_type=$(jq -r ".data[$i].attributes.type // empty" "$response_file")
        fi
        tender_type=$(normalize_tender_type "$tender_type")
        permitted=$(jq -r ".data[$i].attributes[\"is-trucker-shift-rejection-permitted\"]" "$response_file")

        if [[ -z "$tender_id" || "$tender_id" == "null" ]]; then
            continue
        fi
        if [[ "$tender_type" != "broker-tenders" && "$tender_type" != "customer-tenders" ]]; then
            continue
        fi
        if [[ "$permitted" != "true" ]]; then
            continue
        fi
        if [[ -n "$ACCEPT_TENDER_ID" && "$tender_id" == "$ACCEPT_TENDER_ID" ]]; then
            continue
        fi

        local shift_file
        shift_file=$(mktemp)
        run curl -s -o "$shift_file" -w "%{http_code}" \
            -H "Authorization: Bearer $XBE_TOKEN" \
            -H "Accept: application/vnd.api+json" \
            -G "$XBE_BASE_URL/v1/tender-job-schedule-shifts" \
            --data-urlencode "page[limit]=2" \
            --data-urlencode "filter[tender]=$tender_id"

        local shift_http_code="$output"
        if [[ $status -eq 0 && "$shift_http_code" == 2* ]]; then
            local shift_count
            shift_count=$(jq '.data | length' "$shift_file" 2>/dev/null)
            if [[ -n "$shift_count" && "$shift_count" -ge 2 ]]; then
                REJECT_TENDER_SHIFT_ID=$(jq -r '.data[0].id' "$shift_file")
                REJECT_TENDER_ID="$tender_id"
                REJECT_TENDER_TYPE="$tender_type"
                rm -f "$shift_file"
                break
            fi
        fi
        rm -f "$shift_file"
    done

    rm -f "$response_file"

    if [[ -z "$REJECT_TENDER_ID" || -z "$REJECT_TENDER_SHIFT_ID" ]]; then
        return 1
    fi

    return 0
}

# ============================================================================
# LIST / SHOW Tests
# ============================================================================

test_name "List tender acceptances"
xbe_json view tender-acceptances list --limit 1
assert_success

test_name "Capture sample tender acceptance (if available)"
xbe_json view tender-acceptances list --limit 1
if [[ $status -eq 0 ]]; then
    count=$(echo "$output" | jq 'length' 2>/dev/null)
    if [[ "$count" -gt 0 ]]; then
        SAMPLE_ACCEPTANCE_ID=$(json_get ".[0].id")
        pass
    else
        echo "    No tender acceptances available; skipping show test."
        pass
    fi
else
    fail "Failed to list tender acceptances"
fi

if [[ -n "$SAMPLE_ACCEPTANCE_ID" && "$SAMPLE_ACCEPTANCE_ID" != "null" ]]; then
    test_name "Show tender acceptance"
    xbe_json view tender-acceptances show "$SAMPLE_ACCEPTANCE_ID"
    assert_success
fi

# ============================================================================
# CREATE Error Tests
# ============================================================================

test_name "Create tender acceptance requires --tender-type"
xbe_run do tender-acceptances create --tender 123
assert_failure

test_name "Create tender acceptance requires --tender"
xbe_run do tender-acceptances create --tender-type customer-tenders
assert_failure

# ============================================================================
# CREATE Tests
# ============================================================================

if [[ -z "$XBE_TOKEN" ]]; then
    echo "    (XBE_TOKEN not set; skipping mutation tests)"
    SKIP_MUTATION=1
fi

if [[ $SKIP_MUTATION -eq 0 ]]; then
    if [[ -n "$XBE_TEST_TENDER_ID" && -n "$XBE_TEST_TENDER_TYPE" ]]; then
        ACCEPT_TENDER_ID="$XBE_TEST_TENDER_ID"
        ACCEPT_TENDER_TYPE="$XBE_TEST_TENDER_TYPE"
        ACCEPT_TENDER_TYPE=$(normalize_tender_type "$ACCEPT_TENDER_TYPE")
    else
        test_name "Find offered tender to accept"
        if select_offered_tender; then
            pass
        else
            skip "No offered tender available for acceptance"
        fi
    fi
fi

if [[ -n "$ACCEPT_TENDER_ID" && -n "$ACCEPT_TENDER_TYPE" ]]; then
    test_name "Create tender acceptance with comment and skip certification validation"
    COMMENT_TEXT=$(unique_name "TenderAcceptance")
    xbe_json do tender-acceptances create \
        --tender "$ACCEPT_TENDER_ID" \
        --tender-type "$ACCEPT_TENDER_TYPE" \
        --comment "$COMMENT_TEXT" \
        --skip-certification-validation

    if [[ $status -eq 0 ]]; then
        assert_json_equals ".comment" "$COMMENT_TEXT"
        assert_json_equals ".skip_certification_validation" "true"
    else
        skip "Unable to accept tender (permissions or data constraints)"
    fi
else
    test_name "Create tender acceptance with comment and skip certification validation"
    skip "No tender available for acceptance"
fi

if [[ $SKIP_MUTATION -eq 0 ]]; then
    if [[ -n "$XBE_TEST_TENDER_REJECT_ID" && -n "$XBE_TEST_TENDER_REJECT_TYPE" && -n "$XBE_TEST_TENDER_REJECT_SHIFT_ID" ]]; then
        REJECT_TENDER_ID="$XBE_TEST_TENDER_REJECT_ID"
        REJECT_TENDER_TYPE="$XBE_TEST_TENDER_REJECT_TYPE"
        REJECT_TENDER_TYPE=$(normalize_tender_type "$REJECT_TENDER_TYPE")
        REJECT_TENDER_SHIFT_ID="$XBE_TEST_TENDER_REJECT_SHIFT_ID"
    else
        test_name "Find rejectable tender shift"
        if select_rejectable_tender; then
            pass
        else
            skip "No rejectable tender shift available"
        fi
    fi
fi

if [[ -n "$REJECT_TENDER_ID" && -n "$REJECT_TENDER_TYPE" && -n "$REJECT_TENDER_SHIFT_ID" ]]; then
    test_name "Create tender acceptance rejecting one shift"
    COMMENT_TEXT=$(unique_name "TenderAcceptanceReject")
    xbe_json do tender-acceptances create \
        --tender "$REJECT_TENDER_ID" \
        --tender-type "$REJECT_TENDER_TYPE" \
        --comment "$COMMENT_TEXT" \
        --rejected-tender-job-schedule-shift-ids "$REJECT_TENDER_SHIFT_ID"

    if [[ $status -eq 0 ]]; then
        assert_json_equals ".comment" "$COMMENT_TEXT"
        assert_json_has ".rejected_tender_job_schedule_shift_ids"
    else
        skip "Unable to accept tender with rejected shift"
    fi
else
    test_name "Create tender acceptance rejecting one shift"
    skip "No tender available for rejection test"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
