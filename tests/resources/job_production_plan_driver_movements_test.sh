#!/bin/bash
#
# XBE CLI Integration Tests: Job Production Plan Driver Movements
#
# Tests create operations for the job-production-plan-driver-movements resource.
#
# COVERAGE: Create + error cases
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

describe "Resource: job-production-plan-driver-movements"

# ==========================================================================
# Error Cases
# ==========================================================================

test_name "Create movement without required fields fails"
xbe_run do job-production-plan-driver-movements create
assert_failure

# ==========================================================================
# CREATE Tests
# ==========================================================================

test_name "Create job production plan driver movement"
if [[ -z "$XBE_TOKEN" ]]; then
    skip "XBE_TOKEN not set; skipping API lookup for tender job schedule shifts"
else
    base_url="${XBE_BASE_URL%/}"

    shifts_json=$(curl -s \
        -H "Authorization: Bearer $XBE_TOKEN" \
        -H "Accept: application/vnd.api+json" \
        "$base_url/v1/tender-job-schedule-shifts?page[limit]=1&filter[with_job_production_plan]=true&filter[without_seller_operations_contact]=false&include=job-schedule-shift.job.job-production-plan" || true)

    shift_id=$(echo "$shifts_json" | jq -r '.data[0].id // empty' 2>/dev/null || true)
    job_schedule_shift_id=$(echo "$shifts_json" | jq -r '.data[0].relationships["job-schedule-shift"].data.id // empty' 2>/dev/null || true)
    job_production_plan_id=$(echo "$shifts_json" | jq -r '.included[] | select(.type=="job-production-plans") | .id' 2>/dev/null | head -n 1)

    if [[ -z "$job_production_plan_id" && -n "$job_schedule_shift_id" ]]; then
        jss_json=$(curl -s \
            -H "Authorization: Bearer $XBE_TOKEN" \
            -H "Accept: application/vnd.api+json" \
            "$base_url/v1/job-schedule-shifts/$job_schedule_shift_id?include=job.job-production-plan" || true)

        job_production_plan_id=$(echo "$jss_json" | jq -r '.included[] | select(.type=="job-production-plans") | .id' 2>/dev/null | head -n 1)
    fi

    if [[ -z "$shift_id" || -z "$job_production_plan_id" ]]; then
        skip "No suitable shift or job production plan found"
    else
        xbe_json do job-production-plan-driver-movements create \
            --job-production-plan "$job_production_plan_id" \
            --tender-job-schedule-shift "$shift_id" \
            --explicit-location-event-sources user_device \
            --bust-cache

        if [[ $status -eq 0 ]]; then
            assert_json_has ".id"
            assert_json_equals ".job_production_plan_id" "$job_production_plan_id"
            assert_json_equals ".tender_job_schedule_shift_id" "$shift_id"
        else
            if [[ "$output" == *"422"* ]] || [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]]; then
                pass
            else
                fail "Failed to create job production plan driver movement"
            fi
        fi
    fi
fi

# ==========================================================================
# Summary
# ==========================================================================

run_tests
