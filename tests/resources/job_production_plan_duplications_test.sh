#!/bin/bash
#
# XBE CLI Integration Tests: Job Production Plan Duplications
#
# Tests create operations for job production plan duplications.
#
# COVERAGE: Create required fields, optional flags, and error cases.
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_BROKER_ID=""
CREATED_CUSTOMER_ID=""
CREATED_SECOND_CUSTOMER_ID=""
CREATED_TEMPLATE_ID=""
CREATED_DUPLICATION_ID=""
CREATED_DERIVED_PLAN_ID=""

describe "Resource: job-production-plan-duplications"

# ============================================================================
# Prerequisites - Create broker, customers, and template plan
# ============================================================================

test_name "Create prerequisite broker for duplication tests"
BROKER_NAME=$(unique_name "JPPDupBroker")

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
CUSTOMER_NAME=$(unique_name "JPPDupCustomer")

xbe_json do customers create \
    --name "$CUSTOMER_NAME" \
    --broker "$CREATED_BROKER_ID" \
    --can-manage-crew-requirements true

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

test_name "Create second customer for cross-customer duplication"
SECOND_CUSTOMER_NAME=$(unique_name "JPPDupCustomer2")

xbe_json do customers create \
    --name "$SECOND_CUSTOMER_NAME" \
    --broker "$CREATED_BROKER_ID" \
    --can-manage-crew-requirements true

if [[ $status -eq 0 ]]; then
    CREATED_SECOND_CUSTOMER_ID=$(json_get ".id")
    if [[ -n "$CREATED_SECOND_CUSTOMER_ID" && "$CREATED_SECOND_CUSTOMER_ID" != "null" ]]; then
        register_cleanup "customers" "$CREATED_SECOND_CUSTOMER_ID"
        pass
    else
        fail "Created customer but no ID returned"
        echo "Cannot continue without a second customer"
        run_tests
    fi
else
    fail "Failed to create second customer"
    echo "Cannot continue without a second customer"
    run_tests
fi

test_name "Create job production plan template"
TEMPLATE_NAME=$(unique_name "JPPDupTemplate")
TODAY=$(date +%Y-%m-%d)

xbe_json do job-production-plans create \
    --job-name "$TEMPLATE_NAME" \
    --start-on "$TODAY" \
    --start-time "07:00" \
    --customer "$CREATED_CUSTOMER_ID" \
    --is-template \
    --template-name "Duplication Template"

if [[ $status -eq 0 ]]; then
    CREATED_TEMPLATE_ID=$(json_get ".id")
    if [[ -n "$CREATED_TEMPLATE_ID" && "$CREATED_TEMPLATE_ID" != "null" ]]; then
        pass
    else
        fail "Created template but no ID returned"
        echo "Cannot continue without a template ID"
        run_tests
    fi
else
    fail "Failed to create job production plan template"
    echo "Cannot continue without a template"
    run_tests
fi

# ============================================================================
# CREATE Tests - Required
# ============================================================================

test_name "Create duplication with required fields"
xbe_json do job-production-plan-duplications create \
    --job-production-plan-template "$CREATED_TEMPLATE_ID" \
    --start-on "$TODAY"

if [[ $status -eq 0 ]]; then
    CREATED_DUPLICATION_ID=$(json_get ".id")
    CREATED_DERIVED_PLAN_ID=$(json_get ".derived_job_production_plan_id")
    if [[ -n "$CREATED_DUPLICATION_ID" && "$CREATED_DUPLICATION_ID" != "null" ]]; then
        pass
    else
        fail "Created duplication but no ID returned"
    fi
else
    fail "Failed to create job production plan duplication"
fi

# ============================================================================
# CREATE Tests - Optional Flags
# ============================================================================

test_name "Create duplication with optional flags"
xbe_json do job-production-plan-duplications create \
    --job-production-plan-template "$CREATED_TEMPLATE_ID" \
    --start-on "$TODAY" \
    --new-customer "$CREATED_SECOND_CUSTOMER_ID" \
    --derived-job-production-plan-template-name "Dup Template $(date +%s)" \
    --is-async \
    --skip-template-duplication-validation \
    --disable-overlapping-crew-requirement-is-validating-overlapping \
    --skip-job-production-plan-alarms \
    --skip-job-production-plan-locations \
    --skip-job-production-plan-safety-risks \
    --skip-job-production-plan-material-sites \
    --skip-equipment-requirements \
    --skip-labor-requirements \
    --skip-job-production-plan-material-types \
    --skip-job-production-plan-service-type-unit-of-measures \
    --skip-job-production-plan-display-unit-of-measures \
    --skip-job-production-plan-service-type-unit-of-measure-cohorts \
    --skip-job-production-plan-trailer-classifications \
    --skip-job-production-plan-segment-sets \
    --skip-job-production-plan-segments \
    --skip-job-production-plan-subscriptions \
    --skip-job-production-plan-time-card-approvers \
    --skip-job-schedule-shifts \
    --skip-developer-references \
    --skip-job-production-plan-inspectors \
    --skip-job-production-plan-project-phase-revenue-items \
    --skip-equipment-requirements-resource \
    --skip-labor-requirements-resource \
    --skip-labor-requirements-craft-class \
    --skip-job-schedule-shifts-driver-assignment-rule-text-cached \
    --skip-labor-requirements-overlapping-resource \
    --skipped-labor-requirements-overlapping-resource-ids '[[1,2]]' \
    --skipped-labor-requirements-not-valid-to-assign-ids '[[3,4]]'
assert_success

# ============================================================================
# CREATE Tests - Error Cases
# ============================================================================

test_name "Create duplication without template fails"
xbe_json do job-production-plan-duplications create --start-on "$TODAY"
assert_failure

run_tests
