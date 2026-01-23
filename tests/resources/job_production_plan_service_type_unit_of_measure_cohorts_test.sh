#!/bin/bash
#
# XBE CLI Integration Tests: Job Production Plan Service Type Unit Of Measure Cohorts
#
# Tests view and create/delete operations for the job_production_plan_service_type_unit_of_measure_cohorts resource.
# These links attach a service type unit of measure cohort to a job production plan.
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
CREATED_COHORT_ID=""
TRIGGER_STUOM_ID=""
MEMBER_STUOM_ID=""
SKIP_MUTATION=0

TODAY=$(date +%Y-%m-%d)

describe "Resource: job-production-plan-service-type-unit-of-measure-cohorts"

# ============================================================================
# Setup prerequisites (requires XBE_TOKEN for direct API calls)
# ============================================================================

if [[ -z "$XBE_TOKEN" ]]; then
    echo "    (XBE_TOKEN not set; skipping cohort setup and mutation tests)"
    SKIP_MUTATION=1
else
    test_name "Create prerequisite broker"
    BROKER_NAME=$(unique_name "JPPSTUOMCBroker")

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
        CUSTOMER_NAME=$(unique_name "JPPSTUOMCCustomer")

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
        JOB_SITE_NAME=$(unique_name "JPPSTUOMCSite")

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

    if [[ -n "$CREATED_CUSTOMER_ID" && "$CREATED_CUSTOMER_ID" != "null" && -n "$CREATED_JOB_SITE_ID" && "$CREATED_JOB_SITE_ID" != "null" ]]; then
        test_name "Create prerequisite job production plan"
        PLAN_NAME=$(unique_name "JPPSTUOMCPlan")

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

    test_name "Fetch service type unit of measure IDs"
    run curl -s -f \
        -H "Authorization: Bearer $XBE_TOKEN" \
        -H "Accept: application/vnd.api+json" \
        "$XBE_BASE_URL/v1/service-type-unit-of-measures?page[limit]=2"

    if [[ $status -eq 0 ]]; then
        TRIGGER_STUOM_ID=$(echo "$output" | jq -r '.data[0].id // empty')
        MEMBER_STUOM_ID=$(echo "$output" | jq -r '.data[1].id // empty')
        if [[ -n "$TRIGGER_STUOM_ID" && -n "$MEMBER_STUOM_ID" && "$TRIGGER_STUOM_ID" != "$MEMBER_STUOM_ID" ]]; then
            pass
        else
            fail "Not enough service type unit of measure IDs returned"
        fi
    else
        fail "Failed to fetch service type unit of measure IDs"
    fi

    if [[ -n "$CREATED_CUSTOMER_ID" && -n "$TRIGGER_STUOM_ID" && -n "$MEMBER_STUOM_ID" ]]; then
        test_name "Create prerequisite service type unit of measure cohort"
        COHORT_NAME=$(unique_name "JPPSTUOMCCohort")
        cohort_payload=$(cat <<JSON
{"data":{"type":"service-type-unit-of-measure-cohorts","attributes":{"name":"$COHORT_NAME","service-type-unit-of-measure-ids":["$MEMBER_STUOM_ID"],"is-active":true},"relationships":{"customer":{"data":{"type":"customers","id":"$CREATED_CUSTOMER_ID"}},"trigger":{"data":{"type":"service-type-unit-of-measures","id":"$TRIGGER_STUOM_ID"}}}}}
JSON
        )

        run curl -s -f \
            -H "Authorization: Bearer $XBE_TOKEN" \
            -H "Accept: application/vnd.api+json" \
            -H "Content-Type: application/vnd.api+json" \
            -X POST "$XBE_BASE_URL/v1/service-type-unit-of-measure-cohorts" \
            -d "$cohort_payload"

        if [[ $status -eq 0 ]]; then
            CREATED_COHORT_ID=$(echo "$output" | jq -r '.data.id // empty')
            if [[ -n "$CREATED_COHORT_ID" && "$CREATED_COHORT_ID" != "null" ]]; then
                pass
            else
                fail "Created cohort but no ID returned"
            fi
        else
            fail "Failed to create service type unit of measure cohort"
        fi
    fi
fi

# Custom cleanup for cohort (no CLI command available)
cleanup_cohort() {
    if [[ -n "$CREATED_COHORT_ID" && "$CREATED_COHORT_ID" != "null" && -n "$XBE_TOKEN" ]]; then
        curl -s -f \
            -H "Authorization: Bearer $XBE_TOKEN" \
            -H "Accept: application/vnd.api+json" \
            -X DELETE "$XBE_BASE_URL/v1/service-type-unit-of-measure-cohorts/$CREATED_COHORT_ID" \
            >/dev/null 2>&1 || true
    fi
    run_cleanup
}
trap cleanup_cohort EXIT

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create cohort link without required fields fails"
xbe_json do job-production-plan-service-type-unit-of-measure-cohorts create
assert_failure

if [[ $SKIP_MUTATION -eq 1 ]]; then
    skip "Skipping create/delete/filter tests without XBE_TOKEN"
fi

if [[ -n "$CREATED_JPP_ID" && -n "$CREATED_COHORT_ID" ]]; then
    test_name "Create job production plan service type unit of measure cohort link"
    xbe_json do job-production-plan-service-type-unit-of-measure-cohorts create \
        --job-production-plan "$CREATED_JPP_ID" \
        --service-type-unit-of-measure-cohort "$CREATED_COHORT_ID"

    if [[ $status -eq 0 ]]; then
        CREATED_LINK_ID=$(json_get ".id")
        if [[ -n "$CREATED_LINK_ID" && "$CREATED_LINK_ID" != "null" ]]; then
            register_cleanup "job-production-plan-service-type-unit-of-measure-cohorts" "$CREATED_LINK_ID"
            pass
        else
            fail "Created link but no ID returned"
        fi
    else
        fail "Failed to create cohort link"
    fi
else
    skip "Missing job production plan or cohort; skipping create"
fi

# ============================================================================
# LIST / SHOW Tests
# ============================================================================

test_name "List cohort links"
xbe_json view job-production-plan-service-type-unit-of-measure-cohorts list --limit 1
assert_success

if [[ -n "$CREATED_LINK_ID" && "$CREATED_LINK_ID" != "null" ]]; then
    test_name "Show cohort link"
    xbe_json view job-production-plan-service-type-unit-of-measure-cohorts show "$CREATED_LINK_ID"
    assert_success
fi

if [[ -n "$CREATED_JPP_ID" && -n "$CREATED_COHORT_ID" ]]; then
    test_name "List cohort links with --job-production-plan filter"
    xbe_json view job-production-plan-service-type-unit-of-measure-cohorts list --job-production-plan "$CREATED_JPP_ID"
    assert_success

    test_name "List cohort links with --service-type-unit-of-measure-cohort filter"
    xbe_json view job-production-plan-service-type-unit-of-measure-cohorts list --service-type-unit-of-measure-cohort "$CREATED_COHORT_ID"
    assert_success
fi

# ============================================================================
# DELETE Tests
# ============================================================================

if [[ -n "$CREATED_LINK_ID" && "$CREATED_LINK_ID" != "null" ]]; then
    test_name "Delete cohort link"
    xbe_json do job-production-plan-service-type-unit-of-measure-cohorts delete "$CREATED_LINK_ID" --confirm
    assert_success
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
