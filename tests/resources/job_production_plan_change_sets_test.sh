#!/bin/bash
#
# XBE CLI Integration Tests: Job Production Plan Change Sets
#
# Tests list, show, create, update, and delete operations for job production plan change sets.
#
# COVERAGE: All list filters + create attributes/relationships
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_CHANGE_SET_ID=""
CREATED_CHANGE_SET_ID_2=""
CREATED_BROKER_ID=""
CREATED_CUSTOMER_ID=""
CREATED_JOB_SITE_ID_OLD=""
CREATED_JOB_SITE_ID_NEW=""
CREATED_MATERIAL_SUPPLIER_ID=""
CREATED_MATERIAL_SITE_ID_OLD=""
CREATED_MATERIAL_SITE_ID_NEW=""
CREATED_MATERIAL_TYPE_ID_OLD=""
CREATED_MATERIAL_TYPE_ID_NEW=""
CREATED_QC_ID_OLD=""
CREATED_QC_ID_NEW=""
CREATED_MIX_ID_OLD=""
CREATED_MIX_ID_NEW=""
CREATED_COST_CODE_ID_OLD=""
CREATED_COST_CODE_ID_NEW=""
CREATED_USER_PLANNER_ID=""
CREATED_USER_PM_ID=""
CREATED_USER_LABORER_ID=""
CREATED_USER_LABORER_ID_2=""
CREATED_LABOR_CLASSIFICATION_ID=""
CREATED_LABORER_ID_OLD=""
CREATED_LABORER_ID_NEW=""
CREATED_MEMBERSHIP_ID=""
CREATED_MEMBERSHIP_ID_2=""
CREATED_DEVELOPER_ID=""
CREATED_PROJECT_ID=""
CREATED_JPP_ID=""
CREATED_BY_ID=""

TODAY=$(date +%Y-%m-%d)


describe "Resource: job-production-plan-change-sets"

# ==========================================================================
# Prerequisites
# ==========================================================================

test_name "Create prerequisite broker"
BROKER_NAME=$(unique_name "JPPChangeSetBroker")

xbe_json do brokers create --name "$BROKER_NAME"

if [[ $status -eq 0 ]]; then
    CREATED_BROKER_ID=$(json_get ".id")
    if [[ -n "$CREATED_BROKER_ID" && "$CREATED_BROKER_ID" != "null" ]]; then
        register_cleanup "brokers" "$CREATED_BROKER_ID"
        pass
    else
        fail "Created broker but no ID returned"
        run_tests
    fi
else
    if [[ -n "$XBE_TEST_BROKER_ID" ]]; then
        CREATED_BROKER_ID="$XBE_TEST_BROKER_ID"
        echo "    Using XBE_TEST_BROKER_ID: $CREATED_BROKER_ID"
        pass
    else
        fail "Failed to create broker"
        run_tests
    fi
fi

test_name "Create prerequisite customer"
CUSTOMER_NAME=$(unique_name "JPPChangeSetCustomer")

xbe_json do customers create --name "$CUSTOMER_NAME" --broker "$CREATED_BROKER_ID"

if [[ $status -eq 0 ]]; then
    CREATED_CUSTOMER_ID=$(json_get ".id")
    if [[ -n "$CREATED_CUSTOMER_ID" && "$CREATED_CUSTOMER_ID" != "null" ]]; then
        register_cleanup "customers" "$CREATED_CUSTOMER_ID"
        pass
    else
        fail "Created customer but no ID returned"
        run_tests
    fi
else
    fail "Failed to create customer"
    run_tests
fi

test_name "Create prerequisite job sites"
JOB_SITE_NAME_OLD=$(unique_name "JPPChangeSetJobSiteOld")
JOB_SITE_NAME_NEW=$(unique_name "JPPChangeSetJobSiteNew")

xbe_json do job-sites create --name "$JOB_SITE_NAME_OLD" --customer "$CREATED_CUSTOMER_ID" --address "100 Test St, Chicago, IL 60601"
if [[ $status -eq 0 ]]; then
    CREATED_JOB_SITE_ID_OLD=$(json_get ".id")
    if [[ -n "$CREATED_JOB_SITE_ID_OLD" && "$CREATED_JOB_SITE_ID_OLD" != "null" ]]; then
        register_cleanup "job-sites" "$CREATED_JOB_SITE_ID_OLD"
        pass
    else
        fail "Created old job site but no ID returned"
        run_tests
    fi
else
    fail "Failed to create old job site"
    run_tests
fi

xbe_json do job-sites create --name "$JOB_SITE_NAME_NEW" --customer "$CREATED_CUSTOMER_ID" --address "200 Test St, Chicago, IL 60601"
if [[ $status -eq 0 ]]; then
    CREATED_JOB_SITE_ID_NEW=$(json_get ".id")
    if [[ -n "$CREATED_JOB_SITE_ID_NEW" && "$CREATED_JOB_SITE_ID_NEW" != "null" ]]; then
        register_cleanup "job-sites" "$CREATED_JOB_SITE_ID_NEW"
        pass
    else
        fail "Created new job site but no ID returned"
        run_tests
    fi
else
    fail "Failed to create new job site"
    run_tests
fi

test_name "Create prerequisite material supplier"
SUPPLIER_NAME=$(unique_name "JPPChangeSetSupplier")

xbe_json do material-suppliers create --name "$SUPPLIER_NAME" --broker "$CREATED_BROKER_ID"

if [[ $status -eq 0 ]]; then
    CREATED_MATERIAL_SUPPLIER_ID=$(json_get ".id")
    if [[ -n "$CREATED_MATERIAL_SUPPLIER_ID" && "$CREATED_MATERIAL_SUPPLIER_ID" != "null" ]]; then
        register_cleanup "material-suppliers" "$CREATED_MATERIAL_SUPPLIER_ID"
        pass
    else
        fail "Created material supplier but no ID returned"
        run_tests
    fi
else
    fail "Failed to create material supplier"
    run_tests
fi

test_name "Create prerequisite material sites"
MATERIAL_SITE_NAME_OLD=$(unique_name "JPPChangeSetMatSiteOld")
MATERIAL_SITE_NAME_NEW=$(unique_name "JPPChangeSetMatSiteNew")

xbe_json do material-sites create --name "$MATERIAL_SITE_NAME_OLD" --material-supplier "$CREATED_MATERIAL_SUPPLIER_ID" --address "300 Test St, Chicago, IL 60601"
if [[ $status -eq 0 ]]; then
    CREATED_MATERIAL_SITE_ID_OLD=$(json_get ".id")
    if [[ -n "$CREATED_MATERIAL_SITE_ID_OLD" && "$CREATED_MATERIAL_SITE_ID_OLD" != "null" ]]; then
        register_cleanup "material-sites" "$CREATED_MATERIAL_SITE_ID_OLD"
        pass
    else
        fail "Created old material site but no ID returned"
        run_tests
    fi
else
    fail "Failed to create old material site"
    run_tests
fi

xbe_json do material-sites create --name "$MATERIAL_SITE_NAME_NEW" --material-supplier "$CREATED_MATERIAL_SUPPLIER_ID" --address "400 Test St, Chicago, IL 60601"
if [[ $status -eq 0 ]]; then
    CREATED_MATERIAL_SITE_ID_NEW=$(json_get ".id")
    if [[ -n "$CREATED_MATERIAL_SITE_ID_NEW" && "$CREATED_MATERIAL_SITE_ID_NEW" != "null" ]]; then
        register_cleanup "material-sites" "$CREATED_MATERIAL_SITE_ID_NEW"
        pass
    else
        fail "Created new material site but no ID returned"
        run_tests
    fi
else
    fail "Failed to create new material site"
    run_tests
fi

test_name "Create prerequisite material types"
MATERIAL_TYPE_NAME_OLD=$(unique_name "JPPChangeSetMatTypeOld")
MATERIAL_TYPE_NAME_NEW=$(unique_name "JPPChangeSetMatTypeNew")

xbe_json do material-types create --name "$MATERIAL_TYPE_NAME_OLD"
if [[ $status -eq 0 ]]; then
    CREATED_MATERIAL_TYPE_ID_OLD=$(json_get ".id")
    if [[ -n "$CREATED_MATERIAL_TYPE_ID_OLD" && "$CREATED_MATERIAL_TYPE_ID_OLD" != "null" ]]; then
        register_cleanup "material-types" "$CREATED_MATERIAL_TYPE_ID_OLD"
        pass
    else
        fail "Created old material type but no ID returned"
        run_tests
    fi
else
    fail "Failed to create old material type"
    run_tests
fi

xbe_json do material-types create --name "$MATERIAL_TYPE_NAME_NEW"
if [[ $status -eq 0 ]]; then
    CREATED_MATERIAL_TYPE_ID_NEW=$(json_get ".id")
    if [[ -n "$CREATED_MATERIAL_TYPE_ID_NEW" && "$CREATED_MATERIAL_TYPE_ID_NEW" != "null" ]]; then
        register_cleanup "material-types" "$CREATED_MATERIAL_TYPE_ID_NEW"
        pass
    else
        fail "Created new material type but no ID returned"
        run_tests
    fi
else
    fail "Failed to create new material type"
    run_tests
fi

test_name "Create prerequisite quality control classifications"
QC_NAME_OLD=$(unique_name "JPPChangeSetQCFirst")
QC_NAME_NEW=$(unique_name "JPPChangeSetQCSecond")

xbe_json do quality-control-classifications create --name "$QC_NAME_OLD" --broker "$CREATED_BROKER_ID"
if [[ $status -eq 0 ]]; then
    CREATED_QC_ID_OLD=$(json_get ".id")
    if [[ -n "$CREATED_QC_ID_OLD" && "$CREATED_QC_ID_OLD" != "null" ]]; then
        register_cleanup "quality-control-classifications" "$CREATED_QC_ID_OLD"
        pass
    else
        fail "Created old QC classification but no ID returned"
        run_tests
    fi
else
    fail "Failed to create old QC classification"
    run_tests
fi

xbe_json do quality-control-classifications create --name "$QC_NAME_NEW" --broker "$CREATED_BROKER_ID"
if [[ $status -eq 0 ]]; then
    CREATED_QC_ID_NEW=$(json_get ".id")
    if [[ -n "$CREATED_QC_ID_NEW" && "$CREATED_QC_ID_NEW" != "null" ]]; then
        register_cleanup "quality-control-classifications" "$CREATED_QC_ID_NEW"
        pass
    else
        fail "Created new QC classification but no ID returned"
        run_tests
    fi
else
    fail "Failed to create new QC classification"
    run_tests
fi

test_name "Create prerequisite material mix designs"
MIX_ID_OLD="MIX$(unique_suffix)"
MIX_ID_NEW="MIX$(unique_suffix)"
MIX_DESIGNS_AVAILABLE="true"

xbe_json do material-mix-designs create --material-type "$CREATED_MATERIAL_TYPE_ID_OLD" --mix "$MIX_ID_OLD"
if [[ $status -eq 0 ]]; then
    CREATED_MIX_ID_OLD=$(json_get ".id")
    if [[ -n "$CREATED_MIX_ID_OLD" && "$CREATED_MIX_ID_OLD" != "null" ]]; then
        register_cleanup "material-mix-designs" "$CREATED_MIX_ID_OLD"
        pass
    else
        fail "Created old material mix design but no ID returned"
        run_tests
    fi
else
    MIX_DESIGNS_AVAILABLE="false"
    skip "Failed to create old material mix design (server may not support this operation)"
fi

xbe_json do material-mix-designs create --material-type "$CREATED_MATERIAL_TYPE_ID_NEW" --mix "$MIX_ID_NEW"
if [[ $status -eq 0 ]]; then
    CREATED_MIX_ID_NEW=$(json_get ".id")
    if [[ -n "$CREATED_MIX_ID_NEW" && "$CREATED_MIX_ID_NEW" != "null" ]]; then
        register_cleanup "material-mix-designs" "$CREATED_MIX_ID_NEW"
        pass
    else
        fail "Created new material mix design but no ID returned"
        run_tests
    fi
else
    MIX_DESIGNS_AVAILABLE="false"
    skip "Failed to create new material mix design (server may not support this operation)"
fi

if [[ "$MIX_DESIGNS_AVAILABLE" != "true" ]]; then
    CREATED_MIX_ID_OLD=""
    CREATED_MIX_ID_NEW=""
fi

test_name "Create prerequisite cost codes"
COST_CODE_OLD="CC-OLD-$(date +%s%N | tail -c 6)"
COST_CODE_NEW="CC-NEW-$(date +%s%N | tail -c 6)"

xbe_json do cost-codes create --code "$COST_CODE_OLD" --customer "$CREATED_CUSTOMER_ID"
if [[ $status -eq 0 ]]; then
    CREATED_COST_CODE_ID_OLD=$(json_get ".id")
    if [[ -n "$CREATED_COST_CODE_ID_OLD" && "$CREATED_COST_CODE_ID_OLD" != "null" ]]; then
        register_cleanup "cost-codes" "$CREATED_COST_CODE_ID_OLD"
        pass
    else
        fail "Created old cost code but no ID returned"
        run_tests
    fi
else
    fail "Failed to create old cost code"
    run_tests
fi

xbe_json do cost-codes create --code "$COST_CODE_NEW" --customer "$CREATED_CUSTOMER_ID"
if [[ $status -eq 0 ]]; then
    CREATED_COST_CODE_ID_NEW=$(json_get ".id")
    if [[ -n "$CREATED_COST_CODE_ID_NEW" && "$CREATED_COST_CODE_ID_NEW" != "null" ]]; then
        register_cleanup "cost-codes" "$CREATED_COST_CODE_ID_NEW"
        pass
    else
        fail "Created new cost code but no ID returned"
        run_tests
    fi
else
    fail "Failed to create new cost code"
    run_tests
fi

test_name "Create prerequisite users"
PLANNER_EMAIL=$(unique_email)
PLANNER_NAME=$(unique_name "Planner")
PM_EMAIL=$(unique_email)
PM_NAME=$(unique_name "ProjectManager")

xbe_json do users create --name "$PLANNER_NAME" --email "$PLANNER_EMAIL"
if [[ $status -eq 0 ]]; then
    CREATED_USER_PLANNER_ID=$(json_get ".id")
    if [[ -n "$CREATED_USER_PLANNER_ID" && "$CREATED_USER_PLANNER_ID" != "null" ]]; then
        pass
    else
        fail "Created planner user but no ID returned"
        run_tests
    fi
else
    fail "Failed to create planner user"
    run_tests
fi

xbe_json do users create --name "$PM_NAME" --email "$PM_EMAIL"
if [[ $status -eq 0 ]]; then
    CREATED_USER_PM_ID=$(json_get ".id")
    if [[ -n "$CREATED_USER_PM_ID" && "$CREATED_USER_PM_ID" != "null" ]]; then
        pass
    else
        fail "Created project manager user but no ID returned"
        run_tests
    fi
else
    fail "Failed to create project manager user"
    run_tests
fi

test_name "Create prerequisite labor classification"
LABOR_CLASS_NAME=$(unique_name "JPPChangeSetLaborClass")
LABOR_CLASS_ABBR="LC$(date +%s | tail -c 4)"

xbe_json do labor-classifications create --name "$LABOR_CLASS_NAME" --abbreviation "$LABOR_CLASS_ABBR"
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

test_name "Create prerequisite laborer users"
LABORER_EMAIL=$(unique_email)
LABORER_NAME=$(unique_name "Laborer")
LABORER_EMAIL_2=$(unique_email)
LABORER_NAME_2=$(unique_name "Laborer2")

xbe_json do users create --name "$LABORER_NAME" --email "$LABORER_EMAIL"
if [[ $status -eq 0 ]]; then
    CREATED_USER_LABORER_ID=$(json_get ".id")
    if [[ -n "$CREATED_USER_LABORER_ID" && "$CREATED_USER_LABORER_ID" != "null" ]]; then
        pass
    else
        fail "Created laborer user but no ID returned"
        run_tests
    fi
else
    fail "Failed to create laborer user"
    run_tests
fi

xbe_json do users create --name "$LABORER_NAME_2" --email "$LABORER_EMAIL_2"
if [[ $status -eq 0 ]]; then
    CREATED_USER_LABORER_ID_2=$(json_get ".id")
    if [[ -n "$CREATED_USER_LABORER_ID_2" && "$CREATED_USER_LABORER_ID_2" != "null" ]]; then
        pass
    else
        fail "Created second laborer user but no ID returned"
        run_tests
    fi
else
    fail "Failed to create second laborer user"
    run_tests
fi

test_name "Create memberships for laborer users"
xbe_json do memberships create --user "$CREATED_USER_LABORER_ID" --organization "Customer|$CREATED_CUSTOMER_ID"
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

xbe_json do memberships create --user "$CREATED_USER_LABORER_ID_2" --organization "Customer|$CREATED_CUSTOMER_ID"
if [[ $status -eq 0 ]]; then
    CREATED_MEMBERSHIP_ID_2=$(json_get ".id")
    if [[ -n "$CREATED_MEMBERSHIP_ID_2" && "$CREATED_MEMBERSHIP_ID_2" != "null" ]]; then
        register_cleanup "memberships" "$CREATED_MEMBERSHIP_ID_2"
        pass
    else
        fail "Created second membership but no ID returned"
        run_tests
    fi
else
    fail "Failed to create second membership"
    run_tests
fi

test_name "Create laborers"
xbe_json do laborers create \
    --labor-classification "$CREATED_LABOR_CLASSIFICATION_ID" \
    --user "$CREATED_USER_LABORER_ID" \
    --organization-type "customers" \
    --organization-id "$CREATED_CUSTOMER_ID"
if [[ $status -eq 0 ]]; then
    CREATED_LABORER_ID_OLD=$(json_get ".id")
    if [[ -n "$CREATED_LABORER_ID_OLD" && "$CREATED_LABORER_ID_OLD" != "null" ]]; then
        register_cleanup "laborers" "$CREATED_LABORER_ID_OLD"
        pass
    else
        fail "Created laborer but no ID returned"
        run_tests
    fi
else
    fail "Failed to create laborer"
    run_tests
fi

xbe_json do laborers create \
    --labor-classification "$CREATED_LABOR_CLASSIFICATION_ID" \
    --user "$CREATED_USER_LABORER_ID_2" \
    --organization-type "customers" \
    --organization-id "$CREATED_CUSTOMER_ID"
if [[ $status -eq 0 ]]; then
    CREATED_LABORER_ID_NEW=$(json_get ".id")
    if [[ -n "$CREATED_LABORER_ID_NEW" && "$CREATED_LABORER_ID_NEW" != "null" ]]; then
        register_cleanup "laborers" "$CREATED_LABORER_ID_NEW"
        pass
    else
        fail "Created second laborer but no ID returned"
        run_tests
    fi
else
    fail "Failed to create second laborer"
    run_tests
fi

test_name "Create prerequisite developer and project"
DEVELOPER_NAME=$(unique_name "JPPChangeSetDeveloper")
PROJECT_NAME=$(unique_name "JPPChangeSetProject")

xbe_json do developers create --name "$DEVELOPER_NAME" --broker "$CREATED_BROKER_ID"
if [[ $status -eq 0 ]]; then
    CREATED_DEVELOPER_ID=$(json_get ".id")
    if [[ -n "$CREATED_DEVELOPER_ID" && "$CREATED_DEVELOPER_ID" != "null" ]]; then
        register_cleanup "developers" "$CREATED_DEVELOPER_ID"
        pass
    else
        fail "Created developer but no ID returned"
        run_tests
    fi
else
    fail "Failed to create developer"
    run_tests
fi

xbe_json do projects create --name "$PROJECT_NAME" --developer "$CREATED_DEVELOPER_ID"
if [[ $status -eq 0 ]]; then
    CREATED_PROJECT_ID=$(json_get ".id")
    if [[ -n "$CREATED_PROJECT_ID" && "$CREATED_PROJECT_ID" != "null" ]]; then
        register_cleanup "projects" "$CREATED_PROJECT_ID"
        pass
    else
        fail "Created project but no ID returned"
        run_tests
    fi
else
    fail "Failed to create project"
    run_tests
fi

test_name "Create prerequisite job production plan"
JPP_NAME=$(unique_name "JPPChangeSetPlan")

xbe_json do job-production-plans create \
    --job-name "$JPP_NAME" \
    --start-on "$TODAY" \
    --start-time "07:00" \
    --customer "$CREATED_CUSTOMER_ID" \
    --job-site "$CREATED_JOB_SITE_ID_OLD" \
    --cost-codes "$CREATED_COST_CODE_ID_OLD"
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

# ==========================================================================
# CREATE Tests
# ==========================================================================

test_name "Create change set with full scope and change attributes"

CHANGESET_ARGS=(do job-production-plan-change-sets create \
    --broker "$CREATED_BROKER_ID" \
    --customer "$CREATED_CUSTOMER_ID" \
    --scope-job-production-plan-ids "$CREATED_JPP_ID" \
    --scope-project-ids "$CREATED_PROJECT_ID" \
    --scope-planner-ids "$CREATED_USER_PLANNER_ID" \
    --scope-material-site-ids "$CREATED_MATERIAL_SITE_ID_OLD" \
    --scope-statuses "editing" \
    --scope-start-on-min "$TODAY" \
    --scope-start-on-max "$TODAY" \
    --scope-ultimate-material-types "Asphalt" \
    --scope-material-type-ids "$CREATED_MATERIAL_TYPE_ID_OLD" \
    --scope-foreman-ids "$CREATED_LABORER_ID_OLD" \
    --change-job-number-new "JOB-NEW" \
    --change-raw-job-number-new "RAW-NEW" \
    --change-new-status "approved" \
    --change-old-is-schedule-locked "false" \
    --change-new-is-schedule-locked "true" \
    --change-new-days-offset "1" \
    --change-new-offset-skip-saturdays \
    --change-new-offset-skip-sundays \
    --change-old-material-type "$CREATED_MATERIAL_TYPE_ID_OLD" \
    --change-new-material-type "$CREATED_MATERIAL_TYPE_ID_NEW" \
    --change-old-material-site "$CREATED_MATERIAL_SITE_ID_OLD" \
    --change-new-material-site "$CREATED_MATERIAL_SITE_ID_NEW" \
    --change-old-cost-code "$CREATED_COST_CODE_ID_OLD" \
    --change-new-cost-code "$CREATED_COST_CODE_ID_NEW" \
    --change-old-inspector "$CREATED_USER_PLANNER_ID" \
    --change-new-inspector "$CREATED_USER_PM_ID" \
    --change-new-planner "$CREATED_USER_PLANNER_ID" \
    --change-new-project-manager "$CREATED_USER_PM_ID" \
    --change-old-jpp-material-type-quality-control-classification "$CREATED_QC_ID_OLD" \
    --change-new-jpp-material-type-quality-control-classification "$CREATED_QC_ID_NEW" \
    --change-old-laborer "$CREATED_LABORER_ID_OLD" \
    --change-new-laborer "$CREATED_LABORER_ID_NEW" \
    --change-old-job-site "$CREATED_JOB_SITE_ID_OLD" \
    --change-new-job-site "$CREATED_JOB_SITE_ID_NEW")

if [[ -n "$CREATED_MIX_ID_OLD" && "$CREATED_MIX_ID_OLD" != "null" && -n "$CREATED_MIX_ID_NEW" && "$CREATED_MIX_ID_NEW" != "null" ]]; then
    CHANGESET_ARGS+=(--change-old-jpp-material-type-explicit-material-mix-design "$CREATED_MIX_ID_OLD")
    CHANGESET_ARGS+=(--change-new-jpp-material-type-explicit-material-mix-design "$CREATED_MIX_ID_NEW")
else
    echo "    (Skipping mix design change flags; mix design IDs unavailable)"
fi

xbe_json "${CHANGESET_ARGS[@]}"

if [[ $status -eq 0 ]]; then
    CREATED_CHANGE_SET_ID=$(json_get ".id")
    if [[ -n "$CREATED_CHANGE_SET_ID" && "$CREATED_CHANGE_SET_ID" != "null" ]]; then
        pass
    else
        fail "Created change set but no ID returned"
        run_tests
    fi
else
    fail "Failed to create change set"
    run_tests
fi

# Capture created-by for filter tests
xbe_json view job-production-plan-change-sets show "$CREATED_CHANGE_SET_ID"
if [[ $status -eq 0 ]]; then
    CREATED_BY_ID=$(json_get ".created_by_id")
fi

# Create second change set to exercise nullify + should-persist flags

test_name "Create change set with nullify and persist flags"

xbe_json do job-production-plan-change-sets create \
    --customer "$CREATED_CUSTOMER_ID" \
    --scope-statuses "scrapped" \
    --change-new-planner-nullify \
    --change-new-project-manager-nullify \
    --should-persist \
    --skip-invalid-plans

if [[ $status -eq 0 ]]; then
    CREATED_CHANGE_SET_ID_2=$(json_get ".id")
    if [[ -n "$CREATED_CHANGE_SET_ID_2" && "$CREATED_CHANGE_SET_ID_2" != "null" ]]; then
        pass
    else
        fail "Created second change set but no ID returned"
        run_tests
    fi
else
    fail "Failed to create second change set"
    run_tests
fi

# ==========================================================================
# SHOW Tests
# ==========================================================================

test_name "Show change set"
xbe_json view job-production-plan-change-sets show "$CREATED_CHANGE_SET_ID"
if [[ $status -eq 0 ]]; then
    assert_json_equals ".id" "$CREATED_CHANGE_SET_ID"
else
    fail "Failed to show change set"
fi

# ==========================================================================
# LIST Tests - Basic and Filters
# ==========================================================================

test_name "List change sets"
xbe_json view job-production-plan-change-sets list --limit 5
assert_success

test_name "List change sets with broker filter"
xbe_json view job-production-plan-change-sets list --broker "$CREATED_BROKER_ID" --limit 5
assert_success

test_name "List change sets with customer filter"
xbe_json view job-production-plan-change-sets list --customer "$CREATED_CUSTOMER_ID" --limit 5
assert_success

if [[ -n "$CREATED_BY_ID" && "$CREATED_BY_ID" != "null" ]]; then
    test_name "List change sets with created-by filter"
    xbe_json view job-production-plan-change-sets list --created-by "$CREATED_BY_ID" --limit 5
    assert_success
else
    test_name "List change sets with created-by filter"
    skip "No created-by ID available"
fi

test_name "List change sets with change-old-material-type filter"
xbe_json view job-production-plan-change-sets list --change-old-material-type "$CREATED_MATERIAL_TYPE_ID_OLD" --limit 5
assert_success

test_name "List change sets with change-new-material-type filter"
xbe_json view job-production-plan-change-sets list --change-new-material-type "$CREATED_MATERIAL_TYPE_ID_NEW" --limit 5
assert_success

test_name "List change sets with change-new-planner filter"
xbe_json view job-production-plan-change-sets list --change-new-planner "$CREATED_USER_PLANNER_ID" --limit 5
assert_success

test_name "List change sets with change-new-planner-nullify filter"
xbe_json view job-production-plan-change-sets list --change-new-planner-nullify true --limit 5
assert_success

test_name "List change sets with change-new-project-manager filter"
xbe_json view job-production-plan-change-sets list --change-new-project-manager "$CREATED_USER_PM_ID" --limit 5
assert_success

test_name "List change sets with change-new-project-manager-nullify filter"
xbe_json view job-production-plan-change-sets list --change-new-project-manager-nullify true --limit 5
assert_success

test_name "List change sets with should-persist filter"
xbe_json view job-production-plan-change-sets list --should-persist true --limit 5
assert_success

# ==========================================================================
# UPDATE/DELETE Tests (expected failure)
# ==========================================================================

test_name "Update change set should fail (immutable)"
xbe_json do job-production-plan-change-sets update "$CREATED_CHANGE_SET_ID" --change-new-status approved
assert_failure

test_name "Delete change set should fail (immutable)"
xbe_run do job-production-plan-change-sets delete "$CREATED_CHANGE_SET_ID" --confirm
assert_failure

# ==========================================================================
# Summary
# ==========================================================================

run_tests
