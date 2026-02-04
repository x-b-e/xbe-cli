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

type lowestLosingBidPredictionSubjectDetailsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type lowestLosingBidPredictionSubjectDetailDetails struct {
	ID                             string   `json:"id"`
	PredictionSubjectID            string   `json:"prediction_subject_id,omitempty"`
	LowestBidAmount                string   `json:"lowest_bid_amount,omitempty"`
	BidAmount                      string   `json:"bid_amount,omitempty"`
	WalkAwayBidAmount              string   `json:"walk_away_bid_amount,omitempty"`
	EngineerEstimateAmount         string   `json:"engineer_estimate_amount,omitempty"`
	InternalEngineerEstimateAmount string   `json:"internal_engineer_estimate_amount,omitempty"`
	BidDetails                     any      `json:"bid_details,omitempty"`
	PredictionSubjectBidIDs        []string `json:"prediction_subject_bid_ids,omitempty"`
}

func newLowestLosingBidPredictionSubjectDetailsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show lowest losing bid prediction subject detail",
		Long: `Show the full details of a lowest losing bid prediction subject detail.

Output Fields:
  ID                          Detail identifier
  Prediction Subject          Prediction subject ID
  Lowest Bid Amount           Lowest bid amount
  Bid Amount                  Bid amount
  Walk Away Bid Amount        Walk away bid amount
  Engineer Estimate Amount    Engineer estimate amount
  Internal Engineer Estimate  Internal engineer estimate amount
  Bid Details                 Bid details JSON (if present)
  Prediction Subject Bids     Related prediction subject bid IDs

Arguments:
  <id>  The detail ID (required).`,
		Example: `  # Show details
  xbe view lowest-losing-bid-prediction-subject-details show 123

  # Output as JSON
  xbe view lowest-losing-bid-prediction-subject-details show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runLowestLosingBidPredictionSubjectDetailsShow,
	}
	initLowestLosingBidPredictionSubjectDetailsShowFlags(cmd)
	return cmd
}

func init() {
	lowestLosingBidPredictionSubjectDetailsCmd.AddCommand(newLowestLosingBidPredictionSubjectDetailsShowCmd())
}

func initLowestLosingBidPredictionSubjectDetailsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runLowestLosingBidPredictionSubjectDetailsShow(cmd *cobra.Command, args []string) error {
	if handled, err := maybeHandleClientURLShow(cmd, args); err != nil {
		return err
	} else if handled {
		return nil
	}

	opts, err := parseLowestLosingBidPredictionSubjectDetailsShowOptions(cmd)
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
		return fmt.Errorf("detail id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[lowest-losing-bid-prediction-subject-details]", "prediction-subject,prediction-subject-bids,lowest-bid-amount,bid-amount,walk-away-bid-amount,engineer-estimate-amount,internal-engineer-estimate-amount,bid-details")

	body, _, err := client.Get(cmd.Context(), "/v1/lowest-losing-bid-prediction-subject-details/"+id, query)
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

	details := buildLowestLosingBidPredictionSubjectDetailDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderLowestLosingBidPredictionSubjectDetailDetails(cmd, details)
}

func parseLowestLosingBidPredictionSubjectDetailsShowOptions(cmd *cobra.Command) (lowestLosingBidPredictionSubjectDetailsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return lowestLosingBidPredictionSubjectDetailsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildLowestLosingBidPredictionSubjectDetailDetails(resp jsonAPISingleResponse) lowestLosingBidPredictionSubjectDetailDetails {
	attrs := resp.Data.Attributes
	details := lowestLosingBidPredictionSubjectDetailDetails{
		ID:                             resp.Data.ID,
		LowestBidAmount:                stringAttr(attrs, "lowest-bid-amount"),
		BidAmount:                      stringAttr(attrs, "bid-amount"),
		WalkAwayBidAmount:              stringAttr(attrs, "walk-away-bid-amount"),
		EngineerEstimateAmount:         stringAttr(attrs, "engineer-estimate-amount"),
		InternalEngineerEstimateAmount: stringAttr(attrs, "internal-engineer-estimate-amount"),
		BidDetails:                     anyAttr(attrs, "bid-details"),
	}

	if rel, ok := resp.Data.Relationships["prediction-subject"]; ok && rel.Data != nil {
		details.PredictionSubjectID = rel.Data.ID
	}

	if rel, ok := resp.Data.Relationships["prediction-subject-bids"]; ok {
		details.PredictionSubjectBidIDs = relationshipIDList(rel)
	}

	return details
}

func renderLowestLosingBidPredictionSubjectDetailDetails(cmd *cobra.Command, details lowestLosingBidPredictionSubjectDetailDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.PredictionSubjectID != "" {
		fmt.Fprintf(out, "Prediction Subject: %s\n", details.PredictionSubjectID)
	}
	if details.LowestBidAmount != "" {
		fmt.Fprintf(out, "Lowest Bid Amount: %s\n", details.LowestBidAmount)
	}
	if details.BidAmount != "" {
		fmt.Fprintf(out, "Bid Amount: %s\n", details.BidAmount)
	}
	if details.WalkAwayBidAmount != "" {
		fmt.Fprintf(out, "Walk Away Bid Amount: %s\n", details.WalkAwayBidAmount)
	}
	if details.EngineerEstimateAmount != "" {
		fmt.Fprintf(out, "Engineer Estimate Amount: %s\n", details.EngineerEstimateAmount)
	}
	if details.InternalEngineerEstimateAmount != "" {
		fmt.Fprintf(out, "Internal Engineer Estimate Amount: %s\n", details.InternalEngineerEstimateAmount)
	}

	if details.BidDetails != nil {
		fmt.Fprintf(out, "Bid Details: %d\n", countConstraintItems(details.BidDetails))
		if formatted := formatAnyJSON(details.BidDetails); formatted != "" {
			fmt.Fprintln(out, "")
			fmt.Fprintln(out, "Bid Details JSON:")
			fmt.Fprintln(out, strings.Repeat("-", 40))
			fmt.Fprintln(out, formatted)
		}
	}

	if len(details.PredictionSubjectBidIDs) > 0 {
		fmt.Fprintf(out, "Prediction Subject Bids: %s\n", strings.Join(details.PredictionSubjectBidIDs, ", "))
	}

	return nil
}
