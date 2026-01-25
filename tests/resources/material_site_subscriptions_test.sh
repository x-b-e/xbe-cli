#!/bin/bash
#
# XBE CLI Integration Tests: Material Site Subscriptions
#
# Tests CRUD operations for the material_site_subscriptions resource.
#
# COVERAGE: All create/update attributes + list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_SUBSCRIPTION_ID=""
CREATED_MATERIAL_SITE_ID=""
CREATED_MATERIAL_SUPPLIER_ID=""
CREATED_BROKER_ID=""
CREATED_SUBSCRIPTION_USER_ID=""
CREATED_SUBSCRIPTION_USER_EMAIL=""
CREATED_SUBSCRIPTION_USER_MOBILE=""
CURRENT_USER_ID=""
CURRENT_USER_IS_ADMIN=""
CURRENT_USER_EMAIL=""
CURRENT_USER_MOBILE=""

SUBSCRIPTION_CONTACT_METHOD=""

describe "Resource: material_site_subscriptions"

# ============================================================================
# Prerequisites - Current user and material site
# ============================================================================

test_name "Fetch current authenticated user"
xbe_json auth whoami

if [[ $status -eq 0 ]]; then
    CURRENT_USER_ID=$(json_get ".id")
    CURRENT_USER_IS_ADMIN=$(json_get ".is_admin")
    CURRENT_USER_EMAIL=$(json_get ".email")
    CURRENT_USER_MOBILE=$(json_get ".mobile")
    if [[ -n "$CURRENT_USER_ID" && "$CURRENT_USER_ID" != "null" ]]; then
        pass
    else
        fail "Could not determine current user ID"
        run_tests
    fi
else
    fail "Failed to fetch authenticated user"
    run_tests
fi

# Determine if the current user is contactable
if [[ -n "$CURRENT_USER_EMAIL" && "$CURRENT_USER_EMAIL" != "null" ]]; then
    SUBSCRIPTION_CONTACT_METHOD="email_address"
elif [[ -n "$CURRENT_USER_MOBILE" && "$CURRENT_USER_MOBILE" != "null" ]]; then
    SUBSCRIPTION_CONTACT_METHOD="mobile_number"
fi

# If admin, create a dedicated broker/material supplier/material site for testing
if [[ "$CURRENT_USER_IS_ADMIN" == "true" ]]; then
    test_name "Create prerequisite broker for subscription tests"
    BROKER_NAME=$(unique_name "MSSubBroker")
    xbe_json do brokers create --name "$BROKER_NAME"

    if [[ $status -eq 0 ]]; then
        CREATED_BROKER_ID=$(json_get ".id")
        if [[ -n "$CREATED_BROKER_ID" && "$CREATED_BROKER_ID" != "null" ]]; then
            register_cleanup "brokers" "$CREATED_BROKER_ID"
            pass
        else
            fail "Created broker but no ID returned"
            run_tests
        fi
    else
        fail "Failed to create broker"
        run_tests
    fi

    test_name "Create prerequisite material supplier for subscription tests"
    SUPPLIER_NAME=$(unique_name "MSSubSupplier")
    xbe_json do material-suppliers create \
        --name "$SUPPLIER_NAME" \
        --broker "$CREATED_BROKER_ID"

    if [[ $status -eq 0 ]]; then
        CREATED_MATERIAL_SUPPLIER_ID=$(json_get ".id")
        if [[ -n "$CREATED_MATERIAL_SUPPLIER_ID" && "$CREATED_MATERIAL_SUPPLIER_ID" != "null" ]]; then
            register_cleanup "material-suppliers" "$CREATED_MATERIAL_SUPPLIER_ID"
            pass
        else
            fail "Created material supplier but no ID returned"
            run_tests
        fi
    else
        fail "Failed to create material supplier"
        run_tests
    fi

    test_name "Create prerequisite material site for subscription tests"
    SITE_NAME=$(unique_name "MSSubSite")
    xbe_json do material-sites create \
        --name "$SITE_NAME" \
        --material-supplier "$CREATED_MATERIAL_SUPPLIER_ID" \
        --address "100 Subscription Test Rd, Chicago, IL 60601"

    if [[ $status -eq 0 ]]; then
        CREATED_MATERIAL_SITE_ID=$(json_get ".id")
        if [[ -n "$CREATED_MATERIAL_SITE_ID" && "$CREATED_MATERIAL_SITE_ID" != "null" ]]; then
            register_cleanup "material-sites" "$CREATED_MATERIAL_SITE_ID"
            pass
        else
            fail "Created material site but no ID returned"
            run_tests
        fi
    else
        fail "Failed to create material site"
        run_tests
    fi

    if [[ -z "$SUBSCRIPTION_CONTACT_METHOD" ]]; then
        test_name "Create contactable user for subscription tests"
        USER_NAME=$(unique_name "MSSubUser")
        USER_EMAIL=$(unique_email)
        USER_MOBILE=$(unique_mobile)

        xbe_json do users create \
            --name "$USER_NAME" \
            --email "$USER_EMAIL" \
            --mobile "$USER_MOBILE"

        if [[ $status -eq 0 ]]; then
            CREATED_SUBSCRIPTION_USER_ID=$(json_get ".id")
            CREATED_SUBSCRIPTION_USER_EMAIL="$USER_EMAIL"
            CREATED_SUBSCRIPTION_USER_MOBILE="$USER_MOBILE"
            SUBSCRIPTION_CONTACT_METHOD="email_address"
            if [[ -n "$CREATED_SUBSCRIPTION_USER_ID" && "$CREATED_SUBSCRIPTION_USER_ID" != "null" ]]; then
                pass
            else
                fail "Created user but no ID returned"
                run_tests
            fi
        else
            fail "Failed to create contactable user"
            run_tests
        fi

        test_name "Create material supplier membership for contactable user"
        xbe_json do memberships create \
            --user "$CREATED_SUBSCRIPTION_USER_ID" \
            --organization "MaterialSupplier|$CREATED_MATERIAL_SUPPLIER_ID" \
            --kind "operations"

        if [[ $status -eq 0 ]]; then
            MEMBERSHIP_ID=$(json_get ".id")
            if [[ -n "$MEMBERSHIP_ID" && "$MEMBERSHIP_ID" != "null" ]]; then
                register_cleanup "memberships" "$MEMBERSHIP_ID"
                pass
            else
                fail "Created membership but no ID returned"
                run_tests
            fi
        else
            fail "Failed to create material supplier membership"
            run_tests
        fi
    fi
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create material site subscription"

create_subscription() {
    local site_id="$1"
    local user_id="$2"
    xbe_json do material-site-subscriptions create \
        --user "$user_id" \
        --material-site "$site_id" \
        --contact-method "$SUBSCRIPTION_CONTACT_METHOD"
}

if [[ -n "$CREATED_MATERIAL_SITE_ID" ]]; then
    if [[ -n "$CREATED_SUBSCRIPTION_USER_ID" && "$CREATED_SUBSCRIPTION_USER_ID" != "null" ]]; then
        SUBSCRIPTION_USER_ID="$CREATED_SUBSCRIPTION_USER_ID"
    else
        SUBSCRIPTION_USER_ID="$CURRENT_USER_ID"
        CREATED_SUBSCRIPTION_USER_EMAIL="$CURRENT_USER_EMAIL"
        CREATED_SUBSCRIPTION_USER_MOBILE="$CURRENT_USER_MOBILE"
    fi

    create_subscription "$CREATED_MATERIAL_SITE_ID" "$SUBSCRIPTION_USER_ID"
    if [[ $status -eq 0 ]]; then
        CREATED_SUBSCRIPTION_ID=$(json_get ".id")
        if [[ -n "$CREATED_SUBSCRIPTION_ID" && "$CREATED_SUBSCRIPTION_ID" != "null" ]]; then
            register_cleanup "material-site-subscriptions" "$CREATED_SUBSCRIPTION_ID"
            pass
        else
            fail "Created subscription but no ID returned"
        fi
    else
        fail "Failed to create material site subscription"
    fi
else
    # Non-admin path: find a material site the current user can subscribe to
    xbe_json view material-sites list --limit 20
    if [[ $status -ne 0 ]]; then
        fail "Failed to list material sites"
        run_tests
    fi

    SITE_IDS=$(echo "$output" | jq -r '.[].id')
    for site_id in $SITE_IDS; do
        if [[ -z "$SUBSCRIPTION_CONTACT_METHOD" ]]; then
            fail "Current user does not have a contact method configured"
            run_tests
        fi
        create_subscription "$site_id" "$CURRENT_USER_ID"
        if [[ $status -eq 0 ]]; then
            CREATED_SUBSCRIPTION_ID=$(json_get ".id")
            CREATED_MATERIAL_SITE_ID="$site_id"
            CREATED_SUBSCRIPTION_USER_EMAIL="$CURRENT_USER_EMAIL"
            CREATED_SUBSCRIPTION_USER_MOBILE="$CURRENT_USER_MOBILE"
            register_cleanup "material-site-subscriptions" "$CREATED_SUBSCRIPTION_ID"
            pass
            break
        fi
    done

    if [[ -z "$CREATED_SUBSCRIPTION_ID" || "$CREATED_SUBSCRIPTION_ID" == "null" ]]; then
        fail "Unable to create a material site subscription for the current user"
        run_tests
    fi
fi

# Only continue if we successfully created a subscription
if [[ -z "$CREATED_SUBSCRIPTION_ID" || "$CREATED_SUBSCRIPTION_ID" == "null" ]]; then
    echo "Cannot continue without a valid subscription ID"
    run_tests
fi

if [[ -z "$SUBSCRIPTION_CONTACT_METHOD" ]]; then
    echo "Cannot continue without a contact method"
    run_tests
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update subscription contact-method"
xbe_json do material-site-subscriptions update "$CREATED_SUBSCRIPTION_ID" --contact-method "$SUBSCRIPTION_CONTACT_METHOD"
assert_success

HAS_SUBSCRIPTION_USER_EMAIL=false
HAS_SUBSCRIPTION_USER_MOBILE=false
if [[ -n "$CREATED_SUBSCRIPTION_USER_EMAIL" && "$CREATED_SUBSCRIPTION_USER_EMAIL" != "null" ]]; then
    HAS_SUBSCRIPTION_USER_EMAIL=true
fi
if [[ -n "$CREATED_SUBSCRIPTION_USER_MOBILE" && "$CREATED_SUBSCRIPTION_USER_MOBILE" != "null" ]]; then
    HAS_SUBSCRIPTION_USER_MOBILE=true
fi

if [[ "$HAS_SUBSCRIPTION_USER_EMAIL" == true && "$HAS_SUBSCRIPTION_USER_MOBILE" == true ]]; then
    if [[ "$SUBSCRIPTION_CONTACT_METHOD" == "email_address" ]]; then
        test_name "Update subscription contact-method to mobile_number"
        xbe_json do material-site-subscriptions update "$CREATED_SUBSCRIPTION_ID" --contact-method "mobile_number"
        assert_success
    else
        test_name "Update subscription contact-method to email_address"
        xbe_json do material-site-subscriptions update "$CREATED_SUBSCRIPTION_ID" --contact-method "email_address"
        assert_success
    fi
else
    test_name "Update subscription contact-method to alternate method"
    skip "Subscription user does not have both email and mobile"
fi

# ============================================================================
# VIEW Tests
# ============================================================================

test_name "Show material site subscription"
xbe_json view material-site-subscriptions show "$CREATED_SUBSCRIPTION_ID"
if [[ $status -eq 0 ]]; then
    assert_json_equals ".id" "$CREATED_SUBSCRIPTION_ID"
else
    fail "Failed to show material site subscription"
fi

# ============================================================================
# LIST Tests
# ============================================================================

SUBSCRIPTION_USER_FILTER="$CURRENT_USER_ID"
if [[ -n "$CREATED_SUBSCRIPTION_USER_ID" && "$CREATED_SUBSCRIPTION_USER_ID" != "null" ]]; then
    SUBSCRIPTION_USER_FILTER="$CREATED_SUBSCRIPTION_USER_ID"
fi

test_name "List subscriptions filtered by user"
xbe_json view material-site-subscriptions list --user "$SUBSCRIPTION_USER_FILTER"
if [[ $status -eq 0 ]]; then
    assert_json_is_array
    assert_json_array_not_empty
else
    fail "Failed to list subscriptions by user"
fi

test_name "List subscriptions filtered by material site"
xbe_json view material-site-subscriptions list --material-site "$CREATED_MATERIAL_SITE_ID"
if [[ $status -eq 0 ]]; then
    assert_json_is_array
    assert_json_array_not_empty
else
    fail "Failed to list subscriptions by material site"
fi

test_name "List subscriptions filtered by contact method"
xbe_json view material-site-subscriptions list --contact-method "$SUBSCRIPTION_CONTACT_METHOD"
if [[ $status -eq 0 ]]; then
    assert_json_is_array
    assert_json_array_not_empty
else
    fail "Failed to list subscriptions by contact method"
fi

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete material site subscription"
xbe_json do material-site-subscriptions delete "$CREATED_SUBSCRIPTION_ID" --confirm
assert_success

# Run test summary
run_tests
