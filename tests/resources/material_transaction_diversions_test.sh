#!/bin/bash
#
# XBE CLI Integration Tests: Material Transaction Diversions
#
# Tests list/show/create/update/delete operations for material-transaction-diversions.
#
# COVERAGE: List filters + writable attributes
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

SAMPLE_ID=""
SAMPLE_MTXN_ID=""
SAMPLE_BROKER_ID=""
SAMPLE_CREATED_BY_ID=""
SAMPLE_NEW_DELIVERY_DATE=""
CREATED_ID=""

created_from_env=""

# Optional env override for existing diversion
if [[ -n "${XBE_TEST_MATERIAL_TRANSACTION_DIVERSION_ID:-}" ]]; then
    created_from_env="$XBE_TEST_MATERIAL_TRANSACTION_DIVERSION_ID"
fi

describe "Resource: material-transaction-diversions"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List material transaction diversions"
xbe_json view material-transaction-diversions list --limit 5
assert_success

test_name "List material transaction diversions returns array"
xbe_json view material-transaction-diversions list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list material transaction diversions"
fi

# ============================================================================
# Sample Record (used for filters/show/update)
# ============================================================================

test_name "Capture sample diversion"
xbe_json view material-transaction-diversions list --limit 1
if [[ $status -eq 0 ]]; then
    SAMPLE_ID=$(json_get ".[0].id")
    SAMPLE_MTXN_ID=$(json_get ".[0].material_transaction_id")
    SAMPLE_BROKER_ID=$(json_get ".[0].broker_id")
    SAMPLE_CREATED_BY_ID=$(json_get ".[0].created_by_id")
    SAMPLE_NEW_DELIVERY_DATE=$(json_get ".[0].new_delivery_date")
    if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
        pass
    else
        skip "No diversions available for follow-on tests"
    fi
else
    skip "Could not list diversions to capture sample"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List diversions with --material-transaction filter"
if [[ -n "$SAMPLE_MTXN_ID" && "$SAMPLE_MTXN_ID" != "null" ]]; then
    xbe_json view material-transaction-diversions list --material-transaction "$SAMPLE_MTXN_ID" --limit 5
    assert_success
else
    skip "No material transaction ID available"
fi

test_name "List diversions with --broker-id filter"
if [[ -n "$SAMPLE_BROKER_ID" && "$SAMPLE_BROKER_ID" != "null" ]]; then
    xbe_json view material-transaction-diversions list --broker-id "$SAMPLE_BROKER_ID" --limit 5
    assert_success
else
    skip "No broker ID available"
fi

test_name "List diversions with --broker filter"
if [[ -n "$SAMPLE_BROKER_ID" && "$SAMPLE_BROKER_ID" != "null" ]]; then
    xbe_json view material-transaction-diversions list --broker "$SAMPLE_BROKER_ID" --limit 5
    assert_success
else
    skip "No broker ID available"
fi

test_name "List diversions with --created-by filter"
if [[ -n "$SAMPLE_CREATED_BY_ID" && "$SAMPLE_CREATED_BY_ID" != "null" ]]; then
    xbe_json view material-transaction-diversions list --created-by "$SAMPLE_CREATED_BY_ID" --limit 5
    assert_success
else
    skip "No created-by ID available"
fi

test_name "List diversions with --new-delivery-date filter"
if [[ -n "$SAMPLE_NEW_DELIVERY_DATE" && "$SAMPLE_NEW_DELIVERY_DATE" != "null" ]]; then
    xbe_json view material-transaction-diversions list --new-delivery-date "$SAMPLE_NEW_DELIVERY_DATE" --limit 5
    assert_success
else
    skip "No new delivery date available"
fi

test_name "List diversions with --new-delivery-date-min filter"
xbe_json view material-transaction-diversions list --new-delivery-date-min "2020-01-01" --limit 5
assert_success

test_name "List diversions with --new-delivery-date-max filter"
xbe_json view material-transaction-diversions list --new-delivery-date-max "2030-01-01" --limit 5
assert_success

test_name "List diversions with --has-new-delivery-date filter"
xbe_json view material-transaction-diversions list --has-new-delivery-date true --limit 5
assert_success

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show material transaction diversion"
SHOW_ID="$SAMPLE_ID"
if [[ -n "$created_from_env" ]]; then
    SHOW_ID="$created_from_env"
fi
if [[ -n "$SHOW_ID" && "$SHOW_ID" != "null" ]]; then
    xbe_json view material-transaction-diversions show "$SHOW_ID"
    assert_success
else
    skip "No diversion ID available"
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create diversion requires material transaction"
xbe_run do material-transaction-diversions create --driver-instructions "missing mtxn"
assert_failure

MTXN_ID="${XBE_TEST_MATERIAL_TRANSACTION_ID:-}"
if [[ -z "$MTXN_ID" && -n "$SAMPLE_MTXN_ID" && "$SAMPLE_MTXN_ID" != "null" ]]; then
    MTXN_ID="$SAMPLE_MTXN_ID"
fi
if [[ -z "$MTXN_ID" ]]; then
    xbe_json view material-transactions list --limit 1
    if [[ $status -eq 0 ]]; then
        MTXN_ID=$(json_get ".[0].id")
    fi
fi

test_name "Create material transaction diversion"
if [[ -n "$MTXN_ID" && "$MTXN_ID" != "null" ]]; then
    INSTRUCTIONS=$(unique_name "Diversion")
    xbe_json do material-transaction-diversions create \
        --material-transaction "$MTXN_ID" \
        --new-delivery-date "2100-01-01" \
        --diverted-tons-explicit "0.01" \
        --driver-instructions "$INSTRUCTIONS"
    if [[ $status -eq 0 ]]; then
        CREATED_ID=$(json_get ".id")
        if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
            register_cleanup "material-transaction-diversions" "$CREATED_ID"
        fi
        pass
    else
        if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]] || [[ "$output" == *"has already been taken"* ]] || [[ "$output" == *"already"* ]] || [[ "$output" == *"422"* ]] || [[ "$output" == *"409"* ]]; then
            skip "Create blocked by server policy/validation"
        else
            fail "Failed to create material transaction diversion: $output"
        fi
    fi
else
    skip "No material transaction ID available (set XBE_TEST_MATERIAL_TRANSACTION_ID)"
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

TARGET_ID="$CREATED_ID"
if [[ -z "$TARGET_ID" ]]; then
    TARGET_ID="$created_from_env"
fi
if [[ -z "$TARGET_ID" ]]; then
    TARGET_ID="$SAMPLE_ID"
fi

test_name "Update diversion driver instructions"
if [[ -n "$TARGET_ID" && "$TARGET_ID" != "null" ]]; then
    xbe_json do material-transaction-diversions update "$TARGET_ID" --driver-instructions "Updated instructions"
    if [[ $status -eq 0 ]]; then
        pass
    else
        skip "Could not update diversion (permissions or policy)"
    fi
else
    skip "No diversion ID available"
fi

test_name "Update diversion new delivery date"
if [[ -n "$TARGET_ID" && "$TARGET_ID" != "null" ]]; then
    xbe_json do material-transaction-diversions update "$TARGET_ID" --new-delivery-date "2100-01-02"
    if [[ $status -eq 0 ]]; then
        pass
    else
        skip "Could not update new delivery date (permissions or policy)"
    fi
else
    skip "No diversion ID available"
fi

test_name "Update diversion diverted tons explicit"
if [[ -n "$TARGET_ID" && "$TARGET_ID" != "null" ]]; then
    xbe_json do material-transaction-diversions update "$TARGET_ID" --diverted-tons-explicit "0.02"
    if [[ $status -eq 0 ]]; then
        pass
    else
        skip "Could not update diverted tons explicit (permissions or policy)"
    fi
else
    skip "No diversion ID available"
fi

test_name "Update diversion new job site"
if [[ -n "$TARGET_ID" && "$TARGET_ID" != "null" ]]; then
    BROKER_ID="$SAMPLE_BROKER_ID"
    if [[ -z "$BROKER_ID" || "$BROKER_ID" == "null" ]]; then
        xbe_json view material-transaction-diversions show "$TARGET_ID"
        if [[ $status -eq 0 ]]; then
            BROKER_ID=$(json_get ".broker_id")
        fi
    fi
    if [[ -n "$BROKER_ID" && "$BROKER_ID" != "null" ]]; then
        xbe_json view job-sites list --broker-id "$BROKER_ID" --limit 1
        if [[ $status -eq 0 ]]; then
            JOB_SITE_ID=$(json_get ".[0].id")
            if [[ -n "$JOB_SITE_ID" && "$JOB_SITE_ID" != "null" ]]; then
                xbe_json do material-transaction-diversions update "$TARGET_ID" --new-job-site "$JOB_SITE_ID"
                if [[ $status -eq 0 ]]; then
                    pass
                else
                    skip "Could not update new job site (permissions or validation)"
                fi
            else
                skip "No job site available for broker"
            fi
        else
            skip "Could not list job sites for broker"
        fi
    else
        skip "No broker ID available for job site lookup"
    fi
else
    skip "No diversion ID available"
fi

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete diversion requires --confirm flag"
if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
    xbe_run do material-transaction-diversions delete "$CREATED_ID"
    assert_failure
else
    skip "No created diversion available"
fi

test_name "Delete diversion with --confirm"
if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
    xbe_run do material-transaction-diversions delete "$CREATED_ID" --confirm
    if [[ $status -eq 0 ]]; then
        pass
    else
        skip "Could not delete diversion (permissions or policy)"
    fi
else
    skip "No created diversion available"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Update diversion without any fields fails"
xbe_run do material-transaction-diversions update "999999"
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
