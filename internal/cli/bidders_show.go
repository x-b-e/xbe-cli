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

type biddersShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type bidderDetails struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	IsSelfForBroker bool   `json:"is_self_for_broker"`
	BrokerID        string `json:"broker_id,omitempty"`
	BrokerName      string `json:"broker_name,omitempty"`
}

func newBiddersShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show bidder details",
		Long: `Show the full details of a bidder.

Bidders represent entities that submit bids within broker bidding workflows.

Output Fields:
  ID        Bidder identifier
  Name      Bidder name
  Self      Whether the bidder is the broker's self bidder
  Broker    Broker organization

Arguments:
  <id>    The bidder ID (required). You can find IDs using the list command.`,
		Example: `  # Show a bidder
  xbe view bidders show 123

  # Get JSON output
  xbe view bidders show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runBiddersShow,
	}
	initBiddersShowFlags(cmd)
	return cmd
}

func init() {
	biddersCmd.AddCommand(newBiddersShowCmd())
}

func initBiddersShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runBiddersShow(cmd *cobra.Command, args []string) error {
	opts, err := parseBiddersShowOptions(cmd)
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
		return fmt.Errorf("bidder id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[bidders]", "name,is-self-for-broker,broker")
	query.Set("include", "broker")
	query.Set("fields[brokers]", "company-name")

	body, _, err := client.Get(cmd.Context(), "/v1/bidders/"+id, query)
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

	details := buildBidderDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderBidderDetails(cmd, details)
}

func parseBiddersShowOptions(cmd *cobra.Command) (biddersShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return biddersShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildBidderDetails(resp jsonAPISingleResponse) bidderDetails {
	attrs := resp.Data.Attributes
	details := bidderDetails{
		ID:              resp.Data.ID,
		Name:            stringAttr(attrs, "name"),
		IsSelfForBroker: boolAttr(attrs, "is-self-for-broker"),
	}

	included := map[string]map[string]any{}
	for _, inc := range resp.Included {
		if inc.Attributes == nil {
			continue
		}
		included[resourceKey(inc.Type, inc.ID)] = inc.Attributes
	}

	if rel, ok := resp.Data.Relationships["broker"]; ok && rel.Data != nil {
		details.BrokerID = rel.Data.ID
		if attrs, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.BrokerName = stringAttr(attrs, "company-name")
		}
	}

	return details
}

func renderBidderDetails(cmd *cobra.Command, details bidderDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.Name != "" {
		fmt.Fprintf(out, "Name: %s\n", details.Name)
	}
	fmt.Fprintf(out, "Self for Broker: %s\n", yesNo(details.IsSelfForBroker))
	if details.BrokerName != "" {
		fmt.Fprintf(out, "Broker: %s\n", details.BrokerName)
	}
	if details.BrokerID != "" {
		fmt.Fprintf(out, "Broker ID: %s\n", details.BrokerID)
	}

	return nil
}

func yesNo(value bool) string {
	if value {
		return "yes"
	}
	return "no"
}
