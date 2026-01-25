#!/bin/bash
#
# XBE CLI Integration Tests: Efficiency Incidents
#
# Tests CRUD operations for the efficiency-incidents resource.
#
# COVERAGE: Create/update attributes + list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_BROKER_ID=""
CURRENT_USER_ID=""
PARENT_INCIDENT_ID=""
CHILD_INCIDENT_ID=""

START_AT_PARENT="2025-01-01T08:00:00Z"
START_AT_CHILD="2025-01-01T09:00:00Z"
START_AT_UPDATED="2025-01-01T09:30:00Z"
END_AT_CHILD="2025-01-01T10:00:00Z"

START_ON_DATE="2025-01-01"


describe "Resource: efficiency-incidents"

# ============================================================================
# Prerequisites
# ============================================================================

test_name "Create prerequisite broker for incident tests"
BROKER_NAME=$(unique_name "EfficiencyIncidentBroker")

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

# Fetch current user (for assignee/user filters)
test_name "Fetch current user"
xbe_json auth whoami
if [[ $status -eq 0 ]]; then
    CURRENT_USER_ID=$(json_get ".id")
    if [[ -n "$CURRENT_USER_ID" && "$CURRENT_USER_ID" != "null" ]]; then
        pass
    else
        fail "No user ID returned"
    fi
else
    fail "Failed to fetch current user"
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create efficiency incident (parent)"
SUBJECT_VALUE="Broker|$CREATED_BROKER_ID"
xbe_json do efficiency-incidents create \
    --subject "$SUBJECT_VALUE" \
    --start-at "$START_AT_PARENT" \
    --status open \
    --kind over_trucking \
    --severity medium \
    --headline "Efficiency incident parent" \
    --description "Efficiency incident test" \
    --net-impact-dollars 1000 \
    --assignee "$CURRENT_USER_ID"

if [[ $status -eq 0 ]]; then
    PARENT_INCIDENT_ID=$(json_get ".id")
    if [[ -n "$PARENT_INCIDENT_ID" && "$PARENT_INCIDENT_ID" != "null" ]]; then
        register_cleanup "efficiency-incidents" "$PARENT_INCIDENT_ID"
        pass
    else
        fail "Created incident but no ID returned"
    fi
else
    fail "Failed to create efficiency incident"
fi

# Only continue if we successfully created a parent incident
if [[ -z "$PARENT_INCIDENT_ID" || "$PARENT_INCIDENT_ID" == "null" ]]; then
    echo "Cannot continue without a valid parent incident ID"
    run_tests
fi

test_name "Create efficiency incident (child with parent)"
xbe_json do efficiency-incidents create \
    --subject "$SUBJECT_VALUE" \
    --start-at "$START_AT_CHILD" \
    --status open \
    --kind over_trucking \
    --parent "$PARENT_INCIDENT_ID"

if [[ $status -eq 0 ]]; then
    CHILD_INCIDENT_ID=$(json_get ".id")
    if [[ -n "$CHILD_INCIDENT_ID" && "$CHILD_INCIDENT_ID" != "null" ]]; then
        register_cleanup "efficiency-incidents" "$CHILD_INCIDENT_ID"
        pass
    else
        fail "Created child incident but no ID returned"
    fi
else
    fail "Failed to create child efficiency incident"
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update efficiency incident attributes"
xbe_json do efficiency-incidents update "$CHILD_INCIDENT_ID" \
    --start-at "$START_AT_UPDATED" \
    --end-at "$END_AT_CHILD" \
    --status closed \
    --kind over_trucking \
    --severity high \
    --headline "Efficiency incident updated" \
    --description "Updated efficiency incident" \
    --net-impact-dollars 1500 \
    --new-type EfficiencyIncident \
    --assignee "$CURRENT_USER_ID" \
    --parent "$PARENT_INCIDENT_ID"
assert_success

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show efficiency incident"
xbe_json view efficiency-incidents show "$CHILD_INCIDENT_ID"
assert_success

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List efficiency incidents"
xbe_json view efficiency-incidents list --limit 5
assert_success

# ============================================================================
# LIST Tests - Filters (coverage)
# ============================================================================

run_filter() {
    local name="$1"
    shift
    test_name "$name"
    xbe_json view efficiency-incidents list "$@" --limit 5
    assert_success
}

run_filter "Filter by status" --status open
run_filter "Filter by kind" --kind over_trucking
run_filter "Filter by severity" --severity high
run_filter "Filter by broker" --broker "$CREATED_BROKER_ID"
run_filter "Filter by customer" --customer 1
run_filter "Filter by developer" --developer 1
run_filter "Filter by trucker" --trucker 1
run_filter "Filter by contractor" --contractor 1
run_filter "Filter by material-supplier" --material-supplier 1
run_filter "Filter by material-site" --material-site 1
run_filter "Filter by job-production-plan" --job-production-plan 1
run_filter "Filter by job-production-plan-project" --job-production-plan-project 1
run_filter "Filter by equipment" --equipment 1
run_filter "Filter by assignee" --assignee "$CURRENT_USER_ID"
run_filter "Filter by created-by" --created-by "$CURRENT_USER_ID"
run_filter "Filter by parent" --parent "$PARENT_INCIDENT_ID"
run_filter "Filter by subject" --subject "$SUBJECT_VALUE"
run_filter "Filter by subject-type" --subject-type "Broker"
run_filter "Filter by subject-id" --subject-type "Broker" --subject-id "$CREATED_BROKER_ID"
test_name "Filter by not-subject-type (known server issue)"
skip "Server returns 500 for not-subject-type filter"
run_filter "Filter by has-parent" --has-parent true
run_filter "Filter by has-equipment" --has-equipment false
run_filter "Filter by has-live-action-items" --has-live-action-items false
run_filter "Filter by incident-tag" --incident-tag 1
run_filter "Filter by incident-tag-slug" --incident-tag-slug "test"
run_filter "Filter by zero-incident-tags" --zero-incident-tags true
run_filter "Filter by root-causes" --root-causes 1
run_filter "Filter by action-items" --action-items 1
run_filter "Filter by tender-job-schedule-shift" --tender-job-schedule-shift 1
run_filter "Filter by tender-job-schedule-shift-driver" --tender-job-schedule-shift-driver "$CURRENT_USER_ID"
run_filter "Filter by tender-job-schedule-shift-trucker" --tender-job-schedule-shift-trucker 1
run_filter "Filter by job-number" --job-number "JOB-1"
run_filter "Filter by notifiable-to" --notifiable-to "$CURRENT_USER_ID"
run_filter "Filter by user-has-stake" --user-has-stake "$CURRENT_USER_ID"
run_filter "Filter by responsible-person" --responsible-person "$CURRENT_USER_ID"
run_filter "Filter by natures" --natures personal
run_filter "Filter by did-stop-work" --did-stop-work true
run_filter "Filter by start-on" --start-on "$START_ON_DATE"
run_filter "Filter by start-on-min" --start-on-min "$START_ON_DATE"
run_filter "Filter by start-on-max" --start-on-max "$START_ON_DATE"
run_filter "Filter by start-at-min" --start-at-min "$START_AT_PARENT"
run_filter "Filter by start-at-max" --start-at-max "$END_AT_CHILD"
run_filter "Filter by end-at-min" --end-at-min "$START_AT_PARENT"
run_filter "Filter by end-at-max" --end-at-max "$END_AT_CHILD"
run_filter "Filter by net-impact-dollars" --net-impact-dollars 1500
run_filter "Filter by net-impact-dollars-min" --net-impact-dollars-min 500
run_filter "Filter by net-impact-dollars-max" --net-impact-dollars-max 2000
run_filter "Filter by search query" --q "incident"

# ============================================================================
# LIST Tests - Pagination & Sort
# ============================================================================

test_name "List efficiency incidents with --limit"
xbe_json view efficiency-incidents list --limit 3
assert_success

test_name "List efficiency incidents with --offset"
xbe_json view efficiency-incidents list --limit 3 --offset 3
assert_success

test_name "List efficiency incidents with --sort"
xbe_json view efficiency-incidents list --sort start-at
assert_success

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete efficiency incident requires --confirm flag"
xbe_json do efficiency-incidents delete "$CHILD_INCIDENT_ID"
assert_failure

# Create a separate incident for deletion test

test_name "Delete efficiency incident with --confirm"
DELETE_START_AT="2025-01-01T11:00:00Z"
xbe_json do efficiency-incidents create \
    --subject "$SUBJECT_VALUE" \
    --start-at "$DELETE_START_AT" \
    --status open \
    --kind over_trucking

if [[ $status -eq 0 ]]; then
    DELETE_ID=$(json_get ".id")
    xbe_json do efficiency-incidents delete "$DELETE_ID" --confirm
    assert_success
else
    skip "Could not create efficiency incident for deletion test"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create efficiency incident without subject fails"
xbe_json do efficiency-incidents create --start-at "$START_AT_PARENT" --status open
assert_failure

test_name "Update efficiency incident without fields fails"
xbe_json do efficiency-incidents update "$CHILD_INCIDENT_ID"
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
