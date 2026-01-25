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

type doCustomerTenderOfferedBuyerNotificationSubscriptionsCreateOptions struct {
	BaseURL       string
	Token         string
	JSON          bool
	Broker        string
	User          string
	NotifyByTxt   bool
	NotifyByEmail bool
}

func newDoCustomerTenderOfferedBuyerNotificationSubscriptionsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a customer tender offered buyer notification subscription",
		Long: `Create a customer tender offered buyer notification subscription.

Required flags:
  --broker  Broker ID (required)
  --user    User ID (required)

Optional flags:
  --notify-by-txt    Notify by text (true/false)
  --notify-by-email  Notify by email (true/false)`,
		Example: `  # Subscribe a user to customer tender offered buyer notifications
  xbe do customer-tender-offered-buyer-notification-subscriptions create --broker 123 --user 456

  # Enable text notifications
  xbe do customer-tender-offered-buyer-notification-subscriptions create --broker 123 --user 456 --notify-by-txt`,
		Args: cobra.NoArgs,
		RunE: runDoCustomerTenderOfferedBuyerNotificationSubscriptionsCreate,
	}
	initDoCustomerTenderOfferedBuyerNotificationSubscriptionsCreateFlags(cmd)
	return cmd
}

func init() {
	doCustomerTenderOfferedBuyerNotificationSubscriptionsCmd.AddCommand(newDoCustomerTenderOfferedBuyerNotificationSubscriptionsCreateCmd())
}

func initDoCustomerTenderOfferedBuyerNotificationSubscriptionsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("broker", "", "Broker ID (required)")
	cmd.Flags().String("user", "", "User ID (required)")
	cmd.Flags().Bool("notify-by-txt", false, "Notify by text (true/false)")
	cmd.Flags().Bool("notify-by-email", false, "Notify by email (true/false)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoCustomerTenderOfferedBuyerNotificationSubscriptionsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoCustomerTenderOfferedBuyerNotificationSubscriptionsCreateOptions(cmd)
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

	if opts.Broker == "" {
		err := fmt.Errorf("--broker is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if opts.User == "" {
		err := fmt.Errorf("--user is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	relationships := map[string]any{
		"broker": map[string]any{
			"data": map[string]any{
				"type": "brokers",
				"id":   opts.Broker,
			},
		},
		"user": map[string]any{
			"data": map[string]any{
				"type": "users",
				"id":   opts.User,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "customer-tender-offered-buyer-notification-subscriptions",
			"relationships": relationships,
		},
	}

	attributes := map[string]any{}
	if cmd.Flags().Changed("notify-by-txt") {
		attributes["notify-by-txt"] = opts.NotifyByTxt
	}
	if cmd.Flags().Changed("notify-by-email") {
		attributes["notify-by-email"] = opts.NotifyByEmail
	}
	if len(attributes) > 0 {
		requestBody["data"].(map[string]any)["attributes"] = attributes
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/customer-tender-offered-buyer-notification-subscriptions", jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Created customer tender offered buyer notification subscription %s\n", row.ID)
	return nil
}

func parseDoCustomerTenderOfferedBuyerNotificationSubscriptionsCreateOptions(cmd *cobra.Command) (doCustomerTenderOfferedBuyerNotificationSubscriptionsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	broker, _ := cmd.Flags().GetString("broker")
	user, _ := cmd.Flags().GetString("user")
	notifyByTxt, _ := cmd.Flags().GetBool("notify-by-txt")
	notifyByEmail, _ := cmd.Flags().GetBool("notify-by-email")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doCustomerTenderOfferedBuyerNotificationSubscriptionsCreateOptions{
		BaseURL:       baseURL,
		Token:         token,
		JSON:          jsonOut,
		Broker:        broker,
		User:          user,
		NotifyByTxt:   notifyByTxt,
		NotifyByEmail: notifyByEmail,
	}, nil
}
