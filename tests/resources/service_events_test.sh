#!/bin/bash
#
# XBE CLI Integration Tests: Service Events
#
# Tests view and do operations for the service_events resource.
#
# COVERAGE: List filters + show + create/update/delete + error cases
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

SAMPLE_ID=""
SAMPLE_TENDER_JOB_SCHEDULE_SHIFT_ID=""
SAMPLE_KIND=""
SAMPLE_OCCURRED_AT=""
SAMPLE_CREATED_BY_ID=""
SAMPLE_JOB_SCHEDULE_SHIFT_ID=""

CREATED_SERVICE_EVENT_ID=""

NOW_OCCURRED_AT="$(date -u +"%Y-%m-%dT%H:%M:%SZ")"
UPDATED_OCCURRED_AT="$(date -u -v+1m +"%Y-%m-%dT%H:%M:%SZ" 2>/dev/null || date -u -d "+1 minute" +"%Y-%m-%dT%H:%M:%SZ")"
LATITUDE="34.0500"
LONGITUDE="-118.2500"
UPDATED_LATITUDE="34.0600"
UPDATED_LONGITUDE="-118.2600"

DIRECT_API_AVAILABLE=0
if [[ -n "$XBE_TOKEN" ]]; then
    DIRECT_API_AVAILABLE=1
fi

api_get() {
    local path="$1"
    if [[ $DIRECT_API_AVAILABLE -eq 0 ]]; then
        output="Missing XBE_TOKEN for direct API calls"
        status=1
        return
    fi
    run curl -sS -X GET "$XBE_BASE_URL$path" \
        -H "Authorization: Bearer $XBE_TOKEN" \
        -H "Content-Type: application/vnd.api+json"
}

describe "Resource: service-events"

# ==========================================================================
# LIST Tests - Basic
# ==========================================================================

test_name "List service events"
xbe_json view service-events list --limit 5
assert_success

test_name "List service events returns array"
xbe_json view service-events list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list service events"
fi

# ==========================================================================
# Sample Record (used for filters/show)
# ==========================================================================

test_name "Capture sample service event"
xbe_json view service-events list --limit 1
if [[ $status -eq 0 ]]; then
    SAMPLE_ID=$(json_get ".[0].id")
    SAMPLE_TENDER_JOB_SCHEDULE_SHIFT_ID=$(json_get ".[0].tender_job_schedule_shift_id")
    SAMPLE_KIND=$(json_get ".[0].kind")
    SAMPLE_OCCURRED_AT=$(json_get ".[0].occurred_at")
    SAMPLE_CREATED_BY_ID=$(json_get ".[0].created_by_id")
    if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
        pass
    else
        skip "No service events available for follow-on tests"
    fi
else
    skip "Could not list service events to capture sample"
fi

if [[ -n "$SAMPLE_TENDER_JOB_SCHEDULE_SHIFT_ID" && "$SAMPLE_TENDER_JOB_SCHEDULE_SHIFT_ID" != "null" && $DIRECT_API_AVAILABLE -eq 1 ]]; then
    test_name "Lookup job schedule shift for sample tender job schedule shift"
    api_get "/v1/tender-job-schedule-shifts/$SAMPLE_TENDER_JOB_SCHEDULE_SHIFT_ID?fields[tender-job-schedule-shifts]=start-at,job-schedule-shift"
    if [[ $status -eq 0 ]]; then
        SAMPLE_JOB_SCHEDULE_SHIFT_ID=$(echo "$output" | jq -r '.data.relationships["job-schedule-shift"].data.id // empty')
        pass
    else
        skip "Failed to fetch job schedule shift"
    fi
fi

# ==========================================================================
# LIST Tests - Filters
# ==========================================================================

test_name "List service events with --tender-job-schedule-shift filter"
if [[ -n "$SAMPLE_TENDER_JOB_SCHEDULE_SHIFT_ID" && "$SAMPLE_TENDER_JOB_SCHEDULE_SHIFT_ID" != "null" ]]; then
    xbe_json view service-events list --tender-job-schedule-shift "$SAMPLE_TENDER_JOB_SCHEDULE_SHIFT_ID" --limit 5
    assert_success
else
    skip "No tender job schedule shift ID available"
fi

test_name "List service events with --kind filter"
if [[ -n "$SAMPLE_KIND" && "$SAMPLE_KIND" != "null" ]]; then
    xbe_json view service-events list --kind "$SAMPLE_KIND" --limit 5
    assert_success
else
    skip "No kind available"
fi

test_name "List service events with --occurred-at-min filter"
if [[ -n "$SAMPLE_OCCURRED_AT" && "$SAMPLE_OCCURRED_AT" != "null" ]]; then
    xbe_json view service-events list --occurred-at-min "$SAMPLE_OCCURRED_AT" --limit 5
    assert_success
else
    xbe_json view service-events list --occurred-at-min "2024-01-01T00:00:00Z" --limit 5
    assert_success
fi

test_name "List service events with --occurred-at-max filter"
if [[ -n "$SAMPLE_OCCURRED_AT" && "$SAMPLE_OCCURRED_AT" != "null" ]]; then
    xbe_json view service-events list --occurred-at-max "$SAMPLE_OCCURRED_AT" --limit 5
    assert_success
else
    xbe_json view service-events list --occurred-at-max "2030-01-01T00:00:00Z" --limit 5
    assert_success
fi

test_name "List service events with --is-occurred-at filter"
xbe_json view service-events list --is-occurred-at true --limit 5
assert_success

test_name "List service events with --created-by filter"
if [[ -n "$SAMPLE_CREATED_BY_ID" && "$SAMPLE_CREATED_BY_ID" != "null" ]]; then
    xbe_json view service-events list --created-by "$SAMPLE_CREATED_BY_ID" --limit 5
    assert_success
else
    skip "No created-by ID available"
fi

test_name "List service events with --via-gps filter"
xbe_json view service-events list --via-gps true --limit 5
assert_success

test_name "List service events with --via-material-transaction-acceptance filter"
xbe_json view service-events list --via-material-transaction-acceptance true --limit 5
assert_success

test_name "List service events with --job-schedule-shift filter"
if [[ -n "$SAMPLE_JOB_SCHEDULE_SHIFT_ID" ]]; then
    xbe_json view service-events list --job-schedule-shift "$SAMPLE_JOB_SCHEDULE_SHIFT_ID" --limit 5
    assert_success
else
    skip "No job schedule shift ID available"
fi

# ==========================================================================
# SHOW Tests
# ==========================================================================

test_name "Show service event"
if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_json view service-events show "$SAMPLE_ID"
    if [[ $status -eq 0 ]]; then
        assert_json_has ".id"
    else
        fail "Failed to show service event"
    fi
else
    skip "No service event ID available"
fi

# ==========================================================================
# CREATE Tests
# ==========================================================================

test_name "Create service event with required fields"
CREATE_TENDER_JOB_SCHEDULE_SHIFT_ID="$SAMPLE_TENDER_JOB_SCHEDULE_SHIFT_ID"
if [[ -z "$CREATE_TENDER_JOB_SCHEDULE_SHIFT_ID" || "$CREATE_TENDER_JOB_SCHEDULE_SHIFT_ID" == "null" ]]; then
    if [[ -n "$XBE_TEST_TENDER_JOB_SCHEDULE_SHIFT_ID" ]]; then
        CREATE_TENDER_JOB_SCHEDULE_SHIFT_ID="$XBE_TEST_TENDER_JOB_SCHEDULE_SHIFT_ID"
    fi
fi

if [[ -n "$CREATE_TENDER_JOB_SCHEDULE_SHIFT_ID" && "$CREATE_TENDER_JOB_SCHEDULE_SHIFT_ID" != "null" ]]; then
    if [[ $DIRECT_API_AVAILABLE -eq 1 ]]; then
        api_get "/v1/tender-job-schedule-shifts/$CREATE_TENDER_JOB_SCHEDULE_SHIFT_ID?fields[tender-job-schedule-shifts]=start-at"
        if [[ $status -eq 0 ]]; then
            START_AT=$(echo "$output" | jq -r '.data.attributes["start-at"] // empty')
            if [[ -n "$START_AT" ]]; then
                NOW_OCCURRED_AT="$START_AT"
                UPDATED_OCCURRED_AT=$(date -u -d "$START_AT + 1 minute" +"%Y-%m-%dT%H:%M:%SZ" 2>/dev/null || date -u -v+1m -j -f "%Y-%m-%dT%H:%M:%SZ" "$START_AT" +"%Y-%m-%dT%H:%M:%SZ" 2>/dev/null || echo "$START_AT")
            fi
        fi
    fi

    xbe_json do service-events create \
        --tender-job-schedule-shift "$CREATE_TENDER_JOB_SCHEDULE_SHIFT_ID" \
        --occurred-at "$NOW_OCCURRED_AT" \
        --kind ready_to_work \
        --note "Test service event" \
        --occurred-latitude "$LATITUDE" \
        --occurred-longitude "$LONGITUDE"

    if [[ $status -eq 0 ]]; then
        CREATED_SERVICE_EVENT_ID=$(json_get ".id")
        if [[ -n "$CREATED_SERVICE_EVENT_ID" && "$CREATED_SERVICE_EVENT_ID" != "null" ]]; then
            register_cleanup "service-events" "$CREATED_SERVICE_EVENT_ID"
            pass
        else
            fail "Created service event but no ID returned"
        fi
    else
        if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]] || [[ "$output" == *"must be within"* ]] || [[ "$output" == *"already"* ]]; then
            pass
        else
            fail "Failed to create service event"
        fi
    fi
else
    skip "No tender job schedule shift ID available"
fi

# ==========================================================================
# UPDATE Tests
# ==========================================================================

test_name "Update service event note"
if [[ -n "$CREATED_SERVICE_EVENT_ID" && "$CREATED_SERVICE_EVENT_ID" != "null" ]]; then
    xbe_json do service-events update "$CREATED_SERVICE_EVENT_ID" --note "Updated service event"
    if [[ $status -eq 0 ]]; then
        pass
    else
        if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]]; then
            pass
        else
            fail "Failed to update service event"
        fi
    fi
else
    skip "No created service event ID available"
fi

# ==========================================================================
# DELETE Tests
# ==========================================================================

test_name "Delete service event requires --confirm flag"
if [[ -n "$CREATED_SERVICE_EVENT_ID" && "$CREATED_SERVICE_EVENT_ID" != "null" ]]; then
    xbe_run do service-events delete "$CREATED_SERVICE_EVENT_ID"
    assert_failure
else
    skip "No created service event ID available"
fi

test_name "Delete service event with --confirm"
if [[ -n "$CREATED_SERVICE_EVENT_ID" && "$CREATED_SERVICE_EVENT_ID" != "null" ]]; then
    xbe_run do service-events delete "$CREATED_SERVICE_EVENT_ID" --confirm
    if [[ $status -eq 0 ]]; then
        pass
    else
        if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]]; then
            pass
        else
            fail "Failed to delete service event"
        fi
    fi
else
    skip "No created service event ID available"
fi

# ==========================================================================
# Error Cases
# ==========================================================================

test_name "Create service event without required fields fails"
xbe_json do service-events create
assert_failure

test_name "Create service event without --occurred-at fails"
if [[ -n "$CREATE_TENDER_JOB_SCHEDULE_SHIFT_ID" && "$CREATE_TENDER_JOB_SCHEDULE_SHIFT_ID" != "null" ]]; then
    xbe_json do service-events create --tender-job-schedule-shift "$CREATE_TENDER_JOB_SCHEDULE_SHIFT_ID" --kind ready_to_work
    assert_failure
else
    skip "No tender job schedule shift ID available"
fi

test_name "Create service event with only latitude fails"
if [[ -n "$CREATE_TENDER_JOB_SCHEDULE_SHIFT_ID" && "$CREATE_TENDER_JOB_SCHEDULE_SHIFT_ID" != "null" ]]; then
    xbe_json do service-events create \
        --tender-job-schedule-shift "$CREATE_TENDER_JOB_SCHEDULE_SHIFT_ID" \
        --occurred-at "$NOW_OCCURRED_AT" \
        --kind ready_to_work \
        --occurred-latitude "$LATITUDE"
    assert_failure
else
    skip "No tender job schedule shift ID available"
fi

test_name "Update service event without fields fails"
if [[ -n "$CREATED_SERVICE_EVENT_ID" && "$CREATED_SERVICE_EVENT_ID" != "null" ]]; then
    xbe_run do service-events update "$CREATED_SERVICE_EVENT_ID"
    assert_failure
else
    skip "No created service event ID available"
fi

# ==========================================================================
# Summary
# ==========================================================================

run_tests
