#!/bin/bash
#
# XBE CLI Integration Tests: Project Material Type Quality Control Requirements
#
# Tests CRUD operations and filters for the project-material-type-quality-control-requirements resource.
#
# COVERAGE: Full CRUD + list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_ID=""
CREATED_BROKER_ID=""
CREATED_DEVELOPER_ID=""
CREATED_PROJECT_ID=""
CREATED_MATERIAL_TYPE_ID=""
CREATED_PROJECT_MATERIAL_TYPE_ID=""
CREATED_QUALITY_CONTROL_CLASSIFICATION_ID=""
UPDATED_QUALITY_CONTROL_CLASSIFICATION_ID=""

describe "Resource: project-material-type-quality-control-requirements"

cleanup_project_material_types() {
    if [[ -z "$XBE_TOKEN" ]]; then
        return
    fi

    if [[ -n "$CREATED_PROJECT_MATERIAL_TYPE_ID" && "$CREATED_PROJECT_MATERIAL_TYPE_ID" != "null" ]]; then
        curl -sS -X DELETE \
            -H "Authorization: Bearer $XBE_TOKEN" \
            -H "Accept: application/vnd.api+json" \
            "$XBE_BASE_URL/v1/project-material-types/$CREATED_PROJECT_MATERIAL_TYPE_ID" \
            >/dev/null 2>&1 || true
    fi
}

trap 'run_cleanup; cleanup_project_material_types' EXIT

resolve_token_from_store() {
    local base_url="$1"
    local tmp_dir
    local tmp_file
    local token

    tmp_dir=$(mktemp -d "${PROJECT_ROOT}/xbe_token_tmp.XXXXXX")
    tmp_file="${tmp_dir}/token.go"
    cat >"$tmp_file" <<'EOF'
package main

import (
	"fmt"
	"os"

	"github.com/xbe-inc/xbe-cli/internal/auth"
)

func main() {
	if len(os.Args) < 2 {
		os.Exit(1)
	}
	token, _, err := auth.ResolveToken(os.Args[1], "")
	if err != nil || token == "" {
		os.Exit(1)
	}
	fmt.Print(token)
}
EOF

    token=$(cd "$PROJECT_ROOT" && go run "$tmp_file" "$base_url" 2>/dev/null)
    local status=$?
    rm -rf "$tmp_dir"

    if [[ $status -ne 0 ]]; then
        return 1
    fi

    printf "%s" "$token"
}

create_project_material_type() {
    local project_id="$1"
    local material_type_id="$2"
    local display_name="$3"
    local payload
    payload=$(jq -n \
        --arg project_id "$project_id" \
        --arg material_type_id "$material_type_id" \
        --arg display_name "$display_name" \
        '{
            data: {
                type: "project-material-types",
                attributes: {
                    "explicit-display-name": $display_name
                },
                relationships: {
                    project: {
                        data: {
                            type: "projects",
                            id: $project_id
                        }
                    },
                    "material-type": {
                        data: {
                            type: "material-types",
                            id: $material_type_id
                        }
                    }
                }
            }
        }')

    local response
    if ! response=$(curl -sS -w "\n%{http_code}" \
        -H "Authorization: Bearer $XBE_TOKEN" \
        -H "Accept: application/vnd.api+json" \
        -H "Content-Type: application/vnd.api+json" \
        -d "$payload" \
        "$XBE_BASE_URL/v1/project-material-types"); then
        return 1
    fi

    local body
    local status_code
    body=$(printf "%s" "$response" | sed '$d')
    status_code=$(printf "%s" "$response" | tail -n1)

    if [[ "$status_code" == 2* ]]; then
        printf "%s" "$body" | jq -r '.data.id'
        return 0
    fi

    echo "$body" >&2
    return 1
}

# ============================================================================
# Prerequisites
# ============================================================================

test_name "Create prerequisite broker for quality control requirement tests"
BROKER_NAME=$(unique_name "PMTQCRBroker")

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

test_name "Create developer for quality control requirement tests"
if [[ -n "$XBE_TEST_DEVELOPER_ID" ]]; then
    CREATED_DEVELOPER_ID="$XBE_TEST_DEVELOPER_ID"
    echo "    Using XBE_TEST_DEVELOPER_ID: $CREATED_DEVELOPER_ID"
    pass
else
    DEV_NAME=$(unique_name "PMTQCRDeveloper")
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
            echo "Cannot continue without a developer"
            run_tests
        fi
    else
        fail "Failed to create developer"
        echo "Cannot continue without a developer"
        run_tests
    fi
fi

test_name "Create project for quality control requirement tests"
PROJECT_NAME=$(unique_name "PMTQCRProject")

xbe_json do projects create --name "$PROJECT_NAME" --developer "$CREATED_DEVELOPER_ID"

if [[ $status -eq 0 ]]; then
    CREATED_PROJECT_ID=$(json_get ".id")
    if [[ -n "$CREATED_PROJECT_ID" && "$CREATED_PROJECT_ID" != "null" ]]; then
        register_cleanup "projects" "$CREATED_PROJECT_ID"
        pass
    else
        fail "Created project but no ID returned"
        echo "Cannot continue without a project"
        run_tests
    fi
else
    fail "Failed to create project"
    echo "Cannot continue without a project"
    run_tests
fi

test_name "Create material type for project material type tests"
MATERIAL_TYPE_NAME=$(unique_name "PMTQCRMaterialType")

xbe_json do material-types create --name "$MATERIAL_TYPE_NAME"

if [[ $status -eq 0 ]]; then
    CREATED_MATERIAL_TYPE_ID=$(json_get ".id")
    if [[ -n "$CREATED_MATERIAL_TYPE_ID" && "$CREATED_MATERIAL_TYPE_ID" != "null" ]]; then
        register_cleanup "material-types" "$CREATED_MATERIAL_TYPE_ID"
        pass
    else
        fail "Created material type but no ID returned"
        echo "Cannot continue without a material type"
        run_tests
    fi
else
    fail "Failed to create material type"
    echo "Cannot continue without a material type"
    run_tests
fi

if [[ -z "$XBE_TOKEN" && -n "$XBE_USE_STORED_AUTH" ]]; then
    XBE_TOKEN=$(resolve_token_from_store "$XBE_BASE_URL") || true
fi

if [[ -z "$XBE_TOKEN" ]]; then
    fail "XBE_TOKEN is required to create project material types"
    run_tests
fi

test_name "Create project material type for quality control requirement tests"
PROJECT_MATERIAL_TYPE_NAME=$(unique_name "PMTQCRProjectMaterialType")
CREATED_PROJECT_MATERIAL_TYPE_ID=$(create_project_material_type "$CREATED_PROJECT_ID" "$CREATED_MATERIAL_TYPE_ID" "$PROJECT_MATERIAL_TYPE_NAME")
if [[ -n "$CREATED_PROJECT_MATERIAL_TYPE_ID" && "$CREATED_PROJECT_MATERIAL_TYPE_ID" != "null" ]]; then
    pass
else
    fail "Failed to create project material type"
    run_tests
fi

test_name "Create quality control classification"
QCC_NAME=$(unique_name "PMTQCRQCC")

xbe_json do quality-control-classifications create \
    --name "$QCC_NAME" \
    --broker "$CREATED_BROKER_ID"

if [[ $status -eq 0 ]]; then
    CREATED_QUALITY_CONTROL_CLASSIFICATION_ID=$(json_get ".id")
    if [[ -n "$CREATED_QUALITY_CONTROL_CLASSIFICATION_ID" && "$CREATED_QUALITY_CONTROL_CLASSIFICATION_ID" != "null" ]]; then
        register_cleanup "quality-control-classifications" "$CREATED_QUALITY_CONTROL_CLASSIFICATION_ID"
        pass
    else
        fail "Created quality control classification but no ID returned"
        echo "Cannot continue without a quality control classification"
        run_tests
    fi
else
    fail "Failed to create quality control classification"
    echo "Cannot continue without a quality control classification"
    run_tests
fi

test_name "Create second quality control classification for update"
QCC_NAME_UPDATED=$(unique_name "PMTQCRQCCUpdated")

xbe_json do quality-control-classifications create \
    --name "$QCC_NAME_UPDATED" \
    --broker "$CREATED_BROKER_ID"

if [[ $status -eq 0 ]]; then
    UPDATED_QUALITY_CONTROL_CLASSIFICATION_ID=$(json_get ".id")
    if [[ -n "$UPDATED_QUALITY_CONTROL_CLASSIFICATION_ID" && "$UPDATED_QUALITY_CONTROL_CLASSIFICATION_ID" != "null" ]]; then
        register_cleanup "quality-control-classifications" "$UPDATED_QUALITY_CONTROL_CLASSIFICATION_ID"
        pass
    else
        fail "Created quality control classification but no ID returned"
        echo "Cannot continue without a second quality control classification"
        run_tests
    fi
else
    fail "Failed to create second quality control classification"
    echo "Cannot continue without a second quality control classification"
    run_tests
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create project material type quality control requirement"
NOTE_VALUE="Initial note"

xbe_json do project-material-type-quality-control-requirements create \
    --project-material-type "$CREATED_PROJECT_MATERIAL_TYPE_ID" \
    --quality-control-classification "$CREATED_QUALITY_CONTROL_CLASSIFICATION_ID" \
    --note "$NOTE_VALUE"

if [[ $status -eq 0 ]]; then
    CREATED_ID=$(json_get ".id")
    if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
        register_cleanup "project-material-type-quality-control-requirements" "$CREATED_ID"
        pass
    else
        fail "Created requirement but no ID returned"
        run_tests
    fi
else
    fail "Failed to create requirement"
    run_tests
fi

test_name "Create output includes note"
if [[ $status -eq 0 ]]; then
    assert_json_equals ".note" "$NOTE_VALUE"
else
    fail "Create did not succeed"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show project material type quality control requirement"
xbe_json view project-material-type-quality-control-requirements show "$CREATED_ID"
if [[ $status -eq 0 ]]; then
    assert_json_equals ".id" "$CREATED_ID"
    assert_json_equals ".project_material_type_id" "$CREATED_PROJECT_MATERIAL_TYPE_ID"
    assert_json_equals ".quality_control_classification_id" "$CREATED_QUALITY_CONTROL_CLASSIFICATION_ID"
else
    fail "Failed to show requirement"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List requirements filtered by project material type"
xbe_json view project-material-type-quality-control-requirements list --project-material-type "$CREATED_PROJECT_MATERIAL_TYPE_ID" --limit 5
if [[ $status -eq 0 ]]; then
    if echo "$output" | jq -e --arg id "$CREATED_ID" 'map(select(.id == $id)) | length > 0' >/dev/null; then
        pass
    else
        fail "Expected requirement ID not found in filtered list"
    fi
else
    fail "Failed to list requirements by project material type"
fi

test_name "List requirements filtered by quality control classification"
xbe_json view project-material-type-quality-control-requirements list --quality-control-classification "$CREATED_QUALITY_CONTROL_CLASSIFICATION_ID" --limit 5
if [[ $status -eq 0 ]]; then
    if echo "$output" | jq -e --arg id "$CREATED_ID" 'map(select(.id == $id)) | length > 0' >/dev/null; then
        pass
    else
        fail "Expected requirement ID not found in filtered list"
    fi
else
    fail "Failed to list requirements by quality control classification"
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update requirement note and classification"
UPDATED_NOTE="Updated note"
xbe_json do project-material-type-quality-control-requirements update "$CREATED_ID" \
    --note "$UPDATED_NOTE" \
    --quality-control-classification "$UPDATED_QUALITY_CONTROL_CLASSIFICATION_ID"

if [[ $status -eq 0 ]]; then
    assert_json_equals ".id" "$CREATED_ID"
    assert_json_equals ".note" "$UPDATED_NOTE"
else
    fail "Failed to update requirement"
fi

test_name "Show requirement reflects updated classification"
xbe_json view project-material-type-quality-control-requirements show "$CREATED_ID"
if [[ $status -eq 0 ]]; then
    assert_json_equals ".quality_control_classification_id" "$UPDATED_QUALITY_CONTROL_CLASSIFICATION_ID"
    assert_json_equals ".note" "$UPDATED_NOTE"
else
    fail "Failed to show updated requirement"
fi

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete project material type quality control requirement"
xbe_json do project-material-type-quality-control-requirements delete "$CREATED_ID" --confirm
if [[ $status -eq 0 ]]; then
    assert_json_bool ".deleted" "true"
else
    fail "Failed to delete requirement"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
