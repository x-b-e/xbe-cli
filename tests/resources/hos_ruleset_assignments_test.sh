#!/bin/bash
#
# XBE CLI Integration Tests: HOS Ruleset Assignments
#
# Tests list and show operations for the hos_ruleset_assignments resource.
#
# COVERAGE: List + filters + show
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

SAMPLE_ID=""
SAMPLE_USER_ID=""
SAMPLE_HOS_DAY_ID=""
SAMPLE_BROKER_ID=""
SAMPLE_EFFECTIVE_AT=""

describe "Resource: hos-ruleset-assignments"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List HOS ruleset assignments"
xbe_json view hos-ruleset-assignments list --limit 5
assert_success

test_name "List HOS ruleset assignments returns array"
xbe_json view hos-ruleset-assignments list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list HOS ruleset assignments"
fi

# ============================================================================
# Sample Record (used for filters/show)
# ============================================================================

test_name "Capture sample HOS ruleset assignment"
xbe_json view hos-ruleset-assignments list --limit 1
if [[ $status -eq 0 ]]; then
    SAMPLE_ID=$(json_get ".[0].id")
    SAMPLE_USER_ID=$(json_get ".[0].user_id")
    SAMPLE_HOS_DAY_ID=$(json_get ".[0].hos_day_id")
    SAMPLE_BROKER_ID=$(json_get ".[0].broker_id")
    SAMPLE_EFFECTIVE_AT=$(json_get ".[0].effective_at")
    if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
        pass
    else
        skip "No HOS ruleset assignments available for follow-on tests"
    fi
else
    skip "Could not list HOS ruleset assignments to capture sample"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List HOS ruleset assignments with --user filter"
if [[ -n "$SAMPLE_USER_ID" && "$SAMPLE_USER_ID" != "null" ]]; then
    xbe_json view hos-ruleset-assignments list --user "$SAMPLE_USER_ID" --limit 5
    assert_success
else
    skip "No user ID available"
fi

test_name "List HOS ruleset assignments with --driver filter"
if [[ -n "$SAMPLE_USER_ID" && "$SAMPLE_USER_ID" != "null" ]]; then
    xbe_json view hos-ruleset-assignments list --driver "$SAMPLE_USER_ID" --limit 5
    assert_success
else
    skip "No driver ID available"
fi

test_name "List HOS ruleset assignments with --hos-day filter"
if [[ -n "$SAMPLE_HOS_DAY_ID" && "$SAMPLE_HOS_DAY_ID" != "null" ]]; then
    xbe_json view hos-ruleset-assignments list --hos-day "$SAMPLE_HOS_DAY_ID" --limit 5
    assert_success
else
    skip "No HOS day ID available"
fi

test_name "List HOS ruleset assignments with --broker filter"
if [[ -n "$SAMPLE_BROKER_ID" && "$SAMPLE_BROKER_ID" != "null" ]]; then
    xbe_json view hos-ruleset-assignments list --broker "$SAMPLE_BROKER_ID" --limit 5
    assert_success
else
    skip "No broker ID available"
fi

test_name "List HOS ruleset assignments with --effective-at-min filter"
if [[ -n "$SAMPLE_EFFECTIVE_AT" && "$SAMPLE_EFFECTIVE_AT" != "null" ]]; then
    xbe_json view hos-ruleset-assignments list --effective-at-min "$SAMPLE_EFFECTIVE_AT" --limit 5
    assert_success
else
    skip "No effective-at value available"
fi

test_name "List HOS ruleset assignments with --effective-at-max filter"
if [[ -n "$SAMPLE_EFFECTIVE_AT" && "$SAMPLE_EFFECTIVE_AT" != "null" ]]; then
    xbe_json view hos-ruleset-assignments list --effective-at-max "$SAMPLE_EFFECTIVE_AT" --limit 5
    assert_success
else
    skip "No effective-at value available"
fi

test_name "List HOS ruleset assignments with --is-effective-at filter"
if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_json view hos-ruleset-assignments list --is-effective-at true --limit 5
    assert_success
else
    skip "No sample assignment available"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show HOS ruleset assignment"
if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_json view hos-ruleset-assignments show "$SAMPLE_ID"
    assert_success
else
    skip "No HOS ruleset assignment ID available"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
