#!/bin/bash
#
# XBE CLI Integration Tests: Job Production Plan Job Site Changes
#
# Tests show and create operations for the job-production-plan-job-site-changes resource.
#
# COVERAGE: Create + show + error cases
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_BROKER_ID=""
CREATED_CUSTOMER_ID=""
CREATED_OLD_JOB_SITE_ID=""
CREATED_NEW_JOB_SITE_ID=""
CREATED_JPP_ID=""
CREATED_CHANGE_ID=""

describe "Resource: job-production-plan-job-site-changes"

# ============================================================================
# Prerequisites - Create broker, customer, job sites, and job production plan
# ============================================================================

test_name "Create prerequisite broker"
BROKER_NAME=$(unique_name "JPPJobSiteChangeBroker")

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

test_name "Create prerequisite customer"
CUSTOMER_NAME=$(unique_name "JPPJobSiteChangeCustomer")

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

test_name "Create old job site"
OLD_JOB_SITE_NAME=$(unique_name "OldJobSite")
OLD_JOB_SITE_ADDRESS="1001 Test Ave, Chicago, IL 60601"

xbe_json do job-sites create \
    --name "$OLD_JOB_SITE_NAME" \
    --customer "$CREATED_CUSTOMER_ID" \
    --address "$OLD_JOB_SITE_ADDRESS"

if [[ $status -eq 0 ]]; then
    CREATED_OLD_JOB_SITE_ID=$(json_get ".id")
    if [[ -n "$CREATED_OLD_JOB_SITE_ID" && "$CREATED_OLD_JOB_SITE_ID" != "null" ]]; then
        register_cleanup "job-sites" "$CREATED_OLD_JOB_SITE_ID"
        pass
    else
        fail "Created job site but no ID returned"
        echo "Cannot continue without a job site"
        run_tests
    fi
else
    fail "Failed to create old job site"
    echo "Cannot continue without a job site"
    run_tests
fi

test_name "Create new job site"
NEW_JOB_SITE_NAME=$(unique_name "NewJobSite")
NEW_JOB_SITE_ADDRESS="1002 Test Ave, Chicago, IL 60601"

xbe_json do job-sites create \
    --name "$NEW_JOB_SITE_NAME" \
    --customer "$CREATED_CUSTOMER_ID" \
    --address "$NEW_JOB_SITE_ADDRESS"

if [[ $status -eq 0 ]]; then
    CREATED_NEW_JOB_SITE_ID=$(json_get ".id")
    if [[ -n "$CREATED_NEW_JOB_SITE_ID" && "$CREATED_NEW_JOB_SITE_ID" != "null" ]]; then
        register_cleanup "job-sites" "$CREATED_NEW_JOB_SITE_ID"
        pass
    else
        fail "Created job site but no ID returned"
        echo "Cannot continue without a job site"
        run_tests
    fi
else
    fail "Failed to create new job site"
    echo "Cannot continue without a job site"
    run_tests
fi

test_name "Create job production plan for job site change"
TEST_NAME=$(unique_name "JPPJobSiteChange")
TODAY=$(date +%Y-%m-%d)

xbe_json do job-production-plans create \
    --job-name "$TEST_NAME" \
    --start-on "$TODAY" \
    --start-time "07:00" \
    --customer "$CREATED_CUSTOMER_ID" \
    --job-site "$CREATED_OLD_JOB_SITE_ID"

if [[ $status -eq 0 ]]; then
    CREATED_JPP_ID=$(json_get ".id")
    if [[ -n "$CREATED_JPP_ID" && "$CREATED_JPP_ID" != "null" ]]; then
        pass
    else
        fail "Created job production plan but no ID returned"
        echo "Cannot continue without a job production plan"
        run_tests
    fi
else
    fail "Failed to create job production plan"
    echo "Cannot continue without a job production plan"
    run_tests
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create job site change without required fields fails"
xbe_run do job-production-plan-job-site-changes create
assert_failure

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create job site change"
xbe_json do job-production-plan-job-site-changes create \
    --job-production-plan "$CREATED_JPP_ID" \
    --old-job-site "$CREATED_OLD_JOB_SITE_ID" \
    --new-job-site "$CREATED_NEW_JOB_SITE_ID"

if [[ $status -eq 0 ]]; then
    CREATED_CHANGE_ID=$(json_get ".id")
    if [[ -n "$CREATED_CHANGE_ID" && "$CREATED_CHANGE_ID" != "null" ]]; then
        pass
    else
        fail "Created job site change but no ID returned"
    fi
else
    if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]]; then
        pass
    else
        fail "Failed to create job site change: $output"
    fi
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show job site change"
if [[ -n "$CREATED_CHANGE_ID" && "$CREATED_CHANGE_ID" != "null" ]]; then
    xbe_json view job-production-plan-job-site-changes show "$CREATED_CHANGE_ID"
    assert_success
else
    skip "No job site change ID available"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
