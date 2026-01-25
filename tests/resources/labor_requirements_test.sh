#!/bin/bash
#
# XBE CLI Integration Tests: Labor Requirements
#
# Tests CRUD operations for the labor-requirements resource.
# Labor requirements define labor needs for job production plans.
#
# COVERAGE: Create/update attributes + list filters + delete
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_LABOR_REQUIREMENT_ID=""
CREATED_DELETE_LABOR_REQUIREMENT_ID=""
CREATED_BROKER_ID=""
CREATED_CUSTOMER_ID=""
CREATED_JOB_PRODUCTION_PLAN_ID=""
CREATED_LABOR_CLASSIFICATION_ID=""
CREATED_LABOR_CLASSIFICATION_ID_2=""
CREATED_USER_ID=""
CREATED_MEMBERSHIP_ID=""
CREATED_LABORER_ID=""
CREATED_USER_ID_2=""
CREATED_MEMBERSHIP_ID_2=""
CREATED_LABORER_ID_2=""
CREATED_CRAFT_ID=""
CREATED_CRAFT_CLASS_ID=""
CREATED_PROJECT_COST_CLASSIFICATION_ID=""

describe "Resource: labor-requirements"

# ============================================================================
# Prerequisites - Broker, customer, job production plan, classifications
# ============================================================================

test_name "Create prerequisite broker for labor requirements tests"
BROKER_NAME=$(unique_name "LaborReqTestBroker")

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
CUSTOMER_NAME=$(unique_name "LaborReqTestCustomer")
TODAY=$(date +%Y-%m-%d)

xbe_json do customers create \
    --name "$CUSTOMER_NAME" \
    --broker "$CREATED_BROKER_ID" \
    --can-manage-crew-requirements true \
    --is-expecting-crew-requirement-time-sheets true \
    --expecting-crew-requirement-time-sheets-on "$TODAY"

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

test_name "Create prerequisite job production plan"
JOB_NAME=$(unique_name "LaborReqJPP")

xbe_json do job-production-plans create \
    --job-name "$JOB_NAME" \
    --start-on "$TODAY" \
    --start-time "07:00" \
    --customer "$CREATED_CUSTOMER_ID" \
    --is-managing-crew-requirements

if [[ $status -eq 0 ]]; then
    CREATED_JOB_PRODUCTION_PLAN_ID=$(json_get ".id")
    if [[ -n "$CREATED_JOB_PRODUCTION_PLAN_ID" && "$CREATED_JOB_PRODUCTION_PLAN_ID" != "null" ]]; then
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

test_name "Create prerequisite labor classification"
LC_NAME=$(unique_name "LaborReqClass")
LC_ABBR="LR$(date +%s | tail -c 4)"

xbe_json do labor-classifications create \
    --name "$LC_NAME" \
    --abbreviation "$LC_ABBR"

if [[ $status -eq 0 ]]; then
    CREATED_LABOR_CLASSIFICATION_ID=$(json_get ".id")
    if [[ -n "$CREATED_LABOR_CLASSIFICATION_ID" && "$CREATED_LABOR_CLASSIFICATION_ID" != "null" ]]; then
        register_cleanup "labor-classifications" "$CREATED_LABOR_CLASSIFICATION_ID"
        pass
    else
        fail "Created labor classification but no ID returned"
        echo "Cannot continue without a labor classification"
        run_tests
    fi
else
    fail "Failed to create labor classification"
    echo "Cannot continue without a labor classification"
    run_tests
fi

test_name "Create second labor classification for updates"
LC_NAME_2=$(unique_name "LaborReqClass2")
LC_ABBR_2="L2$(date +%s | tail -c 4)"

xbe_json do labor-classifications create \
    --name "$LC_NAME_2" \
    --abbreviation "$LC_ABBR_2"

if [[ $status -eq 0 ]]; then
    CREATED_LABOR_CLASSIFICATION_ID_2=$(json_get ".id")
    if [[ -n "$CREATED_LABOR_CLASSIFICATION_ID_2" && "$CREATED_LABOR_CLASSIFICATION_ID_2" != "null" ]]; then
        register_cleanup "labor-classifications" "$CREATED_LABOR_CLASSIFICATION_ID_2"
        pass
    else
        fail "Created labor classification but no ID returned"
        echo "Cannot continue without a labor classification"
        run_tests
    fi
else
    fail "Failed to create labor classification"
    echo "Cannot continue without a labor classification"
    run_tests
fi

# ============================================================================
# Prerequisites - User, membership, laborer
# ============================================================================

test_name "Create prerequisite user"
USER_EMAIL=$(unique_email)
USER_NAME=$(unique_name "LaborReqUser")

xbe_json do users create \
    --email "$USER_EMAIL" \
    --name "$USER_NAME"

if [[ $status -eq 0 ]]; then
    CREATED_USER_ID=$(json_get ".id")
    if [[ -n "$CREATED_USER_ID" && "$CREATED_USER_ID" != "null" ]]; then
        pass
    else
        fail "Created user but no ID returned"
        echo "Cannot continue without a user"
        run_tests
    fi
else
    fail "Failed to create user"
    echo "Cannot continue without a user"
    run_tests
fi

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
        echo "Cannot continue without a membership"
        run_tests
    fi
else
    fail "Failed to create membership"
    echo "Cannot continue without a membership"
    run_tests
fi

test_name "Create prerequisite laborer"
xbe_json do laborers create \
    --labor-classification "$CREATED_LABOR_CLASSIFICATION_ID" \
    --user "$CREATED_USER_ID" \
    --organization-type "customers" \
    --organization-id "$CREATED_CUSTOMER_ID"

if [[ $status -eq 0 ]]; then
    CREATED_LABORER_ID=$(json_get ".id")
    if [[ -n "$CREATED_LABORER_ID" && "$CREATED_LABORER_ID" != "null" ]]; then
        register_cleanup "laborers" "$CREATED_LABORER_ID"
        pass
    else
        fail "Created laborer but no ID returned"
        echo "Cannot continue without a laborer"
        run_tests
    fi
else
    fail "Failed to create laborer"
    echo "Cannot continue without a laborer"
    run_tests
fi

test_name "Create second user for laborer assignment"
USER_EMAIL_2=$(unique_email)
USER_NAME_2=$(unique_name "LaborReqUser2")

xbe_json do users create \
    --email "$USER_EMAIL_2" \
    --name "$USER_NAME_2"

if [[ $status -eq 0 ]]; then
    CREATED_USER_ID_2=$(json_get ".id")
    if [[ -n "$CREATED_USER_ID_2" && "$CREATED_USER_ID_2" != "null" ]]; then
        pass
    else
        fail "Created user but no ID returned"
        echo "Cannot continue without a user"
        run_tests
    fi
else
    fail "Failed to create user"
    echo "Cannot continue without a user"
    run_tests
fi

test_name "Create membership for second user to customer"
xbe_json do memberships create \
    --user "$CREATED_USER_ID_2" \
    --organization "Customer|$CREATED_CUSTOMER_ID"

if [[ $status -eq 0 ]]; then
    CREATED_MEMBERSHIP_ID_2=$(json_get ".id")
    if [[ -n "$CREATED_MEMBERSHIP_ID_2" && "$CREATED_MEMBERSHIP_ID_2" != "null" ]]; then
        register_cleanup "memberships" "$CREATED_MEMBERSHIP_ID_2"
        pass
    else
        fail "Created membership but no ID returned"
        echo "Cannot continue without a membership"
        run_tests
    fi
else
    fail "Failed to create membership"
    echo "Cannot continue without a membership"
    run_tests
fi

test_name "Create second laborer with updated classification"
xbe_json do laborers create \
    --labor-classification "$CREATED_LABOR_CLASSIFICATION_ID_2" \
    --user "$CREATED_USER_ID_2" \
    --organization-type "customers" \
    --organization-id "$CREATED_CUSTOMER_ID"

if [[ $status -eq 0 ]]; then
    CREATED_LABORER_ID_2=$(json_get ".id")
    if [[ -n "$CREATED_LABORER_ID_2" && "$CREATED_LABORER_ID_2" != "null" ]]; then
        register_cleanup "laborers" "$CREATED_LABORER_ID_2"
        pass
    else
        fail "Created laborer but no ID returned"
        echo "Cannot continue without a laborer"
        run_tests
    fi
else
    fail "Failed to create laborer"
    echo "Cannot continue without a laborer"
    run_tests
fi

# ============================================================================
# Prerequisites - Craft, craft class, project cost classification
# ============================================================================

test_name "Create prerequisite craft"
CRAFT_NAME=$(unique_name "LaborReqCraft")

xbe_json do crafts create \
    --name "$CRAFT_NAME" \
    --broker "$CREATED_BROKER_ID"

if [[ $status -eq 0 ]]; then
    CREATED_CRAFT_ID=$(json_get ".id")
    if [[ -n "$CREATED_CRAFT_ID" && "$CREATED_CRAFT_ID" != "null" ]]; then
        register_cleanup "crafts" "$CREATED_CRAFT_ID"
        pass
    else
        fail "Created craft but no ID returned"
        echo "Cannot continue without a craft"
        run_tests
    fi
else
    fail "Failed to create craft"
    echo "Cannot continue without a craft"
    run_tests
fi

test_name "Create prerequisite craft class"
CRAFT_CLASS_NAME=$(unique_name "LaborReqCraftClass")

xbe_json do craft-classes create \
    --name "$CRAFT_CLASS_NAME" \
    --craft "$CREATED_CRAFT_ID"

if [[ $status -eq 0 ]]; then
    CREATED_CRAFT_CLASS_ID=$(json_get ".id")
    if [[ -n "$CREATED_CRAFT_CLASS_ID" && "$CREATED_CRAFT_CLASS_ID" != "null" ]]; then
        register_cleanup "craft-classes" "$CREATED_CRAFT_CLASS_ID"
        pass
    else
        fail "Created craft class but no ID returned"
        echo "Cannot continue without a craft class"
        run_tests
    fi
else
    fail "Failed to create craft class"
    echo "Cannot continue without a craft class"
    run_tests
fi

test_name "Create prerequisite project cost classification"
PCC_NAME=$(unique_name "LaborReqCostClass")

xbe_json do project-cost-classifications create \
    --name "$PCC_NAME" \
    --broker "$CREATED_BROKER_ID"

if [[ $status -eq 0 ]]; then
    CREATED_PROJECT_COST_CLASSIFICATION_ID=$(json_get ".id")
    if [[ -n "$CREATED_PROJECT_COST_CLASSIFICATION_ID" && "$CREATED_PROJECT_COST_CLASSIFICATION_ID" != "null" ]]; then
        register_cleanup "project-cost-classifications" "$CREATED_PROJECT_COST_CLASSIFICATION_ID"
        pass
    else
        fail "Created project cost classification but no ID returned"
        echo "Cannot continue without a project cost classification"
        run_tests
    fi
else
    fail "Failed to create project cost classification"
    echo "Cannot continue without a project cost classification"
    run_tests
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create labor requirement with required fields and attributes"
START_AT="${TODAY}T08:00:00Z"
END_AT="${TODAY}T12:00:00Z"

xbe_json do labor-requirements create \
    --job-production-plan "$CREATED_JOB_PRODUCTION_PLAN_ID" \
    --labor-classification "$CREATED_LABOR_CLASSIFICATION_ID" \
    --start-at "$START_AT" \
    --end-at "$END_AT" \
    --mobilization-method "crew" \
    --note "Initial labor requirement" \
    --is-validating-overlapping \
    --explicit-inbound-latitude "41.881" \
    --explicit-inbound-longitude "-87.623" \
    --explicit-outbound-latitude "41.882" \
    --explicit-outbound-longitude "-87.624"

if [[ $status -eq 0 ]]; then
    CREATED_LABOR_REQUIREMENT_ID=$(json_get ".id")
    if [[ -n "$CREATED_LABOR_REQUIREMENT_ID" && "$CREATED_LABOR_REQUIREMENT_ID" != "null" ]]; then
        register_cleanup "labor-requirements" "$CREATED_LABOR_REQUIREMENT_ID"
        pass
    else
        fail "Created labor requirement but no ID returned"
        echo "Cannot continue without a labor requirement"
        run_tests
    fi
else
    fail "Failed to create labor requirement"
    echo "Cannot continue without a labor requirement"
    run_tests
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update labor requirement assignments and attributes"
START_AT_2="${TODAY}T09:00:00Z"
END_AT_2="${TODAY}T13:00:00Z"

xbe_json do labor-requirements update "$CREATED_LABOR_REQUIREMENT_ID" \
    --laborer "$CREATED_LABORER_ID_2" \
    --craft-class "$CREATED_CRAFT_CLASS_ID" \
    --project-cost-classification "$CREATED_PROJECT_COST_CLASSIFICATION_ID" \
    --labor-classification "$CREATED_LABOR_CLASSIFICATION_ID_2" \
    --start-at "$START_AT_2" \
    --end-at "$END_AT_2" \
    --mobilization-method "itself" \
    --note "Updated labor requirement" \
    --is-validating-overlapping=false \
    --explicit-inbound-latitude "41.883" \
    --explicit-inbound-longitude "-87.625" \
    --explicit-outbound-latitude "41.884" \
    --explicit-outbound-longitude "-87.626"
assert_success

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show labor requirement"
xbe_json view labor-requirements show "$CREATED_LABOR_REQUIREMENT_ID"
if [[ $status -eq 0 ]]; then
    assert_json_has ".id"
else
    fail "Failed to show labor requirement"
fi

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List labor requirements"
xbe_json view labor-requirements list --limit 5
assert_success

test_name "List labor requirements returns array"
xbe_json view labor-requirements list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list labor requirements"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List labor requirements with --job-production-plan filter"
xbe_json view labor-requirements list --job-production-plan "$CREATED_JOB_PRODUCTION_PLAN_ID" --limit 5
assert_success

test_name "List labor requirements with resource classification filters"
xbe_json view labor-requirements list \
    --resource-classification-type LaborClassification \
    --resource-classification-id "$CREATED_LABOR_CLASSIFICATION_ID_2" \
    --limit 5
assert_success

test_name "List labor requirements with --not-resource-classification-type filter"
xbe_json view labor-requirements list --not-resource-classification-type EquipmentClassification --limit 5
if [[ $status -eq 0 ]]; then
    pass
else
    skip "Server does not support not-resource-classification-type filter"
fi

test_name "List labor requirements with resource filters"
xbe_json view labor-requirements list \
    --resource-type Laborer \
    --resource-id "$CREATED_LABORER_ID_2" \
    --limit 5
assert_success

test_name "List labor requirements with --not-resource-type filter"
xbe_json view labor-requirements list --not-resource-type Equipment --limit 5
if [[ $status -eq 0 ]]; then
    pass
else
    skip "Server does not support not-resource-type filter"
fi

test_name "List labor requirements with --broker filter"
xbe_json view labor-requirements list --broker "$CREATED_BROKER_ID" --limit 5
assert_success

test_name "List labor requirements with --customer filter"
xbe_json view labor-requirements list --customer "$CREATED_CUSTOMER_ID" --limit 5
assert_success

test_name "List labor requirements with --project-manager filter"
xbe_json view labor-requirements list --project-manager "$CREATED_USER_ID" --limit 5
assert_success

test_name "List labor requirements with --project filter"
xbe_json view labor-requirements list --project 999999 --limit 5
assert_success

test_name "List labor requirements with --has-resource filter"
xbe_json view labor-requirements list --has-resource true --limit 5
assert_success

test_name "List labor requirements with start/end filters"
xbe_json view labor-requirements list --start-at-min "$START_AT_2" --start-at-max "$END_AT_2" --limit 5
assert_success
xbe_json view labor-requirements list --end-at-min "$START_AT_2" --end-at-max "$END_AT_2" --limit 5
assert_success

test_name "List labor requirements with is-start-at/is-end-at filters"
xbe_json view labor-requirements list --is-start-at true --limit 5
assert_success
xbe_json view labor-requirements list --is-end-at true --limit 5
assert_success

test_name "List labor requirements with effective time filters"
xbe_json view labor-requirements list --start-at-effective-min "$START_AT_2" --start-at-effective-max "$END_AT_2" --limit 5
assert_success
xbe_json view labor-requirements list --end-at-effective-min "$START_AT_2" --end-at-effective-max "$END_AT_2" --limit 5
assert_success

test_name "List labor requirements with start-on-effective filters"
xbe_json view labor-requirements list --start-on-effective-min "$TODAY" --start-on-effective-max "$TODAY" --limit 5
assert_success

test_name "List labor requirements with --calculated-mobilization-method filter"
xbe_json view labor-requirements list --calculated-mobilization-method itself --limit 5
assert_success

test_name "List labor requirements with --job-production-plan-status filter"
xbe_json view labor-requirements list --job-production-plan-status editing --limit 5
assert_success

test_name "List labor requirements with labor requirement filters"
xbe_json view labor-requirements list --labor-requirement "$CREATED_LABOR_REQUIREMENT_ID" --limit 5
assert_success
xbe_json view labor-requirements list --labor-requirement-laborer "$CREATED_LABORER_ID_2" --limit 5
assert_success
xbe_json view labor-requirements list --labor-requirement-laborer-id "$CREATED_LABORER_ID_2" --limit 5
assert_success
xbe_json view labor-requirements list --labor-requirement-user "$CREATED_USER_ID_2" --limit 5
assert_success
xbe_json view labor-requirements list --labor-requirement-user-id "$CREATED_USER_ID_2" --limit 5
assert_success

test_name "List labor requirements with movement filters"
xbe_json view labor-requirements list --requires-inbound-movement false --limit 5
assert_success
xbe_json view labor-requirements list --requires-outbound-movement false --limit 5
assert_success

test_name "List labor requirements with time sheet filters"
xbe_json view labor-requirements list --without-approved-time-sheet true --limit 5
assert_success
xbe_json view labor-requirements list --without-submitted-time-sheet true --limit 5
assert_success
xbe_json view labor-requirements list --is-expecting-time-sheet true --limit 5
assert_success

test_name "List labor requirements with laborer user filter"
xbe_json view labor-requirements list --laborer-user "$CREATED_USER_ID_2" --limit 5
assert_success

test_name "List labor requirements with assignment candidate filter"
xbe_json view labor-requirements list --is-assignment-candidate-for "$CREATED_LABORER_ID_2" --limit 5
assert_success

test_name "List labor requirements with --is-only-for-equipment-movement filter"
xbe_json view labor-requirements list --is-only-for-equipment-movement false --limit 5
assert_success

# ============================================================================
# LIST Tests - Pagination
# ============================================================================

test_name "List labor requirements with --limit"
xbe_json view labor-requirements list --limit 3
assert_success

test_name "List labor requirements with --offset"
xbe_json view labor-requirements list --limit 3 --offset 3
assert_success

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Create labor requirement for deletion"
xbe_json do labor-requirements create \
    --job-production-plan "$CREATED_JOB_PRODUCTION_PLAN_ID" \
    --labor-classification "$CREATED_LABOR_CLASSIFICATION_ID"

if [[ $status -eq 0 ]]; then
    CREATED_DELETE_LABOR_REQUIREMENT_ID=$(json_get ".id")
    if [[ -n "$CREATED_DELETE_LABOR_REQUIREMENT_ID" && "$CREATED_DELETE_LABOR_REQUIREMENT_ID" != "null" ]]; then
        pass
    else
        fail "Created labor requirement but no ID returned"
    fi
else
    fail "Failed to create labor requirement for deletion"
fi

test_name "Delete labor requirement requires --confirm flag"
xbe_run do labor-requirements delete "$CREATED_DELETE_LABOR_REQUIREMENT_ID"
assert_failure

test_name "Delete labor requirement with --confirm"
xbe_run do labor-requirements delete "$CREATED_DELETE_LABOR_REQUIREMENT_ID" --confirm
assert_success

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create labor requirement without --job-production-plan fails"
xbe_json do labor-requirements create --labor-classification "$CREATED_LABOR_CLASSIFICATION_ID"
assert_failure

test_name "Create labor requirement without --labor-classification fails"
xbe_json do labor-requirements create --job-production-plan "$CREATED_JOB_PRODUCTION_PLAN_ID"
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
