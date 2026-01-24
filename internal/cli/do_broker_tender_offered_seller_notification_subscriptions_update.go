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

type doBrokerTenderOfferedSellerNotificationSubscriptionsUpdateOptions struct {
	BaseURL       string
	Token         string
	JSON          bool
	ID            string
	Trucker       string
	User          string
	NotifyByTxt   bool
	NotifyByEmail bool
}

func newDoBrokerTenderOfferedSellerNotificationSubscriptionsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a broker tender offered seller notification subscription",
		Long: `Update an existing broker tender offered seller notification subscription.

Optional flags:
  --trucker          Trucker ID
  --user             User ID
  --notify-by-txt    Notify by text (true/false)
  --notify-by-email  Notify by email (true/false)`,
		Example: `  # Enable text notifications
  xbe do broker-tender-offered-seller-notification-subscriptions update 123 --notify-by-txt

  # Change the subscribed user
  xbe do broker-tender-offered-seller-notification-subscriptions update 123 --user 456`,
		Args: cobra.ExactArgs(1),
		RunE: runDoBrokerTenderOfferedSellerNotificationSubscriptionsUpdate,
	}
	initDoBrokerTenderOfferedSellerNotificationSubscriptionsUpdateFlags(cmd)
	return cmd
}

func init() {
	doBrokerTenderOfferedSellerNotificationSubscriptionsCmd.AddCommand(newDoBrokerTenderOfferedSellerNotificationSubscriptionsUpdateCmd())
}

func initDoBrokerTenderOfferedSellerNotificationSubscriptionsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("trucker", "", "Trucker ID")
	cmd.Flags().String("user", "", "User ID")
	cmd.Flags().Bool("notify-by-txt", false, "Notify by text (true/false)")
	cmd.Flags().Bool("notify-by-email", false, "Notify by email (true/false)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoBrokerTenderOfferedSellerNotificationSubscriptionsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoBrokerTenderOfferedSellerNotificationSubscriptionsUpdateOptions(cmd, args)
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
	if cmd.Flags().Changed("trucker") {
		if strings.TrimSpace(opts.Trucker) == "" {
			err := fmt.Errorf("--trucker cannot be empty")
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		relationships["trucker"] = map[string]any{
			"data": map[string]any{
				"type": "truckers",
				"id":   opts.Trucker,
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
		"type": "broker-tender-offered-seller-notification-subscriptions",
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

	body, _, err := client.Patch(cmd.Context(), "/v1/broker-tender-offered-seller-notification-subscriptions/"+opts.ID, jsonBody)
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

	row := brokerTenderOfferedSellerNotificationSubscriptionRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated broker tender offered seller notification subscription %s\n", row.ID)
	return nil
}

func parseDoBrokerTenderOfferedSellerNotificationSubscriptionsUpdateOptions(cmd *cobra.Command, args []string) (doBrokerTenderOfferedSellerNotificationSubscriptionsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	trucker, _ := cmd.Flags().GetString("trucker")
	user, _ := cmd.Flags().GetString("user")
	notifyByTxt, _ := cmd.Flags().GetBool("notify-by-txt")
	notifyByEmail, _ := cmd.Flags().GetBool("notify-by-email")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doBrokerTenderOfferedSellerNotificationSubscriptionsUpdateOptions{
		BaseURL:       baseURL,
		Token:         token,
		JSON:          jsonOut,
		ID:            args[0],
		Trucker:       trucker,
		User:          user,
		NotifyByTxt:   notifyByTxt,
		NotifyByEmail: notifyByEmail,
	}, nil
}
