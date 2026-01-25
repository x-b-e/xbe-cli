#!/bin/bash
#
# XBE CLI Integration Tests: Users
#
# Tests CRUD operations for the users resource.
# Note: Users cannot be deleted via API, so cleanup marks them inactive.
#
# COMPLETE COVERAGE: All 22 create/update attributes + 15 list filters
#

# Load test helpers (this also loads config and runs init_tests)
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

# Override cleanup trap since users can't be deleted
trap - EXIT
trap cleanup_users EXIT

CREATED_USER_ID=""

cleanup_users() {
    if [[ -n "$CREATED_USER_ID" && "$CREATED_USER_ID" != "null" ]]; then
        echo ""
        echo -e "${YELLOW}Cleaning up test user...${NC}"
        # Can't delete users, but we can update to mark test users
        # Just leave it - staging can have test users
        echo "  Note: Users cannot be deleted. Test user $CREATED_USER_ID left in staging."
        echo -e "${GREEN}Cleanup complete.${NC}"
    fi
}

describe "Resource: users"

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create user with required fields only"
TEST_EMAIL=$(unique_email)
TEST_NAME=$(unique_name "User")

xbe_json do users create \
    --name "$TEST_NAME" \
    --email "$TEST_EMAIL"

if [[ $status -eq 0 ]]; then
    CREATED_USER_ID=$(json_get ".id")
    if [[ -n "$CREATED_USER_ID" && "$CREATED_USER_ID" != "null" ]]; then
        pass
    else
        fail "Created user but no ID returned"
    fi
else
    fail "Failed to create user"
fi

# Only continue if we successfully created a user
if [[ -z "$CREATED_USER_ID" || "$CREATED_USER_ID" == "null" ]]; then
    echo "Cannot continue without a valid user ID"
    run_tests
fi

# ============================================================================
# UPDATE Tests - String Attributes
# ============================================================================

test_name "Update user name"
UPDATED_NAME=$(unique_name "UpdatedUser")
xbe_json do users update "$CREATED_USER_ID" --name "$UPDATED_NAME"
assert_success

test_name "Update user email"
UPDATED_EMAIL=$(unique_email)
xbe_json do users update "$CREATED_USER_ID" --email "$UPDATED_EMAIL"
assert_success

test_name "Update user mobile"
TEST_MOBILE=$(unique_mobile)
xbe_json do users update "$CREATED_USER_ID" --mobile "$TEST_MOBILE"
assert_success

test_name "Update user default-contact-method to email_address"
xbe_json do users update "$CREATED_USER_ID" --default-contact-method "email_address"
assert_success

test_name "Update user default-contact-method to mobile_number"
xbe_json do users update "$CREATED_USER_ID" --default-contact-method "mobile_number"
assert_success

test_name "Update user dark-mode to os"
xbe_json do users update "$CREATED_USER_ID" --dark-mode "os"
assert_success

test_name "Update user dark-mode to on"
xbe_json do users update "$CREATED_USER_ID" --dark-mode "on"
assert_success

test_name "Update user dark-mode to off"
xbe_json do users update "$CREATED_USER_ID" --dark-mode "off"
assert_success

test_name "Update user slack-id"
xbe_json do users update "$CREATED_USER_ID" --slack-id "U123TEST456"
assert_success

test_name "Update user notification-preferences-explicit"
xbe_json do users update "$CREATED_USER_ID" --notification-preferences-explicit "custom_pref"
assert_success

test_name "Update user explicit-time-zone-id"
xbe_json do users update "$CREATED_USER_ID" --explicit-time-zone-id "America/New_York"
assert_success

test_name "Update user reference-data (JSON)"
xbe_json do users update "$CREATED_USER_ID" --reference-data '{"test_key": "test_value", "number": 123}'
assert_success

# ============================================================================
# UPDATE Tests - Boolean Attributes (true then false for each)
# ============================================================================

test_name "Update is-suspended-from-driving to true"
xbe_json do users update "$CREATED_USER_ID" --is-suspended-from-driving true
assert_success

test_name "Update is-suspended-from-driving to false"
xbe_json do users update "$CREATED_USER_ID" --is-suspended-from-driving false
assert_success

test_name "Update is-available-for-question to true"
xbe_json do users update "$CREATED_USER_ID" --is-available-for-question true
assert_success

test_name "Update is-available-for-question to false"
xbe_json do users update "$CREATED_USER_ID" --is-available-for-question false
assert_success

test_name "Update is-admin to true"
xbe_json do users update "$CREATED_USER_ID" --is-admin true
assert_success

test_name "Update is-admin to false"
xbe_json do users update "$CREATED_USER_ID" --is-admin false
assert_success

test_name "Update is-potential-trucker-referrer to true"
xbe_json do users update "$CREATED_USER_ID" --is-potential-trucker-referrer true
assert_success

test_name "Update is-potential-trucker-referrer to false"
xbe_json do users update "$CREATED_USER_ID" --is-potential-trucker-referrer false
assert_success

test_name "Update opt-out-of-check-in-request-notifications to true"
xbe_json do users update "$CREATED_USER_ID" --opt-out-of-check-in-request-notifications true
assert_success

test_name "Update opt-out-of-check-in-request-notifications to false"
xbe_json do users update "$CREATED_USER_ID" --opt-out-of-check-in-request-notifications false
assert_success

test_name "Update opt-out-of-shift-starting-notifications to true"
xbe_json do users update "$CREATED_USER_ID" --opt-out-of-shift-starting-notifications true
assert_success

test_name "Update opt-out-of-shift-starting-notifications to false"
xbe_json do users update "$CREATED_USER_ID" --opt-out-of-shift-starting-notifications false
assert_success

test_name "Update opt-out-of-pre-approval-notifications to true"
xbe_json do users update "$CREATED_USER_ID" --opt-out-of-pre-approval-notifications true
assert_success

test_name "Update opt-out-of-pre-approval-notifications to false"
xbe_json do users update "$CREATED_USER_ID" --opt-out-of-pre-approval-notifications false
assert_success

test_name "Update is-contact-method-required to true"
xbe_json do users update "$CREATED_USER_ID" --is-contact-method-required true
assert_success

test_name "Update is-contact-method-required to false"
xbe_json do users update "$CREATED_USER_ID" --is-contact-method-required false
assert_success

test_name "Update notify-when-gps-not-available to true"
xbe_json do users update "$CREATED_USER_ID" --notify-when-gps-not-available true
assert_success

test_name "Update notify-when-gps-not-available to false"
xbe_json do users update "$CREATED_USER_ID" --notify-when-gps-not-available false
assert_success

test_name "Update opt-out-of-time-card-approver-notifications to true"
xbe_json do users update "$CREATED_USER_ID" --opt-out-of-time-card-approver-notifications true
assert_success

test_name "Update opt-out-of-time-card-approver-notifications to false"
xbe_json do users update "$CREATED_USER_ID" --opt-out-of-time-card-approver-notifications false
assert_success

test_name "Update is-generating-notification-posts-explicit to true"
xbe_json do users update "$CREATED_USER_ID" --is-generating-notification-posts-explicit true
assert_success

test_name "Update is-generating-notification-posts-explicit to false"
xbe_json do users update "$CREATED_USER_ID" --is-generating-notification-posts-explicit false
assert_success

test_name "Update is-read-only-mode-enabled to true"
xbe_json do users update "$CREATED_USER_ID" --is-read-only-mode-enabled true
assert_success

test_name "Update is-read-only-mode-enabled to false"
xbe_json do users update "$CREATED_USER_ID" --is-read-only-mode-enabled false
assert_success

test_name "Update is-notifiable to true"
xbe_json do users update "$CREATED_USER_ID" --is-notifiable true
assert_success

test_name "Update is-notifiable to false"
xbe_json do users update "$CREATED_USER_ID" --is-notifiable false
assert_success

test_name "Update is-sales to true"
xbe_json do users update "$CREATED_USER_ID" --is-sales true
assert_success

test_name "Update is-sales to false"
xbe_json do users update "$CREATED_USER_ID" --is-sales false
assert_success

test_name "Update is-customer-success to true"
xbe_json do users update "$CREATED_USER_ID" --is-customer-success true
assert_success

test_name "Update is-customer-success to false"
xbe_json do users update "$CREATED_USER_ID" --is-customer-success false
assert_success

# ============================================================================
# LIST Tests - Basic
# ============================================================================

# Note: Users resource does not have a "show" command

test_name "List users"
xbe_json view users list --limit 5
assert_success

test_name "List users returns array"
xbe_json view users list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list users"
fi

# ============================================================================
# LIST Tests - All Filters
# ============================================================================

test_name "List users with --name filter"
xbe_json view users list --name "$UPDATED_NAME" --limit 10
assert_success

test_name "List users with --email-address filter"
xbe_json view users list --email-address "$UPDATED_EMAIL" --limit 10
assert_success

test_name "List users with --email-address-like filter (partial match)"
EMAIL_PARTIAL="${UPDATED_EMAIL%%@*}"
xbe_json view users list --email-address-like "$EMAIL_PARTIAL" --limit 10
assert_success

test_name "List users with --mobile-number filter"
xbe_json view users list --mobile-number "$TEST_MOBILE" --limit 10
assert_success

test_name "List users with --slack-id filter"
xbe_json view users list --slack-id "U123TEST456" --limit 10
assert_success

test_name "List users with --is-admin flag"
xbe_run view users list --is-admin --limit 5
assert_success

test_name "List users with --is-driver filter (true)"
xbe_json view users list --is-driver true --limit 5
assert_success

test_name "List users with --is-driver filter (false)"
xbe_json view users list --is-driver false --limit 5
assert_success

test_name "List users with --is-suspended-from-driving filter (true)"
xbe_json view users list --is-suspended-from-driving true --limit 5
assert_success

test_name "List users with --is-suspended-from-driving filter (false)"
xbe_json view users list --is-suspended-from-driving false --limit 5
assert_success

test_name "List users with --is-available-for-question-assignment filter (true)"
xbe_json view users list --is-available-for-question-assignment true --limit 5
assert_success

test_name "List users with --is-available-for-question-assignment filter (false)"
xbe_json view users list --is-available-for-question-assignment false --limit 5
assert_success

test_name "List users with --dark-mode filter"
xbe_json view users list --dark-mode "off" --limit 5
assert_success

test_name "List users with --has-notifai-text filter (true)"
xbe_json view users list --has-notifai-text true --limit 5
assert_success

test_name "List users with --has-notifai-text filter (false)"
xbe_json view users list --has-notifai-text false --limit 5
assert_success

# Note: These filters require valid IDs to be meaningful, but we test the command works
test_name "List users with --having-customer-membership-with filter"
xbe_json view users list --having-customer-membership-with "1" --limit 5
assert_success

test_name "List users with --having-trucker-membership-with filter"
xbe_json view users list --having-trucker-membership-with "1" --limit 5
assert_success

test_name "List users with --having-manager-trucker-membership-with filter"
xbe_json view users list --having-manager-trucker-membership-with "1" --limit 5
assert_success

test_name "List users with --can-manage-projects-for filter"
xbe_json view users list --can-manage-projects-for "1" --limit 5
assert_success

# ============================================================================
# LIST Tests - Pagination
# ============================================================================

test_name "List users with --limit"
xbe_json view users list --limit 3
assert_success

test_name "List users with --offset"
xbe_json view users list --limit 3 --offset 3
assert_success

test_name "List users with pagination (limit + offset)"
xbe_json view users list --limit 5 --offset 10
assert_success

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create user without name fails"
xbe_json do users create --email "test@example.com"
assert_failure

test_name "Create user without email fails"
xbe_json do users create --name "Test User"
assert_failure

test_name "Update without any fields fails"
xbe_json do users update "$CREATED_USER_ID"
assert_failure

test_name "Update with invalid reference-data JSON fails"
xbe_json do users update "$CREATED_USER_ID" --reference-data "not valid json"
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
