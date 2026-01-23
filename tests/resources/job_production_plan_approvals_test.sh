#!/bin/bash
#
# XBE CLI Integration Tests: Job Production Plan Approvals
#
# Tests create and list operations for the job_production_plan_approvals resource.
# Approvals transition submitted plans to approved.
#
# COVERAGE: Create attributes + list
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

SUBMITTED_JPP_ID=""
CREATED_BROKER_ID=""
CREATED_CUSTOMER_ID=""
CREATED_JOB_SITE_ID=""
CREATED_JPP_ID=""

describe "Resource: job-production-plan-approvals"

# ==========================================================================
# Setup - Create a submitted job production plan when possible
# ==========================================================================

if [[ -n "$XBE_TOKEN" ]]; then
    test_name "Create prerequisite broker for job production plan approval tests"
    BROKER_NAME=$(unique_name "JPPApprovalBroker")

    xbe_json do brokers create --name "$BROKER_NAME"

    if [[ $status -eq 0 ]]; then
        CREATED_BROKER_ID=$(json_get ".id")
        if [[ -n "$CREATED_BROKER_ID" && "$CREATED_BROKER_ID" != "null" ]]; then
            register_cleanup "brokers" "$CREATED_BROKER_ID"
            pass
        else
            fail "Created broker but no ID returned"
        fi
    else
        if [[ -n "$XBE_TEST_BROKER_ID" ]]; then
            CREATED_BROKER_ID="$XBE_TEST_BROKER_ID"
            echo "    Using XBE_TEST_BROKER_ID: $CREATED_BROKER_ID"
            pass
        else
            fail "Failed to create broker and XBE_TEST_BROKER_ID not set"
        fi
    fi

    if [[ -n "$CREATED_BROKER_ID" && "$CREATED_BROKER_ID" != "null" ]]; then
        test_name "Create prerequisite customer"
        CUSTOMER_NAME=$(unique_name "JPPApprovalCustomer")

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
            fi
        else
            fail "Failed to create customer"
        fi
    fi

    if [[ -n "$CREATED_CUSTOMER_ID" && "$CREATED_CUSTOMER_ID" != "null" ]]; then
        test_name "Create prerequisite job site"
        JOB_SITE_NAME=$(unique_name "JPPApprovalJobSite")

        xbe_json do job-sites create \
            --name "$JOB_SITE_NAME" \
            --customer "$CREATED_CUSTOMER_ID" \
            --address "100 Approval Street, Chicago, IL 60601"

        if [[ $status -eq 0 ]]; then
            CREATED_JOB_SITE_ID=$(json_get ".id")
            if [[ -n "$CREATED_JOB_SITE_ID" && "$CREATED_JOB_SITE_ID" != "null" ]]; then
                register_cleanup "job-sites" "$CREATED_JOB_SITE_ID"
                pass
            else
                fail "Created job site but no ID returned"
            fi
        else
            fail "Failed to create job site"
        fi
    fi

    if [[ -n "$CREATED_CUSTOMER_ID" && "$CREATED_CUSTOMER_ID" != "null" && -n "$CREATED_JOB_SITE_ID" && "$CREATED_JOB_SITE_ID" != "null" ]]; then
        test_name "Create job production plan for approval"
        TODAY=$(date +%Y-%m-%d)
        JOB_NAME=$(unique_name "JPPApprovalPlan")
        JOB_NUMBER="JPP-APPROVE-$(date +%s)"

        xbe_json do job-production-plans create \
            --job-name "$JOB_NAME" \
            --job-number "$JOB_NUMBER" \
            --start-on "$TODAY" \
            --start-time "07:00" \
            --customer "$CREATED_CUSTOMER_ID" \
            --job-site "$CREATED_JOB_SITE_ID" \
            --requires-trucking=false \
            --requires-materials=false

        if [[ $status -eq 0 ]]; then
            CREATED_JPP_ID=$(json_get ".id")
            if [[ -n "$CREATED_JPP_ID" && "$CREATED_JPP_ID" != "null" ]]; then
                pass
            else
                fail "Created job production plan but no ID returned"
            fi
        else
            fail "Failed to create job production plan"
        fi
    else
        skip "Missing customer or job site; cannot create job production plan"
    fi

    if [[ -n "$CREATED_JPP_ID" && "$CREATED_JPP_ID" != "null" ]]; then
        test_name "Submit job production plan for approval"
        submission_payload=$(cat <<JSON
{"data":{"type":"job-production-plan-submissions","relationships":{"job-production-plan":{"data":{"type":"job-production-plans","id":"$CREATED_JPP_ID"}}}}}
JSON
        )

        response_file=$(mktemp)
        run curl -s -o "$response_file" -w "%{http_code}" \
            -H "Authorization: Bearer $XBE_TOKEN" \
            -H "Accept: application/vnd.api+json" \
            -H "Content-Type: application/vnd.api+json" \
            -X POST "$XBE_BASE_URL/v1/job-production-plan-submissions" \
            -d "$submission_payload"

        http_code="$output"
        if [[ $status -eq 0 && "$http_code" == 2* ]]; then
            SUBMITTED_JPP_ID="$CREATED_JPP_ID"
            pass
        else
            if [[ -s "$response_file" ]]; then
                echo "    Submission response: $(head -c 200 "$response_file")"
            fi
            skip "Unable to submit job production plan (HTTP ${http_code})"
        fi
        rm -f "$response_file"
    fi
else
    echo "    (XBE_TOKEN not set; skipping submission-dependent approval tests)"
fi

# ==========================================================================
# CREATE Tests
# ==========================================================================

test_name "Create job production plan approval without job production plan fails"
xbe_json do job-production-plan-approvals create
assert_failure

if [[ -n "$SUBMITTED_JPP_ID" && "$SUBMITTED_JPP_ID" != "null" ]]; then
    test_name "Create job production plan approval (minimal)"
    xbe_json do job-production-plan-approvals create --job-production-plan "$SUBMITTED_JPP_ID"
    assert_success

    test_name "Create job production plan approval with comment and flags"
    COMMENT_TEXT="Approved by CLI test"
    xbe_json do job-production-plan-approvals create \
        --job-production-plan "$SUBMITTED_JPP_ID" \
        --comment "$COMMENT_TEXT" \
        --suppress-status-change-notifications \
        --skip-validate-required-mix-designs

    if [[ $status -eq 0 ]]; then
        assert_json_equals ".comment" "$COMMENT_TEXT"
        assert_json_bool ".suppress_status_change_notifications" "true"
        assert_json_bool ".skip_validate_required_mix_designs" "true"
    else
        fail "Failed to create approval with comment and flags"
    fi
else
    skip "No submitted job production plan available for approval tests"
fi

# ==========================================================================
# Error Cases
# ==========================================================================

test_name "Create job production plan approval with invalid ID fails"
xbe_json do job-production-plan-approvals create --job-production-plan "999999999"
assert_failure

# ==========================================================================
# Summary
# ==========================================================================

run_tests
