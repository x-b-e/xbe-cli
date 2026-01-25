#!/bin/bash
#
# XBE CLI Integration Tests: Project Unabandonments
#
# Tests view and create operations for project_unabandonments.
# Unabandonments restore abandoned projects to their previous status.
#
# COVERAGE: Create attributes
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

ABANDONED_PROJECT_ID=""
CREATED_BROKER_ID=""
CREATED_DEVELOPER_ID=""
CREATED_PROJECT_ID=""
SAMPLE_UNABANDONMENT_ID=""
SKIP_MUTATION=0

describe "Resource: project-unabandonments"

# ============================================================================
# LIST / SHOW Tests
# ============================================================================

test_name "List project unabandonments"
xbe_json view project-unabandonments list --limit 1
assert_success

test_name "Capture sample project unabandonment (if available)"
xbe_json view project-unabandonments list --limit 1
if [[ $status -eq 0 ]]; then
    count=$(echo "$output" | jq 'length' 2>/dev/null)
    if [[ "$count" -gt 0 ]]; then
        SAMPLE_UNABANDONMENT_ID=$(json_get ".[0].id")
        pass
    else
        echo "    No project unabandonments available; skipping show test."
        pass
    fi
else
    fail "Failed to list project unabandonments"
fi

if [[ -n "$SAMPLE_UNABANDONMENT_ID" && "$SAMPLE_UNABANDONMENT_ID" != "null" ]]; then
    test_name "Show project unabandonment"
    xbe_json view project-unabandonments show "$SAMPLE_UNABANDONMENT_ID"
    assert_success
fi

# ============================================================================
# CREATE Error Tests
# ============================================================================

test_name "Create project unabandonment requires --project"
xbe_run do project-unabandonments create
assert_failure

# ============================================================================
# CREATE Tests
# ============================================================================

if [[ -z "$XBE_TOKEN" ]]; then
    echo "    (XBE_TOKEN not set; skipping mutation tests)"
    SKIP_MUTATION=1
fi

if [[ $SKIP_MUTATION -eq 0 ]]; then
    test_name "Create prerequisite broker for project unabandonment tests"
    BROKER_NAME=$(unique_name "ProjectUnabandonBroker")

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
        test_name "Create prerequisite developer"
        if [[ -n "$XBE_TEST_DEVELOPER_ID" ]]; then
            CREATED_DEVELOPER_ID="$XBE_TEST_DEVELOPER_ID"
            echo "    Using XBE_TEST_DEVELOPER_ID: $CREATED_DEVELOPER_ID"
            pass
        else
            DEV_NAME=$(unique_name "ProjectUnabandonDev")
            xbe_json do developers create \
                --name "$DEV_NAME" \
                --broker "$CREATED_BROKER_ID"

            if [[ $status -eq 0 ]]; then
                CREATED_DEVELOPER_ID=$(json_get ".id")
                if [[ -n "$CREATED_DEVELOPER_ID" && "$CREATED_DEVELOPER_ID" != "null" ]]; then
                    register_cleanup "developers" "$CREATED_DEVELOPER_ID"
                    pass
                else
                    fail "Created developer but no ID returned"
                fi
            else
                fail "Failed to create developer"
            fi
        fi
    fi

    if [[ -n "$CREATED_DEVELOPER_ID" && "$CREATED_DEVELOPER_ID" != "null" ]]; then
        test_name "Create project for unabandonment"
        PROJECT_NAME=$(unique_name "ProjectUnabandon")

        xbe_json do projects create \
            --name "$PROJECT_NAME" \
            --developer "$CREATED_DEVELOPER_ID"

        if [[ $status -eq 0 ]]; then
            CREATED_PROJECT_ID=$(json_get ".id")
            if [[ -n "$CREATED_PROJECT_ID" && "$CREATED_PROJECT_ID" != "null" ]]; then
                register_cleanup "projects" "$CREATED_PROJECT_ID"
                pass
            else
                fail "Created project but no ID returned"
            fi
        else
            fail "Failed to create project"
        fi
    fi

    if [[ -n "$CREATED_PROJECT_ID" && "$CREATED_PROJECT_ID" != "null" ]]; then
        test_name "Abandon project for unabandonment"
        abandonment_payload=$(cat <<JSON
{"data":{"type":"project-abandonments","relationships":{"project":{"data":{"type":"projects","id":"$CREATED_PROJECT_ID"}}}}}
JSON
        )

        response_file=$(mktemp)
        run curl -s -o "$response_file" -w "%{http_code}" \
            -H "Authorization: Bearer $XBE_TOKEN" \
            -H "Accept: application/vnd.api+json" \
            -H "Content-Type: application/vnd.api+json" \
            -X POST "$XBE_BASE_URL/v1/project-abandonments" \
            -d "$abandonment_payload"

        http_code="$output"
        if [[ $status -eq 0 && "$http_code" == 2* ]]; then
            ABANDONED_PROJECT_ID="$CREATED_PROJECT_ID"
            pass
        else
            if [[ -s "$response_file" ]]; then
                echo "    Abandonment response: $(head -c 200 "$response_file")"
            fi
            skip "Unable to abandon project (HTTP ${http_code})"
        fi
        rm -f "$response_file"
    fi
else
    echo "    (Missing prerequisites; skipping unabandonment creation)"
fi

if [[ -n "$ABANDONED_PROJECT_ID" && "$ABANDONED_PROJECT_ID" != "null" ]]; then
    test_name "Create project unabandonment (minimal)"
    xbe_json do project-unabandonments create --project "$ABANDONED_PROJECT_ID"
    assert_success

    test_name "Create project unabandonment with comment"
    COMMENT_TEXT="Restored by CLI test"
    xbe_json do project-unabandonments create \
        --project "$ABANDONED_PROJECT_ID" \
        --comment "$COMMENT_TEXT"

    if [[ $status -eq 0 ]]; then
        assert_json_equals ".comment" "$COMMENT_TEXT"
    else
        fail "Failed to create project unabandonment with comment"
    fi
else
    skip "No abandoned project available for unabandonment tests"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create project unabandonment with invalid ID fails"
xbe_run do project-unabandonments create --project "999999999"
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
