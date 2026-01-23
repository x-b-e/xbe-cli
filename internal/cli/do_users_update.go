package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type doUsersUpdateOptions struct {
	BaseURL                               string
	Token                                 string
	JSON                                  bool
	Name                                  string
	Email                                 string
	Mobile                                string
	DefaultContactMethod                  string
	IsSuspendedFromDriving                string
	DarkMode                              string
	IsAvailableForQuestion                string
	SlackID                               string
	IsAdmin                               string
	IsPotentialTruckerReferrer            string
	OptOutOfCheckInRequestNotifications   string
	OptOutOfShiftStartingNotifications    string
	OptOutOfPreApprovalNotifications      string
	IsContactMethodRequired               string
	NotifyWhenGpsNotAvailable             string
	OptOutOfTimeCardApproverNotifications string
	NotificationPreferencesExplicit       string
	ExplicitTimeZoneID                    string
	IsGeneratingNotificationPostsExplicit string
	ReferenceData                         string
	IsReadOnlyModeEnabled                 string
	IsNotifiable                          string
	IsSales                               string
	IsCustomerSuccess                     string
}

func newDoUsersUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a user",
		Long: `Update an existing user.

Only the fields you specify will be updated. Fields not provided will remain unchanged.

Arguments:
  <id>    The user ID (required)

Flags:
  --name                                      Update the name
  --email                                     Update the email address
  --mobile                                    Update mobile phone number
  --default-contact-method                    Update default contact method (email, sms, push)
  --is-suspended-from-driving                 Update driving suspension (true/false)
  --dark-mode                                 Update dark mode preference
  --is-available-for-question                 Update question assignment availability (true/false)
  --slack-id                                  Update Slack user ID (admin only)
  --is-admin                                  Update admin status (true/false)
  --is-potential-trucker-referrer             Update potential trucker referrer (true/false)
  --opt-out-of-check-in-request-notifications Update check-in notification opt-out (true/false)
  --opt-out-of-shift-starting-notifications   Update shift starting notification opt-out (true/false)
  --opt-out-of-pre-approval-notifications     Update pre-approval notification opt-out (true/false)
  --is-contact-method-required                Update contact method required (true/false)
  --notify-when-gps-not-available             Update GPS notification (true/false)
  --opt-out-of-time-card-approver-notifications Update time card approver notification opt-out (true/false)
  --notification-preferences-explicit         Update explicit notification preferences
  --explicit-time-zone-id                     Update explicit time zone ID
  --is-generating-notification-posts-explicit Update notification posts generation (true/false)
  --reference-data                            Update reference data (JSON string)
  --is-read-only-mode-enabled                 Update read-only mode (true/false)
  --is-notifiable                             Update notifiable status (true/false)
  --is-sales                                  Update sales role (true/false, admin only)
  --is-customer-success                       Update customer success role (true/false, admin only)`,
		Example: `  # Update the name
  xbe do users update 123 --name "Jane Doe"

  # Update email and mobile
  xbe do users update 123 --email "new@example.com" --mobile "+15559876543"

  # Update contact preferences
  xbe do users update 123 --default-contact-method push

  # Suspend from driving
  xbe do users update 123 --is-suspended-from-driving true

  # Update admin status
  xbe do users update 123 --is-admin true

  # Get JSON output
  xbe do users update 123 --name "Updated" --json`,
		Args: cobra.ExactArgs(1),
		RunE: runDoUsersUpdate,
	}
	initDoUsersUpdateFlags(cmd)
	return cmd
}

func init() {
	doUsersCmd.AddCommand(newDoUsersUpdateCmd())
}

func initDoUsersUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("name", "", "New name")
	cmd.Flags().String("email", "", "New email address")
	cmd.Flags().String("mobile", "", "New mobile phone number")
	cmd.Flags().String("default-contact-method", "", "New default contact method (email, sms, push)")
	cmd.Flags().String("is-suspended-from-driving", "", "Update driving suspension (true/false)")
	cmd.Flags().String("dark-mode", "", "New dark mode preference")
	cmd.Flags().String("is-available-for-question", "", "Update question assignment availability (true/false)")
	cmd.Flags().String("slack-id", "", "New Slack user ID (admin only)")
	cmd.Flags().String("is-admin", "", "Update admin status (true/false)")
	cmd.Flags().String("is-potential-trucker-referrer", "", "Update potential trucker referrer (true/false)")
	cmd.Flags().String("opt-out-of-check-in-request-notifications", "", "Update check-in notification opt-out (true/false)")
	cmd.Flags().String("opt-out-of-shift-starting-notifications", "", "Update shift starting notification opt-out (true/false)")
	cmd.Flags().String("opt-out-of-pre-approval-notifications", "", "Update pre-approval notification opt-out (true/false)")
	cmd.Flags().String("is-contact-method-required", "", "Update contact method required (true/false)")
	cmd.Flags().String("notify-when-gps-not-available", "", "Update GPS notification (true/false)")
	cmd.Flags().String("opt-out-of-time-card-approver-notifications", "", "Update time card approver notification opt-out (true/false)")
	cmd.Flags().String("notification-preferences-explicit", "", "Update explicit notification preferences")
	cmd.Flags().String("explicit-time-zone-id", "", "Update explicit time zone ID")
	cmd.Flags().String("is-generating-notification-posts-explicit", "", "Update notification posts generation (true/false)")
	cmd.Flags().String("reference-data", "", "Update reference data (JSON string)")
	cmd.Flags().String("is-read-only-mode-enabled", "", "Update read-only mode (true/false)")
	cmd.Flags().String("is-notifiable", "", "Update notifiable status (true/false)")
	cmd.Flags().String("is-sales", "", "Update sales role (true/false, admin only)")
	cmd.Flags().String("is-customer-success", "", "Update customer success role (true/false, admin only)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoUsersUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoUsersUpdateOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	// Require authentication for write operations
	if strings.TrimSpace(opts.Token) == "" {
		if token, _, err := auth.ResolveToken(opts.BaseURL, ""); err == nil {
			opts.Token = token
		} else if errors.Is(err, auth.ErrNotFound) {
			fmt.Fprintln(cmd.ErrOrStderr(), "Authentication required. Run 'xbe auth login' first.")
			return err
		} else {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	id := strings.TrimSpace(args[0])
	if id == "" {
		return fmt.Errorf("user id is required")
	}

	// Require at least one field to update
	if opts.Name == "" && opts.Email == "" && opts.Mobile == "" &&
		opts.DefaultContactMethod == "" && opts.IsSuspendedFromDriving == "" &&
		opts.DarkMode == "" && opts.IsAvailableForQuestion == "" && opts.SlackID == "" &&
		opts.IsAdmin == "" && opts.IsPotentialTruckerReferrer == "" &&
		opts.OptOutOfCheckInRequestNotifications == "" && opts.OptOutOfShiftStartingNotifications == "" &&
		opts.OptOutOfPreApprovalNotifications == "" && opts.IsContactMethodRequired == "" &&
		opts.NotifyWhenGpsNotAvailable == "" && opts.OptOutOfTimeCardApproverNotifications == "" &&
		opts.NotificationPreferencesExplicit == "" && opts.ExplicitTimeZoneID == "" &&
		opts.IsGeneratingNotificationPostsExplicit == "" && opts.ReferenceData == "" &&
		opts.IsReadOnlyModeEnabled == "" && opts.IsNotifiable == "" &&
		opts.IsSales == "" && opts.IsCustomerSuccess == "" {
		err := fmt.Errorf("at least one field to update is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	// Build attributes
	attributes := map[string]any{}
	if opts.Name != "" {
		attributes["name"] = opts.Name
	}
	if opts.Email != "" {
		attributes["email-address"] = opts.Email
	}
	if opts.Mobile != "" {
		attributes["mobile-number"] = opts.Mobile
	}
	if opts.DefaultContactMethod != "" {
		attributes["default-contact-method"] = opts.DefaultContactMethod
	}
	if opts.IsSuspendedFromDriving != "" {
		attributes["is-suspended-from-driving"] = opts.IsSuspendedFromDriving == "true"
	}
	if opts.DarkMode != "" {
		attributes["dark-mode"] = opts.DarkMode
	}
	if opts.IsAvailableForQuestion != "" {
		attributes["is-available-for-question-assignment"] = opts.IsAvailableForQuestion == "true"
	}
	if opts.SlackID != "" {
		attributes["slack-id"] = opts.SlackID
	}
	if opts.IsAdmin != "" {
		attributes["is-admin"] = opts.IsAdmin == "true"
	}
	if opts.IsPotentialTruckerReferrer != "" {
		attributes["is-potential-trucker-referrer"] = opts.IsPotentialTruckerReferrer == "true"
	}
	if opts.OptOutOfCheckInRequestNotifications != "" {
		attributes["opt-out-of-check-in-request-notifications"] = opts.OptOutOfCheckInRequestNotifications == "true"
	}
	if opts.OptOutOfShiftStartingNotifications != "" {
		attributes["opt-out-of-shift-starting-notifications"] = opts.OptOutOfShiftStartingNotifications == "true"
	}
	if opts.OptOutOfPreApprovalNotifications != "" {
		attributes["opt-out-of-pre-approval-notifications"] = opts.OptOutOfPreApprovalNotifications == "true"
	}
	if opts.IsContactMethodRequired != "" {
		attributes["is-contact-method-required"] = opts.IsContactMethodRequired == "true"
	}
	if opts.NotifyWhenGpsNotAvailable != "" {
		attributes["notify-when-gps-not-available"] = opts.NotifyWhenGpsNotAvailable == "true"
	}
	if opts.OptOutOfTimeCardApproverNotifications != "" {
		attributes["opt-out-of-time-card-approver-notifications"] = opts.OptOutOfTimeCardApproverNotifications == "true"
	}
	if opts.NotificationPreferencesExplicit != "" {
		attributes["notification-preferences-explicit"] = opts.NotificationPreferencesExplicit
	}
	if opts.ExplicitTimeZoneID != "" {
		attributes["explicit-time-zone-id"] = opts.ExplicitTimeZoneID
	}
	if opts.IsGeneratingNotificationPostsExplicit != "" {
		attributes["is-generating-notification-posts-explicit"] = opts.IsGeneratingNotificationPostsExplicit == "true"
	}
	if opts.ReferenceData != "" {
		var refData map[string]any
		if err := json.Unmarshal([]byte(opts.ReferenceData), &refData); err != nil {
			return fmt.Errorf("invalid reference-data JSON: %w", err)
		}
		attributes["reference-data"] = refData
	}
	if opts.IsReadOnlyModeEnabled != "" {
		attributes["is-read-only-mode-enabled"] = opts.IsReadOnlyModeEnabled == "true"
	}
	if opts.IsNotifiable != "" {
		attributes["is-notifiable"] = opts.IsNotifiable == "true"
	}
	if opts.IsSales != "" {
		attributes["is-sales"] = opts.IsSales == "true"
	}
	if opts.IsCustomerSuccess != "" {
		attributes["is-customer-success"] = opts.IsCustomerSuccess == "true"
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"id":         id,
			"type":       "users",
			"attributes": attributes,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Patch(cmd.Context(), "/v1/users/"+id, jsonBody)
	if err != nil {
		if len(body) > 0 {
			fmt.Fprintln(cmd.ErrOrStderr(), string(body))
		}
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	var resp jsonAPISingleResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	row := userRow{
		ID:    resp.Data.ID,
		Name:  stringAttr(resp.Data.Attributes, "name"),
		Email: stringAttr(resp.Data.Attributes, "email-address"),
	}

	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated user %s (%s)\n", row.ID, row.Name)
	return nil
}

func parseDoUsersUpdateOptions(cmd *cobra.Command) (doUsersUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	name, _ := cmd.Flags().GetString("name")
	email, _ := cmd.Flags().GetString("email")
	mobile, _ := cmd.Flags().GetString("mobile")
	defaultContactMethod, _ := cmd.Flags().GetString("default-contact-method")
	isSuspendedFromDriving, _ := cmd.Flags().GetString("is-suspended-from-driving")
	darkMode, _ := cmd.Flags().GetString("dark-mode")
	isAvailableForQuestion, _ := cmd.Flags().GetString("is-available-for-question")
	slackID, _ := cmd.Flags().GetString("slack-id")
	isAdmin, _ := cmd.Flags().GetString("is-admin")
	isPotentialTruckerReferrer, _ := cmd.Flags().GetString("is-potential-trucker-referrer")
	optOutOfCheckInRequestNotifications, _ := cmd.Flags().GetString("opt-out-of-check-in-request-notifications")
	optOutOfShiftStartingNotifications, _ := cmd.Flags().GetString("opt-out-of-shift-starting-notifications")
	optOutOfPreApprovalNotifications, _ := cmd.Flags().GetString("opt-out-of-pre-approval-notifications")
	isContactMethodRequired, _ := cmd.Flags().GetString("is-contact-method-required")
	notifyWhenGpsNotAvailable, _ := cmd.Flags().GetString("notify-when-gps-not-available")
	optOutOfTimeCardApproverNotifications, _ := cmd.Flags().GetString("opt-out-of-time-card-approver-notifications")
	notificationPreferencesExplicit, _ := cmd.Flags().GetString("notification-preferences-explicit")
	explicitTimeZoneID, _ := cmd.Flags().GetString("explicit-time-zone-id")
	isGeneratingNotificationPostsExplicit, _ := cmd.Flags().GetString("is-generating-notification-posts-explicit")
	referenceData, _ := cmd.Flags().GetString("reference-data")
	isReadOnlyModeEnabled, _ := cmd.Flags().GetString("is-read-only-mode-enabled")
	isNotifiable, _ := cmd.Flags().GetString("is-notifiable")
	isSales, _ := cmd.Flags().GetString("is-sales")
	isCustomerSuccess, _ := cmd.Flags().GetString("is-customer-success")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doUsersUpdateOptions{
		BaseURL:                               baseURL,
		Token:                                 token,
		JSON:                                  jsonOut,
		Name:                                  name,
		Email:                                 email,
		Mobile:                                mobile,
		DefaultContactMethod:                  defaultContactMethod,
		IsSuspendedFromDriving:                isSuspendedFromDriving,
		DarkMode:                              darkMode,
		IsAvailableForQuestion:                isAvailableForQuestion,
		SlackID:                               slackID,
		IsAdmin:                               isAdmin,
		IsPotentialTruckerReferrer:            isPotentialTruckerReferrer,
		OptOutOfCheckInRequestNotifications:   optOutOfCheckInRequestNotifications,
		OptOutOfShiftStartingNotifications:    optOutOfShiftStartingNotifications,
		OptOutOfPreApprovalNotifications:      optOutOfPreApprovalNotifications,
		IsContactMethodRequired:               isContactMethodRequired,
		NotifyWhenGpsNotAvailable:             notifyWhenGpsNotAvailable,
		OptOutOfTimeCardApproverNotifications: optOutOfTimeCardApproverNotifications,
		NotificationPreferencesExplicit:       notificationPreferencesExplicit,
		ExplicitTimeZoneID:                    explicitTimeZoneID,
		IsGeneratingNotificationPostsExplicit: isGeneratingNotificationPostsExplicit,
		ReferenceData:                         referenceData,
		IsReadOnlyModeEnabled:                 isReadOnlyModeEnabled,
		IsNotifiable:                          isNotifiable,
		IsSales:                               isSales,
		IsCustomerSuccess:                     isCustomerSuccess,
	}, nil
}
