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

type doCustomerTenderOfferedBuyerNotificationSubscriptionsUpdateOptions struct {
	BaseURL       string
	Token         string
	JSON          bool
	ID            string
	Broker        string
	User          string
	NotifyByTxt   bool
	NotifyByEmail bool
}

func newDoCustomerTenderOfferedBuyerNotificationSubscriptionsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a customer tender offered buyer notification subscription",
		Long: `Update an existing customer tender offered buyer notification subscription.

Optional flags:
  --broker           Broker ID
  --user             User ID
  --notify-by-txt    Notify by text (true/false)
  --notify-by-email  Notify by email (true/false)`,
		Example: `  # Enable text notifications
  xbe do customer-tender-offered-buyer-notification-subscriptions update 123 --notify-by-txt

  # Change the subscribed user
  xbe do customer-tender-offered-buyer-notification-subscriptions update 123 --user 456`,
		Args: cobra.ExactArgs(1),
		RunE: runDoCustomerTenderOfferedBuyerNotificationSubscriptionsUpdate,
	}
	initDoCustomerTenderOfferedBuyerNotificationSubscriptionsUpdateFlags(cmd)
	return cmd
}

func init() {
	doCustomerTenderOfferedBuyerNotificationSubscriptionsCmd.AddCommand(newDoCustomerTenderOfferedBuyerNotificationSubscriptionsUpdateCmd())
}

func initDoCustomerTenderOfferedBuyerNotificationSubscriptionsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("broker", "", "Broker ID")
	cmd.Flags().String("user", "", "User ID")
	cmd.Flags().Bool("notify-by-txt", false, "Notify by text (true/false)")
	cmd.Flags().Bool("notify-by-email", false, "Notify by email (true/false)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoCustomerTenderOfferedBuyerNotificationSubscriptionsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoCustomerTenderOfferedBuyerNotificationSubscriptionsUpdateOptions(cmd, args)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

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

	attributes := map[string]any{}
	if cmd.Flags().Changed("notify-by-txt") {
		attributes["notify-by-txt"] = opts.NotifyByTxt
	}
	if cmd.Flags().Changed("notify-by-email") {
		attributes["notify-by-email"] = opts.NotifyByEmail
	}

	relationships := map[string]any{}
	if cmd.Flags().Changed("broker") {
		if strings.TrimSpace(opts.Broker) == "" {
			err := fmt.Errorf("--broker cannot be empty")
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		relationships["broker"] = map[string]any{
			"data": map[string]any{
				"type": "brokers",
				"id":   opts.Broker,
			},
		}
	}
	if cmd.Flags().Changed("user") {
		if strings.TrimSpace(opts.User) == "" {
			err := fmt.Errorf("--user cannot be empty")
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		relationships["user"] = map[string]any{
			"data": map[string]any{
				"type": "users",
				"id":   opts.User,
			},
		}
	}

	if len(attributes) == 0 && len(relationships) == 0 {
		err := fmt.Errorf("no fields to update; specify at least one field flag")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	data := map[string]any{
		"type": "customer-tender-offered-buyer-notification-subscriptions",
		"id":   opts.ID,
	}
	if len(attributes) > 0 {
		data["attributes"] = attributes
	}
	if len(relationships) > 0 {
		data["relationships"] = relationships
	}

	requestBody := map[string]any{
		"data": data,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Patch(cmd.Context(), "/v1/customer-tender-offered-buyer-notification-subscriptions/"+opts.ID, jsonBody)
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

	row := customerTenderOfferedBuyerNotificationSubscriptionRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated customer tender offered buyer notification subscription %s\n", row.ID)
	return nil
}

func parseDoCustomerTenderOfferedBuyerNotificationSubscriptionsUpdateOptions(cmd *cobra.Command, args []string) (doCustomerTenderOfferedBuyerNotificationSubscriptionsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	broker, _ := cmd.Flags().GetString("broker")
	user, _ := cmd.Flags().GetString("user")
	notifyByTxt, _ := cmd.Flags().GetBool("notify-by-txt")
	notifyByEmail, _ := cmd.Flags().GetBool("notify-by-email")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doCustomerTenderOfferedBuyerNotificationSubscriptionsUpdateOptions{
		BaseURL:       baseURL,
		Token:         token,
		JSON:          jsonOut,
		ID:            args[0],
		Broker:        broker,
		User:          user,
		NotifyByTxt:   notifyByTxt,
		NotifyByEmail: notifyByEmail,
	}, nil
}
