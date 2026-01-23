#!/bin/bash
#
# XBE CLI Integration Tests: Time Sheet Line Items
#
# COVERAGE: list/show + filters + create/update/delete + error cases
#

source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

describe "Resource: time_sheet_line_items"

SHOW_ID=""
LIST_TIME_SHEET_ID=""
TIME_SHEET_ID=""
COST_CODE_ID=""
CRAFT_CLASS_ID=""
TIME_CARD_ID=""
MAINTENANCE_REQUIREMENT_ID=""
CLASSIFICATION_ID=""
PROJECT_COST_CLASSIFICATION_ID=""
EQUIPMENT_REQUIREMENT_ID=""
EXPLICIT_JOB_PRODUCTION_PLAN_ID=""
CRAFT_CLASS_EFFECTIVE_ID=""
START_AT=""
END_AT=""
BROKER_ID=""
TRUCKER_ID=""
CUSTOMER_ID=""
CREATED_LINE_ITEM_ID=""

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List time sheet line items"
xbe_json view time-sheet-line-items list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
    SHOW_ID=$(json_get '.[0].id // empty')
    LIST_TIME_SHEET_ID=$(json_get '.[0].time_sheet_id // empty')
    TIME_SHEET_ID="$LIST_TIME_SHEET_ID"
    COST_CODE_ID=$(json_get '.[0].cost_code_id // empty')
    CRAFT_CLASS_ID=$(json_get '.[0].craft_class_id // empty')
    TIME_CARD_ID=$(json_get '.[0].time_card_id // empty')
    MAINTENANCE_REQUIREMENT_ID=$(json_get '.[0].maintenance_requirement_id // empty')
    CLASSIFICATION_ID=$(json_get '.[0].time_sheet_line_item_classification_id // empty')
    PROJECT_COST_CLASSIFICATION_ID=$(json_get '.[0].project_cost_classification_id // empty')
    EQUIPMENT_REQUIREMENT_ID=$(json_get '.[0].equipment_requirement_id // empty')
    EXPLICIT_JOB_PRODUCTION_PLAN_ID=$(json_get '.[0].explicit_job_production_plan_id // empty')
    CRAFT_CLASS_EFFECTIVE_ID=$(json_get '.[0].craft_class_effective_id // empty')
    START_AT=$(json_get '.[0].start_at // empty')
    END_AT=$(json_get '.[0].end_at // empty')
else
    fail "Failed to list time sheet line items"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show time sheet line item"
if [[ -n "$SHOW_ID" && "$SHOW_ID" != "null" ]]; then
    xbe_json view time-sheet-line-items show "$SHOW_ID"
    assert_success
else
    skip "No time sheet line item ID available"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List time sheet line items with --time-sheet filter"
if [[ -n "$LIST_TIME_SHEET_ID" && "$LIST_TIME_SHEET_ID" != "null" ]]; then
    xbe_json view time-sheet-line-items list --time-sheet "$LIST_TIME_SHEET_ID" --limit 5
    assert_success
else
    skip "No time sheet ID available"
fi

test_name "List time sheet line items with --cost-code filter"
if [[ -n "$COST_CODE_ID" && "$COST_CODE_ID" != "null" ]]; then
    xbe_json view time-sheet-line-items list --cost-code "$COST_CODE_ID" --limit 5
    assert_success
else
    skip "No cost code ID available"
fi

test_name "List time sheet line items with --craft-class filter"
if [[ -n "$CRAFT_CLASS_ID" && "$CRAFT_CLASS_ID" != "null" ]]; then
    xbe_json view time-sheet-line-items list --craft-class "$CRAFT_CLASS_ID" --limit 5
    assert_success
else
    skip "No craft class ID available"
fi

test_name "List time sheet line items with --time-card filter"
if [[ -n "$TIME_CARD_ID" && "$TIME_CARD_ID" != "null" ]]; then
    xbe_json view time-sheet-line-items list --time-card "$TIME_CARD_ID" --limit 5
    assert_success
else
    skip "No time card ID available"
fi

test_name "List time sheet line items with --maintenance-requirement filter"
if [[ -n "$MAINTENANCE_REQUIREMENT_ID" && "$MAINTENANCE_REQUIREMENT_ID" != "null" ]]; then
    xbe_json view time-sheet-line-items list --maintenance-requirement "$MAINTENANCE_REQUIREMENT_ID" --limit 5
    assert_success
else
    skip "No maintenance requirement ID available"
fi

test_name "List time sheet line items with --craft-class-effective filter"
if [[ -n "$CRAFT_CLASS_EFFECTIVE_ID" && "$CRAFT_CLASS_EFFECTIVE_ID" != "null" ]]; then
    xbe_json view time-sheet-line-items list --craft-class-effective "$CRAFT_CLASS_EFFECTIVE_ID" --limit 5
    assert_success
else
    skip "No craft class effective ID available"
fi

test_name "List time sheet line items with --craft-class-effective-id filter"
if [[ -n "$CRAFT_CLASS_EFFECTIVE_ID" && "$CRAFT_CLASS_EFFECTIVE_ID" != "null" ]]; then
    xbe_json view time-sheet-line-items list --craft-class-effective-id "$CRAFT_CLASS_EFFECTIVE_ID" --limit 5
    assert_success
else
    skip "No craft class effective ID available"
fi

test_name "List time sheet line items with --start-at-min filter"
if [[ -n "$START_AT" && "$START_AT" != "null" ]]; then
    xbe_json view time-sheet-line-items list --start-at-min "$START_AT" --limit 5
    assert_success
else
    skip "No start_at value available"
fi

test_name "List time sheet line items with --end-at-max filter"
if [[ -n "$END_AT" && "$END_AT" != "null" ]]; then
    xbe_json view time-sheet-line-items list --end-at-max "$END_AT" --limit 5
    assert_success
else
    skip "No end_at value available"
fi

test_name "List time sheet line items with --broker filter"
if [[ -n "$BROKER_ID" && "$BROKER_ID" != "null" ]]; then
    xbe_json view time-sheet-line-items list --broker "$BROKER_ID" --limit 5
    assert_success
else
    skip "No broker ID available"
fi

test_name "List time sheet line items with --trucker filter"
if [[ -n "$TRUCKER_ID" && "$TRUCKER_ID" != "null" ]]; then
    xbe_json view time-sheet-line-items list --trucker "$TRUCKER_ID" --limit 5
    assert_success
else
    skip "No trucker ID available"
fi

test_name "List time sheet line items with --customer filter"
if [[ -n "$CUSTOMER_ID" && "$CUSTOMER_ID" != "null" ]]; then
    xbe_json view time-sheet-line-items list --customer "$CUSTOMER_ID" --limit 5
    assert_success
else
    skip "No customer ID available"
fi

test_name "List time sheet line items with --is-start-at filter"
if [[ -n "$START_AT" && "$START_AT" != "null" ]]; then
    filter_val="true"
else
    filter_val="false"
fi
xbe_json view time-sheet-line-items list --is-start-at "$filter_val" --limit 5
assert_success

test_name "List time sheet line items with --is-end-at filter"
if [[ -n "$END_AT" && "$END_AT" != "null" ]]; then
    filter_val="true"
else
    filter_val="false"
fi
xbe_json view time-sheet-line-items list --is-end-at "$filter_val" --limit 5
assert_success

# ============================================================================
# API Lookup (optional)
# ============================================================================

test_name "Lookup time sheet line item dependencies via API"
if [[ -z "$XBE_TOKEN" ]]; then
    skip "XBE_TOKEN not set; skipping API lookup"
else
    base_url="${XBE_BASE_URL%/}"

    if [[ -z "$TIME_SHEET_ID" || "$TIME_SHEET_ID" == "null" ]]; then
        time_sheets_json=$(curl -s \
            -H "Authorization: Bearer $XBE_TOKEN" \
            -H "Accept: application/vnd.api+json" \
            "$base_url/v1/time-sheets?page[limit]=1&filter[status]=editing" || true)
        TIME_SHEET_ID=$(echo "$time_sheets_json" | jq -r '.data[0].id // empty' 2>/dev/null || true)
    fi

    if [[ -z "$BROKER_ID" || "$BROKER_ID" == "null" ]]; then
        brokers_json=$(curl -s \
            -H "Authorization: Bearer $XBE_TOKEN" \
            -H "Accept: application/vnd.api+json" \
            "$base_url/v1/brokers?page[limit]=1" || true)
        BROKER_ID=$(echo "$brokers_json" | jq -r '.data[0].id // empty' 2>/dev/null || true)
    fi

    if [[ -z "$TRUCKER_ID" || "$TRUCKER_ID" == "null" ]]; then
        truckers_json=$(curl -s \
            -H "Authorization: Bearer $XBE_TOKEN" \
            -H "Accept: application/vnd.api+json" \
            "$base_url/v1/truckers?page[limit]=1" || true)
        TRUCKER_ID=$(echo "$truckers_json" | jq -r '.data[0].id // empty' 2>/dev/null || true)
    fi

    if [[ -z "$CUSTOMER_ID" || "$CUSTOMER_ID" == "null" ]]; then
        customers_json=$(curl -s \
            -H "Authorization: Bearer $XBE_TOKEN" \
            -H "Accept: application/vnd.api+json" \
            "$base_url/v1/customers?page[limit]=1" || true)
        CUSTOMER_ID=$(echo "$customers_json" | jq -r '.data[0].id // empty' 2>/dev/null || true)
    fi

    pass
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create time sheet line item without required fields fails"
xbe_run do time-sheet-line-items create
assert_failure

test_name "Create time sheet line item with required fields"
if [[ -n "$TIME_SHEET_ID" && "$TIME_SHEET_ID" != "null" ]]; then
    xbe_json do time-sheet-line-items create \
        --time-sheet "$TIME_SHEET_ID" \
        --start-at "2000-01-01T00:00:00Z" \
        --end-at "2000-01-01T01:00:00Z" \
        --break-minutes 10 \
        --description "CLI test line item" \
        --is-non-job-line-item false
    if [[ $status -eq 0 ]]; then
        CREATED_LINE_ITEM_ID=$(json_get ".id")
        register_cleanup "time-sheet-line-items" "$CREATED_LINE_ITEM_ID"
        assert_json_has ".id"
    else
        if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]] || [[ "$output" == *"403"* ]] || [[ "$output" == *"422"* ]]; then
            skip "Not authorized or validation failed for create"
        else
            fail "Failed to create time sheet line item"
        fi
    fi
else
    skip "No time sheet ID available for create"
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update time sheet line item without fields fails"
if [[ -n "$SHOW_ID" && "$SHOW_ID" != "null" ]]; then
    xbe_run do time-sheet-line-items update "$SHOW_ID"
    assert_failure
else
    skip "No time sheet line item ID available"
fi

test_name "Update time sheet line item description"
if [[ -n "$CREATED_LINE_ITEM_ID" && "$CREATED_LINE_ITEM_ID" != "null" ]]; then
    xbe_json do time-sheet-line-items update "$CREATED_LINE_ITEM_ID" --description "Updated description"
    assert_success
else
    skip "No created time sheet line item ID available"
fi

test_name "Update time sheet line item break minutes"
if [[ -n "$CREATED_LINE_ITEM_ID" && "$CREATED_LINE_ITEM_ID" != "null" ]]; then
    xbe_json do time-sheet-line-items update "$CREATED_LINE_ITEM_ID" --break-minutes 20
    assert_success
else
    skip "No created time sheet line item ID available"
fi

test_name "Update time sheet line item flags"
if [[ -n "$CREATED_LINE_ITEM_ID" && "$CREATED_LINE_ITEM_ID" != "null" ]]; then
    xbe_json do time-sheet-line-items update "$CREATED_LINE_ITEM_ID" --skip-validate-overlap true --is-non-job-line-item true
    assert_success
else
    skip "No created time sheet line item ID available"
fi

test_name "Update time sheet line item cost code"
if [[ -n "$CREATED_LINE_ITEM_ID" && "$CREATED_LINE_ITEM_ID" != "null" && "$TIME_SHEET_ID" == "$LIST_TIME_SHEET_ID" && -n "$COST_CODE_ID" && "$COST_CODE_ID" != "null" ]]; then
    xbe_json do time-sheet-line-items update "$CREATED_LINE_ITEM_ID" --cost-code "$COST_CODE_ID"
    assert_success
else
    skip "No cost code ID available for update"
fi

test_name "Update time sheet line item craft class"
if [[ -n "$CREATED_LINE_ITEM_ID" && "$CREATED_LINE_ITEM_ID" != "null" && "$TIME_SHEET_ID" == "$LIST_TIME_SHEET_ID" && -n "$CRAFT_CLASS_ID" && "$CRAFT_CLASS_ID" != "null" ]]; then
    xbe_json do time-sheet-line-items update "$CREATED_LINE_ITEM_ID" --craft-class "$CRAFT_CLASS_ID"
    assert_success
else
    skip "No craft class ID available for update"
fi

test_name "Update time sheet line item equipment requirement"
if [[ -n "$CREATED_LINE_ITEM_ID" && "$CREATED_LINE_ITEM_ID" != "null" && "$TIME_SHEET_ID" == "$LIST_TIME_SHEET_ID" && -n "$EQUIPMENT_REQUIREMENT_ID" && "$EQUIPMENT_REQUIREMENT_ID" != "null" ]]; then
    xbe_json do time-sheet-line-items update "$CREATED_LINE_ITEM_ID" --equipment-requirement "$EQUIPMENT_REQUIREMENT_ID"
    assert_success
else
    skip "No equipment requirement ID available for update"
fi

test_name "Update time sheet line item maintenance requirement"
if [[ -n "$CREATED_LINE_ITEM_ID" && "$CREATED_LINE_ITEM_ID" != "null" && "$TIME_SHEET_ID" == "$LIST_TIME_SHEET_ID" && -n "$MAINTENANCE_REQUIREMENT_ID" && "$MAINTENANCE_REQUIREMENT_ID" != "null" ]]; then
    xbe_json do time-sheet-line-items update "$CREATED_LINE_ITEM_ID" --maintenance-requirement "$MAINTENANCE_REQUIREMENT_ID"
    assert_success
else
    skip "No maintenance requirement ID available for update"
fi

test_name "Update time sheet line item classification"
if [[ -n "$CREATED_LINE_ITEM_ID" && "$CREATED_LINE_ITEM_ID" != "null" && "$TIME_SHEET_ID" == "$LIST_TIME_SHEET_ID" && -n "$CLASSIFICATION_ID" && "$CLASSIFICATION_ID" != "null" ]]; then
    xbe_json do time-sheet-line-items update "$CREATED_LINE_ITEM_ID" --time-sheet-line-item-classification "$CLASSIFICATION_ID"
    assert_success
else
    skip "No classification ID available for update"
fi

test_name "Update time sheet line item project cost classification"
if [[ -n "$CREATED_LINE_ITEM_ID" && "$CREATED_LINE_ITEM_ID" != "null" && "$TIME_SHEET_ID" == "$LIST_TIME_SHEET_ID" && -n "$PROJECT_COST_CLASSIFICATION_ID" && "$PROJECT_COST_CLASSIFICATION_ID" != "null" ]]; then
    xbe_json do time-sheet-line-items update "$CREATED_LINE_ITEM_ID" --project-cost-classification "$PROJECT_COST_CLASSIFICATION_ID"
    assert_success
else
    skip "No project cost classification ID available for update"
fi

test_name "Update time sheet line item explicit job production plan"
if [[ -n "$CREATED_LINE_ITEM_ID" && "$CREATED_LINE_ITEM_ID" != "null" && "$TIME_SHEET_ID" == "$LIST_TIME_SHEET_ID" && -n "$EXPLICIT_JOB_PRODUCTION_PLAN_ID" && "$EXPLICIT_JOB_PRODUCTION_PLAN_ID" != "null" ]]; then
    xbe_json do time-sheet-line-items update "$CREATED_LINE_ITEM_ID" --explicit-job-production-plan "$EXPLICIT_JOB_PRODUCTION_PLAN_ID"
    assert_success
else
    skip "No explicit job production plan ID available for update"
fi

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete time sheet line item requires --confirm"
if [[ -n "$CREATED_LINE_ITEM_ID" && "$CREATED_LINE_ITEM_ID" != "null" ]]; then
    xbe_run do time-sheet-line-items delete "$CREATED_LINE_ITEM_ID"
    assert_failure
else
    skip "No created time sheet line item ID available"
fi

test_name "Delete time sheet line item with --confirm"
if [[ -n "$CREATED_LINE_ITEM_ID" && "$CREATED_LINE_ITEM_ID" != "null" ]]; then
    xbe_run do time-sheet-line-items delete "$CREATED_LINE_ITEM_ID" --confirm
    assert_success
else
    skip "No created time sheet line item ID available"
fi

run_tests
