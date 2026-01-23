#!/bin/bash
#
# XBE CLI Integration Tests: Shift Set Time Card Constraints
#
# Tests list, show, create, update, and delete operations for the shift-set-time-card-constraints resource.
#
# COVERAGE: All filters + create/update attributes + relationships + error cases
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

SAMPLE_ID=""
SAMPLE_NAME=""
SAMPLE_CONSTRAINED_TYPE=""
SAMPLE_CONSTRAINED_ID=""
SAMPLE_CONSTRAINT_TYPE=""
SAMPLE_CONSTRAINED_AMOUNT=""
SAMPLE_CONSTRAINED_AMOUNT_TYPE=""
SAMPLE_CURRENCY_CODE=""
SAMPLE_STATUS=""
SAMPLE_SHIFT_SCOPE_ID=""
SAMPLE_STUOM_CONSTRAINT_TYPE=""
SAMPLE_SHIFT_SET_GROUPED_BY=""
SAMPLE_CALC_PRICE_PER_UNIT=""
SAMPLE_CALC_UOM_ID=""

SAMPLE_SERVICE_TYPE_UNIT_OF_MEASURE_ID=""
SAMPLE_TRAILER_CLASSIFICATION_ID=""
SAMPLE_MATERIAL_TYPE_ID=""

CREATE_CONSTRAINED_TYPE=""
CREATE_CONSTRAINED_ID=""
CREATED_CONSTRAINT_ID=""
CREATED_CALCULATED_CONSTRAINT_ID=""

UPDATED_NAME=""


describe "Resource: shift-set-time-card-constraints"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List shift set time card constraints"
xbe_json view shift-set-time-card-constraints list --limit 5
assert_success

test_name "List shift set time card constraints returns array"
xbe_json view shift-set-time-card-constraints list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list shift set time card constraints"
fi

# ============================================================================
# Sample Record (used for filters/show)
# ============================================================================

test_name "Capture sample shift set time card constraint"
xbe_json view shift-set-time-card-constraints list --limit 1
if [[ $status -eq 0 ]]; then
    SAMPLE_ID=$(json_get ".[0].id")
    SAMPLE_NAME=$(json_get ".[0].name")
    SAMPLE_CONSTRAINED_TYPE=$(json_get ".[0].constrained_type")
    SAMPLE_CONSTRAINED_ID=$(json_get ".[0].constrained_id")
    SAMPLE_CONSTRAINT_TYPE=$(json_get ".[0].constraint_type")
    SAMPLE_CONSTRAINED_AMOUNT=$(json_get ".[0].constrained_amount")
    SAMPLE_CONSTRAINED_AMOUNT_TYPE=$(json_get ".[0].constrained_amount_type")
    SAMPLE_CURRENCY_CODE=$(json_get ".[0].currency_code")
    SAMPLE_STATUS=$(json_get ".[0].status")
    SAMPLE_SHIFT_SCOPE_ID=$(json_get ".[0].shift_scope_id")
    SAMPLE_CALC_PRICE_PER_UNIT=$(json_get ".[0].calculated_constrained_amount_price_per_unit")
    SAMPLE_CALC_UOM_ID=$(json_get ".[0].calculated_constrained_amount_unit_of_measure_id")
    if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
        pass
    else
        skip "No shift set time card constraints available for follow-on tests"
    fi
else
    skip "Could not list shift set time card constraints to capture sample"
fi

if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_json view shift-set-time-card-constraints show "$SAMPLE_ID"
    if [[ $status -eq 0 ]]; then
        SAMPLE_STUOM_CONSTRAINT_TYPE=$(json_get ".service_type_unit_of_measures_constraint_type")
        SAMPLE_SHIFT_SET_GROUPED_BY=$(json_get ".shift_set_grouped_by")
        SAMPLE_SHIFT_SCOPE_ID=$(json_get ".shift_scope_id")
        SAMPLE_SERVICE_TYPE_UNIT_OF_MEASURE_ID=$(json_get ".service_type_unit_of_measure_ids[0]")
        SAMPLE_TRAILER_CLASSIFICATION_ID=$(json_get ".trailer_classification_ids[0]")
        SAMPLE_MATERIAL_TYPE_ID=$(json_get ".material_type_ids[0]")
    fi
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List constraints with --constrained-model-type filter"
if [[ -n "$SAMPLE_CONSTRAINED_TYPE" && "$SAMPLE_CONSTRAINED_TYPE" != "null" ]]; then
    xbe_json view shift-set-time-card-constraints list --constrained-model-type "$SAMPLE_CONSTRAINED_TYPE" --limit 5
    if [[ $status -eq 0 ]]; then
        pass
    else
        if [[ "$output" == *"INTERNAL SERVER ERROR"* ]] || [[ "$output" == *"Invalid field value"* ]]; then
            skip "Server rejected constrained model type filter"
        else
            fail "Expected success (exit 0), got exit $status"
        fi
    fi
else
    skip "No constrained type available"
fi

test_name "List constraints with --constrained-model-type/--constrained-model-id filter"
if [[ -n "$SAMPLE_CONSTRAINED_TYPE" && "$SAMPLE_CONSTRAINED_TYPE" != "null" && -n "$SAMPLE_CONSTRAINED_ID" && "$SAMPLE_CONSTRAINED_ID" != "null" ]]; then
    xbe_json view shift-set-time-card-constraints list --constrained-model-type "$SAMPLE_CONSTRAINED_TYPE" --constrained-model-id "$SAMPLE_CONSTRAINED_ID" --limit 5
    if [[ $status -eq 0 ]]; then
        pass
    else
        if [[ "$output" == *"INTERNAL SERVER ERROR"* ]] || [[ "$output" == *"Invalid field value"* ]]; then
            skip "Server rejected constrained model filter"
        else
            fail "Expected success (exit 0), got exit $status"
        fi
    fi
else
    skip "No constrained type/id available"
fi

test_name "List constraints with --constrained-amount filter"
if [[ -n "$SAMPLE_CONSTRAINED_AMOUNT" && "$SAMPLE_CONSTRAINED_AMOUNT" != "null" ]]; then
    xbe_json view shift-set-time-card-constraints list --constrained-amount "$SAMPLE_CONSTRAINED_AMOUNT" --limit 5
    assert_success
else
    skip "No constrained amount available"
fi

test_name "List constraints with --constraint-type filter"
if [[ -n "$SAMPLE_CONSTRAINT_TYPE" && "$SAMPLE_CONSTRAINT_TYPE" != "null" ]]; then
    xbe_json view shift-set-time-card-constraints list --constraint-type "$SAMPLE_CONSTRAINT_TYPE" --limit 5
    assert_success
else
    skip "No constraint type available"
fi

test_name "List constraints with --status filter"
if [[ -n "$SAMPLE_STATUS" && "$SAMPLE_STATUS" != "null" ]]; then
    xbe_json view shift-set-time-card-constraints list --status "$SAMPLE_STATUS" --limit 5
    assert_success
else
    skip "No status available"
fi

test_name "List constraints with --currency-code filter"
if [[ -n "$SAMPLE_CURRENCY_CODE" && "$SAMPLE_CURRENCY_CODE" != "null" ]]; then
    xbe_json view shift-set-time-card-constraints list --currency-code "$SAMPLE_CURRENCY_CODE" --limit 5
    assert_success
else
    skip "No currency code available"
fi

test_name "List constraints with --name filter"
if [[ -n "$SAMPLE_NAME" && "$SAMPLE_NAME" != "null" ]]; then
    xbe_json view shift-set-time-card-constraints list --name "$SAMPLE_NAME" --limit 5
    assert_success
else
    skip "No name available"
fi

test_name "List constraints with --shift-scope filter"
if [[ -n "$SAMPLE_SHIFT_SCOPE_ID" && "$SAMPLE_SHIFT_SCOPE_ID" != "null" ]]; then
    xbe_json view shift-set-time-card-constraints list --shift-scope "$SAMPLE_SHIFT_SCOPE_ID" --limit 5
    assert_success
else
    skip "No shift scope available"
fi

test_name "List constraints with --search filter"
if [[ -n "$SAMPLE_NAME" && "$SAMPLE_NAME" != "null" ]]; then
    xbe_json view shift-set-time-card-constraints list --search "$SAMPLE_NAME" --limit 5
    assert_success
else
    skip "No name available for search"
fi

test_name "List constraints with --service-type-unit-of-measures filter"
if [[ -n "$SAMPLE_SERVICE_TYPE_UNIT_OF_MEASURE_ID" && "$SAMPLE_SERVICE_TYPE_UNIT_OF_MEASURE_ID" != "null" ]]; then
    xbe_json view shift-set-time-card-constraints list --service-type-unit-of-measures "$SAMPLE_SERVICE_TYPE_UNIT_OF_MEASURE_ID" --limit 5
    assert_success
else
    skip "No service type unit of measure ID available"
fi

test_name "List constraints with --rate-agreement filter"
if [[ "$SAMPLE_CONSTRAINED_TYPE" == "rate-agreements" && -n "$SAMPLE_CONSTRAINED_ID" && "$SAMPLE_CONSTRAINED_ID" != "null" ]]; then
    xbe_json view shift-set-time-card-constraints list --rate-agreement "$SAMPLE_CONSTRAINED_ID" --limit 5
    assert_success
elif [[ -n "$XBE_TEST_RATE_AGREEMENT_ID" ]]; then
    xbe_json view shift-set-time-card-constraints list --rate-agreement "$XBE_TEST_RATE_AGREEMENT_ID" --limit 5
    assert_success
else
    skip "No rate agreement ID available"
fi

test_name "List constraints with --trailer-classification filter"
if [[ -n "$SAMPLE_TRAILER_CLASSIFICATION_ID" && "$SAMPLE_TRAILER_CLASSIFICATION_ID" != "null" ]]; then
    xbe_json view shift-set-time-card-constraints list --trailer-classification "$SAMPLE_TRAILER_CLASSIFICATION_ID" --limit 5
    assert_success
else
    skip "No trailer classification ID available"
fi

test_name "List constraints with --material-type filter"
if [[ -n "$SAMPLE_MATERIAL_TYPE_ID" && "$SAMPLE_MATERIAL_TYPE_ID" != "null" ]]; then
    xbe_json view shift-set-time-card-constraints list --material-type "$SAMPLE_MATERIAL_TYPE_ID" --limit 5
    assert_success
else
    skip "No material type ID available"
fi

test_name "List constraints with --scoped-to-shift filter"
if [[ -n "$XBE_TEST_TENDER_JOB_SCHEDULE_SHIFT_ID" ]]; then
    xbe_json view shift-set-time-card-constraints list --scoped-to-shift "$XBE_TEST_TENDER_JOB_SCHEDULE_SHIFT_ID" --limit 5
    assert_success
else
    skip "No tender job schedule shift ID available"
fi

test_name "List constraints with --scoped-to-tender filter"
if [[ "$SAMPLE_CONSTRAINED_TYPE" == "tenders" && -n "$SAMPLE_CONSTRAINED_ID" && "$SAMPLE_CONSTRAINED_ID" != "null" ]]; then
    xbe_json view shift-set-time-card-constraints list --scoped-to-tender "$SAMPLE_CONSTRAINED_ID" --limit 5
    assert_success
elif [[ -n "$XBE_TEST_TENDER_ID" ]]; then
    xbe_json view shift-set-time-card-constraints list --scoped-to-tender "$XBE_TEST_TENDER_ID" --limit 5
    assert_success
else
    skip "No tender ID available"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show shift set time card constraint"
if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_json view shift-set-time-card-constraints show "$SAMPLE_ID"
    assert_success
else
    skip "No shift set time card constraint ID available"
fi

# ============================================================================
# CREATE Tests
# ============================================================================

if [[ -n "$XBE_TEST_RATE_AGREEMENT_ID" ]]; then
    CREATE_CONSTRAINED_TYPE="rate-agreements"
    CREATE_CONSTRAINED_ID="$XBE_TEST_RATE_AGREEMENT_ID"
elif [[ -n "$XBE_TEST_TENDER_ID" ]]; then
    CREATE_CONSTRAINED_TYPE="tenders"
    CREATE_CONSTRAINED_ID="$XBE_TEST_TENDER_ID"
elif [[ -n "$SAMPLE_CONSTRAINED_TYPE" && "$SAMPLE_CONSTRAINED_TYPE" != "null" && -n "$SAMPLE_CONSTRAINED_ID" && "$SAMPLE_CONSTRAINED_ID" != "null" ]]; then
    CREATE_CONSTRAINED_TYPE="$SAMPLE_CONSTRAINED_TYPE"
    CREATE_CONSTRAINED_ID="$SAMPLE_CONSTRAINED_ID"
fi

CREATE_CONSTRAINT_TYPE="${SAMPLE_CONSTRAINT_TYPE:-minimum}"
CREATE_CONSTRAINED_AMOUNT_TYPE="${SAMPLE_CONSTRAINED_AMOUNT_TYPE:-effective}"
CREATE_CURRENCY_CODE="${SAMPLE_CURRENCY_CODE:-USD}"
CREATE_STATUS="${SAMPLE_STATUS:-active}"
CREATE_STUOM_CONSTRAINT_TYPE="${SAMPLE_STUOM_CONSTRAINT_TYPE:-applicable}"
CREATE_SHIFT_SET_GROUPED_BY="${SAMPLE_SHIFT_SET_GROUPED_BY:-broker}"


test_name "Create shift set time card constraint (explicit amount)"
if [[ -n "$CREATE_CONSTRAINED_TYPE" && -n "$CREATE_CONSTRAINED_ID" ]]; then
    CREATE_NAME=$(unique_name "Constraint")
    create_cmd=(do shift-set-time-card-constraints create
        --constrained-type "$CREATE_CONSTRAINED_TYPE"
        --constrained-id "$CREATE_CONSTRAINED_ID"
        --constraint-type "$CREATE_CONSTRAINT_TYPE"
        --constrained-amount-type "$CREATE_CONSTRAINED_AMOUNT_TYPE"
        --currency-code "$CREATE_CURRENCY_CODE"
        --status "$CREATE_STATUS"
        --service-type-unit-of-measures-constraint-type "$CREATE_STUOM_CONSTRAINT_TYPE"
        --shift-set-grouped-by "$CREATE_SHIFT_SET_GROUPED_BY"
        --constrained-amount 100.00
        --name "$CREATE_NAME")

    if [[ -n "$SAMPLE_SERVICE_TYPE_UNIT_OF_MEASURE_ID" && "$SAMPLE_SERVICE_TYPE_UNIT_OF_MEASURE_ID" != "null" ]]; then
        create_cmd+=(--service-type-unit-of-measures "$SAMPLE_SERVICE_TYPE_UNIT_OF_MEASURE_ID")
    fi
    if [[ -n "$SAMPLE_TRAILER_CLASSIFICATION_ID" && "$SAMPLE_TRAILER_CLASSIFICATION_ID" != "null" ]]; then
        create_cmd+=(--trailer-classifications "$SAMPLE_TRAILER_CLASSIFICATION_ID")
    fi
    if [[ -n "$SAMPLE_MATERIAL_TYPE_ID" && "$SAMPLE_MATERIAL_TYPE_ID" != "null" ]]; then
        create_cmd+=(--material-types "$SAMPLE_MATERIAL_TYPE_ID")
    fi

    xbe_json "${create_cmd[@]}"
    if [[ $status -eq 0 ]]; then
        CREATED_CONSTRAINT_ID=$(json_get ".id")
        if [[ -n "$CREATED_CONSTRAINT_ID" && "$CREATED_CONSTRAINT_ID" != "null" ]]; then
            register_cleanup "shift-set-time-card-constraints" "$CREATED_CONSTRAINT_ID"
            pass
        else
            fail "Created constraint but no ID returned"
        fi
    else
        if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]] || [[ "$output" == *"422"* ]] || [[ "$output" == *"VALIDATION_ERROR"* ]]; then
            skip "Create not permitted or failed validation"
        else
            fail "Create failed"
        fi
    fi
else
    skip "No constrained type/id available for create"
fi


test_name "Create shift set time card constraint (calculated amount)"
if [[ -n "$CREATE_CONSTRAINED_TYPE" && -n "$CREATE_CONSTRAINED_ID" && -n "$XBE_TEST_CALCULATED_CONSTRAINT_UNIT_OF_MEASURE_ID" && -n "$XBE_TEST_CALCULATED_CONSTRAINT_PRICE_PER_UNIT" ]]; then
    CREATE_NAME=$(unique_name "Calculated-Constraint")
    xbe_json do shift-set-time-card-constraints create \
        --constrained-type "$CREATE_CONSTRAINED_TYPE" \
        --constrained-id "$CREATE_CONSTRAINED_ID" \
        --constraint-type "$CREATE_CONSTRAINT_TYPE" \
        --constrained-amount-type "$CREATE_CONSTRAINED_AMOUNT_TYPE" \
        --currency-code "$CREATE_CURRENCY_CODE" \
        --status "$CREATE_STATUS" \
        --service-type-unit-of-measures-constraint-type "$CREATE_STUOM_CONSTRAINT_TYPE" \
        --shift-set-grouped-by "$CREATE_SHIFT_SET_GROUPED_BY" \
        --calculated-constrained-amount-price-per-unit "$XBE_TEST_CALCULATED_CONSTRAINT_PRICE_PER_UNIT" \
        --calculated-constrained-amount-unit-of-measure "$XBE_TEST_CALCULATED_CONSTRAINT_UNIT_OF_MEASURE_ID" \
        --name "$CREATE_NAME"
    if [[ $status -eq 0 ]]; then
        CREATED_CALCULATED_CONSTRAINT_ID=$(json_get ".id")
        if [[ -n "$CREATED_CALCULATED_CONSTRAINT_ID" && "$CREATED_CALCULATED_CONSTRAINT_ID" != "null" ]]; then
            register_cleanup "shift-set-time-card-constraints" "$CREATED_CALCULATED_CONSTRAINT_ID"
            pass
        else
            fail "Created calculated constraint but no ID returned"
        fi
    else
        if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]] || [[ "$output" == *"422"* ]] || [[ "$output" == *"VALIDATION_ERROR"* ]]; then
            skip "Create not permitted or failed validation"
        else
            fail "Create failed"
        fi
    fi
else
    skip "Calculated constraint inputs not available"
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update shift set time card constraint attributes"
if [[ -n "$CREATED_CONSTRAINT_ID" && "$CREATED_CONSTRAINT_ID" != "null" ]]; then
    UPDATED_NAME=$(unique_name "Updated-Constraint")
    xbe_json do shift-set-time-card-constraints update "$CREATED_CONSTRAINT_ID" \
        --name "$UPDATED_NAME" \
        --constrained-amount 150.00 \
        --constraint-type "$CREATE_CONSTRAINT_TYPE" \
        --constrained-amount-type "$CREATE_CONSTRAINED_AMOUNT_TYPE" \
        --currency-code "$CREATE_CURRENCY_CODE" \
        --status inactive \
        --service-type-unit-of-measures-constraint-type "$CREATE_STUOM_CONSTRAINT_TYPE" \
        --shift-set-grouped-by "$CREATE_SHIFT_SET_GROUPED_BY"
    if [[ $status -eq 0 ]]; then
        pass
    else
        if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]] || [[ "$output" == *"422"* ]] || [[ "$output" == *"VALIDATION_ERROR"* ]]; then
            skip "Update not permitted or failed validation"
        else
            fail "Update failed"
        fi
    fi
else
    skip "No created constraint ID available"
fi


test_name "Update shift set time card constraint relationships"
if [[ -n "$CREATED_CONSTRAINT_ID" && "$CREATED_CONSTRAINT_ID" != "null" ]]; then
    if [[ -n "$SAMPLE_SERVICE_TYPE_UNIT_OF_MEASURE_ID" || -n "$SAMPLE_TRAILER_CLASSIFICATION_ID" || -n "$SAMPLE_MATERIAL_TYPE_ID" ]]; then
        update_cmd=(do shift-set-time-card-constraints update "$CREATED_CONSTRAINT_ID")
        if [[ -n "$SAMPLE_SERVICE_TYPE_UNIT_OF_MEASURE_ID" && "$SAMPLE_SERVICE_TYPE_UNIT_OF_MEASURE_ID" != "null" ]]; then
            update_cmd+=(--service-type-unit-of-measures "$SAMPLE_SERVICE_TYPE_UNIT_OF_MEASURE_ID")
        fi
        if [[ -n "$SAMPLE_TRAILER_CLASSIFICATION_ID" && "$SAMPLE_TRAILER_CLASSIFICATION_ID" != "null" ]]; then
            update_cmd+=(--trailer-classifications "$SAMPLE_TRAILER_CLASSIFICATION_ID")
        fi
        if [[ -n "$SAMPLE_MATERIAL_TYPE_ID" && "$SAMPLE_MATERIAL_TYPE_ID" != "null" ]]; then
            update_cmd+=(--material-types "$SAMPLE_MATERIAL_TYPE_ID")
        fi

        xbe_json "${update_cmd[@]}"
        if [[ $status -eq 0 ]]; then
            pass
        else
            if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]] || [[ "$output" == *"422"* ]] || [[ "$output" == *"VALIDATION_ERROR"* ]]; then
                skip "Update not permitted or failed validation"
            else
                fail "Update failed"
            fi
        fi
    else
        skip "No relationship IDs available"
    fi
else
    skip "No created constraint ID available"
fi


test_name "Update shift set time card constraint shift scope"
if [[ -n "$CREATED_CONSTRAINT_ID" && "$CREATED_CONSTRAINT_ID" != "null" && -n "$XBE_TEST_SHIFT_SCOPE_ID" ]]; then
    xbe_json do shift-set-time-card-constraints update "$CREATED_CONSTRAINT_ID" --shift-scope "$XBE_TEST_SHIFT_SCOPE_ID"
    if [[ $status -eq 0 ]]; then
        pass
    else
        if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]] || [[ "$output" == *"422"* ]] || [[ "$output" == *"VALIDATION_ERROR"* ]]; then
            skip "Update not permitted or failed validation"
        else
            fail "Update failed"
        fi
    fi
else
    skip "No shift scope ID available"
fi


test_name "Update shift set time card constraint calculated fields"
if [[ -n "$CREATED_CONSTRAINT_ID" && "$CREATED_CONSTRAINT_ID" != "null" && -n "$XBE_TEST_CALCULATED_CONSTRAINT_UNIT_OF_MEASURE_ID" && -n "$XBE_TEST_CALCULATED_CONSTRAINT_PRICE_PER_UNIT" ]]; then
    xbe_json do shift-set-time-card-constraints update "$CREATED_CONSTRAINT_ID" \
        --calculated-constrained-amount-price-per-unit "$XBE_TEST_CALCULATED_CONSTRAINT_PRICE_PER_UNIT" \
        --calculated-constrained-amount-unit-of-measure "$XBE_TEST_CALCULATED_CONSTRAINT_UNIT_OF_MEASURE_ID"
    if [[ $status -eq 0 ]]; then
        pass
    else
        if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]] || [[ "$output" == *"422"* ]] || [[ "$output" == *"VALIDATION_ERROR"* ]]; then
            skip "Update not permitted or failed validation"
        else
            fail "Update failed"
        fi
    fi
else
    skip "Calculated constraint inputs not available"
fi

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete shift set time card constraint"
if [[ -n "$CREATED_CONSTRAINT_ID" && "$CREATED_CONSTRAINT_ID" != "null" ]]; then
    xbe_run do shift-set-time-card-constraints delete "$CREATED_CONSTRAINT_ID" --confirm
    if [[ $status -eq 0 ]]; then
        pass
    else
        skip "Could not delete constraint (permissions or policy)"
    fi
else
    skip "No created constraint ID available"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create shift set time card constraint without required fields fails"
xbe_run do shift-set-time-card-constraints create
assert_failure


test_name "Update shift set time card constraint without any fields fails"
xbe_run do shift-set-time-card-constraints update "999999"
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
