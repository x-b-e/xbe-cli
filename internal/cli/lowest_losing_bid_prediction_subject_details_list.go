package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type lowestLosingBidPredictionSubjectDetailsListOptions struct {
	BaseURL           string
	Token             string
	JSON              bool
	NoAuth            bool
	Limit             int
	Offset            int
	Sort              string
	PredictionSubject string
	Bidder            string
}

type lowestLosingBidPredictionSubjectDetailRow struct {
	ID                             string `json:"id"`
	PredictionSubjectID            string `json:"prediction_subject_id,omitempty"`
	LowestBidAmount                string `json:"lowest_bid_amount,omitempty"`
	BidAmount                      string `json:"bid_amount,omitempty"`
	WalkAwayBidAmount              string `json:"walk_away_bid_amount,omitempty"`
	EngineerEstimateAmount         string `json:"engineer_estimate_amount,omitempty"`
	InternalEngineerEstimateAmount string `json:"internal_engineer_estimate_amount,omitempty"`
}

func newLowestLosingBidPredictionSubjectDetailsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List lowest losing bid prediction subject details",
		Long: `List lowest losing bid prediction subject details with filtering and pagination.

Output Columns:
  ID          Detail identifier
  SUBJECT     Prediction subject ID
  BID         Bid amount
  LOWEST      Lowest bid amount
  WALK AWAY   Walk away bid amount
  ENG EST     Engineer estimate amount
  INT ENG EST Internal engineer estimate amount

Filters:
  --prediction-subject  Filter by prediction subject ID
  --bidder              Filter by bidder ID

Sorting:
  Use --sort to specify sort order. Prefix with - for descending.

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List lowest losing bid prediction subject details
  xbe view lowest-losing-bid-prediction-subject-details list

  # Filter by prediction subject
  xbe view lowest-losing-bid-prediction-subject-details list --prediction-subject 123

  # Filter by bidder
  xbe view lowest-losing-bid-prediction-subject-details list --bidder 456

  # Output as JSON
  xbe view lowest-losing-bid-prediction-subject-details list --json`,
		Args: cobra.NoArgs,
		RunE: runLowestLosingBidPredictionSubjectDetailsList,
	}
	initLowestLosingBidPredictionSubjectDetailsListFlags(cmd)
	return cmd
}

func init() {
	lowestLosingBidPredictionSubjectDetailsCmd.AddCommand(newLowestLosingBidPredictionSubjectDetailsListCmd())
}

func initLowestLosingBidPredictionSubjectDetailsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("prediction-subject", "", "Filter by prediction subject ID")
	cmd.Flags().String("bidder", "", "Filter by bidder ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runLowestLosingBidPredictionSubjectDetailsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseLowestLosingBidPredictionSubjectDetailsListOptions(cmd)
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
			fmt.Fprintln(cmd.ErrOrStderr(), "Authentication required. Run 'xbe auth login' first.")
			return err
		} else {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[lowest-losing-bid-prediction-subject-details]", "prediction-subject,lowest-bid-amount,bid-amount,walk-away-bid-amount,engineer-estimate-amount,internal-engineer-estimate-amount")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if strings.TrimSpace(opts.Sort) != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[prediction_subject]", opts.PredictionSubject)
	setFilterIfPresent(query, "filter[bidder]", opts.Bidder)

	body, _, err := client.Get(cmd.Context(), "/v1/lowest-losing-bid-prediction-subject-details", query)
	if err != nil {
		if len(body) > 0 {
			fmt.Fprintln(cmd.ErrOrStderr(), string(body))
		}
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	var resp jsonAPIResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	rows := buildLowestLosingBidPredictionSubjectDetailRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderLowestLosingBidPredictionSubjectDetailsTable(cmd, rows)
}

func parseLowestLosingBidPredictionSubjectDetailsListOptions(cmd *cobra.Command) (lowestLosingBidPredictionSubjectDetailsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	predictionSubject, _ := cmd.Flags().GetString("prediction-subject")
	bidder, _ := cmd.Flags().GetString("bidder")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return lowestLosingBidPredictionSubjectDetailsListOptions{
		BaseURL:           baseURL,
		Token:             token,
		JSON:              jsonOut,
		NoAuth:            noAuth,
		Limit:             limit,
		Offset:            offset,
		Sort:              sort,
		PredictionSubject: predictionSubject,
		Bidder:            bidder,
	}, nil
}

func buildLowestLosingBidPredictionSubjectDetailRows(resp jsonAPIResponse) []lowestLosingBidPredictionSubjectDetailRow {
	rows := make([]lowestLosingBidPredictionSubjectDetailRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, buildLowestLosingBidPredictionSubjectDetailRow(resource))
	}
	return rows
}

func lowestLosingBidPredictionSubjectDetailRowFromSingle(resp jsonAPISingleResponse) lowestLosingBidPredictionSubjectDetailRow {
	return buildLowestLosingBidPredictionSubjectDetailRow(resp.Data)
}

func buildLowestLosingBidPredictionSubjectDetailRow(resource jsonAPIResource) lowestLosingBidPredictionSubjectDetailRow {
	attrs := resource.Attributes
	row := lowestLosingBidPredictionSubjectDetailRow{
		ID:                             resource.ID,
		LowestBidAmount:                stringAttr(attrs, "lowest-bid-amount"),
		BidAmount:                      stringAttr(attrs, "bid-amount"),
		WalkAwayBidAmount:              stringAttr(attrs, "walk-away-bid-amount"),
		EngineerEstimateAmount:         stringAttr(attrs, "engineer-estimate-amount"),
		InternalEngineerEstimateAmount: stringAttr(attrs, "internal-engineer-estimate-amount"),
	}

	if rel, ok := resource.Relationships["prediction-subject"]; ok && rel.Data != nil {
		row.PredictionSubjectID = rel.Data.ID
	}

	return row
}

func renderLowestLosingBidPredictionSubjectDetailsTable(cmd *cobra.Command, rows []lowestLosingBidPredictionSubjectDetailRow) error {
	out := cmd.OutOrStdout()
	w := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)

	fmt.Fprintln(w, "ID\tSUBJECT\tBID\tLOWEST\tWALK AWAY\tENG EST\tINT ENG EST")
	for _, row := range rows {
		fmt.Fprintf(
			w,
			"%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.PredictionSubjectID,
			row.BidAmount,
			row.LowestBidAmount,
			row.WalkAwayBidAmount,
			row.EngineerEstimateAmount,
			row.InternalEngineerEstimateAmount,
		)
	}

	return w.Flush()
}
