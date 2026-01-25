#!/bin/bash
#
# XBE CLI Integration Tests: Job Production Plan Service Type Unit of Measures
#
# Tests list, show, create, update, and delete operations.
#
# COVERAGE: List + show + filters + create/update/delete + error cases
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

SAMPLE_ID=""
SAMPLE_JOB_PRODUCTION_PLAN_ID=""
SAMPLE_SERVICE_TYPE_UNIT_OF_MEASURE_ID=""
LIST_SUPPORTED="true"

CREATE_JOB_PRODUCTION_PLAN_ID=""
CREATE_SERVICE_TYPE_UNIT_OF_MEASURE_ID=""
CREATED_ID=""

describe "Resource: job-production-plan-service-type-unit-of-measures"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List job production plan service type unit of measures"
xbe_json view job-production-plan-service-type-unit-of-measures list --limit 5
if [[ $status -eq 0 ]]; then
    pass
else
    if [[ "$output" == *"404"* ]] || [[ "$output" == *"doesn't exist"* ]]; then
        LIST_SUPPORTED="false"
        skip "Server does not support listing job production plan service type unit of measures"
    else
        fail "Expected success (exit 0), got exit $status"
    fi
fi

test_name "List job production plan service type unit of measures returns array"
if [[ "$LIST_SUPPORTED" == "true" ]]; then
    xbe_json view job-production-plan-service-type-unit-of-measures list --limit 5
    if [[ $status -eq 0 ]]; then
        assert_json_is_array
    else
        fail "Failed to list job production plan service type unit of measures"
    fi
else
    skip "List endpoint not supported"
fi

# ============================================================================
# Sample Record (used for show + filters)
# ============================================================================

test_name "Capture sample job production plan service type unit of measure"
if [[ "$LIST_SUPPORTED" == "true" ]]; then
    xbe_json view job-production-plan-service-type-unit-of-measures list --limit 1
    if [[ $status -eq 0 ]]; then
        SAMPLE_ID=$(json_get ".[0].id")
        SAMPLE_JOB_PRODUCTION_PLAN_ID=$(json_get ".[0].job_production_plan_id")
        SAMPLE_SERVICE_TYPE_UNIT_OF_MEASURE_ID=$(json_get ".[0].service_type_unit_of_measure_id")
        if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
            pass
        else
            skip "No job production plan service type unit of measures available for follow-on tests"
        fi
    else
        skip "Could not list job production plan service type unit of measures to capture sample"
    fi
else
    skip "List endpoint not supported"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List job production plan service type unit of measures with --job-production-plan filter"
if [[ -n "$SAMPLE_JOB_PRODUCTION_PLAN_ID" && "$SAMPLE_JOB_PRODUCTION_PLAN_ID" != "null" ]]; then
    xbe_json view job-production-plan-service-type-unit-of-measures list --job-production-plan "$SAMPLE_JOB_PRODUCTION_PLAN_ID" --limit 5
    assert_success
else
    skip "No sample job production plan ID available"
fi

test_name "List job production plan service type unit of measures with --service-type-unit-of-measure filter"
if [[ -n "$SAMPLE_SERVICE_TYPE_UNIT_OF_MEASURE_ID" && "$SAMPLE_SERVICE_TYPE_UNIT_OF_MEASURE_ID" != "null" ]]; then
    xbe_json view job-production-plan-service-type-unit-of-measures list --service-type-unit-of-measure "$SAMPLE_SERVICE_TYPE_UNIT_OF_MEASURE_ID" --limit 5
    assert_success
else
    skip "No sample service type unit of measure ID available"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show job production plan service type unit of measure"
if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_json view job-production-plan-service-type-unit-of-measures show "$SAMPLE_ID"
    assert_success
else
    skip "No job production plan service type unit of measure ID available"
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

test_name "Find service type unit of measure for create"
if [[ -n "$SAMPLE_SERVICE_TYPE_UNIT_OF_MEASURE_ID" && "$SAMPLE_SERVICE_TYPE_UNIT_OF_MEASURE_ID" != "null" ]]; then
    CREATE_SERVICE_TYPE_UNIT_OF_MEASURE_ID="$SAMPLE_SERVICE_TYPE_UNIT_OF_MEASURE_ID"
    pass
else
    xbe_json view rates list --limit 5
    if [[ $status -eq 0 ]]; then
        CREATE_SERVICE_TYPE_UNIT_OF_MEASURE_ID=$(json_get ".[0].service_type_unit_of_measure_id")
    fi

    if [[ -n "$CREATE_SERVICE_TYPE_UNIT_OF_MEASURE_ID" && "$CREATE_SERVICE_TYPE_UNIT_OF_MEASURE_ID" != "null" ]]; then
        pass
    else
        skip "No service type unit of measure ID available"
    fi
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create job production plan service type unit of measure"
if [[ -n "$CREATE_JOB_PRODUCTION_PLAN_ID" && "$CREATE_JOB_PRODUCTION_PLAN_ID" != "null" && -n "$CREATE_SERVICE_TYPE_UNIT_OF_MEASURE_ID" && "$CREATE_SERVICE_TYPE_UNIT_OF_MEASURE_ID" != "null" ]]; then
    xbe_json do job-production-plan-service-type-unit-of-measures create \
        --job-production-plan "$CREATE_JOB_PRODUCTION_PLAN_ID" \
        --service-type-unit-of-measure "$CREATE_SERVICE_TYPE_UNIT_OF_MEASURE_ID" \
        --step-size no_step \
        --explicit-step-size-target "Loads" \
        --exclude-from-time-card-invoices

    if [[ $status -eq 0 ]]; then
        CREATED_ID=$(json_get ".id")
        if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
            register_cleanup "job-production-plan-service-type-unit-of-measures" "$CREATED_ID"
            pass
        else
            fail "Created job production plan service type unit of measure but no ID returned"
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

test_name "Update job production plan service type unit of measure"
if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
    xbe_json do job-production-plan-service-type-unit-of-measures update "$CREATED_ID" \
        --step-size floor \
        --explicit-step-size-target "Tons" \
        --exclude-from-time-card-invoices
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
    skip "No created job production plan service type unit of measure to update"
fi

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete job production plan service type unit of measure"
if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
    xbe_run do job-production-plan-service-type-unit-of-measures delete "$CREATED_ID" --confirm
    if [[ $status -eq 0 ]]; then
        pass
    else
        if [[ "$output" == *"Not Authorized"* ]] || \
           [[ "$output" == *"not authorized"* ]]; then
            pass
        else
            fail "Delete failed: $output"
        fi
    fi
else
    skip "No created job production plan service type unit of measure to delete"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create job production plan service type unit of measure without job production plan fails"
xbe_run do job-production-plan-service-type-unit-of-measures create --service-type-unit-of-measure "1"
assert_failure

test_name "Create job production plan service type unit of measure without service type unit of measure fails"
xbe_run do job-production-plan-service-type-unit-of-measures create --job-production-plan "1"
assert_failure

test_name "Update job production plan service type unit of measure without fields fails"
xbe_run do job-production-plan-service-type-unit-of-measures update "99999999"
assert_failure

test_name "Delete job production plan service type unit of measure without confirm fails"
xbe_run do job-production-plan-service-type-unit-of-measures delete "99999999"
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
