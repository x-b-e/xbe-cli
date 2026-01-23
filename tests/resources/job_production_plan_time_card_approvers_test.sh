#!/bin/bash
#
# XBE CLI Integration Tests: Job Production Plan Time Card Approvers
#
# Tests view and create/delete operations for the job_production_plan_time_card_approvers resource.
# These links attach approver users to job production plans.
#
# COVERAGE: Create attributes + list filters + delete
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_LINK_ID=""
CREATED_BROKER_ID=""
CREATED_CUSTOMER_ID=""
CREATED_JOB_SITE_ID=""
CREATED_JPP_ID=""
CREATED_USER_ID=""
CREATED_MEMBERSHIP_ID=""
SKIP_MUTATION=0

TODAY=$(date +%Y-%m-%d)

describe "Resource: job-production-plan-time-card-approvers"

# ============================================================================
# Setup prerequisites (requires XBE_TOKEN for direct API calls)
# ============================================================================

if [[ -z "$XBE_TOKEN" ]]; then
    echo "    (XBE_TOKEN not set; skipping setup and mutation tests)"
    SKIP_MUTATION=1
else
    test_name "Create prerequisite broker"
    BROKER_NAME=$(unique_name "JPPApproverBroker")

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
        CUSTOMER_NAME=$(unique_name "JPPApproverCustomer")

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
        JOB_SITE_NAME=$(unique_name "JPPApproverSite")

        xbe_json do job-sites create \
            --name "$JOB_SITE_NAME" \
            --customer "$CREATED_CUSTOMER_ID" \
            --address "100 Main St, Chicago, IL 60601"

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

    if [[ -n "$CREATED_CUSTOMER_ID" && -n "$CREATED_JOB_SITE_ID" ]]; then
        test_name "Create prerequisite job production plan"
        PLAN_NAME=$(unique_name "JPPApproverPlan")

        xbe_json do job-production-plans create \
            --job-name "$PLAN_NAME" \
            --start-on "$TODAY" \
            --start-time "07:00" \
            --customer "$CREATED_CUSTOMER_ID" \
            --job-site "$CREATED_JOB_SITE_ID" \
            --requires-trucking=false \
            --requires-materials=false

        if [[ $status -eq 0 ]]; then
            CREATED_JPP_ID=$(json_get ".id")
            if [[ -n "$CREATED_JPP_ID" && "$CREATED_JPP_ID" != "null" ]]; then
                register_cleanup "job-production-plans" "$CREATED_JPP_ID"
                pass
            else
                fail "Created job production plan but no ID returned"
            fi
        else
            fail "Failed to create job production plan"
        fi
    fi

    test_name "Create prerequisite user"
    USER_NAME=$(unique_name "JPPApproverUser")
    USER_EMAIL=$(unique_email)

    xbe_json do users create \
        --name "$USER_NAME" \
        --email "$USER_EMAIL"

    if [[ $status -eq 0 ]]; then
        CREATED_USER_ID=$(json_get ".id")
        if [[ -n "$CREATED_USER_ID" && "$CREATED_USER_ID" != "null" ]]; then
            pass
        else
            fail "Created user but no ID returned"
        fi
    else
        fail "Failed to create user"
    fi

    if [[ -n "$CREATED_USER_ID" && -n "$CREATED_CUSTOMER_ID" ]]; then
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
            fi
        else
            fail "Failed to create membership"
        fi
    fi
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create time card approver without required fields fails"
xbe_json do job-production-plan-time-card-approvers create
assert_failure

if [[ $SKIP_MUTATION -eq 1 ]]; then
    skip "Skipping create/delete/filter tests without XBE_TOKEN"
fi

if [[ -n "$CREATED_JPP_ID" && -n "$CREATED_USER_ID" ]]; then
    test_name "Create job production plan time card approver"
    xbe_json do job-production-plan-time-card-approvers create \
        --job-production-plan "$CREATED_JPP_ID" \
        --user "$CREATED_USER_ID"

    if [[ $status -eq 0 ]]; then
        CREATED_LINK_ID=$(json_get ".id")
        if [[ -n "$CREATED_LINK_ID" && "$CREATED_LINK_ID" != "null" ]]; then
            register_cleanup "job-production-plan-time-card-approvers" "$CREATED_LINK_ID"
            pass
        else
            fail "Created time card approver but no ID returned"
        fi
    else
        fail "Failed to create time card approver"
    fi
else
    skip "Missing job production plan or user; skipping create"
fi

# ============================================================================
# LIST / SHOW Tests
# ============================================================================

test_name "List time card approvers"
xbe_json view job-production-plan-time-card-approvers list --limit 1
assert_success

if [[ -n "$CREATED_LINK_ID" && "$CREATED_LINK_ID" != "null" ]]; then
    test_name "Show time card approver"
    xbe_json view job-production-plan-time-card-approvers show "$CREATED_LINK_ID"
    assert_success
fi

if [[ -n "$CREATED_JPP_ID" && -n "$CREATED_USER_ID" ]]; then
    test_name "List time card approvers with --job-production-plan filter"
    xbe_json view job-production-plan-time-card-approvers list --job-production-plan "$CREATED_JPP_ID"
    assert_success

    test_name "List time card approvers with --user filter"
    xbe_json view job-production-plan-time-card-approvers list --user "$CREATED_USER_ID"
    assert_success
fi

# ============================================================================
# DELETE Tests
# ============================================================================

if [[ -n "$CREATED_LINK_ID" && "$CREATED_LINK_ID" != "null" ]]; then
    test_name "Delete time card approver"
    xbe_json do job-production-plan-time-card-approvers delete "$CREATED_LINK_ID" --confirm
    assert_success
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
