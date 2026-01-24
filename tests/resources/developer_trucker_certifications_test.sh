#!/bin/bash
#
# XBE CLI Integration Tests: Developer Trucker Certifications
#
# Tests list, show, create, update, and delete operations for developer-trucker-certifications.
#
# COVERAGE: All list filters + writable attributes
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_BROKER_ID=""
CREATED_DEVELOPER_ID=""
CREATED_TRUCKER_ID=""
CREATED_CLASSIFICATION_ID=""
CREATED_CLASSIFICATION_UPDATE_ID=""
CREATED_DEVELOPER_TRUCKER_CERTIFICATION_ID=""

SAMPLE_TENDER_JOB_SCHEDULE_SHIFT_ID=""

describe "Resource: developer-trucker-certifications"

# ============================================================================
# Prerequisites - Create broker, developer, trucker, classifications
# ============================================================================

test_name "Create broker for developer trucker certification tests"
BROKER_NAME=$(unique_name "DTCBroker")

xbe_json do brokers create --name "$BROKER_NAME"

if [[ $status -eq 0 ]]; then
    CREATED_BROKER_ID=$(json_get ".id")
    if [[ -n "$CREATED_BROKER_ID" && "$CREATED_BROKER_ID" != "null" ]]; then
        register_cleanup "brokers" "$CREATED_BROKER_ID"
        pass
    else
        fail "Created broker but no ID returned"
        echo "Cannot continue without a broker"
        run_tests
    fi
else
    if [[ -n "$XBE_TEST_BROKER_ID" ]]; then
        CREATED_BROKER_ID="$XBE_TEST_BROKER_ID"
        echo "    Using XBE_TEST_BROKER_ID: $CREATED_BROKER_ID"
        pass
    else
        fail "Failed to create broker and XBE_TEST_BROKER_ID not set"
        echo "Cannot continue without a broker"
        run_tests
    fi
fi

test_name "Create developer for developer trucker certification tests"
DEVELOPER_NAME=$(unique_name "DTCDeveloper")

xbe_json do developers create --name "$DEVELOPER_NAME" --broker "$CREATED_BROKER_ID"

if [[ $status -eq 0 ]]; then
    CREATED_DEVELOPER_ID=$(json_get ".id")
    if [[ -n "$CREATED_DEVELOPER_ID" && "$CREATED_DEVELOPER_ID" != "null" ]]; then
        register_cleanup "developers" "$CREATED_DEVELOPER_ID"
        pass
    else
        fail "Created developer but no ID returned"
        echo "Cannot continue without a developer"
        run_tests
    fi
else
    if [[ -n "$XBE_TEST_DEVELOPER_ID" ]]; then
        CREATED_DEVELOPER_ID="$XBE_TEST_DEVELOPER_ID"
        echo "    Using XBE_TEST_DEVELOPER_ID: $CREATED_DEVELOPER_ID"
        pass
    else
        fail "Failed to create developer and XBE_TEST_DEVELOPER_ID not set"
        echo "Cannot continue without a developer"
        run_tests
    fi
fi

test_name "Create trucker for developer trucker certification tests"
TRUCKER_NAME=$(unique_name "DTCTrucker")
TRUCKER_ADDRESS="123 $(unique_name "DTC Address") St"

xbe_json do truckers create --name "$TRUCKER_NAME" --broker "$CREATED_BROKER_ID" --company-address "$TRUCKER_ADDRESS"

if [[ $status -eq 0 ]]; then
    CREATED_TRUCKER_ID=$(json_get ".id")
    if [[ -n "$CREATED_TRUCKER_ID" && "$CREATED_TRUCKER_ID" != "null" ]]; then
        register_cleanup "truckers" "$CREATED_TRUCKER_ID"
        pass
    else
        fail "Created trucker but no ID returned"
        echo "Cannot continue without a trucker"
        run_tests
    fi
else
    if [[ -n "$XBE_TEST_TRUCKER_ID" ]]; then
        CREATED_TRUCKER_ID="$XBE_TEST_TRUCKER_ID"
        echo "    Using XBE_TEST_TRUCKER_ID: $CREATED_TRUCKER_ID"
        pass
    else
        fail "Failed to create trucker and XBE_TEST_TRUCKER_ID not set"
        echo "Cannot continue without a trucker"
        run_tests
    fi
fi

test_name "Create developer trucker certification classification"
CLASS_NAME=$(unique_name "DTCClass")

xbe_json do developer-trucker-certification-classifications create --name "$CLASS_NAME" --developer "$CREATED_DEVELOPER_ID"

if [[ $status -eq 0 ]]; then
    CREATED_CLASSIFICATION_ID=$(json_get ".id")
    if [[ -n "$CREATED_CLASSIFICATION_ID" && "$CREATED_CLASSIFICATION_ID" != "null" ]]; then
        register_cleanup "developer-trucker-certification-classifications" "$CREATED_CLASSIFICATION_ID"
        pass
    else
        fail "Created classification but no ID returned"
        echo "Cannot continue without a classification"
        run_tests
    fi
else
    fail "Failed to create developer trucker certification classification"
    echo "Cannot continue without a classification"
    run_tests
fi

test_name "Create second classification for update test"
CLASS_NAME_UPDATE=$(unique_name "DTCClassUpdate")

xbe_json do developer-trucker-certification-classifications create --name "$CLASS_NAME_UPDATE" --developer "$CREATED_DEVELOPER_ID"

if [[ $status -eq 0 ]]; then
    CREATED_CLASSIFICATION_UPDATE_ID=$(json_get ".id")
    if [[ -n "$CREATED_CLASSIFICATION_UPDATE_ID" && "$CREATED_CLASSIFICATION_UPDATE_ID" != "null" ]]; then
        register_cleanup "developer-trucker-certification-classifications" "$CREATED_CLASSIFICATION_UPDATE_ID"
        pass
    else
        fail "Created update classification but no ID returned"
        echo "Cannot continue without a classification for update"
        run_tests
    fi
else
    fail "Failed to create update classification"
    echo "Cannot continue without a classification for update"
    run_tests
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create developer trucker certification"
xbe_json do developer-trucker-certifications create \
    --developer "$CREATED_DEVELOPER_ID" \
    --trucker "$CREATED_TRUCKER_ID" \
    --classification "$CREATED_CLASSIFICATION_ID" \
    --start-on "2024-01-01" \
    --end-on "2024-12-31" \
    --default-multiplier "1.1"

if [[ $status -eq 0 ]]; then
    CREATED_DEVELOPER_TRUCKER_CERTIFICATION_ID=$(json_get ".id")
    if [[ -n "$CREATED_DEVELOPER_TRUCKER_CERTIFICATION_ID" && "$CREATED_DEVELOPER_TRUCKER_CERTIFICATION_ID" != "null" ]]; then
        register_cleanup "developer-trucker-certifications" "$CREATED_DEVELOPER_TRUCKER_CERTIFICATION_ID"
        pass
    else
        fail "Created developer trucker certification but no ID returned"
    fi
else
    fail "Failed to create developer trucker certification"
fi

if [[ -z "$CREATED_DEVELOPER_TRUCKER_CERTIFICATION_ID" || "$CREATED_DEVELOPER_TRUCKER_CERTIFICATION_ID" == "null" ]]; then
    echo "Cannot continue without a valid developer trucker certification ID"
    run_tests
fi

# ============================================================================
# LIST Tests
# ============================================================================

test_name "List developer trucker certifications"
xbe_json view developer-trucker-certifications list --limit 50
assert_json_is_array

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show developer trucker certification"
xbe_json view developer-trucker-certifications show "$CREATED_DEVELOPER_TRUCKER_CERTIFICATION_ID"
assert_success

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "Filter by developer"
xbe_json view developer-trucker-certifications list --developer "$CREATED_DEVELOPER_ID" --limit 5
assert_success

test_name "Filter by trucker"
xbe_json view developer-trucker-certifications list --trucker "$CREATED_TRUCKER_ID" --limit 5
assert_success

test_name "Filter by classification"
xbe_json view developer-trucker-certifications list --classification "$CREATED_CLASSIFICATION_ID" --limit 5
assert_success

test_name "Capture tender job schedule shift for filter"
SAMPLE_TENDER_JOB_SCHEDULE_SHIFT_ID="${XBE_TEST_TENDER_JOB_SCHEDULE_SHIFT_ID:-}"

if [[ -z "$SAMPLE_TENDER_JOB_SCHEDULE_SHIFT_ID" ]]; then
    xbe_json view time-cards list --limit 1
    if [[ $status -eq 0 ]]; then
        SAMPLE_TENDER_JOB_SCHEDULE_SHIFT_ID=$(json_get ".[0].tender_job_schedule_shift_id")
    fi
fi

test_name "Filter by tender job schedule shift"
if [[ -n "$SAMPLE_TENDER_JOB_SCHEDULE_SHIFT_ID" && "$SAMPLE_TENDER_JOB_SCHEDULE_SHIFT_ID" != "null" ]]; then
    xbe_json view developer-trucker-certifications list --tender-job-schedule-shift "$SAMPLE_TENDER_JOB_SCHEDULE_SHIFT_ID" --limit 5
    assert_success
else
    skip "No tender job schedule shift ID available"
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update developer trucker certification"
xbe_json do developer-trucker-certifications update "$CREATED_DEVELOPER_TRUCKER_CERTIFICATION_ID" \
    --classification "$CREATED_CLASSIFICATION_UPDATE_ID" \
    --start-on "2024-02-01" \
    --end-on "2024-11-30" \
    --default-multiplier "1.25"

assert_success

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete developer trucker certification"
xbe_run do developer-trucker-certifications delete "$CREATED_DEVELOPER_TRUCKER_CERTIFICATION_ID" --confirm
assert_success

run_tests
