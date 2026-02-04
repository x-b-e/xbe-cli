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

type predictionSubjectBidsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type predictionSubjectBidDetails struct {
	ID                                       string   `json:"id"`
	Amount                                   *float64 `json:"amount,omitempty"`
	BidderID                                 string   `json:"bidder_id,omitempty"`
	BidderName                               string   `json:"bidder_name,omitempty"`
	BidderIsSelfForBroker                    *bool    `json:"bidder_is_self_for_broker,omitempty"`
	LowestLosingBidPredictionSubjectDetailID string   `json:"lowest_losing_bid_prediction_subject_detail_id,omitempty"`
	PredictionSubjectID                      string   `json:"prediction_subject_id,omitempty"`
}

func newPredictionSubjectBidsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show prediction subject bid details",
		Long: `Show the full details of a prediction subject bid.

Output Fields:
  ID        Prediction subject bid identifier
  Amount    Bid amount
  Bidder    Bidder name
  Bidder ID Bidder identifier
  Self      Whether the bidder is the broker's self bidder
  Detail ID Lowest losing bid detail ID
  Subject   Prediction subject ID

Arguments:
  <id>    The prediction subject bid ID (required). Use the list command to find IDs.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a prediction subject bid
  xbe view prediction-subject-bids show 123

  # JSON output
  xbe view prediction-subject-bids show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runPredictionSubjectBidsShow,
	}
	initPredictionSubjectBidsShowFlags(cmd)
	return cmd
}

func init() {
	predictionSubjectBidsCmd.AddCommand(newPredictionSubjectBidsShowCmd())
}

func initPredictionSubjectBidsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runPredictionSubjectBidsShow(cmd *cobra.Command, args []string) error {
	if handled, err := maybeHandleClientURLShow(cmd, args); err != nil {
		return err
	} else if handled {
		return nil
	}

	opts, err := parsePredictionSubjectBidsShowOptions(cmd)
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
		return fmt.Errorf("prediction subject bid id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[prediction-subject-bids]", "amount,bidder,lowest-losing-bid-prediction-subject-detail")
	query.Set("include", "bidder,lowest-losing-bid-prediction-subject-detail")
	query.Set("fields[bidders]", "name,is-self-for-broker")
	query.Set("fields[lowest-losing-bid-prediction-subject-details]", "prediction-subject")

	body, _, err := client.Get(cmd.Context(), "/v1/prediction-subject-bids/"+id, query)
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

	details := buildPredictionSubjectBidDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderPredictionSubjectBidDetails(cmd, details)
}

func parsePredictionSubjectBidsShowOptions(cmd *cobra.Command) (predictionSubjectBidsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return predictionSubjectBidsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildPredictionSubjectBidDetails(resp jsonAPISingleResponse) predictionSubjectBidDetails {
	attrs := resp.Data.Attributes
	details := predictionSubjectBidDetails{ID: resp.Data.ID}
	if amount, ok := floatAttrValue(attrs, "amount"); ok {
		details.Amount = &amount
	}

	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	if rel, ok := resp.Data.Relationships["bidder"]; ok && rel.Data != nil {
		details.BidderID = rel.Data.ID
		if inc, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.BidderName = stringAttr(inc.Attributes, "name")
			if inc.Attributes != nil {
				if _, ok := inc.Attributes["is-self-for-broker"]; ok {
					value := boolAttr(inc.Attributes, "is-self-for-broker")
					details.BidderIsSelfForBroker = &value
				}
			}
		}
	}

	if rel, ok := resp.Data.Relationships["lowest-losing-bid-prediction-subject-detail"]; ok && rel.Data != nil {
		details.LowestLosingBidPredictionSubjectDetailID = rel.Data.ID
		if inc, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			if detailRel, ok := inc.Relationships["prediction-subject"]; ok && detailRel.Data != nil {
				details.PredictionSubjectID = detailRel.Data.ID
			}
		}
	}

	return details
}

func renderPredictionSubjectBidDetails(cmd *cobra.Command, details predictionSubjectBidDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.Amount != nil {
		fmt.Fprintf(out, "Amount: %s\n", formatPredictionSubjectBidAmount(details.Amount))
	}
	if details.BidderName != "" {
		fmt.Fprintf(out, "Bidder: %s\n", details.BidderName)
	}
	if details.BidderID != "" {
		fmt.Fprintf(out, "Bidder ID: %s\n", details.BidderID)
	}
	if details.BidderIsSelfForBroker != nil {
		fmt.Fprintf(out, "Self Bidder: %s\n", yesNo(*details.BidderIsSelfForBroker))
	}
	if details.LowestLosingBidPredictionSubjectDetailID != "" {
		fmt.Fprintf(out, "Detail ID: %s\n", details.LowestLosingBidPredictionSubjectDetailID)
	}
	if details.PredictionSubjectID != "" {
		fmt.Fprintf(out, "Prediction Subject ID: %s\n", details.PredictionSubjectID)
	}

	return nil
}
