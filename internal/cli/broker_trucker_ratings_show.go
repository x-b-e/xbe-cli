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

type brokerTruckerRatingsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type brokerTruckerRatingDetails struct {
	ID        string `json:"id"`
	Rating    int    `json:"rating,omitempty"`
	BrokerID  string `json:"broker_id,omitempty"`
	TruckerID string `json:"trucker_id,omitempty"`
}

func newBrokerTruckerRatingsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show broker trucker rating details",
		Long: `Show the full details of a broker trucker rating.

Output Fields:
  ID         Broker trucker rating identifier
  Rating     Rating (1-5)
  Broker ID  Broker ID
  Trucker ID Trucker ID

Arguments:
  <id>  The broker trucker rating ID (required). Use the list command to find IDs.`,
		Example: `  # Show a broker trucker rating
  xbe view broker-trucker-ratings show 123

  # Show as JSON
  xbe view broker-trucker-ratings show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runBrokerTruckerRatingsShow,
	}
	initBrokerTruckerRatingsShowFlags(cmd)
	return cmd
}

func init() {
	brokerTruckerRatingsCmd.AddCommand(newBrokerTruckerRatingsShowCmd())
}

func initBrokerTruckerRatingsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runBrokerTruckerRatingsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseBrokerTruckerRatingsShowOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.NoAuth {
		opts.Token = ""
	} else if strings.TrimSpace(opts.Token) == "" {
		if token, _, err := auth.ResolveToken(opts.BaseURL, ""); err == nil {
			opts.Token = token
		} else if errors.Is(err, auth.ErrNotFound) {
			fmt.Fprintln(cmd.ErrOrStderr(), "Authentication required. Run \"xbe auth login\" first.")
			return err
		} else {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	id := strings.TrimSpace(args[0])
	if id == "" {
		return fmt.Errorf("broker trucker rating id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[broker-trucker-ratings]", "rating,broker,trucker")

	body, _, err := client.Get(cmd.Context(), "/v1/broker-trucker-ratings/"+id, query)
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

	details := buildBrokerTruckerRatingDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderBrokerTruckerRatingDetails(cmd, details)
}

func parseBrokerTruckerRatingsShowOptions(cmd *cobra.Command) (brokerTruckerRatingsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return brokerTruckerRatingsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildBrokerTruckerRatingDetails(resp jsonAPISingleResponse) brokerTruckerRatingDetails {
	details := brokerTruckerRatingDetails{
		ID:     resp.Data.ID,
		Rating: intAttr(resp.Data.Attributes, "rating"),
	}

	if rel, ok := resp.Data.Relationships["broker"]; ok && rel.Data != nil {
		details.BrokerID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["trucker"]; ok && rel.Data != nil {
		details.TruckerID = rel.Data.ID
	}

	return details
}

func renderBrokerTruckerRatingDetails(cmd *cobra.Command, details brokerTruckerRatingDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	fmt.Fprintf(out, "Rating: %d\n", details.Rating)
	if details.BrokerID != "" {
		fmt.Fprintf(out, "Broker ID: %s\n", details.BrokerID)
	}
	if details.TruckerID != "" {
		fmt.Fprintf(out, "Trucker ID: %s\n", details.TruckerID)
	}

	return nil
}
