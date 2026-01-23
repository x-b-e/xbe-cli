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

type doUsersCreateOptions struct {
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

func newDoUsersCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new user",
		Long: `Create a new user.

Required flags:
  --name     The user's name (required)
  --email    The user's email address (required)

Optional flags:
  --mobile                                    Mobile phone number
  --default-contact-method                    Default contact method (email, sms, push)
  --is-suspended-from-driving                 Suspend user from driving (true/false)
  --dark-mode                                 Dark mode preference
  --is-available-for-question                 Available for question assignment (true/false)
  --slack-id                                  Slack user ID (admin only)
  --is-admin                                  Admin status (true/false)
  --is-potential-trucker-referrer             Potential trucker referrer (true/false)
  --opt-out-of-check-in-request-notifications Opt out of check-in notifications (true/false)
  --opt-out-of-shift-starting-notifications   Opt out of shift starting notifications (true/false)
  --opt-out-of-pre-approval-notifications     Opt out of pre-approval notifications (true/false)
  --is-contact-method-required                Contact method required (true/false)
  --notify-when-gps-not-available             Notify when GPS not available (true/false)
  --opt-out-of-time-card-approver-notifications Opt out of time card approver notifications (true/false)
  --notification-preferences-explicit         Explicit notification preferences
  --explicit-time-zone-id                     Explicit time zone ID
  --is-generating-notification-posts-explicit Generating notification posts (true/false)
  --reference-data                            Reference data (JSON string)
  --is-read-only-mode-enabled                 Read-only mode enabled (true/false)
  --is-notifiable                             Notifiable (true/false)
  --is-sales                                  Sales role (true/false, admin only)
  --is-customer-success                       Customer success role (true/false, admin only)`,
		Example: `  # Create a user
  xbe do users create --name "John Doe" --email "john@example.com"

  # Create with mobile number
  xbe do users create --name "Jane Doe" --email "jane@example.com" --mobile "+15551234567"

  # Create with contact preferences
  xbe do users create --name "Bob" --email "bob@example.com" --default-contact-method sms

  # Create an admin user
  xbe do users create --name "Admin" --email "admin@example.com" --is-admin true

  # Get JSON output
  xbe do users create --name "Test User" --email "test@example.com" --json`,
		Args: cobra.NoArgs,
		RunE: runDoUsersCreate,
	}
	initDoUsersCreateFlags(cmd)
	return cmd
}

func init() {
	doUsersCmd.AddCommand(newDoUsersCreateCmd())
}

func initDoUsersCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("name", "", "User name (required)")
	cmd.Flags().String("email", "", "Email address (required)")
	cmd.Flags().String("mobile", "", "Mobile phone number")
	cmd.Flags().String("default-contact-method", "", "Default contact method (email, sms, push)")
	cmd.Flags().String("is-suspended-from-driving", "", "Suspend from driving (true/false)")
	cmd.Flags().String("dark-mode", "", "Dark mode preference")
	cmd.Flags().String("is-available-for-question", "", "Available for question assignment (true/false)")
	cmd.Flags().String("slack-id", "", "Slack user ID (admin only)")
	cmd.Flags().String("is-admin", "", "Admin status (true/false)")
	cmd.Flags().String("is-potential-trucker-referrer", "", "Potential trucker referrer (true/false)")
	cmd.Flags().String("opt-out-of-check-in-request-notifications", "", "Opt out of check-in notifications (true/false)")
	cmd.Flags().String("opt-out-of-shift-starting-notifications", "", "Opt out of shift starting notifications (true/false)")
	cmd.Flags().String("opt-out-of-pre-approval-notifications", "", "Opt out of pre-approval notifications (true/false)")
	cmd.Flags().String("is-contact-method-required", "", "Contact method required (true/false)")
	cmd.Flags().String("notify-when-gps-not-available", "", "Notify when GPS not available (true/false)")
	cmd.Flags().String("opt-out-of-time-card-approver-notifications", "", "Opt out of time card approver notifications (true/false)")
	cmd.Flags().String("notification-preferences-explicit", "", "Explicit notification preferences")
	cmd.Flags().String("explicit-time-zone-id", "", "Explicit time zone ID")
	cmd.Flags().String("is-generating-notification-posts-explicit", "", "Generating notification posts (true/false)")
	cmd.Flags().String("reference-data", "", "Reference data (JSON string)")
	cmd.Flags().String("is-read-only-mode-enabled", "", "Read-only mode enabled (true/false)")
	cmd.Flags().String("is-notifiable", "", "Notifiable (true/false)")
	cmd.Flags().String("is-sales", "", "Sales role (true/false, admin only)")
	cmd.Flags().String("is-customer-success", "", "Customer success role (true/false, admin only)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoUsersCreate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoUsersCreateOptions(cmd)
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

	// Require name
	if opts.Name == "" {
		err := fmt.Errorf("--name is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	// Require email
	if opts.Email == "" {
		err := fmt.Errorf("--email is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	// Build attributes
	attributes := map[string]any{
		"name":          opts.Name,
		"email-address": opts.Email,
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

	body, _, err := client.Post(cmd.Context(), "/v1/users", jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Created user %s (%s)\n", row.ID, row.Name)
	return nil
}

func parseDoUsersCreateOptions(cmd *cobra.Command) (doUsersCreateOptions, error) {
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

	return doUsersCreateOptions{
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
