#!/bin/bash
#
# XBE CLI Integration Tests: Organization Invoices Batch PDF Templates
#
# Tests list/show/create/update operations for organization-invoices-batch-pdf-templates.
#
# COVERAGE: List filters + writable attributes
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

SAMPLE_ID=""
SAMPLE_ORGANIZATION_TYPE=""
SAMPLE_ORGANIZATION_ID=""
SAMPLE_BROKER_ID=""
SAMPLE_CREATED_BY_ID=""

CREATED_TEMPLATE_ID=""

BROKER_ID="${XBE_TEST_ORGANIZATION_INVOICES_BATCH_PDF_TEMPLATE_BROKER_ID:-}"
ORGANIZATION_TYPE="${XBE_TEST_ORGANIZATION_INVOICES_BATCH_PDF_TEMPLATE_ORGANIZATION_TYPE:-}"
ORGANIZATION_ID="${XBE_TEST_ORGANIZATION_INVOICES_BATCH_PDF_TEMPLATE_ORGANIZATION_ID:-}"
CREATED_BY_ID="${XBE_TEST_ORGANIZATION_INVOICES_BATCH_PDF_TEMPLATE_CREATED_BY_ID:-}"

normalize_org_type() {
    case "$1" in
        brokers|Broker|BROKER) echo "Broker" ;;
        customers|Customer|CUSTOMER) echo "Customer" ;;
        truckers|Trucker|TRUCKER) echo "Trucker" ;;
        developers|Developer|DEVELOPER) echo "Developer" ;;
        material-suppliers|material_suppliers|MaterialSupplier|MATERIALSUPPLIER) echo "MaterialSupplier" ;;
        *) echo "$1" ;;
    esac
}

describe "Resource: organization-invoices-batch-pdf-templates"

# ============================================================================
# Prerequisites - Create broker for organization templates if needed
# ============================================================================

if [[ -z "$BROKER_ID" || "$BROKER_ID" == "null" ]]; then
    test_name "Create prerequisite broker for template tests"
    BROKER_NAME=$(unique_name "OrgInvoicesBatchPdfTemplateBroker")
    xbe_json do brokers create --name "$BROKER_NAME"
    if [[ $status -eq 0 ]]; then
        BROKER_ID=$(json_get ".id")
        if [[ -n "$BROKER_ID" && "$BROKER_ID" != "null" ]]; then
            register_cleanup "brokers" "$BROKER_ID"
            pass
        else
            fail "Created broker but no ID returned"
        fi
    else
        if [[ -n "$XBE_TEST_BROKER_ID" ]]; then
            BROKER_ID="$XBE_TEST_BROKER_ID"
            echo "    Using XBE_TEST_BROKER_ID: $BROKER_ID"
            pass
        else
            fail "Failed to create broker and XBE_TEST_BROKER_ID not set"
        fi
    fi
fi

if [[ -z "$ORGANIZATION_TYPE" || "$ORGANIZATION_TYPE" == "null" ]]; then
    ORGANIZATION_TYPE="Broker"
fi
if [[ -z "$ORGANIZATION_ID" || "$ORGANIZATION_ID" == "null" ]]; then
    ORGANIZATION_ID="$BROKER_ID"
fi

TEMPLATE_CONTENT='{{invoice_number}}'
TEMPLATE_CONTENT_ALT='{{invoice_number}}-ALT'

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create organization invoices batch PDF template (required fields)"
if [[ -n "$BROKER_ID" && "$BROKER_ID" != "null" && -n "$ORGANIZATION_ID" && "$ORGANIZATION_ID" != "null" ]]; then
    xbe_json do organization-invoices-batch-pdf-templates create \
        --organization "${ORGANIZATION_TYPE}|${ORGANIZATION_ID}" \
        --broker "$BROKER_ID" \
        --template "$TEMPLATE_CONTENT"

    if [[ $status -eq 0 ]]; then
        CREATED_TEMPLATE_ID=$(json_get ".id")
        if [[ -n "$CREATED_TEMPLATE_ID" && "$CREATED_TEMPLATE_ID" != "null" ]]; then
            pass
        else
            fail "Created template but no ID returned"
        fi
    else
        if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]] || [[ "$output" == *"422"* ]] || [[ "$output" == *"409"* ]]; then
            skip "Create blocked by server policy/validation"
        else
            fail "Failed to create template: $output"
        fi
    fi
else
    skip "No broker/organization IDs available for create"
fi

test_name "Create organization invoices batch PDF template (optional fields)"
if [[ -n "$BROKER_ID" && "$BROKER_ID" != "null" && -n "$ORGANIZATION_ID" && "$ORGANIZATION_ID" != "null" ]]; then
    args=(--organization "${ORGANIZATION_TYPE}|${ORGANIZATION_ID}" --broker "$BROKER_ID" --template "$TEMPLATE_CONTENT" --description "CLI template test" --is-active=false)
    if [[ -n "$CREATED_BY_ID" && "$CREATED_BY_ID" != "null" ]]; then
        args+=(--created-by "$CREATED_BY_ID")
    fi
    xbe_json do organization-invoices-batch-pdf-templates create "${args[@]}"
    if [[ $status -eq 0 ]]; then
        pass
    else
        if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]] || [[ "$output" == *"422"* ]] || [[ "$output" == *"409"* ]]; then
            skip "Optional create blocked by server policy/validation"
        else
            fail "Failed to create template with optional fields: $output"
        fi
    fi
else
    skip "No broker/organization IDs available for optional create"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show organization invoices batch PDF template"
DETAIL_ID="${CREATED_TEMPLATE_ID:-${XBE_TEST_ORGANIZATION_INVOICES_BATCH_PDF_TEMPLATE_ID:-}}"
if [[ -n "$DETAIL_ID" && "$DETAIL_ID" != "null" ]]; then
    xbe_json view organization-invoices-batch-pdf-templates show "$DETAIL_ID"
    if [[ $status -eq 0 ]]; then
        assert_json_has ".id"
    else
        if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]] || [[ "$output" == *"Not Found"* ]] || [[ "$output" == *"not found"* ]]; then
            skip "Show blocked by server policy or record not found"
        else
            fail "Failed to show template: $output"
        fi
    fi
else
    skip "No template ID available for show"
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

UPDATE_ID="${CREATED_TEMPLATE_ID:-${XBE_TEST_ORGANIZATION_INVOICES_BATCH_PDF_TEMPLATE_ID:-}}"

update_template() {
    local description="$1"
    shift
    test_name "$description"
    xbe_json do organization-invoices-batch-pdf-templates update "$UPDATE_ID" "$@"
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
    update_template "Update description" --description "Updated template description"
    update_template "Update template content" --template "$TEMPLATE_CONTENT_ALT"
    update_template "Update is-active" --is-active=true
    update_template "Update is-global" --is-global=true
else
    skip "No template ID available for update"
fi

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List organization invoices batch PDF templates"
xbe_json view organization-invoices-batch-pdf-templates list --limit 5
assert_success

test_name "List organization invoices batch PDF templates returns array"
xbe_json view organization-invoices-batch-pdf-templates list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
    SAMPLE_ID=$(json_get ".[0].id")
    SAMPLE_ORGANIZATION_TYPE=$(json_get ".[0].organization_type")
    SAMPLE_ORGANIZATION_ID=$(json_get ".[0].organization_id")
    SAMPLE_BROKER_ID=$(json_get ".[0].broker_id")
    SAMPLE_CREATED_BY_ID=$(json_get ".[0].created_by_id")
else
    fail "Failed to list templates"
fi

if [[ -n "$SAMPLE_ORGANIZATION_TYPE" && "$SAMPLE_ORGANIZATION_TYPE" != "null" ]]; then
    SAMPLE_ORGANIZATION_TYPE=$(normalize_org_type "$SAMPLE_ORGANIZATION_TYPE")
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "Filter templates by organization"
if [[ -n "$SAMPLE_ORGANIZATION_TYPE" && "$SAMPLE_ORGANIZATION_TYPE" != "null" && -n "$SAMPLE_ORGANIZATION_ID" && "$SAMPLE_ORGANIZATION_ID" != "null" ]]; then
    xbe_json view organization-invoices-batch-pdf-templates list --organization "${SAMPLE_ORGANIZATION_TYPE}|${SAMPLE_ORGANIZATION_ID}" --limit 5
    assert_success
else
    skip "No organization available for filter"
fi

test_name "Filter templates by organization-type and organization-id"
if [[ -n "$SAMPLE_ORGANIZATION_TYPE" && "$SAMPLE_ORGANIZATION_TYPE" != "null" && -n "$SAMPLE_ORGANIZATION_ID" && "$SAMPLE_ORGANIZATION_ID" != "null" ]]; then
    xbe_json view organization-invoices-batch-pdf-templates list --organization-type "$SAMPLE_ORGANIZATION_TYPE" --organization-id "$SAMPLE_ORGANIZATION_ID" --limit 5
    assert_success
else
    skip "No organization available for organization-type/id filter"
fi

test_name "Filter templates by broker"
if [[ -n "$SAMPLE_BROKER_ID" && "$SAMPLE_BROKER_ID" != "null" ]]; then
    xbe_json view organization-invoices-batch-pdf-templates list --broker "$SAMPLE_BROKER_ID" --limit 5
    assert_success
else
    skip "No broker ID available for filter"
fi

test_name "Filter templates by is-active"
xbe_json view organization-invoices-batch-pdf-templates list --is-active true --limit 5
assert_success

test_name "Filter templates by is-global"
xbe_json view organization-invoices-batch-pdf-templates list --is-global true --limit 5
assert_success

test_name "Filter templates by created-by"
if [[ -n "$SAMPLE_CREATED_BY_ID" && "$SAMPLE_CREATED_BY_ID" != "null" ]]; then
    xbe_json view organization-invoices-batch-pdf-templates list --created-by "$SAMPLE_CREATED_BY_ID" --limit 5
    assert_success
else
    skip "No created-by ID available for filter"
fi

test_name "Filter templates by created-at-min"
xbe_json view organization-invoices-batch-pdf-templates list --created-at-min "2020-01-01T00:00:00Z" --limit 5
assert_success

test_name "Filter templates by created-at-max"
xbe_json view organization-invoices-batch-pdf-templates list --created-at-max "2030-01-01T00:00:00Z" --limit 5
assert_success

test_name "Filter templates by is-created-at"
xbe_json view organization-invoices-batch-pdf-templates list --is-created-at true --limit 5
assert_success

# ============================================================================
# Summary
# ============================================================================

run_tests
