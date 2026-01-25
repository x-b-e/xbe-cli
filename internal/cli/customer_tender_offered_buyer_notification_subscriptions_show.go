package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type customerTenderOfferedBuyerNotificationSubscriptionsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

func newCustomerTenderOfferedBuyerNotificationSubscriptionsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show customer tender offered buyer notification subscription details",
		Long: `Show the full details of a customer tender offered buyer notification subscription.

Includes the associated broker and user information.

Arguments:
  <id>  The subscription ID (required).`,
		Example: `  # Show a subscription
  xbe view customer-tender-offered-buyer-notification-subscriptions show 123

  # Output as JSON
  xbe view customer-tender-offered-buyer-notification-subscriptions show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runCustomerTenderOfferedBuyerNotificationSubscriptionsShow,
	}
	initCustomerTenderOfferedBuyerNotificationSubscriptionsShowFlags(cmd)
	return cmd
}

func init() {
	customerTenderOfferedBuyerNotificationSubscriptionsCmd.AddCommand(newCustomerTenderOfferedBuyerNotificationSubscriptionsShowCmd())
}

func initCustomerTenderOfferedBuyerNotificationSubscriptionsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runCustomerTenderOfferedBuyerNotificationSubscriptionsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseCustomerTenderOfferedBuyerNotificationSubscriptionsShowOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.NoAuth {
		opts.Token = ""
	} else if strings.TrimSpace(opts.Token) == "" {
		if token, _, err := auth.ResolveToken(opts.BaseURL, ""); err == nil {
			opts.Token = token
		} else if !errors.Is(err, auth.ErrNotFound) {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	id := strings.TrimSpace(args[0])
	if id == "" {
		return fmt.Errorf("subscription id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[customer-tender-offered-buyer-notification-subscriptions]", "broker,user,notify-by-txt,notify-by-email")
	query.Set("include", "broker,user")
	query.Set("fields[brokers]", "company-name")
	query.Set("fields[users]", "name,email-address")

	body, _, err := client.Get(cmd.Context(), "/v1/customer-tender-offered-buyer-notification-subscriptions/"+id, query)
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

	handled, err := renderSparseShowIfRequested(cmd, resp)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if handled {
		return nil
	}

	details := customerTenderOfferedBuyerNotificationSubscriptionRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderCustomerTenderOfferedBuyerNotificationSubscriptionDetails(cmd, details)
}

func parseCustomerTenderOfferedBuyerNotificationSubscriptionsShowOptions(cmd *cobra.Command) (customerTenderOfferedBuyerNotificationSubscriptionsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return customerTenderOfferedBuyerNotificationSubscriptionsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func renderCustomerTenderOfferedBuyerNotificationSubscriptionDetails(cmd *cobra.Command, details customerTenderOfferedBuyerNotificationSubscriptionRow) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.BrokerID != "" {
		fmt.Fprintf(out, "Broker ID: %s\n", details.BrokerID)
	}
	if details.BrokerName != "" {
		fmt.Fprintf(out, "Broker Name: %s\n", details.BrokerName)
	}
	if details.UserID != "" {
		fmt.Fprintf(out, "User ID: %s\n", details.UserID)
	}
	if details.UserName != "" {
		fmt.Fprintf(out, "User Name: %s\n", details.UserName)
	}
	if details.UserEmail != "" {
		fmt.Fprintf(out, "User Email: %s\n", details.UserEmail)
	}
	fmt.Fprintf(out, "Notify By Txt: %s\n", formatBool(details.NotifyByTxt))
	fmt.Fprintf(out, "Notify By Email: %s\n", formatBool(details.NotifyByEmail))

	return nil
}
