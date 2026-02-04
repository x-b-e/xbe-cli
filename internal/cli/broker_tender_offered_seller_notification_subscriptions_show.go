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

type brokerTenderOfferedSellerNotificationSubscriptionsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

func newBrokerTenderOfferedSellerNotificationSubscriptionsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show broker tender offered seller notification subscription details",
		Long: `Show the full details of a broker tender offered seller notification subscription.

Includes the associated trucker and user information.

Arguments:
  <id>  The subscription ID (required).`,
		Example: `  # Show a subscription
  xbe view broker-tender-offered-seller-notification-subscriptions show 123

  # Output as JSON
  xbe view broker-tender-offered-seller-notification-subscriptions show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runBrokerTenderOfferedSellerNotificationSubscriptionsShow,
	}
	initBrokerTenderOfferedSellerNotificationSubscriptionsShowFlags(cmd)
	return cmd
}

func init() {
	brokerTenderOfferedSellerNotificationSubscriptionsCmd.AddCommand(newBrokerTenderOfferedSellerNotificationSubscriptionsShowCmd())
}

func initBrokerTenderOfferedSellerNotificationSubscriptionsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runBrokerTenderOfferedSellerNotificationSubscriptionsShow(cmd *cobra.Command, args []string) error {
	if handled, err := maybeHandleClientURLShow(cmd, args); err != nil {
		return err
	} else if handled {
		return nil
	}

	opts, err := parseBrokerTenderOfferedSellerNotificationSubscriptionsShowOptions(cmd)
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
	query.Set("fields[broker-tender-offered-seller-notification-subscriptions]", "trucker,user,notify-by-txt,notify-by-email")
	query.Set("include", "trucker,user")
	query.Set("fields[truckers]", "company-name")
	query.Set("fields[users]", "name,email-address")

	body, _, err := client.Get(cmd.Context(), "/v1/broker-tender-offered-seller-notification-subscriptions/"+id, query)
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

	details := brokerTenderOfferedSellerNotificationSubscriptionRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderBrokerTenderOfferedSellerNotificationSubscriptionDetails(cmd, details)
}

func parseBrokerTenderOfferedSellerNotificationSubscriptionsShowOptions(cmd *cobra.Command) (brokerTenderOfferedSellerNotificationSubscriptionsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return brokerTenderOfferedSellerNotificationSubscriptionsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func renderBrokerTenderOfferedSellerNotificationSubscriptionDetails(cmd *cobra.Command, details brokerTenderOfferedSellerNotificationSubscriptionRow) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.TruckerID != "" {
		fmt.Fprintf(out, "Trucker ID: %s\n", details.TruckerID)
	}
	if details.TruckerName != "" {
		fmt.Fprintf(out, "Trucker Name: %s\n", details.TruckerName)
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
