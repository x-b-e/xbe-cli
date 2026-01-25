#!/bin/bash
#
# XBE CLI Integration Tests: Job Schedule Shift Start Site Changes
#
# Tests create/list/show for job schedule shift start site changes.
#
# COVERAGE: Create + list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_BROKER_ID=""
CREATED_CUSTOMER_ID=""
CREATED_JOB_SITE_ID=""
CREATED_MATERIAL_SUPPLIER_ID=""
CREATED_MATERIAL_SITE_ID=""
CREATED_JPP_ID=""
CREATED_JOB_ID=""
CREATED_SHIFT_ID=""
CREATED_CHANGE_ID=""

describe "Resource: job-schedule-shift-start-site-changes"

# ============================================================================
# Setup - Create prerequisite records
# ============================================================================

if [[ -n "$XBE_TOKEN" ]]; then
    test_name "Create prerequisite broker"
    BROKER_NAME=$(unique_name "JSSSChangeBroker")
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
        test_name "Create prerequisite customer"
        CUSTOMER_NAME=$(unique_name "JSSSChangeCustomer")
        xbe_json do customers create --name "$CUSTOMER_NAME" --broker "$CREATED_BROKER_ID"

        if [[ $status -eq 0 ]]; then
            CREATED_CUSTOMER_ID=$(json_get ".id")
            if [[ -n "$CREATED_CUSTOMER_ID" && "$CREATED_CUSTOMER_ID" != "null" ]]; then
                register_cleanup "customers" "$CREATED_CUSTOMER_ID"
                pass
            else
                fail "Created customer but no ID returned"
            fi
        else
            fail "Failed to create customer"
        fi
    fi

    if [[ -n "$CREATED_CUSTOMER_ID" && "$CREATED_CUSTOMER_ID" != "null" ]]; then
        test_name "Create prerequisite job site"
        JOB_SITE_NAME=$(unique_name "JSSSChangeJobSite")
        xbe_json do job-sites create \
            --name "$JOB_SITE_NAME" \
            --customer "$CREATED_CUSTOMER_ID" \
            --address "100 Shift Change St, Chicago, IL 60601"

        if [[ $status -eq 0 ]]; then
            CREATED_JOB_SITE_ID=$(json_get ".id")
            if [[ -n "$CREATED_JOB_SITE_ID" && "$CREATED_JOB_SITE_ID" != "null" ]]; then
                register_cleanup "job-sites" "$CREATED_JOB_SITE_ID"
                pass
            else
                fail "Created job site but no ID returned"
            fi
        else
            fail "Failed to create job site"
        fi
    fi

    if [[ -n "$CREATED_BROKER_ID" && "$CREATED_BROKER_ID" != "null" ]]; then
        test_name "Create prerequisite material supplier"
        MATERIAL_SUPPLIER_NAME=$(unique_name "JSSSChangeSupplier")
        xbe_json do material-suppliers create --name "$MATERIAL_SUPPLIER_NAME" --broker "$CREATED_BROKER_ID"

        if [[ $status -eq 0 ]]; then
            CREATED_MATERIAL_SUPPLIER_ID=$(json_get ".id")
            if [[ -n "$CREATED_MATERIAL_SUPPLIER_ID" && "$CREATED_MATERIAL_SUPPLIER_ID" != "null" ]]; then
                register_cleanup "material-suppliers" "$CREATED_MATERIAL_SUPPLIER_ID"
                pass
            else
                fail "Created material supplier but no ID returned"
            fi
        else
            fail "Failed to create material supplier"
        fi
    fi

    if [[ -n "$CREATED_MATERIAL_SUPPLIER_ID" && "$CREATED_MATERIAL_SUPPLIER_ID" != "null" ]]; then
        test_name "Create prerequisite material site"
        MATERIAL_SITE_NAME=$(unique_name "JSSSChangeMaterialSite")
        xbe_json do material-sites create \
            --name "$MATERIAL_SITE_NAME" \
            --material-supplier "$CREATED_MATERIAL_SUPPLIER_ID" \
            --can-be-job-material-site \
            --can-be-start-site

        if [[ $status -eq 0 ]]; then
            CREATED_MATERIAL_SITE_ID=$(json_get ".id")
            if [[ -n "$CREATED_MATERIAL_SITE_ID" && "$CREATED_MATERIAL_SITE_ID" != "null" ]]; then
                register_cleanup "material-sites" "$CREATED_MATERIAL_SITE_ID"
                pass
            else
                fail "Created material site but no ID returned"
            fi
        else
            fail "Failed to create material site"
        fi
    fi

    if [[ -n "$CREATED_CUSTOMER_ID" && "$CREATED_CUSTOMER_ID" != "null" && -n "$CREATED_JOB_SITE_ID" && "$CREATED_JOB_SITE_ID" != "null" ]]; then
        test_name "Create job production plan"
        TODAY=$(date +%Y-%m-%d)
        JOB_NAME=$(unique_name "JSSSChangePlan")
        JOB_NUMBER="JSSS-CHANGE-$(date +%s)"

        xbe_json do job-production-plans create \
            --job-name "$JOB_NAME" \
            --job-number "$JOB_NUMBER" \
            --start-on "$TODAY" \
            --start-time "07:00" \
            --customer "$CREATED_CUSTOMER_ID" \
            --job-site "$CREATED_JOB_SITE_ID" \
            --requires-trucking=false \
            --requires-materials=false

        if [[ $status -eq 0 ]]; then
            CREATED_JPP_ID=$(json_get ".id")
            if [[ -n "$CREATED_JPP_ID" && "$CREATED_JPP_ID" != "null" ]]; then
                register_cleanup "job-production-plans" "$CREATED_JPP_ID"
                pass
            else
                fail "Created job production plan but no ID returned"
            fi
        else
            fail "Failed to create job production plan"
        fi
    fi

    if [[ -n "$CREATED_JPP_ID" && "$CREATED_JPP_ID" != "null" && -n "$CREATED_MATERIAL_SITE_ID" && "$CREATED_MATERIAL_SITE_ID" != "null" ]]; then
        test_name "Link material site to job production plan"
        payload=$(cat <<JSON
{"data":{"type":"job-production-plan-material-sites","attributes":{"is-default":true},"relationships":{"job-production-plan":{"data":{"type":"job-production-plans","id":"$CREATED_JPP_ID"}},"material-site":{"data":{"type":"material-sites","id":"$CREATED_MATERIAL_SITE_ID"}}}}}
JSON
        )

        response_file=$(mktemp)
        run curl -s -o "$response_file" -w "%{http_code}" \
            -H "Authorization: Bearer $XBE_TOKEN" \
            -H "Accept: application/vnd.api+json" \
            -H "Content-Type: application/vnd.api+json" \
            -X POST "$XBE_BASE_URL/v1/job-production-plan-material-sites" \
            -d "$payload"

        http_code="$output"
        if [[ $status -eq 0 && "$http_code" == 2* ]]; then
            pass
        else
            if [[ -s "$response_file" ]]; then
                echo "    Response: $(head -c 200 "$response_file")"
            fi
            skip "Unable to link material site to job production plan (HTTP ${http_code})"
        fi
        rm -f "$response_file"
    fi

    if [[ -n "$CREATED_JPP_ID" && "$CREATED_JPP_ID" != "null" ]]; then
        test_name "Fetch job and shift IDs"
        response_file=$(mktemp)
        run curl -s -o "$response_file" -w "%{http_code}" \
            -H "Authorization: Bearer $XBE_TOKEN" \
            -H "Accept: application/vnd.api+json" \
            "$XBE_BASE_URL/v1/job-production-plans/$CREATED_JPP_ID?include=jobs"

        http_code="$output"
        if [[ $status -eq 0 && "$http_code" == 2* ]]; then
            CREATED_JOB_ID=$(jq -r '[.included[]? | select(.type=="jobs") | .id] | first // empty' "$response_file")
            if [[ -n "$CREATED_JOB_ID" && "$CREATED_JOB_ID" != "null" ]]; then
                pass
            else
                skip "Job ID not found in job production plan response"
            fi
        else
            skip "Unable to fetch job production plan details (HTTP ${http_code})"
        fi
        rm -f "$response_file"
    fi

    if [[ -n "$CREATED_JOB_ID" && "$CREATED_JOB_ID" != "null" ]]; then
        test_name "Fetch job schedule shift"
        response_file=$(mktemp)
        run curl -s -o "$response_file" -w "%{http_code}" \
            -H "Authorization: Bearer $XBE_TOKEN" \
            -H "Accept: application/vnd.api+json" \
            "$XBE_BASE_URL/v1/job-schedule-shifts?filter[job]=$CREATED_JOB_ID&page[limit]=1"

        http_code="$output"
        if [[ $status -eq 0 && "$http_code" == 2* ]]; then
            CREATED_SHIFT_ID=$(jq -r '.data[0].id // empty' "$response_file")
            if [[ -n "$CREATED_SHIFT_ID" && "$CREATED_SHIFT_ID" != "null" ]]; then
                pass
            else
                skip "No job schedule shift found for job"
            fi
        else
            skip "Unable to fetch job schedule shift (HTTP ${http_code})"
        fi
        rm -f "$response_file"
    fi

    if [[ -n "$CREATED_SHIFT_ID" && "$CREATED_SHIFT_ID" != "null" ]]; then
        test_name "Create job schedule shift start site change"
        response_file=$(mktemp)
        run curl -s -o "$response_file" -w "%{http_code}" \
            -H "Authorization: Bearer $XBE_TOKEN" \
            -H "Accept: application/vnd.api+json" \
            "$XBE_BASE_URL/v1/job-schedule-shifts/$CREATED_SHIFT_ID?include=start-site"

        http_code="$output"
        if [[ $status -eq 0 && "$http_code" == 2* ]]; then
            old_start_site_type=$(jq -r '.data.relationships["start-site"].data.type // empty' "$response_file")
            old_start_site_id=$(jq -r '.data.relationships["start-site"].data.id // empty' "$response_file")
            new_start_site_type=""
            new_start_site_id=""

            if [[ "$old_start_site_type" == "job-sites" && -n "$CREATED_MATERIAL_SITE_ID" ]]; then
                new_start_site_type="material-sites"
                new_start_site_id="$CREATED_MATERIAL_SITE_ID"
            elif [[ "$old_start_site_type" == "material-sites" && -n "$CREATED_JOB_SITE_ID" ]]; then
                new_start_site_type="job-sites"
                new_start_site_id="$CREATED_JOB_SITE_ID"
            fi

            if [[ -n "$new_start_site_type" && -n "$new_start_site_id" && "$old_start_site_id" != "$new_start_site_id" ]]; then
                xbe_json do job-schedule-shift-start-site-changes create \
                    --job-schedule-shift "$CREATED_SHIFT_ID" \
                    --new-start-site-type "$new_start_site_type" \
                    --new-start-site-id "$new_start_site_id"

                if [[ $status -eq 0 ]]; then
                    CREATED_CHANGE_ID=$(json_get ".id")
                    if [[ -n "$CREATED_CHANGE_ID" && "$CREATED_CHANGE_ID" != "null" ]]; then
                        pass
                    else
                        fail "Created start site change but no ID returned"
                    fi
                else
                    fail "Failed to create start site change"
                fi
            else
                skip "Unable to determine a valid new start site"
            fi
        else
            skip "Unable to fetch job schedule shift details (HTTP ${http_code})"
        fi
        rm -f "$response_file"
    fi
else
    echo "    (XBE_TOKEN not set; skipping create/show tests that require direct API calls)"
fi

# ============================================================================
# CREATE Tests - Validation
# ============================================================================

test_name "Create start site change without required fields fails"
xbe_json do job-schedule-shift-start-site-changes create
assert_failure

# ============================================================================
# LIST Tests
# ============================================================================

test_name "List start site changes"
xbe_json view job-schedule-shift-start-site-changes list
assert_success

test_name "List start site changes returns array"
xbe_json view job-schedule-shift-start-site-changes list
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list start site changes"
fi

test_name "List start site changes with --created-at-min"
xbe_json view job-schedule-shift-start-site-changes list --created-at-min "2024-01-01T00:00:00Z"
assert_success

test_name "List start site changes with --updated-at-max"
xbe_json view job-schedule-shift-start-site-changes list --updated-at-max "2030-01-01T00:00:00Z"
assert_success

# ============================================================================
# SHOW Tests
# ============================================================================

if [[ -n "$CREATED_CHANGE_ID" && "$CREATED_CHANGE_ID" != "null" ]]; then
    test_name "Show start site change"
    xbe_json view job-schedule-shift-start-site-changes show "$CREATED_CHANGE_ID"
    assert_success
else
    skip "No start site change ID available for show tests"
fi

run_tests
