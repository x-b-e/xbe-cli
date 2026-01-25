#!/bin/bash
#
# XBE CLI Integration Tests: Project Transport Plans
#
# Tests list, show, create, update, delete, and filters for project_transport_plans.
#
# COVERAGE: All list filters + writable attributes
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

SEED_PTP_ID="${XBE_TEST_PROJECT_TRANSPORT_PLAN_ID:-}"
PROJECT_ID="${XBE_TEST_PROJECT_ID:-}"
BROKER_ID="${XBE_TEST_BROKER_ID:-}"
PROJECT_TRANSPORT_ORG_ID="${XBE_TEST_PROJECT_TRANSPORT_ORGANIZATION_ID:-}"
PROJECT_OFFICE_ID="${XBE_TEST_PROJECT_OFFICE_ID:-}"
PROJECT_CATEGORY_ID="${XBE_TEST_PROJECT_CATEGORY_ID:-}"
NEAREST_PROJECT_OFFICE_ID="${XBE_TEST_NEAREST_PROJECT_OFFICE_ID:-}"
SEGMENT_MILES=""
EVENT_TIMES_AT_MIN=""
EVENT_TIMES_AT_MAX=""
EVENT_TIMES_ON_MIN=""
EVENT_TIMES_ON_MAX=""
CREATED_PTP_ID=""

describe "Resource: project-transport-plans"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List project transport plans"
xbe_json view project-transport-plans list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
    if [[ -z "$SEED_PTP_ID" || "$SEED_PTP_ID" == "null" ]]; then
        SEED_PTP_ID=$(json_get ".[0].id")
    fi
    if [[ -z "$PROJECT_ID" || "$PROJECT_ID" == "null" ]]; then
        PROJECT_ID=$(json_get ".[0].project_id")
    fi
    if [[ -z "$BROKER_ID" || "$BROKER_ID" == "null" ]]; then
        BROKER_ID=$(json_get ".[0].broker_id")
    fi
    SEGMENT_MILES=$(json_get ".[0].segment_miles")
    EVENT_TIMES_AT_MIN=$(json_get ".[0].event_times_at_min")
    EVENT_TIMES_AT_MAX=$(json_get ".[0].event_times_at_max")
    EVENT_TIMES_ON_MIN=$(json_get ".[0].event_times_on_min")
    EVENT_TIMES_ON_MAX=$(json_get ".[0].event_times_on_max")
    if [[ "$SEGMENT_MILES" == "null" ]]; then
        SEGMENT_MILES=""
    fi
    if [[ "$EVENT_TIMES_AT_MIN" == "null" ]]; then
        EVENT_TIMES_AT_MIN=""
    fi
    if [[ "$EVENT_TIMES_AT_MAX" == "null" ]]; then
        EVENT_TIMES_AT_MAX=""
    fi
    if [[ "$EVENT_TIMES_ON_MIN" == "null" ]]; then
        EVENT_TIMES_ON_MIN=""
    fi
    if [[ "$EVENT_TIMES_ON_MAX" == "null" ]]; then
        EVENT_TIMES_ON_MAX=""
    fi
else
    fail "Failed to list project transport plans"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show project transport plan"
if [[ -n "$SEED_PTP_ID" && "$SEED_PTP_ID" != "null" ]]; then
    xbe_json view project-transport-plans show "$SEED_PTP_ID"
    assert_success
else
    skip "No project transport plan available to show"
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create project transport plan"
if [[ -n "$PROJECT_ID" && "$PROJECT_ID" != "null" ]]; then
    xbe_json do project-transport-plans create \
        --project "$PROJECT_ID" \
        --skip-actualization \
        --skip-assignment-rules-validation \
        --assignment-rule-override-reason "CLI test override"
    if [[ $status -eq 0 ]]; then
        CREATED_PTP_ID=$(json_get ".id")
        if [[ -n "$CREATED_PTP_ID" && "$CREATED_PTP_ID" != "null" ]]; then
            register_cleanup "project-transport-plans" "$CREATED_PTP_ID"
            pass
        else
            fail "Created project transport plan but no ID returned"
        fi
    else
        if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]] || [[ "$output" == *"422"* ]] || [[ "$output" == *"409"* ]]; then
            skip "Create blocked by server policy/validation"
        else
            fail "Failed to create project transport plan"
        fi
    fi
else
    skip "Missing project ID (set XBE_TEST_PROJECT_ID)"
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update project transport plan attributes"
if [[ -n "$CREATED_PTP_ID" && "$CREATED_PTP_ID" != "null" ]]; then
    xbe_json do project-transport-plans update "$CREATED_PTP_ID" \
        --status editing \
        --skip-actualization=false \
        --skip-assignment-rules-validation=false \
        --assignment-rule-override-reason ""
    if [[ $status -eq 0 ]]; then
        pass
    else
        if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]] || [[ "$output" == *"422"* ]] || [[ "$output" == *"409"* ]]; then
            skip "Update blocked by server policy/validation"
        else
            fail "Failed to update project transport plan"
        fi
    fi
else
    skip "No created project transport plan to update"
fi

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete project transport plan requires --confirm flag"
if [[ -n "$CREATED_PTP_ID" && "$CREATED_PTP_ID" != "null" ]]; then
    xbe_run do project-transport-plans delete "$CREATED_PTP_ID"
    assert_failure
else
    skip "No created project transport plan to delete"
fi

test_name "Delete project transport plan"
if [[ -n "$CREATED_PTP_ID" && "$CREATED_PTP_ID" != "null" ]]; then
    xbe_run do project-transport-plans delete "$CREATED_PTP_ID" --confirm
    if [[ $status -eq 0 ]]; then
        pass
    else
        if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]] || [[ "$output" == *"422"* ]] || [[ "$output" == *"409"* ]]; then
            skip "Delete blocked by server policy/validation"
        else
            fail "Failed to delete project transport plan"
        fi
    fi
else
    skip "No created project transport plan to delete"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "Filter by project"
FILTER_PROJECT="${PROJECT_ID:-1}"
xbe_json view project-transport-plans list --project "$FILTER_PROJECT" --limit 5
assert_success

test_name "Filter by broker"
FILTER_BROKER="${BROKER_ID:-1}"
xbe_json view project-transport-plans list --broker "$FILTER_BROKER" --limit 5
assert_success

test_name "Filter by project transport organization"
FILTER_PTO="${PROJECT_TRANSPORT_ORG_ID:-1}"
xbe_json view project-transport-plans list --project-transport-organization "$FILTER_PTO" --limit 5
assert_success

test_name "Filter by broker-id"
FILTER_BROKER_ID="${BROKER_ID:-1}"
xbe_json view project-transport-plans list --broker-id "$FILTER_BROKER_ID" --limit 5
assert_success

test_name "Filter by search query"
xbe_json view project-transport-plans list --q "test" --limit 5
assert_success

test_name "Filter by segment miles min"
FILTER_SEGMENT_MILES_MIN="${SEGMENT_MILES:-0}"
xbe_json view project-transport-plans list --segment-miles-min "$FILTER_SEGMENT_MILES_MIN" --limit 5
assert_success

test_name "Filter by segment miles max"
FILTER_SEGMENT_MILES_MAX="${SEGMENT_MILES:-1000}"
xbe_json view project-transport-plans list --segment-miles-max "$FILTER_SEGMENT_MILES_MAX" --limit 5
assert_success

test_name "Filter by event times at min min"
FILTER_EVENT_AT_MIN="${EVENT_TIMES_AT_MIN:-2025-01-01T00:00:00Z}"
xbe_json view project-transport-plans list --event-times-at-min-min "$FILTER_EVENT_AT_MIN" --limit 5
assert_success

test_name "Filter by event times at min max"
FILTER_EVENT_AT_MIN_MAX="${EVENT_TIMES_AT_MIN:-2025-12-31T23:59:59Z}"
xbe_json view project-transport-plans list --event-times-at-min-max "$FILTER_EVENT_AT_MIN_MAX" --limit 5
assert_success

test_name "Filter by is-event-times-at-min"
xbe_json view project-transport-plans list --is-event-times-at-min true --limit 5
assert_success

test_name "Filter by event times at max min"
FILTER_EVENT_AT_MAX_MIN="${EVENT_TIMES_AT_MAX:-2025-01-01T00:00:00Z}"
xbe_json view project-transport-plans list --event-times-at-max-min "$FILTER_EVENT_AT_MAX_MIN" --limit 5
assert_success

test_name "Filter by event times at max max"
FILTER_EVENT_AT_MAX="${EVENT_TIMES_AT_MAX:-2025-12-31T23:59:59Z}"
xbe_json view project-transport-plans list --event-times-at-max-max "$FILTER_EVENT_AT_MAX" --limit 5
assert_success

test_name "Filter by is-event-times-at-max"
xbe_json view project-transport-plans list --is-event-times-at-max true --limit 5
assert_success

test_name "Filter by event times on min"
FILTER_EVENT_ON_MIN="${EVENT_TIMES_ON_MIN:-2025-01-01}"
xbe_json view project-transport-plans list --event-times-on-min "$FILTER_EVENT_ON_MIN" --limit 5
assert_success

test_name "Filter by event times on min min"
FILTER_EVENT_ON_MIN_MIN="${EVENT_TIMES_ON_MIN:-2025-01-01}"
xbe_json view project-transport-plans list --event-times-on-min-min "$FILTER_EVENT_ON_MIN_MIN" --limit 5
assert_success

test_name "Filter by event times on min max"
FILTER_EVENT_ON_MIN_MAX="${EVENT_TIMES_ON_MIN:-2025-12-31}"
xbe_json view project-transport-plans list --event-times-on-min-max "$FILTER_EVENT_ON_MIN_MAX" --limit 5
assert_success

test_name "Filter by has-event-times-on-min"
xbe_json view project-transport-plans list --has-event-times-on-min true --limit 5
assert_success

test_name "Filter by event times on max"
FILTER_EVENT_ON_MAX="${EVENT_TIMES_ON_MAX:-2025-12-31}"
xbe_json view project-transport-plans list --event-times-on-max "$FILTER_EVENT_ON_MAX" --limit 5
assert_success

test_name "Filter by event times on max min"
FILTER_EVENT_ON_MAX_MIN="${EVENT_TIMES_ON_MAX:-2025-01-01}"
xbe_json view project-transport-plans list --event-times-on-max-min "$FILTER_EVENT_ON_MAX_MIN" --limit 5
assert_success

test_name "Filter by event times on max max"
FILTER_EVENT_ON_MAX_MAX="${EVENT_TIMES_ON_MAX:-2025-12-31}"
xbe_json view project-transport-plans list --event-times-on-max-max "$FILTER_EVENT_ON_MAX_MAX" --limit 5
assert_success

test_name "Filter by has-event-times-on-max"
xbe_json view project-transport-plans list --has-event-times-on-max true --limit 5
assert_success

test_name "Filter by maybe active"
xbe_json view project-transport-plans list --maybe-active true --limit 5
assert_success

test_name "Filter by project office"
FILTER_PROJECT_OFFICE="${PROJECT_OFFICE_ID:-1}"
xbe_json view project-transport-plans list --project-office "$FILTER_PROJECT_OFFICE" --limit 5
assert_success

test_name "Filter by project category"
FILTER_PROJECT_CATEGORY="${PROJECT_CATEGORY_ID:-1}"
xbe_json view project-transport-plans list --project-category "$FILTER_PROJECT_CATEGORY" --limit 5
assert_success

test_name "Filter by is-managed"
xbe_json view project-transport-plans list --is-managed true --limit 5
assert_success

test_name "Filter by nearest project office IDs"
FILTER_NEAREST_PROJECT_OFFICE_IDS="${NEAREST_PROJECT_OFFICE_ID:-1}"
xbe_json view project-transport-plans list --nearest-project-office-ids "$FILTER_NEAREST_PROJECT_OFFICE_IDS" --limit 5
assert_success

test_name "Filter by external order number"
xbe_json view project-transport-plans list --external-order-number "CLI-TEST-ORDER" --limit 5
assert_success

test_name "Filter by transport order project office or nearest project offices"
FILTER_TO_OFFICES="${PROJECT_OFFICE_ID:-1}"
xbe_json view project-transport-plans list --transport-order-project-office-or-nearest-project-offices "$FILTER_TO_OFFICES" --limit 5
assert_success

test_name "Filter by pickup address state codes"
xbe_json view project-transport-plans list --pickup-address-state-codes "TX" --limit 5
assert_success

test_name "Filter by delivery address state codes"
xbe_json view project-transport-plans list --delivery-address-state-codes "CA" --limit 5
assert_success

run_tests
