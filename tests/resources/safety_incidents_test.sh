#!/bin/bash
#
# XBE CLI Integration Tests: Safety Incidents
#
# Tests CRUD operations for the safety-incidents resource.
#
# COVERAGE: All filters + writable attributes
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_BROKER_ID=""
PARENT_SAFETY_INCIDENT_ID=""
CREATED_SAFETY_INCIDENT_ID=""
ASSIGNEE_ID=""
EQUIPMENT_ID="${XBE_TEST_EQUIPMENT_ID:-}"
JOB_PRODUCTION_PLAN_ID="${XBE_TEST_JOB_PRODUCTION_PLAN_ID:-}"
TENDER_JOB_SCHEDULE_SHIFT_ID="${XBE_TEST_TENDER_JOB_SCHEDULE_SHIFT_ID:-${XBE_TEST_TIME_CARD_TENDER_JOB_SCHEDULE_SHIFT_ID:-}}"

describe "Resource: safety-incidents"

# ============================================================================
# Prerequisites - Create broker for subject
# ============================================================================

test_name "Create prerequisite broker for safety incident tests"
BROKER_NAME=$(unique_name "SafetyIncidentTestBroker")

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
        fail "Failed to create broker and XBE_TEST_BROKER_ID not set"
        run_tests
    fi
fi

# ============================================================================
# Optional IDs for relationship tests
# ============================================================================

test_name "Lookup user ID for safety incident tests"
xbe_json view users list --limit 1
if [[ $status -eq 0 ]]; then
    ASSIGNEE_ID=$(echo "$output" | jq -r '.[0].id // empty')
    if [[ -n "$ASSIGNEE_ID" ]]; then
        pass
    else
        skip "No users available for assignee tests"
    fi
else
    skip "Failed to list users for assignee tests"
fi

if [[ -z "$EQUIPMENT_ID" ]]; then
    test_name "Lookup equipment ID for safety incident tests"
    xbe_json view equipment list --limit 1
    if [[ $status -eq 0 ]]; then
        EQUIPMENT_ID=$(echo "$output" | jq -r '.[0].id // empty')
        if [[ -n "$EQUIPMENT_ID" ]]; then
            pass
        else
            skip "No equipment available for relationship tests"
        fi
    else
        skip "Failed to list equipment"
    fi
fi

if [[ -z "$JOB_PRODUCTION_PLAN_ID" ]]; then
    test_name "Lookup job production plan ID for safety incident tests"
    xbe_json view job-production-plans list --limit 1
    if [[ $status -eq 0 ]]; then
        JOB_PRODUCTION_PLAN_ID=$(echo "$output" | jq -r '.[0].id // empty')
        if [[ -n "$JOB_PRODUCTION_PLAN_ID" ]]; then
            pass
        else
            skip "No job production plans available for relationship tests"
        fi
    else
        skip "Failed to list job production plans"
    fi
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create parent safety incident"
xbe_json do safety-incidents create \
    --subject-type brokers \
    --subject-id "$CREATED_BROKER_ID" \
    --start-at "2025-01-10T08:00:00Z" \
    --status open \
    --kind near_miss \
    --headline "Parent safety incident"

if [[ $status -eq 0 ]]; then
    PARENT_SAFETY_INCIDENT_ID=$(json_get ".id")
    if [[ -n "$PARENT_SAFETY_INCIDENT_ID" && "$PARENT_SAFETY_INCIDENT_ID" != "null" ]]; then
        register_cleanup "safety-incidents" "$PARENT_SAFETY_INCIDENT_ID"
        pass
    else
        fail "Created parent safety incident but no ID returned"
    fi
else
    fail "Failed to create parent safety incident"
fi

test_name "Create safety incident with required fields"
create_args=(
    do safety-incidents create
    --subject-type brokers
    --subject-id "$CREATED_BROKER_ID"
    --start-at "2025-01-11T08:00:00Z"
    --status open
    --kind near_miss
    --headline "Safety incident test"
    --description "Test description"
    --severity high
    --natures personal
    --did-stop-work
)
if [[ -n "$PARENT_SAFETY_INCIDENT_ID" ]]; then
    create_args+=(--parent "$PARENT_SAFETY_INCIDENT_ID")
fi

xbe_json "${create_args[@]}"

if [[ $status -eq 0 ]]; then
    CREATED_SAFETY_INCIDENT_ID=$(json_get ".id")
    if [[ -n "$CREATED_SAFETY_INCIDENT_ID" && "$CREATED_SAFETY_INCIDENT_ID" != "null" ]]; then
        register_cleanup "safety-incidents" "$CREATED_SAFETY_INCIDENT_ID"
        pass
    else
        fail "Created safety incident but no ID returned"
    fi
else
    fail "Failed to create safety incident"
fi

if [[ -z "$CREATED_SAFETY_INCIDENT_ID" || "$CREATED_SAFETY_INCIDENT_ID" == "null" ]]; then
    echo "Cannot continue without a valid safety incident ID"
    run_tests
fi

# ============================================================================
# UPDATE Tests - All writable attributes
# ============================================================================

test_name "Update safety incident attributes"
update_args=(
    do safety-incidents update "$CREATED_SAFETY_INCIDENT_ID"
    --start-at "2025-01-11T09:00:00Z"
    --end-at "2025-01-11T10:00:00Z"
    --status closed
    --kind overloading
    --severity medium
    --headline "Updated safety incident"
    --description "Updated description"
    --natures property
    --did-stop-work=false
    --net-impact-tons 8.5
    --new-type SafetyIncident
)
if [[ -n "$ASSIGNEE_ID" ]]; then
    update_args+=(--assignee "$ASSIGNEE_ID")
fi
if [[ -n "$JOB_PRODUCTION_PLAN_ID" ]]; then
    update_args+=(--job-production-plan "$JOB_PRODUCTION_PLAN_ID")
    if [[ -n "$EQUIPMENT_ID" ]]; then
        update_args+=(--equipment "$EQUIPMENT_ID")
    fi
fi
if [[ -n "$TENDER_JOB_SCHEDULE_SHIFT_ID" ]]; then
    update_args+=(--tender-job-schedule-shift "$TENDER_JOB_SCHEDULE_SHIFT_ID")
fi
if [[ -n "$PARENT_SAFETY_INCIDENT_ID" ]]; then
    update_args+=(--parent "$PARENT_SAFETY_INCIDENT_ID")
fi

xbe_json "${update_args[@]}"
assert_success

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show safety incident"
xbe_json view safety-incidents show "$CREATED_SAFETY_INCIDENT_ID"
assert_success

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List safety incidents"
xbe_json view safety-incidents list --limit 5
assert_success

test_name "List safety incidents returns array"
xbe_json view safety-incidents list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list safety incidents"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

FILTER_BROKER_ID="$CREATED_BROKER_ID"
FILTER_ASSIGNEE_ID="${ASSIGNEE_ID:-1}"
FILTER_EQUIPMENT_ID="${EQUIPMENT_ID:-1}"
FILTER_JOB_PRODUCTION_PLAN_ID="${JOB_PRODUCTION_PLAN_ID:-1}"
FILTER_TENDER_SHIFT_ID="${TENDER_JOB_SCHEDULE_SHIFT_ID:-1}"
FILTER_PARENT_ID="${PARENT_SAFETY_INCIDENT_ID:-1}"

test_name "List safety incidents with --status filter"
xbe_json view safety-incidents list --status "open" --limit 10
assert_success

test_name "List safety incidents with --kind filter"
xbe_json view safety-incidents list --kind "near_miss" --limit 10
assert_success

test_name "List safety incidents with --severity filter"
xbe_json view safety-incidents list --severity "high" --limit 10
assert_success

test_name "List safety incidents with --broker filter"
xbe_json view safety-incidents list --broker "$FILTER_BROKER_ID" --limit 10
assert_success

test_name "List safety incidents with --customer filter"
xbe_json view safety-incidents list --customer "1" --limit 10
assert_success

test_name "List safety incidents with --developer filter"
xbe_json view safety-incidents list --developer "1" --limit 10
assert_success

test_name "List safety incidents with --trucker filter"
xbe_json view safety-incidents list --trucker "1" --limit 10
assert_success

test_name "List safety incidents with --contractor filter"
xbe_json view safety-incidents list --contractor "1" --limit 10
assert_success

test_name "List safety incidents with --material-supplier filter"
xbe_json view safety-incidents list --material-supplier "1" --limit 10
assert_success

test_name "List safety incidents with --material-site filter"
xbe_json view safety-incidents list --material-site "1" --limit 10
assert_success

test_name "List safety incidents with --job-production-plan filter"
xbe_json view safety-incidents list --job-production-plan "$FILTER_JOB_PRODUCTION_PLAN_ID" --limit 10
assert_success

test_name "List safety incidents with --job-production-plan-project filter"
xbe_json view safety-incidents list --job-production-plan-project "1" --limit 10
assert_success

test_name "List safety incidents with --equipment filter"
xbe_json view safety-incidents list --equipment "$FILTER_EQUIPMENT_ID" --limit 10
assert_success

test_name "List safety incidents with --assignee filter"
xbe_json view safety-incidents list --assignee "$FILTER_ASSIGNEE_ID" --limit 10
assert_success

test_name "List safety incidents with --created-by filter"
xbe_json view safety-incidents list --created-by "$FILTER_ASSIGNEE_ID" --limit 10
assert_success

test_name "List safety incidents with --start-on filter"
xbe_json view safety-incidents list --start-on "2024-01-01" --limit 10
assert_success

test_name "List safety incidents with --start-on-min filter"
xbe_json view safety-incidents list --start-on-min "2024-01-01" --limit 10
assert_success

test_name "List safety incidents with --start-on-max filter"
xbe_json view safety-incidents list --start-on-max "2025-12-31" --limit 10
assert_success

test_name "List safety incidents with --start-at filter"
xbe_json view safety-incidents list --start-at "2025-01-01T00:00:00Z" --limit 10
assert_success

test_name "List safety incidents with --start-at-min filter"
xbe_json view safety-incidents list --start-at-min "2025-01-01T00:00:00Z" --limit 10
assert_success

test_name "List safety incidents with --start-at-max filter"
xbe_json view safety-incidents list --start-at-max "2025-12-31T23:59:59Z" --limit 10
assert_success

test_name "List safety incidents with --end-at filter"
xbe_json view safety-incidents list --end-at "2025-01-02T00:00:00Z" --limit 10
assert_success

test_name "List safety incidents with --end-at-min filter"
xbe_json view safety-incidents list --end-at-min "2025-01-01T00:00:00Z" --limit 10
assert_success

test_name "List safety incidents with --end-at-max filter"
xbe_json view safety-incidents list --end-at-max "2025-12-31T23:59:59Z" --limit 10
assert_success

test_name "List safety incidents with --subject filter"
xbe_json view safety-incidents list --subject "Broker|$FILTER_BROKER_ID" --limit 10
assert_success

test_name "List safety incidents with --subject-type filter"
xbe_json view safety-incidents list --subject-type "Broker" --limit 10
assert_success

test_name "List safety incidents with --subject-id filter"
xbe_json view safety-incidents list --subject-id "Broker|$FILTER_BROKER_ID" --limit 10
assert_success

test_name "List safety incidents with --parent filter"
xbe_json view safety-incidents list --parent "$FILTER_PARENT_ID" --limit 10
assert_success

test_name "List safety incidents with --has-parent filter"
xbe_json view safety-incidents list --has-parent "true" --limit 10
assert_success

test_name "List safety incidents with --has-equipment filter"
xbe_json view safety-incidents list --has-equipment "false" --limit 10
assert_success

test_name "List safety incidents with --has-live-action-items filter"
xbe_json view safety-incidents list --has-live-action-items "false" --limit 10
assert_success

test_name "List safety incidents with --incident-tag filter"
xbe_json view safety-incidents list --incident-tag "1" --limit 10
assert_success

test_name "List safety incidents with --incident-tag-slug filter"
xbe_json view safety-incidents list --incident-tag-slug "test-tag" --limit 10
assert_success

test_name "List safety incidents with --zero-incident-tags filter"
xbe_json view safety-incidents list --zero-incident-tags "true" --limit 10
assert_success

test_name "List safety incidents with --root-causes filter"
xbe_json view safety-incidents list --root-causes "1" --limit 10
assert_success

test_name "List safety incidents with --action-items filter"
xbe_json view safety-incidents list --action-items "1" --limit 10
assert_success

test_name "List safety incidents with --tender-job-schedule-shift filter"
xbe_json view safety-incidents list --tender-job-schedule-shift "$FILTER_TENDER_SHIFT_ID" --limit 10
assert_success

test_name "List safety incidents with --tender-job-schedule-shift-driver filter"
xbe_json view safety-incidents list --tender-job-schedule-shift-driver "1" --limit 10
assert_success

test_name "List safety incidents with --tender-job-schedule-shift-trucker filter"
xbe_json view safety-incidents list --tender-job-schedule-shift-trucker "1" --limit 10
assert_success

test_name "List safety incidents with --job-number filter"
xbe_json view safety-incidents list --job-number "TEST-123" --limit 10
assert_success

test_name "List safety incidents with --notifiable-to filter"
xbe_json view safety-incidents list --notifiable-to "$FILTER_ASSIGNEE_ID" --limit 10
assert_success

test_name "List safety incidents with --user-has-stake filter"
xbe_json view safety-incidents list --user-has-stake "$FILTER_ASSIGNEE_ID" --limit 10
assert_success

test_name "List safety incidents with --responsible-person filter"
xbe_json view safety-incidents list --responsible-person "$FILTER_ASSIGNEE_ID" --limit 10
assert_success

test_name "List safety incidents with --did-stop-work filter"
xbe_json view safety-incidents list --did-stop-work "true" --limit 10
assert_success

test_name "List safety incidents with --natures filter"
xbe_json view safety-incidents list --natures "personal" --limit 10
assert_success

test_name "List safety incidents with --q filter"
xbe_json view safety-incidents list --q "test" --limit 10
assert_success

test_name "List safety incidents with --net-impact-tons filter"
xbe_json view safety-incidents list --net-impact-tons "5" --limit 10
assert_success

test_name "List safety incidents with --net-impact-tons-min filter"
xbe_json view safety-incidents list --net-impact-tons-min "1" --limit 10
assert_success

test_name "List safety incidents with --net-impact-tons-max filter"
xbe_json view safety-incidents list --net-impact-tons-max "10" --limit 10
assert_success

test_name "List safety incidents with --sort"
xbe_json view safety-incidents list --sort -start-at --limit 5
assert_success

# ============================================================================
# LIST Tests - Pagination
# ============================================================================

test_name "List safety incidents with --limit"
xbe_json view safety-incidents list --limit 3
assert_success

test_name "List safety incidents with --offset"
xbe_json view safety-incidents list --limit 3 --offset 3
assert_success

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete safety incident"
xbe_run do safety-incidents delete "$CREATED_SAFETY_INCIDENT_ID" --confirm
assert_success

# ============================================================================
# Summary
# ============================================================================

run_tests
