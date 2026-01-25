#!/bin/bash
#
# XBE CLI Integration Tests: Maintenance Requirement Parts
#
# Tests view and create/update/delete operations for the maintenance_requirement_maintenance_requirement_parts resource.
# These links attach parts to maintenance requirements.
#
# COVERAGE: Create attributes + update attributes/relationships + list filters + delete
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_LINK_ID=""
CREATED_BROKER_ID=""
CREATED_REQUIREMENT_ID=""
CREATED_REQUIREMENT_ID_2=""
CREATED_PART_ID=""
CREATED_PART_ID_2=""
SKIP_MUTATION=0

describe "Resource: maintenance-requirement-maintenance-requirement-parts"

# ============================================================================
# Setup prerequisites (requires XBE_TOKEN for direct API calls)
# ============================================================================

if [[ -z "$XBE_TOKEN" ]]; then
    echo "    (XBE_TOKEN not set; skipping setup and mutation tests)"
    SKIP_MUTATION=1
else
    test_name "Create prerequisite broker"
    BROKER_NAME=$(unique_name "MaintReqPartBroker")

    xbe_json do brokers create --name "$BROKER_NAME"

    if [[ $status -eq 0 ]]; then
        CREATED_BROKER_ID=$(json_get ".id")
        if [[ -n "$CREATED_BROKER_ID" && "$CREATED_BROKER_ID" != "null" ]]; then
            register_cleanup "brokers" "$CREATED_BROKER_ID"
            pass
        else
            fail "Created broker but no ID returned"
        fi
    else
        if [[ -n "$XBE_TEST_BROKER_ID" ]]; then
            CREATED_BROKER_ID="$XBE_TEST_BROKER_ID"
            echo "    Using XBE_TEST_BROKER_ID: $CREATED_BROKER_ID"
            pass
        else
            fail "Failed to create broker and XBE_TEST_BROKER_ID not set"
        fi
    fi

    if [[ -n "$CREATED_BROKER_ID" && "$CREATED_BROKER_ID" != "null" ]]; then
        test_name "Create prerequisite maintenance requirement"
        REQUIREMENT_DESC=$(unique_name "MaintReq")
        REQUIREMENT_TEMPLATE=$(unique_name "MaintReqTemplate")

        requirement_payload=$(cat <<JSON
{"data":{"type":"maintenance-requirements","attributes":{"description":"$REQUIREMENT_DESC","template-name":"$REQUIREMENT_TEMPLATE","is-template":true,"is-required":true},"relationships":{"broker":{"data":{"type":"brokers","id":"$CREATED_BROKER_ID"}}}}}
JSON
        )

        run curl -s -f \
            -H "Authorization: Bearer $XBE_TOKEN" \
            -H "Accept: application/vnd.api+json" \
            -H "Content-Type: application/vnd.api+json" \
            -X POST "$XBE_BASE_URL/v1/maintenance-requirements" \
            -d "$requirement_payload"

        if [[ $status -eq 0 ]]; then
            CREATED_REQUIREMENT_ID=$(echo "$output" | jq -r '.data.id // empty')
            if [[ -n "$CREATED_REQUIREMENT_ID" && "$CREATED_REQUIREMENT_ID" != "null" ]]; then
                pass
            else
                fail "Created maintenance requirement but no ID returned"
            fi
        else
            fail "Failed to create maintenance requirement"
        fi

        test_name "Create second maintenance requirement"
        REQUIREMENT_DESC_2=$(unique_name "MaintReqB")
        REQUIREMENT_TEMPLATE_2=$(unique_name "MaintReqTemplateB")

        requirement_payload_2=$(cat <<JSON
{"data":{"type":"maintenance-requirements","attributes":{"description":"$REQUIREMENT_DESC_2","template-name":"$REQUIREMENT_TEMPLATE_2","is-template":true,"is-required":true},"relationships":{"broker":{"data":{"type":"brokers","id":"$CREATED_BROKER_ID"}}}}}
JSON
        )

        run curl -s -f \
            -H "Authorization: Bearer $XBE_TOKEN" \
            -H "Accept: application/vnd.api+json" \
            -H "Content-Type: application/vnd.api+json" \
            -X POST "$XBE_BASE_URL/v1/maintenance-requirements" \
            -d "$requirement_payload_2"

        if [[ $status -eq 0 ]]; then
            CREATED_REQUIREMENT_ID_2=$(echo "$output" | jq -r '.data.id // empty')
            if [[ -n "$CREATED_REQUIREMENT_ID_2" && "$CREATED_REQUIREMENT_ID_2" != "null" ]]; then
                pass
            else
                fail "Created second maintenance requirement but no ID returned"
            fi
        else
            fail "Failed to create second maintenance requirement"
        fi
    fi

    test_name "Create prerequisite maintenance requirement part"
    PART_NAME=$(unique_name "MaintReqPart")
    PART_NUMBER="PN-$(unique_suffix)"

    part_payload=$(cat <<JSON
{"data":{"type":"maintenance-requirement-parts","attributes":{"name":"$PART_NAME","part-number":"$PART_NUMBER"}}}
JSON
    )

    run curl -s -f \
        -H "Authorization: Bearer $XBE_TOKEN" \
        -H "Accept: application/vnd.api+json" \
        -H "Content-Type: application/vnd.api+json" \
        -X POST "$XBE_BASE_URL/v1/maintenance-requirement-parts" \
        -d "$part_payload"

    if [[ $status -eq 0 ]]; then
        CREATED_PART_ID=$(echo "$output" | jq -r '.data.id // empty')
        if [[ -n "$CREATED_PART_ID" && "$CREATED_PART_ID" != "null" ]]; then
            pass
        else
            fail "Created maintenance requirement part but no ID returned"
        fi
    else
        fail "Failed to create maintenance requirement part"
    fi

    test_name "Create second maintenance requirement part"
    PART_NAME_2=$(unique_name "MaintReqPartB")
    PART_NUMBER_2="PN-$(unique_suffix)"

    part_payload_2=$(cat <<JSON
{"data":{"type":"maintenance-requirement-parts","attributes":{"name":"$PART_NAME_2","part-number":"$PART_NUMBER_2"}}}
JSON
    )

    run curl -s -f \
        -H "Authorization: Bearer $XBE_TOKEN" \
        -H "Accept: application/vnd.api+json" \
        -H "Content-Type: application/vnd.api+json" \
        -X POST "$XBE_BASE_URL/v1/maintenance-requirement-parts" \
        -d "$part_payload_2"

    if [[ $status -eq 0 ]]; then
        CREATED_PART_ID_2=$(echo "$output" | jq -r '.data.id // empty')
        if [[ -n "$CREATED_PART_ID_2" && "$CREATED_PART_ID_2" != "null" ]]; then
            pass
        else
            fail "Created second maintenance requirement part but no ID returned"
        fi
    else
        fail "Failed to create second maintenance requirement part"
    fi
fi

# Custom cleanup for resources created via direct API
cleanup_maintenance_requirement_parts() {
    if [[ -n "$CREATED_REQUIREMENT_ID_2" && "$CREATED_REQUIREMENT_ID_2" != "null" && -n "$XBE_TOKEN" ]]; then
        curl -s -f \
            -H "Authorization: Bearer $XBE_TOKEN" \
            -H "Accept: application/vnd.api+json" \
            -X DELETE "$XBE_BASE_URL/v1/maintenance-requirements/$CREATED_REQUIREMENT_ID_2" \
            >/dev/null 2>&1 || true
    fi

    if [[ -n "$CREATED_REQUIREMENT_ID" && "$CREATED_REQUIREMENT_ID" != "null" && -n "$XBE_TOKEN" ]]; then
        curl -s -f \
            -H "Authorization: Bearer $XBE_TOKEN" \
            -H "Accept: application/vnd.api+json" \
            -X DELETE "$XBE_BASE_URL/v1/maintenance-requirements/$CREATED_REQUIREMENT_ID" \
            >/dev/null 2>&1 || true
    fi

    if [[ -n "$CREATED_PART_ID_2" && "$CREATED_PART_ID_2" != "null" && -n "$XBE_TOKEN" ]]; then
        curl -s -f \
            -H "Authorization: Bearer $XBE_TOKEN" \
            -H "Accept: application/vnd.api+json" \
            -X DELETE "$XBE_BASE_URL/v1/maintenance-requirement-parts/$CREATED_PART_ID_2" \
            >/dev/null 2>&1 || true
    fi

    if [[ -n "$CREATED_PART_ID" && "$CREATED_PART_ID" != "null" && -n "$XBE_TOKEN" ]]; then
        curl -s -f \
            -H "Authorization: Bearer $XBE_TOKEN" \
            -H "Accept: application/vnd.api+json" \
            -X DELETE "$XBE_BASE_URL/v1/maintenance-requirement-parts/$CREATED_PART_ID" \
            >/dev/null 2>&1 || true
    fi

    run_cleanup
}
trap cleanup_maintenance_requirement_parts EXIT

# ==========================================================================
# CREATE Tests
# ============================================================================

test_name "Create maintenance requirement part link without required fields fails"
xbe_json do maintenance-requirement-maintenance-requirement-parts create
assert_failure

if [[ $SKIP_MUTATION -eq 1 ]]; then
    skip "Skipping create/update/delete/filter tests without XBE_TOKEN"
fi

if [[ -n "$CREATED_REQUIREMENT_ID" && -n "$CREATED_PART_ID" ]]; then
    test_name "Create maintenance requirement part link"
    xbe_json do maintenance-requirement-maintenance-requirement-parts create \
        --maintenance-requirement "$CREATED_REQUIREMENT_ID" \
        --maintenance-requirement-part "$CREATED_PART_ID" \
        --quantity 2 \
        --unit-cost 15.75 \
        --source purchase

    if [[ $status -eq 0 ]]; then
        CREATED_LINK_ID=$(json_get ".id")
        if [[ -n "$CREATED_LINK_ID" && "$CREATED_LINK_ID" != "null" ]]; then
            register_cleanup "maintenance-requirement-maintenance-requirement-parts" "$CREATED_LINK_ID"
            pass
        else
            fail "Created maintenance requirement part link but no ID returned"
        fi
    else
        fail "Failed to create maintenance requirement part link"
    fi
else
    skip "Missing maintenance requirement or part; skipping create"
fi

# ==========================================================================
# LIST / SHOW Tests
# ============================================================================

test_name "List maintenance requirement part links"
xbe_json view maintenance-requirement-maintenance-requirement-parts list --limit 1
assert_success

if [[ -n "$CREATED_LINK_ID" && "$CREATED_LINK_ID" != "null" ]]; then
    test_name "Show maintenance requirement part link"
    xbe_json view maintenance-requirement-maintenance-requirement-parts show "$CREATED_LINK_ID"
    assert_success
fi

if [[ -n "$CREATED_REQUIREMENT_ID" && -n "$CREATED_PART_ID" ]]; then
    test_name "List maintenance requirement part links with --maintenance-requirement filter"
    xbe_json view maintenance-requirement-maintenance-requirement-parts list --maintenance-requirement "$CREATED_REQUIREMENT_ID"
    assert_success

    test_name "List maintenance requirement part links with --maintenance-requirement-part filter"
    xbe_json view maintenance-requirement-maintenance-requirement-parts list --maintenance-requirement-part "$CREATED_PART_ID"
    assert_success
fi

# ==========================================================================
# UPDATE Tests
# ============================================================================

if [[ -n "$CREATED_LINK_ID" && "$CREATED_LINK_ID" != "null" ]]; then
    test_name "Update maintenance requirement part link without fields fails"
    xbe_json do maintenance-requirement-maintenance-requirement-parts update "$CREATED_LINK_ID"
    assert_failure

    test_name "Update maintenance requirement part link attributes"
    xbe_json do maintenance-requirement-maintenance-requirement-parts update "$CREATED_LINK_ID" \
        --quantity 3 \
        --unit-cost 20 \
        --source stock
    assert_success

    if [[ -n "$CREATED_PART_ID_2" ]]; then
        test_name "Update maintenance requirement part link part relationship"
        xbe_json do maintenance-requirement-maintenance-requirement-parts update "$CREATED_LINK_ID" \
            --maintenance-requirement-part "$CREATED_PART_ID_2"
        assert_success
    fi

    if [[ -n "$CREATED_REQUIREMENT_ID_2" ]]; then
        test_name "Update maintenance requirement part link requirement relationship"
        xbe_json do maintenance-requirement-maintenance-requirement-parts update "$CREATED_LINK_ID" \
            --maintenance-requirement "$CREATED_REQUIREMENT_ID_2"
        assert_success
    fi
fi

# ==========================================================================
# DELETE Tests
# ============================================================================

if [[ -n "$CREATED_LINK_ID" && "$CREATED_LINK_ID" != "null" ]]; then
    test_name "Delete maintenance requirement part link"
    xbe_json do maintenance-requirement-maintenance-requirement-parts delete "$CREATED_LINK_ID" --confirm
    assert_success
fi

# ==========================================================================
# Summary
# ============================================================================

run_tests
