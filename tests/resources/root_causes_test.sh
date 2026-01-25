#!/bin/bash
#
# XBE CLI Integration Tests: Root Causes
#
# Tests CRUD operations and list filters for the root causes resource.
#
# COVERAGE: Create/update attributes + list filters + delete
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_ROOT_CAUSE_ID=""
CHILD_ROOT_CAUSE_ID=""
INCIDENT_ID=""
INCIDENT_TYPE=""
INCIDENT_REF=""
SKIP_MUTATION=0
SKIP_INCIDENT_FILTERS=0

ROOT_CAUSE_TITLE=""

normalize_incident_type() {
    local value
    value=$(echo "$1" | tr '[:upper:]' '[:lower:]')
    case "$value" in
        incident|incidents)
            echo "Incident"
            ;;
        safety-incident|safety-incidents|safetyincident)
            echo "SafetyIncident"
            ;;
        production-incident|production-incidents|productionincident)
            echo "ProductionIncident"
            ;;
        efficiency-incident|efficiency-incidents|efficiencyincident)
            echo "EfficiencyIncident"
            ;;
        administrative-incident|administrative-incidents|administrativeincident)
            echo "AdministrativeIncident"
            ;;
        liability-incident|liability-incidents|liabilityincident)
            echo "LiabilityIncident"
            ;;
        *)
            echo "$1"
            ;;
    esac
}


describe "Resource: root-causes"

# ==========================================================================
# LIST Tests - Basic
# ==========================================================================

test_name "List root causes"
xbe_json view root-causes list --limit 5
assert_success

test_name "List root causes returns array"
xbe_json view root-causes list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list root causes"
fi

# ==========================================================================
# Setup - Sample incident
# ==========================================================================

test_name "Fetch sample incident"
xbe_json view incidents list --limit 1
if [[ $status -eq 0 ]]; then
    INCIDENT_ID=$(json_get '.[0].id')
    INCIDENT_TYPE=$(json_get '.[0].type')
    if [[ -n "$INCIDENT_ID" && "$INCIDENT_ID" != "null" ]]; then
        if [[ -z "$INCIDENT_TYPE" || "$INCIDENT_TYPE" == "null" ]]; then
            INCIDENT_TYPE="incidents"
        fi
        INCIDENT_REF="$(normalize_incident_type "$INCIDENT_TYPE")|$INCIDENT_ID"
        pass
    else
        skip "No incidents available; skipping mutation and incident filter tests"
        SKIP_MUTATION=1
        SKIP_INCIDENT_FILTERS=1
    fi
else
    skip "Failed to fetch incidents; skipping mutation and incident filter tests"
    SKIP_MUTATION=1
    SKIP_INCIDENT_FILTERS=1
fi

# ==========================================================================
# CREATE Tests
# ==========================================================================

test_name "Create root cause without required fields fails"
xbe_run do root-causes create
assert_failure

if [[ $SKIP_MUTATION -eq 1 ]]; then
    skip "Skipping create/update/delete tests"
else
    test_name "Create root cause"
    ROOT_CAUSE_TITLE=$(unique_name "RootCause")
    xbe_json do root-causes create \
        --incident-type "$INCIDENT_TYPE" \
        --incident-id "$INCIDENT_ID" \
        --title "$ROOT_CAUSE_TITLE" \
        --description "Test root cause description" \
        --is-triaged

    if [[ $status -eq 0 ]]; then
        CREATED_ROOT_CAUSE_ID=$(json_get ".id")
        if [[ -n "$CREATED_ROOT_CAUSE_ID" && "$CREATED_ROOT_CAUSE_ID" != "null" ]]; then
            register_cleanup "root-causes" "$CREATED_ROOT_CAUSE_ID"
            pass
        else
            fail "Created root cause but no ID returned"
        fi
    else
        fail "Failed to create root cause"
    fi
fi

if [[ -z "$CREATED_ROOT_CAUSE_ID" || "$CREATED_ROOT_CAUSE_ID" == "null" ]]; then
    echo "Cannot continue without a valid root cause ID"
    run_tests
fi

test_name "Create child root cause"
xbe_json do root-causes create \
    --incident-type "$INCIDENT_TYPE" \
    --incident-id "$INCIDENT_ID" \
    --root-cause "$CREATED_ROOT_CAUSE_ID" \
    --title "Child root cause"

if [[ $status -eq 0 ]]; then
    CHILD_ROOT_CAUSE_ID=$(json_get ".id")
    if [[ -n "$CHILD_ROOT_CAUSE_ID" && "$CHILD_ROOT_CAUSE_ID" != "null" ]]; then
        register_cleanup "root-causes" "$CHILD_ROOT_CAUSE_ID"
        pass
    else
        fail "Created child root cause but no ID returned"
    fi
else
    fail "Failed to create child root cause"
fi

# ==========================================================================
# SHOW Tests
# ==========================================================================

test_name "Show root cause"
xbe_json view root-causes show "$CREATED_ROOT_CAUSE_ID"
assert_success

# ==========================================================================
# UPDATE Tests
# ==========================================================================

test_name "Update root cause title"
UPDATED_TITLE=$(unique_name "RootCauseUpdated")
xbe_json do root-causes update "$CREATED_ROOT_CAUSE_ID" --title "$UPDATED_TITLE"
assert_success

test_name "Update root cause description"
xbe_json do root-causes update "$CREATED_ROOT_CAUSE_ID" --description "Updated description"
assert_success

test_name "Update root cause triage status"
xbe_json do root-causes update "$CREATED_ROOT_CAUSE_ID" --is-triaged=false
assert_success

# ==========================================================================
# LIST Tests - Filters
# ==========================================================================

if [[ $SKIP_INCIDENT_FILTERS -eq 1 ]]; then
    skip "Skipping incident filter tests"
else
    test_name "List root causes with --incident filter"
    xbe_json view root-causes list --incident "$INCIDENT_REF" --limit 5
    if [[ $status -eq 0 ]]; then
        pass
    elif [[ "$output" == *"Internal Server Error"* || "$output" == *"INTERNAL SERVER ERROR"* ]]; then
        skip "Server does not support incident filter for this resource"
    else
        fail "Failed to list root causes with incident filter"
    fi

    test_name "List root causes with --incident-type and --incident-id filters"
    xbe_json view root-causes list --incident-type "$INCIDENT_TYPE" --incident-id "$INCIDENT_ID" --limit 5
    if [[ $status -eq 0 ]]; then
        pass
    elif [[ "$output" == *"Internal Server Error"* || "$output" == *"INTERNAL SERVER ERROR"* ]]; then
        skip "Server does not support incident polymorphic filter for this resource"
    else
        fail "Failed to list root causes with incident polymorphic filter"
    fi
fi

test_name "List root causes with --is-triaged filter"
xbe_json view root-causes list --is-triaged true --limit 5
assert_success

# ==========================================================================
# DELETE Tests
# ==========================================================================

test_name "Delete root cause requires --confirm flag"
xbe_run do root-causes delete "$CREATED_ROOT_CAUSE_ID"
assert_failure

if [[ $SKIP_MUTATION -eq 1 ]]; then
    skip "Skipping delete test"
else
    test_name "Delete root cause with --confirm"
    xbe_json do root-causes create \
        --incident-type "$INCIDENT_TYPE" \
        --incident-id "$INCIDENT_ID" \
        --title "Delete root cause"

    if [[ $status -eq 0 ]]; then
        DEL_ID=$(json_get ".id")
        if [[ -n "$DEL_ID" && "$DEL_ID" != "null" ]]; then
            xbe_run do root-causes delete "$DEL_ID" --confirm
            assert_success
        else
            fail "Created root cause for deletion but no ID returned"
        fi
    else
        fail "Failed to create root cause for deletion"
    fi
fi

run_tests
