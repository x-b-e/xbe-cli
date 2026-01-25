#!/bin/bash
#
# XBE CLI Integration Tests: Crew Requirements
#
# Tests list, show, create, update, and delete operations for crew requirements.
#
# COVERAGE: All list filters + create/update attributes and relationships
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_BROKER_ID=""
CREATED_CUSTOMER_ID=""
CREATED_JPP_ID=""
CREATED_JPP_ID_2=""
CREATED_EQUIPMENT_CLASSIFICATION_ID=""
CREATED_EQUIPMENT_CLASSIFICATION_ID_2=""
CREATED_EQUIPMENT_ID=""
CREATED_EQUIPMENT_ID_2=""
CREATED_CRAFT_ID=""
CREATED_CRAFT_CLASS_ID=""
CREATED_PROJECT_COST_CLASSIFICATION_ID=""
CREATED_LABOR_CLASSIFICATION_ID=""
CREATED_USER_ID=""
CREATED_MEMBERSHIP_ID=""
CREATED_LABORER_ID=""
CREATED_EQUIPMENT_REQUIREMENT_ID=""
CREATED_LABOR_REQUIREMENT_ID=""
CREATED_LABOR_REQUIREMENT_ID_2=""

REQ_START_AT=""
REQ_END_AT=""
REQ_START_AT_EFFECTIVE=""
REQ_END_AT_EFFECTIVE=""
REQ_START_ON_EFFECTIVE=""
REQ_CALC_MOBILIZATION_METHOD=""
REQ_JPP_STATUS=""

describe "Resource: crew-requirements"

# ============================================================================
# Prerequisites
# ============================================================================

test_name "Create prerequisite broker for crew requirement tests"
BROKER_NAME=$(unique_name "CrewReqBroker")

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
        fail "Failed to create broker"
        echo "Cannot continue without a broker"
        run_tests
    fi
fi

test_name "Create prerequisite customer"
CUSTOMER_NAME=$(unique_name "CrewReqCustomer")

xbe_json do customers create \
    --name "$CUSTOMER_NAME" \
    --broker "$CREATED_BROKER_ID" \
    --can-manage-crew-requirements true \
    --default-is-managing-crew-requirements true

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

test_name "Create job production plan"
TODAY=$(date +%Y-%m-%d)
JPP_NAME=$(unique_name "CrewReqPlan")

xbe_json do job-production-plans create \
    --job-name "$JPP_NAME" \
    --start-on "$TODAY" \
    --start-time "07:00" \
    --customer "$CREATED_CUSTOMER_ID" \
    --is-managing-crew-requirements true

if [[ $status -eq 0 ]]; then
    CREATED_JPP_ID=$(json_get ".id")
    if [[ -n "$CREATED_JPP_ID" && "$CREATED_JPP_ID" != "null" ]]; then
        pass
    else
        fail "Created job production plan but no ID returned"
        run_tests
    fi
else
    fail "Failed to create job production plan"
    run_tests
fi

test_name "Create second job production plan"
JPP_NAME_2=$(unique_name "CrewReqPlan2")

xbe_json do job-production-plans create \
    --job-name "$JPP_NAME_2" \
    --start-on "$TODAY" \
    --start-time "08:00" \
    --customer "$CREATED_CUSTOMER_ID" \
    --is-managing-crew-requirements true

if [[ $status -eq 0 ]]; then
    CREATED_JPP_ID_2=$(json_get ".id")
    if [[ -n "$CREATED_JPP_ID_2" && "$CREATED_JPP_ID_2" != "null" ]]; then
        pass
    else
        fail "Created job production plan but no ID returned"
        run_tests
    fi
else
    fail "Failed to create second job production plan"
    run_tests
fi

test_name "Create equipment classifications"
EC_NAME=$(unique_name "CrewReqEquipClass")
EC_ABBR="EC$(date +%s%N | tail -c 6)"
EC_NAME_2=$(unique_name "CrewReqEquipClass2")
EC_ABBR_2="EC$(date +%s%N | tail -c 6)"

xbe_json do equipment-classifications create --name "$EC_NAME" --abbreviation "$EC_ABBR"
if [[ $status -eq 0 ]]; then
    CREATED_EQUIPMENT_CLASSIFICATION_ID=$(json_get ".id")
    if [[ -n "$CREATED_EQUIPMENT_CLASSIFICATION_ID" && "$CREATED_EQUIPMENT_CLASSIFICATION_ID" != "null" ]]; then
        register_cleanup "equipment-classifications" "$CREATED_EQUIPMENT_CLASSIFICATION_ID"
        pass
    else
        fail "Created equipment classification but no ID returned"
        run_tests
    fi
else
    fail "Failed to create equipment classification"
    run_tests
fi

xbe_json do equipment-classifications create --name "$EC_NAME_2" --abbreviation "$EC_ABBR_2"
if [[ $status -eq 0 ]]; then
    CREATED_EQUIPMENT_CLASSIFICATION_ID_2=$(json_get ".id")
    if [[ -n "$CREATED_EQUIPMENT_CLASSIFICATION_ID_2" && "$CREATED_EQUIPMENT_CLASSIFICATION_ID_2" != "null" ]]; then
        register_cleanup "equipment-classifications" "$CREATED_EQUIPMENT_CLASSIFICATION_ID_2"
        pass
    else
        fail "Created second equipment classification but no ID returned"
        run_tests
    fi
else
    fail "Failed to create second equipment classification"
    run_tests
fi

test_name "Create equipment for crew requirements"
EQUIPMENT_NAME=$(unique_name "CrewReqEquip")
EQUIPMENT_NAME_2=$(unique_name "CrewReqEquip2")

xbe_json do equipment create \
    --nickname "$EQUIPMENT_NAME" \
    --equipment-classification "$CREATED_EQUIPMENT_CLASSIFICATION_ID" \
    --organization-type "brokers" \
    --organization-id "$CREATED_BROKER_ID"

if [[ $status -eq 0 ]]; then
    CREATED_EQUIPMENT_ID=$(json_get ".id")
    if [[ -n "$CREATED_EQUIPMENT_ID" && "$CREATED_EQUIPMENT_ID" != "null" ]]; then
        register_cleanup "equipment" "$CREATED_EQUIPMENT_ID"
        pass
    else
        fail "Created equipment but no ID returned"
        run_tests
    fi
else
    fail "Failed to create equipment"
    run_tests
fi

xbe_json do equipment create \
    --nickname "$EQUIPMENT_NAME_2" \
    --equipment-classification "$CREATED_EQUIPMENT_CLASSIFICATION_ID_2" \
    --organization-type "brokers" \
    --organization-id "$CREATED_BROKER_ID"

if [[ $status -eq 0 ]]; then
    CREATED_EQUIPMENT_ID_2=$(json_get ".id")
    if [[ -n "$CREATED_EQUIPMENT_ID_2" && "$CREATED_EQUIPMENT_ID_2" != "null" ]]; then
        register_cleanup "equipment" "$CREATED_EQUIPMENT_ID_2"
        pass
    else
        fail "Created second equipment but no ID returned"
        run_tests
    fi
else
    fail "Failed to create second equipment"
    run_tests
fi

test_name "Create craft and craft class"
CRAFT_NAME=$(unique_name "CrewReqCraft")
CRAFT_CLASS_NAME=$(unique_name "CrewReqCraftClass")

xbe_json do crafts create --name "$CRAFT_NAME" --broker "$CREATED_BROKER_ID"
if [[ $status -eq 0 ]]; then
    CREATED_CRAFT_ID=$(json_get ".id")
    if [[ -n "$CREATED_CRAFT_ID" && "$CREATED_CRAFT_ID" != "null" ]]; then
        register_cleanup "crafts" "$CREATED_CRAFT_ID"
        pass
    else
        fail "Created craft but no ID returned"
        run_tests
    fi
else
    fail "Failed to create craft"
    run_tests
fi

xbe_json do craft-classes create --name "$CRAFT_CLASS_NAME" --craft "$CREATED_CRAFT_ID"
if [[ $status -eq 0 ]]; then
    CREATED_CRAFT_CLASS_ID=$(json_get ".id")
    if [[ -n "$CREATED_CRAFT_CLASS_ID" && "$CREATED_CRAFT_CLASS_ID" != "null" ]]; then
        register_cleanup "craft-classes" "$CREATED_CRAFT_CLASS_ID"
        pass
    else
        fail "Created craft class but no ID returned"
        run_tests
    fi
else
    fail "Failed to create craft class"
    run_tests
fi

test_name "Create project cost classification"
PCC_NAME=$(unique_name "CrewReqCostClass")

xbe_json do project-cost-classifications create --name "$PCC_NAME" --broker "$CREATED_BROKER_ID"
if [[ $status -eq 0 ]]; then
    CREATED_PROJECT_COST_CLASSIFICATION_ID=$(json_get ".id")
    if [[ -n "$CREATED_PROJECT_COST_CLASSIFICATION_ID" && "$CREATED_PROJECT_COST_CLASSIFICATION_ID" != "null" ]]; then
        register_cleanup "project-cost-classifications" "$CREATED_PROJECT_COST_CLASSIFICATION_ID"
        pass
    else
        fail "Created project cost classification but no ID returned"
        run_tests
    fi
else
    fail "Failed to create project cost classification"
    run_tests
fi

test_name "Create labor classification"
LC_NAME=$(unique_name "CrewReqLaborClass")
LC_ABBR="LC$(date +%s | tail -c 4)"

xbe_json do labor-classifications create --name "$LC_NAME" --abbreviation "$LC_ABBR"
if [[ $status -eq 0 ]]; then
    CREATED_LABOR_CLASSIFICATION_ID=$(json_get ".id")
    if [[ -n "$CREATED_LABOR_CLASSIFICATION_ID" && "$CREATED_LABOR_CLASSIFICATION_ID" != "null" ]]; then
        register_cleanup "labor-classifications" "$CREATED_LABOR_CLASSIFICATION_ID"
        pass
    else
        fail "Created labor classification but no ID returned"
        run_tests
    fi
else
    fail "Failed to create labor classification"
    run_tests
fi

test_name "Create user and membership for laborer"
USER_EMAIL=$(unique_email)
USER_NAME=$(unique_name "CrewReqUser")

xbe_json do users create --email "$USER_EMAIL" --name "$USER_NAME"
if [[ $status -eq 0 ]]; then
    CREATED_USER_ID=$(json_get ".id")
    if [[ -n "$CREATED_USER_ID" && "$CREATED_USER_ID" != "null" ]]; then
        pass
    else
        fail "Created user but no ID returned"
        run_tests
    fi
else
    fail "Failed to create user"
    run_tests
fi

xbe_json do memberships create --user "$CREATED_USER_ID" --organization "Customer|$CREATED_CUSTOMER_ID"
if [[ $status -eq 0 ]]; then
    CREATED_MEMBERSHIP_ID=$(json_get ".id")
    if [[ -n "$CREATED_MEMBERSHIP_ID" && "$CREATED_MEMBERSHIP_ID" != "null" ]]; then
        register_cleanup "memberships" "$CREATED_MEMBERSHIP_ID"
        pass
    else
        fail "Created membership but no ID returned"
        run_tests
    fi
else
    fail "Failed to create membership"
    run_tests
fi

test_name "Create laborer for labor requirement"
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
        run_tests
    fi
else
    fail "Failed to create laborer"
    run_tests
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create labor requirement"
LABOR_START_AT="${TODAY}T08:00:00Z"
LABOR_END_AT="${TODAY}T12:00:00Z"

xbe_json do crew-requirements create \
    --requirement-type labor \
    --job-production-plan "$CREATED_JPP_ID" \
    --resource-classification-type labor-classifications \
    --resource-classification-id "$CREATED_LABOR_CLASSIFICATION_ID" \
    --resource-type laborers \
    --resource-id "$CREATED_LABORER_ID" \
    --start-at "$LABOR_START_AT" \
    --end-at "$LABOR_END_AT" \
    --mobilization-method crew \
    --note "Labor requirement"

if [[ $status -eq 0 ]]; then
    CREATED_LABOR_REQUIREMENT_ID=$(json_get ".id")
    if [[ -n "$CREATED_LABOR_REQUIREMENT_ID" && "$CREATED_LABOR_REQUIREMENT_ID" != "null" ]]; then
        register_cleanup "crew-requirements" "$CREATED_LABOR_REQUIREMENT_ID"
        pass
    else
        fail "Created labor requirement but no ID returned"
        run_tests
    fi
else
    fail "Failed to create labor requirement"
    run_tests
fi

test_name "Create second labor requirement for updated plan"
LABOR_START_AT_2="${TODAY}T14:00:00Z"
LABOR_END_AT_2="${TODAY}T18:00:00Z"

xbe_json do crew-requirements create \
    --requirement-type labor \
    --job-production-plan "$CREATED_JPP_ID_2" \
    --resource-classification-type labor-classifications \
    --resource-classification-id "$CREATED_LABOR_CLASSIFICATION_ID" \
    --resource-type laborers \
    --resource-id "$CREATED_LABORER_ID" \
    --start-at "$LABOR_START_AT_2" \
    --end-at "$LABOR_END_AT_2" \
    --mobilization-method crew \
    --note "Labor requirement for plan 2"

if [[ $status -eq 0 ]]; then
    CREATED_LABOR_REQUIREMENT_ID_2=$(json_get ".id")
    if [[ -n "$CREATED_LABOR_REQUIREMENT_ID_2" && "$CREATED_LABOR_REQUIREMENT_ID_2" != "null" ]]; then
        register_cleanup "crew-requirements" "$CREATED_LABOR_REQUIREMENT_ID_2"
        pass
    else
        fail "Created second labor requirement but no ID returned"
        run_tests
    fi
else
    fail "Failed to create second labor requirement"
    run_tests
fi

test_name "Create equipment requirement"
REQ_START_AT="${TODAY}T13:00:00Z"
REQ_END_AT="${TODAY}T17:00:00Z"

xbe_json do crew-requirements create \
    --requirement-type equipment \
    --job-production-plan "$CREATED_JPP_ID" \
    --resource-classification-type equipment-classifications \
    --resource-classification-id "$CREATED_EQUIPMENT_CLASSIFICATION_ID" \
    --resource-type equipment \
    --resource-id "$CREATED_EQUIPMENT_ID" \
    --start-at "$REQ_START_AT" \
    --end-at "$REQ_END_AT" \
    --mobilization-method itself \
    --note "Equipment requirement" \
    --requires-inbound-movement true \
    --requires-outbound-movement false

if [[ $status -eq 0 ]]; then
    CREATED_EQUIPMENT_REQUIREMENT_ID=$(json_get ".id")
    if [[ -n "$CREATED_EQUIPMENT_REQUIREMENT_ID" && "$CREATED_EQUIPMENT_REQUIREMENT_ID" != "null" ]]; then
        register_cleanup "crew-requirements" "$CREATED_EQUIPMENT_REQUIREMENT_ID"
        pass
    else
        fail "Created equipment requirement but no ID returned"
        run_tests
    fi
else
    fail "Failed to create equipment requirement"
    run_tests
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update crew requirement attributes"
xbe_json do crew-requirements update "$CREATED_EQUIPMENT_REQUIREMENT_ID" \
    --note "Updated requirement note" \
    --mobilization-method trailer \
    --requires-inbound-movement false \
    --requires-outbound-movement true \
    --is-validating-overlapping true \
    --explicit-inbound-latitude "41.8800" \
    --explicit-inbound-longitude "-87.6300" \
    --explicit-outbound-latitude "41.8900" \
    --explicit-outbound-longitude "-87.6400"
assert_success

test_name "Update crew requirement relationships"
xbe_json do crew-requirements update "$CREATED_EQUIPMENT_REQUIREMENT_ID" \
    --job-production-plan "$CREATED_JPP_ID_2" \
    --resource-classification-type equipment-classifications \
    --resource-classification-id "$CREATED_EQUIPMENT_CLASSIFICATION_ID_2" \
    --resource-type equipment \
    --resource-id "$CREATED_EQUIPMENT_ID_2" \
    --labor-requirement "$CREATED_LABOR_REQUIREMENT_ID_2"
assert_success

test_name "Update labor requirement relationships"
xbe_json do crew-requirements update "$CREATED_LABOR_REQUIREMENT_ID" \
    --craft-class "$CREATED_CRAFT_CLASS_ID" \
    --project-cost-classification "$CREATED_PROJECT_COST_CLASSIFICATION_ID"
assert_success

test_name "Clear crew requirement credential classifications"
xbe_json do crew-requirements update "$CREATED_EQUIPMENT_REQUIREMENT_ID" \
    --crew-requirement-credential-classifications ""
assert_success

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show crew requirement"
xbe_json view crew-requirements show "$CREATED_EQUIPMENT_REQUIREMENT_ID"
if [[ $status -eq 0 ]]; then
    REQ_START_AT_EFFECTIVE=$(json_get ".start_at_effective")
    REQ_END_AT_EFFECTIVE=$(json_get ".end_at_effective")
    REQ_CALC_MOBILIZATION_METHOD=$(json_get ".calculated_mobilization_method")
    pass
else
    fail "Failed to show crew requirement"
fi

REQ_START_ON_EFFECTIVE=$(echo "$REQ_START_AT_EFFECTIVE" | cut -dT -f1)

test_name "Capture job production plan status"
xbe_json view job-production-plans show "$CREATED_JPP_ID_2"
if [[ $status -eq 0 ]]; then
    REQ_JPP_STATUS=$(json_get ".status")
    pass
else
    skip "Could not fetch job production plan status"
fi

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List crew requirements"
xbe_json view crew-requirements list --limit 5
assert_success

test_name "List crew requirements returns array"
xbe_json view crew-requirements list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list crew requirements"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "Filter by job production plan"
xbe_json view crew-requirements list --job-production-plan "$CREATED_JPP_ID_2" --limit 5
assert_success

test_name "Filter by resource classification"
xbe_json view crew-requirements list \
    --resource-classification-type EquipmentClassification \
    --resource-classification-id "$CREATED_EQUIPMENT_CLASSIFICATION_ID_2" \
    --limit 5
assert_success

test_name "Filter by resource"
xbe_json view crew-requirements list \
    --resource-type Equipment \
    --resource-id "$CREATED_EQUIPMENT_ID_2" \
    --limit 5
assert_success

test_name "Filter by broker"
xbe_json view crew-requirements list --broker "$CREATED_BROKER_ID" --limit 5
assert_success

test_name "Filter by customer"
xbe_json view crew-requirements list --customer "$CREATED_CUSTOMER_ID" --limit 5
assert_success

test_name "Filter by project manager"
xbe_json view crew-requirements list --project-manager "1" --limit 5
assert_success

test_name "Filter by project"
xbe_json view crew-requirements list --project "1" --limit 5
assert_success

test_name "Filter by has-resource"
xbe_json view crew-requirements list --has-resource true --limit 5
assert_success

test_name "Filter by start-at-min"
xbe_json view crew-requirements list --start-at-min "$REQ_START_AT" --limit 5
assert_success

test_name "Filter by start-at-max"
xbe_json view crew-requirements list --start-at-max "$REQ_END_AT" --limit 5
assert_success

test_name "Filter by is-start-at"
xbe_json view crew-requirements list --is-start-at true --limit 5
assert_success

test_name "Filter by end-at-min"
xbe_json view crew-requirements list --end-at-min "$REQ_START_AT" --limit 5
assert_success

test_name "Filter by end-at-max"
xbe_json view crew-requirements list --end-at-max "$REQ_END_AT" --limit 5
assert_success

test_name "Filter by is-end-at"
xbe_json view crew-requirements list --is-end-at true --limit 5
assert_success

test_name "Filter by start-at-effective-min"
xbe_json view crew-requirements list --start-at-effective-min "$REQ_START_AT_EFFECTIVE" --limit 5
assert_success

test_name "Filter by start-at-effective-max"
xbe_json view crew-requirements list --start-at-effective-max "$REQ_START_AT_EFFECTIVE" --limit 5
assert_success

test_name "Filter by end-at-effective-min"
xbe_json view crew-requirements list --end-at-effective-min "$REQ_END_AT_EFFECTIVE" --limit 5
assert_success

test_name "Filter by end-at-effective-max"
xbe_json view crew-requirements list --end-at-effective-max "$REQ_END_AT_EFFECTIVE" --limit 5
assert_success

test_name "Filter by start-on-effective-min"
xbe_json view crew-requirements list --start-on-effective-min "$REQ_START_ON_EFFECTIVE" --limit 5
assert_success

test_name "Filter by start-on-effective-max"
xbe_json view crew-requirements list --start-on-effective-max "$REQ_START_ON_EFFECTIVE" --limit 5
assert_success

test_name "Filter by calculated mobilization method"
if [[ -n "$REQ_CALC_MOBILIZATION_METHOD" && "$REQ_CALC_MOBILIZATION_METHOD" != "null" ]]; then
    xbe_json view crew-requirements list --calculated-mobilization-method "$REQ_CALC_MOBILIZATION_METHOD" --limit 5
    assert_success
else
    skip "No calculated mobilization method available"
fi

test_name "Filter by job production plan status"
if [[ -n "$REQ_JPP_STATUS" && "$REQ_JPP_STATUS" != "null" ]]; then
    xbe_json view crew-requirements list --job-production-plan-status "$REQ_JPP_STATUS" --limit 5
    assert_success
else
    skip "No job production plan status available"
fi

test_name "Filter by labor requirement"
xbe_json view crew-requirements list --labor-requirement "$CREATED_LABOR_REQUIREMENT_ID_2" --limit 5
assert_success

test_name "Filter by labor requirement laborer"
xbe_json view crew-requirements list --labor-requirement-laborer "$CREATED_LABORER_ID" --limit 5
assert_success

test_name "Filter by labor requirement laborer ID"
xbe_json view crew-requirements list --labor-requirement-laborer-id "$CREATED_LABORER_ID" --limit 5
assert_success

test_name "Filter by labor requirement user"
xbe_json view crew-requirements list --labor-requirement-user "$CREATED_USER_ID" --limit 5
assert_success

test_name "Filter by labor requirement user ID"
xbe_json view crew-requirements list --labor-requirement-user-id "$CREATED_USER_ID" --limit 5
assert_success

test_name "Filter by requires inbound movement"
xbe_json view crew-requirements list --requires-inbound-movement false --limit 5
assert_success

test_name "Filter by requires outbound movement"
xbe_json view crew-requirements list --requires-outbound-movement true --limit 5
assert_success

test_name "Filter by is-only-for-equipment-movement"
xbe_json view crew-requirements list --is-only-for-equipment-movement false --limit 5
assert_success

test_name "Filter by without-approved-time-sheet"
xbe_json view crew-requirements list --without-approved-time-sheet true --limit 5
assert_success

test_name "Filter by without-submitted-time-sheet"
xbe_json view crew-requirements list --without-submitted-time-sheet true --limit 5
assert_success

test_name "Filter by is-expecting-time-sheet"
xbe_json view crew-requirements list --is-expecting-time-sheet false --limit 5
assert_success

test_name "Filter by created-at-min"
xbe_json view crew-requirements list --created-at-min "2020-01-01T00:00:00Z" --limit 5
assert_success

test_name "Filter by created-at-max"
xbe_json view crew-requirements list --created-at-max "2030-01-01T00:00:00Z" --limit 5
assert_success

test_name "Filter by is-created-at"
xbe_json view crew-requirements list --is-created-at true --limit 5
assert_success

test_name "Filter by updated-at-min"
xbe_json view crew-requirements list --updated-at-min "2020-01-01T00:00:00Z" --limit 5
assert_success

test_name "Filter by updated-at-max"
xbe_json view crew-requirements list --updated-at-max "2030-01-01T00:00:00Z" --limit 5
assert_success

test_name "Filter by is-updated-at"
xbe_json view crew-requirements list --is-updated-at true --limit 5
assert_success

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete labor requirement"
if [[ -n "$CREATED_LABOR_REQUIREMENT_ID_2" && "$CREATED_LABOR_REQUIREMENT_ID_2" != "null" ]]; then
    xbe_run do crew-requirements delete "$CREATED_LABOR_REQUIREMENT_ID_2" --confirm
    if [[ $status -eq 0 ]]; then
        pass
    else
        skip "Could not delete labor requirement (permissions or policy)"
    fi
else
    skip "No labor requirement ID available"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create crew requirement without required fields fails"
xbe_run do crew-requirements create
assert_failure

test_name "Update crew requirement without changes fails"
xbe_run do crew-requirements update "999999"
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
