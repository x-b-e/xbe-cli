#!/bin/bash
#
# XBE CLI Integration Tests: Job Production Plan Trailer Classifications
#
# Tests list, show, create, update, and delete operations.
#
# COVERAGE: List + show + filters + create/update/delete + error cases
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

SAMPLE_ID=""
SAMPLE_JOB_PRODUCTION_PLAN_ID=""
SAMPLE_TRAILER_CLASSIFICATION_ID=""
LIST_SUPPORTED="true"

CREATE_JOB_PRODUCTION_PLAN_ID=""
CREATE_TRAILER_CLASSIFICATION_ID=""
CREATED_ID=""

describe "Resource: job-production-plan-trailer-classifications"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List job production plan trailer classifications"
xbe_json view job-production-plan-trailer-classifications list --limit 5
if [[ $status -eq 0 ]]; then
    pass
else
    if [[ "$output" == *"404"* ]] || [[ "$output" == *"doesn't exist"* ]]; then
        LIST_SUPPORTED="false"
        skip "Server does not support listing job production plan trailer classifications"
    else
        fail "Expected success (exit 0), got exit $status"
    fi
fi

test_name "List job production plan trailer classifications returns array"
if [[ "$LIST_SUPPORTED" == "true" ]]; then
    xbe_json view job-production-plan-trailer-classifications list --limit 5
    if [[ $status -eq 0 ]]; then
        assert_json_is_array
    else
        fail "Failed to list job production plan trailer classifications"
    fi
else
    skip "List endpoint not supported"
fi

# ============================================================================
# Sample Record (used for show + filters)
# ============================================================================

test_name "Capture sample job production plan trailer classification"
if [[ "$LIST_SUPPORTED" == "true" ]]; then
    xbe_json view job-production-plan-trailer-classifications list --limit 1
    if [[ $status -eq 0 ]]; then
        SAMPLE_ID=$(json_get ".[0].id")
        SAMPLE_JOB_PRODUCTION_PLAN_ID=$(json_get ".[0].job_production_plan_id")
        SAMPLE_TRAILER_CLASSIFICATION_ID=$(json_get ".[0].trailer_classification_id")
        if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
            pass
        else
            skip "No job production plan trailer classifications available for follow-on tests"
        fi
    else
        skip "Could not list job production plan trailer classifications to capture sample"
    fi
else
    skip "List endpoint not supported"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List job production plan trailer classifications with --job-production-plan filter"
if [[ -n "$SAMPLE_JOB_PRODUCTION_PLAN_ID" && "$SAMPLE_JOB_PRODUCTION_PLAN_ID" != "null" ]]; then
    xbe_json view job-production-plan-trailer-classifications list --job-production-plan "$SAMPLE_JOB_PRODUCTION_PLAN_ID" --limit 5
    assert_success
else
    skip "No sample job production plan ID available"
fi

test_name "List job production plan trailer classifications with --trailer-classification filter"
if [[ -n "$SAMPLE_TRAILER_CLASSIFICATION_ID" && "$SAMPLE_TRAILER_CLASSIFICATION_ID" != "null" ]]; then
    xbe_json view job-production-plan-trailer-classifications list --trailer-classification "$SAMPLE_TRAILER_CLASSIFICATION_ID" --limit 5
    assert_success
else
    skip "No sample trailer classification ID available"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show job production plan trailer classification"
if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_json view job-production-plan-trailer-classifications show "$SAMPLE_ID"
    assert_success
else
    skip "No job production plan trailer classification ID available"
fi

# ============================================================================
# Prerequisites for Create
# ============================================================================

test_name "Find job production plan for create"
xbe_json view job-production-plans list --limit 5
if [[ $status -eq 0 ]]; then
    CREATE_JOB_PRODUCTION_PLAN_ID=$(json_get ".[0].id")
    if [[ -n "$CREATE_JOB_PRODUCTION_PLAN_ID" && "$CREATE_JOB_PRODUCTION_PLAN_ID" != "null" ]]; then
        pass
    else
        skip "No job production plans available"
    fi
else
    skip "Could not list job production plans"
fi

test_name "Find trailer classification for create"
xbe_json view trailer-classifications list --limit 5
if [[ $status -eq 0 ]]; then
    CREATE_TRAILER_CLASSIFICATION_ID=$(json_get ".[0].id")
fi

if [[ -n "$CREATE_TRAILER_CLASSIFICATION_ID" && "$CREATE_TRAILER_CLASSIFICATION_ID" != "null" ]]; then
    pass
else
    skip "No trailer classifications available"
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create job production plan trailer classification"
if [[ -n "$CREATE_JOB_PRODUCTION_PLAN_ID" && "$CREATE_JOB_PRODUCTION_PLAN_ID" != "null" && -n "$CREATE_TRAILER_CLASSIFICATION_ID" && "$CREATE_TRAILER_CLASSIFICATION_ID" != "null" ]]; then
    xbe_json do job-production-plan-trailer-classifications create \
        --job-production-plan "$CREATE_JOB_PRODUCTION_PLAN_ID" \
        --trailer-classification "$CREATE_TRAILER_CLASSIFICATION_ID" \
        --gross-weight-legal-limit-lbs-explicit 80000 \
        --explicit-material-transaction-tons-max 20

    if [[ $status -eq 0 ]]; then
        CREATED_ID=$(json_get ".id")
        if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
            register_cleanup "job-production-plan-trailer-classifications" "$CREATED_ID"
            pass
        else
            fail "Created job production plan trailer classification but no ID returned"
        fi
    else
        if [[ "$output" == *"Not Authorized"* ]] || \
           [[ "$output" == *"not authorized"* ]] || \
           [[ "$output" == *"must be unique"* ]] || \
           [[ "$output" == *"Record Invalid"* ]] || \
           [[ "$output" == *"422"* ]]; then
            pass
        else
            fail "Create failed: $output"
        fi
    fi
else
    skip "Missing prerequisites for create"
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update job production plan trailer classification"
if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
    xbe_json do job-production-plan-trailer-classifications update "$CREATED_ID" \
        --gross-weight-legal-limit-lbs-explicit 90000 \
        --explicit-material-transaction-tons-max 22
    if [[ $status -eq 0 ]]; then
        pass
    else
        if [[ "$output" == *"Not Authorized"* ]] || \
           [[ "$output" == *"not authorized"* ]] || \
           [[ "$output" == *"Record Invalid"* ]] || \
           [[ "$output" == *"422"* ]]; then
            pass
        else
            fail "Update failed: $output"
        fi
    fi
else
    skip "No created job production plan trailer classification to update"
fi

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete job production plan trailer classification"
if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
    xbe_run do job-production-plan-trailer-classifications delete "$CREATED_ID" --confirm
    if [[ $status -eq 0 ]]; then
        pass
    else
        if [[ "$output" == *"Not Authorized"* ]] || \
           [[ "$output" == *"not authorized"* ]] || \
           [[ "$output" == *"cannot be deleted"* ]]; then
            pass
        else
            fail "Delete failed: $output"
        fi
    fi
else
    skip "No created job production plan trailer classification to delete"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create job production plan trailer classification without job production plan fails"
xbe_run do job-production-plan-trailer-classifications create --trailer-classification "1"
assert_failure

test_name "Create job production plan trailer classification without trailer classification fails"
xbe_run do job-production-plan-trailer-classifications create --job-production-plan "1"
assert_failure

test_name "Update job production plan trailer classification without fields fails"
xbe_run do job-production-plan-trailer-classifications update "99999999"
assert_failure

test_name "Delete job production plan trailer classification without confirm fails"
xbe_run do job-production-plan-trailer-classifications delete "99999999"
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
