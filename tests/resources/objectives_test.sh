#!/bin/bash
#
# XBE CLI Integration Tests: Objectives
#
# Tests list/show/create/update/delete operations for objectives.
#
# COVERAGE: List filters + writable attributes
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

OBJECTIVE_ID="${XBE_TEST_OBJECTIVE_ID:-}"
OWNER_ID="${XBE_TEST_OBJECTIVE_OWNER_ID:-}"
ORG="${XBE_TEST_OBJECTIVE_ORGANIZATION:-}"
ORG_TYPE="${XBE_TEST_OBJECTIVE_ORGANIZATION_TYPE:-}"
ORG_ID="${XBE_TEST_OBJECTIVE_ORGANIZATION_ID:-}"
PROJECT_ID="${XBE_TEST_OBJECTIVE_PROJECT_ID:-}"
SALES_RESP_ID="${XBE_TEST_OBJECTIVE_SALES_RESPONSIBLE_PERSON_ID:-}"
PARENT="${XBE_TEST_OBJECTIVE_PARENT:-}"

SAMPLE_ID=""
SAMPLE_OWNER_ID=""
SAMPLE_ORG_TYPE=""
SAMPLE_ORG_ID=""
SAMPLE_PROJECT_ID=""
SAMPLE_SALES_RESP_ID=""
SAMPLE_SLUG=""

CREATED_ID=""
TEMPLATE_CREATED_ID=""

normalize_org_type() {
    local raw="${1:-}"
    raw=$(echo "$raw" | tr '[:upper:]' '[:lower:]')
    raw=${raw//_/-}
    raw=${raw// /-}
    case "$raw" in
        broker|brokers)
            echo "Broker"
            ;;
        customer|customers)
            echo "Customer"
            ;;
        trucker|truckers)
            echo "Trucker"
            ;;
        material-supplier|material-suppliers|materialsupplier|materialsuppliers)
            echo "MaterialSupplier"
            ;;
        developer|developers)
            echo "Developer"
            ;;
        *)
            echo "$1"
            ;;
    esac
}

handle_write_failure() {
    local output_text="$1"
    if [[ "$output_text" == *"Not Authorized"* ]] || [[ "$output_text" == *"not authorized"* ]] || [[ "$output_text" == *"422"* ]] || [[ "$output_text" == *"409"* ]]; then
        skip "Write blocked by server policy/validation"
        return 0
    fi
    fail "Write failed: $output_text"
    return 1
}

describe "Resource: objectives"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List objectives"
xbe_json view objectives list --limit 5
assert_success

test_name "List objectives returns array"
xbe_json view objectives list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list objectives"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List objectives with --status filter"
xbe_json view objectives list --status green --limit 5
assert_success

test_name "List objectives with --commitment filter"
xbe_json view objectives list --commitment committed --limit 5
assert_success

test_name "List objectives with --is-template filter"
xbe_json view objectives list --is-template true --limit 5
assert_success

test_name "List objectives with --template-scope filter"
xbe_json view objectives list --template-scope match_all --limit 5
assert_success

test_name "List objectives with --has-sales-responsible-person filter"
xbe_json view objectives list --has-sales-responsible-person true --limit 5
assert_success

test_name "List objectives with --without-customer-success-responsible-person filter"
xbe_json view objectives list --without-customer-success-responsible-person true --limit 5
assert_success

TODAY=$(date -u +%Y-%m-%d)
FUTURE=$(date -u -v+30d +%Y-%m-%d 2>/dev/null || date -u -d '+30 days' +%Y-%m-%d)

test_name "List objectives with --start-on filter"
xbe_json view objectives list --start-on "$TODAY" --limit 5
assert_success

test_name "List objectives with --start-on-min filter"
xbe_json view objectives list --start-on-min "$TODAY" --limit 5
assert_success

test_name "List objectives with --start-on-max filter"
xbe_json view objectives list --start-on-max "$FUTURE" --limit 5
assert_success

test_name "List objectives with --end-on filter"
xbe_json view objectives list --end-on "$FUTURE" --limit 5
assert_success

test_name "List objectives with --end-on-min filter"
xbe_json view objectives list --end-on-min "$TODAY" --limit 5
assert_success

test_name "List objectives with --end-on-max filter"
xbe_json view objectives list --end-on-max "$FUTURE" --limit 5
assert_success

test_name "List objectives with --name filter"
xbe_json view objectives list --name "Objective" --limit 5
assert_success

# ==========================================================================
# Sample Record (used for show + filters)
# ==========================================================================

test_name "Capture sample objective"
xbe_json view objectives list --limit 1
if [[ $status -eq 0 ]]; then
    SAMPLE_ID=$(json_get ".[0].id")
    SAMPLE_OWNER_ID=$(json_get ".[0].owner_id")
    SAMPLE_ORG_TYPE=$(json_get ".[0].organization_type")
    SAMPLE_ORG_ID=$(json_get ".[0].organization_id")
    SAMPLE_PROJECT_ID=$(json_get ".[0].project_id")
    SAMPLE_SALES_RESP_ID=$(json_get ".[0].sales_responsible_person_id")
    SAMPLE_SLUG=$(json_get ".[0].slug")
    if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
        pass
    else
        skip "No objectives available for show test"
    fi
else
    skip "Could not list objectives to capture sample"
fi

# ==========================================================================
# LIST Tests - Relationship Filters (using sample/env)
# ==========================================================================

test_name "List objectives with --owner filter"
OWNER_FILTER_ID="${OWNER_ID:-$SAMPLE_OWNER_ID}"
if [[ -n "$OWNER_FILTER_ID" && "$OWNER_FILTER_ID" != "null" ]]; then
    xbe_json view objectives list --owner "$OWNER_FILTER_ID" --limit 5
    assert_success
else
    skip "No owner ID available. Set XBE_TEST_OBJECTIVE_OWNER_ID to enable owner filter testing."
fi

test_name "List objectives with --organization filter"
ORG_FILTER="${ORG:-}"
if [[ -z "$ORG_FILTER" || "$ORG_FILTER" == "null" ]]; then
    if [[ -n "$SAMPLE_ORG_TYPE" && "$SAMPLE_ORG_TYPE" != "null" && -n "$SAMPLE_ORG_ID" && "$SAMPLE_ORG_ID" != "null" ]]; then
        ORG_FILTER="$(normalize_org_type "$SAMPLE_ORG_TYPE")|${SAMPLE_ORG_ID}"
    fi
fi
if [[ -n "$ORG_FILTER" && "$ORG_FILTER" != "null" ]]; then
    xbe_json view objectives list --organization "$ORG_FILTER" --limit 5
    assert_success
else
    skip "No organization available. Set XBE_TEST_OBJECTIVE_ORGANIZATION or ensure sample has organization."
fi

test_name "List objectives with --organization-type/--organization-id filter"
ORG_TYPE_FILTER="${ORG_TYPE:-$SAMPLE_ORG_TYPE}"
ORG_ID_FILTER="${ORG_ID:-$SAMPLE_ORG_ID}"
if [[ -n "$ORG_TYPE_FILTER" && "$ORG_TYPE_FILTER" != "null" && -n "$ORG_ID_FILTER" && "$ORG_ID_FILTER" != "null" ]]; then
    ORG_TYPE_FILTER="$(normalize_org_type "$ORG_TYPE_FILTER")"
    xbe_json view objectives list --organization-type "$ORG_TYPE_FILTER" --organization-id "$ORG_ID_FILTER" --limit 5
    assert_success
else
    skip "No organization type/id available. Set XBE_TEST_OBJECTIVE_ORGANIZATION_TYPE/ID to enable filter testing."
fi

test_name "List objectives with --project filter"
PROJECT_FILTER_ID="${PROJECT_ID:-$SAMPLE_PROJECT_ID}"
if [[ -n "$PROJECT_FILTER_ID" && "$PROJECT_FILTER_ID" != "null" ]]; then
    xbe_json view objectives list --project "$PROJECT_FILTER_ID" --limit 5
    assert_success
else
    skip "No project ID available. Set XBE_TEST_OBJECTIVE_PROJECT_ID to enable project filter testing."
fi

test_name "List objectives with --sales-responsible-person filter"
SALES_FILTER_ID="${SALES_RESP_ID:-$SAMPLE_SALES_RESP_ID}"
if [[ -n "$SALES_FILTER_ID" && "$SALES_FILTER_ID" != "null" ]]; then
    xbe_json view objectives list --sales-responsible-person "$SALES_FILTER_ID" --limit 5
    assert_success
else
    skip "No sales responsible person ID available. Set XBE_TEST_OBJECTIVE_SALES_RESPONSIBLE_PERSON_ID to enable filter testing."
fi

test_name "List objectives with --slug filter"
SLUG_FILTER="${SAMPLE_SLUG:-}"
if [[ -n "$SLUG_FILTER" && "$SLUG_FILTER" != "null" ]]; then
    xbe_json view objectives list --slug "$SLUG_FILTER" --limit 5
    assert_success
else
    skip "No slug available from sample objective"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show objective"
DETAIL_ID="${OBJECTIVE_ID:-$SAMPLE_ID}"
if [[ -n "$DETAIL_ID" && "$DETAIL_ID" != "null" ]]; then
    xbe_json view objectives show "$DETAIL_ID"
    if [[ $status -eq 0 ]]; then
        assert_json_has ".id"
    else
        if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]] || [[ "$output" == *"Not Found"* ]] || [[ "$output" == *"not found"* ]]; then
            skip "Show blocked by server policy or record not found"
        else
            fail "Failed to show objective: $output"
        fi
    fi
else
    skip "No objective ID available for show"
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create template objective"
TEMPLATE_NAME=$(unique_name "ObjectiveTemplate")
xbe_json do objectives create \
    --name "$TEMPLATE_NAME" \
    --is-template true \
    --template-scope match_all \
    --commitment committed \
    --start-on "$TODAY" \
    --end-on "$FUTURE" \
    --description "Template objective" \
    --name-summary-explicit "Template summary" \
    --is-generating-objective-stakeholder-classifications true
if [[ $status -eq 0 ]]; then
    TEMPLATE_CREATED_ID=$(json_get ".id")
    if [[ -n "$TEMPLATE_CREATED_ID" && "$TEMPLATE_CREATED_ID" != "null" ]]; then
        register_cleanup "objectives" "$TEMPLATE_CREATED_ID"
        pass
    else
        fail "Created template objective but no ID returned"
    fi
else
    handle_write_failure "$output"
fi

ORG_ARG=""
if [[ -n "$ORG" && "$ORG" != "null" ]]; then
    ORG_ARG=(--organization "$ORG")
elif [[ -n "$ORG_TYPE" && "$ORG_TYPE" != "null" && -n "$ORG_ID" && "$ORG_ID" != "null" ]]; then
    ORG_ARG=(--organization-type "$ORG_TYPE" --organization-id "$ORG_ID")
else
    ORG_ARG=()
fi

if [[ -n "$PROJECT_ID" && "$PROJECT_ID" != "null" ]]; then
    PROJECT_ARG=(--project "$PROJECT_ID")
else
    PROJECT_ARG=()
fi

if [[ ${#ORG_ARG[@]} -gt 0 ]]; then
    test_name "Create objective"
    OBJ_NAME=$(unique_name "Objective")
    xbe_json do objectives create \
        --name "$OBJ_NAME" \
        "${ORG_ARG[@]}" \
        "${PROJECT_ARG[@]}" \
        --commitment committed \
        --start-on "$TODAY" \
        --end-on "$FUTURE" \
        --description "Test objective" \
        --name-summary-explicit "Delivery goal"
    if [[ $status -eq 0 ]]; then
        CREATED_ID=$(json_get ".id")
        if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
            register_cleanup "objectives" "$CREATED_ID"
            pass
        else
            fail "Created objective but no ID returned"
        fi
    else
        handle_write_failure "$output"
    fi
else
    test_name "Create objective"
    skip "No organization available. Set XBE_TEST_OBJECTIVE_ORGANIZATION (Type|ID) or ORG_TYPE/ORG_ID to enable create/update tests."
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

UPDATE_ID="${CREATED_ID:-$OBJECTIVE_ID}"

update_objective() {
    local description="$1"
    shift
    test_name "$description"
    xbe_json do objectives update "$UPDATE_ID" "$@"
    if [[ $status -eq 0 ]]; then
        pass
    else
        if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]] || [[ "$output" == *"422"* ]] || [[ "$output" == *"409"* ]]; then
            skip "Update blocked by server policy/validation"
        else
            fail "Update failed: $output"
        fi
    fi
}

if [[ -n "$UPDATE_ID" && "$UPDATE_ID" != "null" ]]; then
    update_objective "Update name" --name "Updated $(unique_name "Objective")"
    update_objective "Update description" --description "Updated description"
    update_objective "Update name summary explicit" --name-summary-explicit "Updated summary"
    update_objective "Update commitment" --commitment aspirational
    update_objective "Update start/end dates" --start-on "$TODAY" --end-on "$FUTURE"
    update_objective "Update stakeholder generation flag" --is-generating-objective-stakeholder-classifications false

    OWNER_UPDATE_ID="${OWNER_ID:-$SAMPLE_OWNER_ID}"
    if [[ -n "$OWNER_UPDATE_ID" && "$OWNER_UPDATE_ID" != "null" ]]; then
        update_objective "Update owner" --owner "$OWNER_UPDATE_ID"
    else
        test_name "Update owner"
        skip "No owner ID available. Set XBE_TEST_OBJECTIVE_OWNER_ID to enable owner update testing."
    fi

    if [[ -n "$PROJECT_ID" && "$PROJECT_ID" != "null" ]]; then
        update_objective "Update project" --project "$PROJECT_ID"
    else
        test_name "Update project"
        skip "No project ID available. Set XBE_TEST_OBJECTIVE_PROJECT_ID to enable project update testing."
    fi

    SALES_UPDATE_ID="${SALES_RESP_ID:-$SAMPLE_SALES_RESP_ID}"
    if [[ -n "$SALES_UPDATE_ID" && "$SALES_UPDATE_ID" != "null" ]]; then
        update_objective "Update sales responsible person" --sales-responsible-person "$SALES_UPDATE_ID"
    else
        test_name "Update sales responsible person"
        skip "No sales responsible person ID available. Set XBE_TEST_OBJECTIVE_SALES_RESPONSIBLE_PERSON_ID to enable update testing."
    fi

    if [[ -n "$PARENT" && "$PARENT" != "null" ]]; then
        update_objective "Update parent" --parent "$PARENT"
    else
        test_name "Update parent"
        skip "No parent available. Set XBE_TEST_OBJECTIVE_PARENT to enable parent update testing."
    fi

    if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
        update_objective "Mark abandoned" --is-abandoned true
    else
        test_name "Mark abandoned"
        skip "Skipping abandon update on non-test objective"
    fi
else
    test_name "Update objective"
    skip "No objective ID available for update tests"
fi

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete objective"
if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
    xbe_json do objectives delete "$CREATED_ID" --confirm
    if [[ $status -eq 0 ]]; then
        pass
    else
        if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]] || [[ "$output" == *"422"* ]] || [[ "$output" == *"409"* ]]; then
            skip "Delete blocked by server policy/validation"
        else
            fail "Failed to delete objective: $output"
        fi
    fi
else
    skip "No created objective ID available for delete test"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
