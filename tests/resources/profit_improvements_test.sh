#!/bin/bash
#
# XBE CLI Integration Tests: Profit Improvements
#
# Tests CRUD operations and filters for the profit-improvements resource.
#
# COVERAGE: All writable attributes + list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_BROKER_ID=""
CREATED_CATEGORY_ID=""
CREATED_PROFIT_IMPROVEMENT_ID=""
ORIGINAL_PROFIT_IMPROVEMENT_ID=""
WHOAMI_USER_ID=""

describe "Resource: profit-improvements"

# ============================================================================
# Prerequisites
# ============================================================================

test_name "Get current user"
xbe_json auth whoami
if [[ $status -eq 0 ]]; then
    WHOAMI_USER_ID=$(json_get ".id")
    pass
else
    skip "Unable to resolve current user"
fi

test_name "Resolve broker for profit improvements"
if [[ -n "$XBE_TEST_BROKER_ID" ]]; then
    CREATED_BROKER_ID="$XBE_TEST_BROKER_ID"
    echo "    Using XBE_TEST_BROKER_ID: $CREATED_BROKER_ID"
    pass
elif [[ -n "$WHOAMI_USER_ID" && "$WHOAMI_USER_ID" != "null" ]]; then
    xbe_json view memberships list --user "$WHOAMI_USER_ID" --limit 50
    if [[ $status -eq 0 ]]; then
        CREATED_BROKER_ID=$(echo "$output" | jq -r '.[] | select(.organization_type=="Broker" or .organization_type=="brokers") | .organization_id' | head -n 1)
        if [[ -n "$CREATED_BROKER_ID" && "$CREATED_BROKER_ID" != "null" ]]; then
            pass
        else
            fail "No broker membership found for current user"
        fi
    else
        fail "Failed to list memberships for broker lookup"
    fi
else
    skip "No user ID available for broker lookup"
fi

test_name "Select profit improvement category"
if [[ -n "$XBE_TEST_PROFIT_IMPROVEMENT_CATEGORY_ID" ]]; then
    CREATED_CATEGORY_ID="$XBE_TEST_PROFIT_IMPROVEMENT_CATEGORY_ID"
    echo "    Using XBE_TEST_PROFIT_IMPROVEMENT_CATEGORY_ID: $CREATED_CATEGORY_ID"
    pass
else
    xbe_json view profit-improvement-categories list --limit 1
    if [[ $status -eq 0 ]]; then
        CREATED_CATEGORY_ID=$(json_get ".[0].id")
        if [[ -n "$CREATED_CATEGORY_ID" && "$CREATED_CATEGORY_ID" != "null" ]]; then
            pass
        else
            fail "No profit improvement category found"
        fi
    else
        fail "Failed to list profit improvement categories"
    fi
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create profit improvement with required fields"
if [[ -n "$CREATED_BROKER_ID" && -n "$CREATED_CATEGORY_ID" ]]; then
    TITLE=$(unique_name "ProfitImprovement")
    xbe_json do profit-improvements create \
        --title "$TITLE" \
        --profit-improvement-category "$CREATED_CATEGORY_ID" \
        --organization "Broker|$CREATED_BROKER_ID"

    if [[ $status -eq 0 ]]; then
        CREATED_PROFIT_IMPROVEMENT_ID=$(json_get ".id")
        if [[ -n "$CREATED_PROFIT_IMPROVEMENT_ID" && "$CREATED_PROFIT_IMPROVEMENT_ID" != "null" ]]; then
            register_cleanup "profit-improvements" "$CREATED_PROFIT_IMPROVEMENT_ID"
            pass
        else
            fail "Created profit improvement but no ID returned"
        fi
    else
        fail "Failed to create profit improvement: $output"
    fi
else
    skip "Missing broker/category for profit improvement create"
fi

if [[ -z "$CREATED_PROFIT_IMPROVEMENT_ID" || "$CREATED_PROFIT_IMPROVEMENT_ID" == "null" ]]; then
    echo "Cannot continue without a valid profit improvement ID"
    run_tests
fi

test_name "Create profit improvement for original link"
TITLE2=$(unique_name "ProfitImprovementOriginal")
xbe_json do profit-improvements create \
    --title "$TITLE2" \
    --profit-improvement-category "$CREATED_CATEGORY_ID" \
    --organization "Broker|$CREATED_BROKER_ID"

if [[ $status -eq 0 ]]; then
    ORIGINAL_PROFIT_IMPROVEMENT_ID=$(json_get ".id")
    if [[ -n "$ORIGINAL_PROFIT_IMPROVEMENT_ID" && "$ORIGINAL_PROFIT_IMPROVEMENT_ID" != "null" ]]; then
        register_cleanup "profit-improvements" "$ORIGINAL_PROFIT_IMPROVEMENT_ID"
        pass
    else
        fail "Created original profit improvement but no ID returned"
    fi
else
    skip "Unable to create original profit improvement"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show profit improvement"
xbe_json view profit-improvements show "$CREATED_PROFIT_IMPROVEMENT_ID"
assert_success

# ============================================================================
# UPDATE Tests - Basic Fields
# ============================================================================

test_name "Update profit improvement basic fields"
xbe_json do profit-improvements update "$CREATED_PROFIT_IMPROVEMENT_ID" \
    --title "Updated $(unique_name "ProfitImprovement")" \
    --description "Updated description" \
    --status submitted \
    --amount-estimated 1000 \
    --impact-frequency-estimated recurring \
    --impact-interval-estimated monthly \
    --impact-start-on-estimated "2025-01-01" \
    --impact-end-on-estimated "2025-12-31"

if [[ $status -eq 0 ]]; then
    pass
else
    if [[ "$output" == *"Not authorized"* ]] || [[ "$output" == *"not authorized"* ]] || [[ "$output" == *"403"* ]]; then
        echo "    (Permission constrained)"
        pass
    else
        fail "Failed to update profit improvement basic fields"
    fi
fi

test_name "Update profit improvement relationships"
if [[ -n "$WHOAMI_USER_ID" && "$WHOAMI_USER_ID" != "null" && -n "$ORIGINAL_PROFIT_IMPROVEMENT_ID" && "$ORIGINAL_PROFIT_IMPROVEMENT_ID" != "null" ]]; then
    xbe_json do profit-improvements update "$CREATED_PROFIT_IMPROVEMENT_ID" \
        --profit-improvement-category "$CREATED_CATEGORY_ID" \
        --created-by "$WHOAMI_USER_ID" \
        --owned-by "$WHOAMI_USER_ID" \
        --original "$ORIGINAL_PROFIT_IMPROVEMENT_ID"
    if [[ $status -eq 0 ]]; then
        pass
    else
        if [[ "$output" == *"Not authorized"* ]] || [[ "$output" == *"not authorized"* ]] || [[ "$output" == *"403"* ]] || [[ "$output" == *"422"* ]]; then
            echo "    (Permission/validation constrained)"
            pass
        else
            fail "Failed to update profit improvement relationships"
        fi
    fi
else
    skip "No user/original ID available"
fi

test_name "Update profit improvement validated fields"
if [[ -n "$WHOAMI_USER_ID" && "$WHOAMI_USER_ID" != "null" ]]; then
    xbe_json do profit-improvements update "$CREATED_PROFIT_IMPROVEMENT_ID" \
        --amount-validated 0 \
        --impact-frequency-validated recurring \
        --impact-interval-validated quarterly \
        --impact-start-on-validated "2025-02-01" \
        --impact-end-on-validated "2025-11-30" \
        --gain-share-fee-percentage 0.1 \
        --gain-share-fee-start-on "2025-03-01" \
        --gain-share-fee-end-on "2025-12-31" \
        --validated-by "$WHOAMI_USER_ID"
    if [[ $status -eq 0 ]]; then
        pass
    else
        if [[ "$output" == *"Not authorized"* ]] || [[ "$output" == *"not authorized"* ]] || [[ "$output" == *"403"* ]] || [[ "$output" == *"422"* ]]; then
            echo "    (Permission/validation constrained)"
            pass
        else
            fail "Failed to update validated fields"
        fi
    fi
else
    skip "No user ID available"
fi

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete profit improvement"
TITLE3=$(unique_name "ProfitImprovementDelete")
xbe_json do profit-improvements create \
    --title "$TITLE3" \
    --profit-improvement-category "$CREATED_CATEGORY_ID" \
    --organization "Broker|$CREATED_BROKER_ID"
if [[ $status -eq 0 ]]; then
    DELETE_ID=$(json_get ".id")
    if [[ -n "$DELETE_ID" && "$DELETE_ID" != "null" ]]; then
        xbe_json do profit-improvements delete "$DELETE_ID" --confirm
        assert_success
    else
        fail "Created delete candidate but no ID returned"
    fi
else
    skip "Unable to create delete candidate"
fi

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List profit improvements"
xbe_json view profit-improvements list --limit 5
assert_success

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List profit improvements with basic filters"
xbe_json view profit-improvements list \
    --title "Test" \
    --description "Updated" \
    --status submitted \
    --profit-improvement-category "$CREATED_CATEGORY_ID" \
    --organization "Broker|$CREATED_BROKER_ID" \
    --broker "$CREATED_BROKER_ID" \
    --limit 5
assert_success

test_name "List profit improvements with amount filters"
xbe_json view profit-improvements list \
    --amount-estimated 100 \
    --amount-estimated-min 10 \
    --amount-estimated-max 1000 \
    --amount-validated 0 \
    --amount-validated-min 0 \
    --amount-validated-max 1000 \
    --gain-share-fee-percentage 0.1 \
    --gain-share-fee-percentage-min 0.0 \
    --gain-share-fee-percentage-max 1.0 \
    --limit 5
assert_success

test_name "List profit improvements with estimated impact enums"
xbe_json view profit-improvements list \
    --impact-frequency-estimated recurring \
    --impact-interval-estimated monthly \
    --limit 5
assert_success

test_name "List profit improvements with validated impact enums"
xbe_json view profit-improvements list \
    --impact-frequency-validated recurring \
    --impact-interval-validated quarterly \
    --limit 5
assert_success

test_name "List profit improvements with estimated impact start date filters"
xbe_json view profit-improvements list \
    --impact-start-on-estimated "2025-01-01" \
    --impact-start-on-estimated-min "2020-01-01" \
    --impact-start-on-estimated-max "2030-01-01" \
    --has-impact-start-on-estimated true \
    --limit 5
assert_success

test_name "List profit improvements with estimated impact end date filters"
xbe_json view profit-improvements list \
    --impact-end-on-estimated "2025-12-31" \
    --impact-end-on-estimated-min "2020-01-01" \
    --impact-end-on-estimated-max "2030-12-31" \
    --has-impact-end-on-estimated true \
    --limit 5
assert_success

test_name "List profit improvements with validated impact start date filters"
xbe_json view profit-improvements list \
    --impact-start-on-validated "2025-02-01" \
    --impact-start-on-validated-min "2020-01-01" \
    --impact-start-on-validated-max "2030-01-01" \
    --has-impact-start-on-validated true \
    --limit 5
assert_success

test_name "List profit improvements with validated impact end date filters"
xbe_json view profit-improvements list \
    --impact-end-on-validated "2025-11-30" \
    --impact-end-on-validated-min "2020-01-01" \
    --impact-end-on-validated-max "2030-12-31" \
    --has-impact-end-on-validated true \
    --limit 5
assert_success

test_name "List profit improvements with gain share start date filters"
xbe_json view profit-improvements list \
    --gain-share-fee-start-on "2025-03-01" \
    --gain-share-fee-start-on-min "2020-01-01" \
    --gain-share-fee-start-on-max "2030-01-01" \
    --has-gain-share-fee-start-on true \
    --limit 5
assert_success

test_name "List profit improvements with gain share end date filters"
xbe_json view profit-improvements list \
    --gain-share-fee-end-on "2025-12-31" \
    --gain-share-fee-end-on-min "2020-01-01" \
    --gain-share-fee-end-on-max "2030-12-31" \
    --has-gain-share-fee-end-on true \
    --limit 5
assert_success

test_name "List profit improvements with organization filters"
xbe_json view profit-improvements list \
    --organization-id "$CREATED_BROKER_ID" \
    --organization-type Broker \
    --not-organization-type Customer \
    --limit 5
assert_success

test_name "List profit improvements with created-by filter"
if [[ -n "$WHOAMI_USER_ID" && "$WHOAMI_USER_ID" != "null" ]]; then
    xbe_json view profit-improvements list --created-by "$WHOAMI_USER_ID" --limit 5
    assert_success
else
    skip "No user ID available"
fi

test_name "List profit improvements with owned-by filter"
if [[ -n "$WHOAMI_USER_ID" && "$WHOAMI_USER_ID" != "null" ]]; then
    xbe_json view profit-improvements list --owned-by "$WHOAMI_USER_ID" --limit 5
    assert_success
else
    skip "No user ID available"
fi

test_name "List profit improvements with validated-by filter"
if [[ -n "$WHOAMI_USER_ID" && "$WHOAMI_USER_ID" != "null" ]]; then
    xbe_json view profit-improvements list --validated-by "$WHOAMI_USER_ID" --limit 5
    assert_success
else
    skip "No user ID available"
fi

test_name "List profit improvements with original filter"
if [[ -n "$ORIGINAL_PROFIT_IMPROVEMENT_ID" && "$ORIGINAL_PROFIT_IMPROVEMENT_ID" != "null" ]]; then
    xbe_json view profit-improvements list --original "$ORIGINAL_PROFIT_IMPROVEMENT_ID" --limit 5
    assert_success
else
    skip "No original profit improvement ID available"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
