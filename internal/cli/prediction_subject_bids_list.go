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

type predictionSubjectBidsListOptions struct {
	BaseURL                                string
	Token                                  string
	JSON                                   bool
	NoAuth                                 bool
	Limit                                  int
	Offset                                 int
	Sort                                   string
	Bidder                                 string
	LowestLosingBidPredictionSubjectDetail string
	Broker                                 string
	PredictionSubject                      string
}

type predictionSubjectBidRow struct {
	ID                                       string   `json:"id"`
	Amount                                   *float64 `json:"amount,omitempty"`
	BidderID                                 string   `json:"bidder_id,omitempty"`
	BidderName                               string   `json:"bidder_name,omitempty"`
	LowestLosingBidPredictionSubjectDetailID string   `json:"lowest_losing_bid_prediction_subject_detail_id,omitempty"`
	PredictionSubjectID                      string   `json:"prediction_subject_id,omitempty"`
}

func newPredictionSubjectBidsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List prediction subject bids",
		Long: `List prediction subject bids with filtering and pagination.

Prediction subject bids capture bidder amounts tied to a prediction subject's
lowest losing bid detail.

Output Columns:
  ID       Prediction subject bid identifier
  AMOUNT   Bid amount
  BIDDER   Bidder name or ID
  DETAIL   Lowest losing bid detail ID
  SUBJECT  Prediction subject ID (when available)

Filters:
  --bidder                                   Filter by bidder ID
  --lowest-losing-bid-prediction-subject-detail  Filter by lowest losing bid detail ID
  --broker                                   Filter by broker ID
  --prediction-subject                       Filter by prediction subject ID

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List prediction subject bids
  xbe view prediction-subject-bids list

  # Filter by bidder
  xbe view prediction-subject-bids list --bidder 123

  # Filter by lowest losing bid detail
  xbe view prediction-subject-bids list --lowest-losing-bid-prediction-subject-detail 456

  # Filter by broker
  xbe view prediction-subject-bids list --broker 789

  # Filter by prediction subject
  xbe view prediction-subject-bids list --prediction-subject 321

  # Output as JSON
  xbe view prediction-subject-bids list --json`,
		Args: cobra.NoArgs,
		RunE: runPredictionSubjectBidsList,
	}
	initPredictionSubjectBidsListFlags(cmd)
	return cmd
}

func init() {
	predictionSubjectBidsCmd.AddCommand(newPredictionSubjectBidsListCmd())
}

func initPredictionSubjectBidsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("bidder", "", "Filter by bidder ID")
	cmd.Flags().String("lowest-losing-bid-prediction-subject-detail", "", "Filter by lowest losing bid detail ID")
	cmd.Flags().String("broker", "", "Filter by broker ID")
	cmd.Flags().String("prediction-subject", "", "Filter by prediction subject ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runPredictionSubjectBidsList(cmd *cobra.Command, _ []string) error {
	opts, err := parsePredictionSubjectBidsListOptions(cmd)
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
	query.Set("fields[prediction-subject-bids]", "amount,bidder,lowest-losing-bid-prediction-subject-detail")
	query.Set("include", "bidder,lowest-losing-bid-prediction-subject-detail")
	query.Set("fields[bidders]", "name,is-self-for-broker")
	query.Set("fields[lowest-losing-bid-prediction-subject-details]", "prediction-subject")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[bidder]", opts.Bidder)
	setFilterIfPresent(query, "filter[lowest-losing-bid-prediction-subject-detail]", opts.LowestLosingBidPredictionSubjectDetail)
	setFilterIfPresent(query, "filter[broker]", opts.Broker)
	setFilterIfPresent(query, "filter[prediction-subject]", opts.PredictionSubject)

	body, _, err := client.Get(cmd.Context(), "/v1/prediction-subject-bids", query)
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

	handled, err := renderSparseListIfRequested(cmd, resp)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if handled {
		return nil
	}

	rows := buildPredictionSubjectBidRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderPredictionSubjectBidsTable(cmd, rows)
}

func parsePredictionSubjectBidsListOptions(cmd *cobra.Command) (predictionSubjectBidsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	bidder, _ := cmd.Flags().GetString("bidder")
	lowestDetail, _ := cmd.Flags().GetString("lowest-losing-bid-prediction-subject-detail")
	broker, _ := cmd.Flags().GetString("broker")
	predictionSubject, _ := cmd.Flags().GetString("prediction-subject")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return predictionSubjectBidsListOptions{
		BaseURL:                                baseURL,
		Token:                                  token,
		JSON:                                   jsonOut,
		NoAuth:                                 noAuth,
		Limit:                                  limit,
		Offset:                                 offset,
		Sort:                                   sort,
		Bidder:                                 bidder,
		LowestLosingBidPredictionSubjectDetail: lowestDetail,
		Broker:                                 broker,
		PredictionSubject:                      predictionSubject,
	}, nil
}

func buildPredictionSubjectBidRows(resp jsonAPIResponse) []predictionSubjectBidRow {
	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	rows := make([]predictionSubjectBidRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := predictionSubjectBidRow{ID: resource.ID}
		if amount, ok := floatAttrValue(resource.Attributes, "amount"); ok {
			row.Amount = &amount
		}

		if rel, ok := resource.Relationships["bidder"]; ok && rel.Data != nil {
			row.BidderID = rel.Data.ID
			if inc, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
				row.BidderName = stringAttr(inc.Attributes, "name")
			}
		}

		if rel, ok := resource.Relationships["lowest-losing-bid-prediction-subject-detail"]; ok && rel.Data != nil {
			row.LowestLosingBidPredictionSubjectDetailID = rel.Data.ID
			if inc, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
				if detailRel, ok := inc.Relationships["prediction-subject"]; ok && detailRel.Data != nil {
					row.PredictionSubjectID = detailRel.Data.ID
				}
			}
		}

		rows = append(rows, row)
	}
	return rows
}

func renderPredictionSubjectBidsTable(cmd *cobra.Command, rows []predictionSubjectBidRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No prediction subject bids found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tAMOUNT\tBIDDER\tDETAIL\tSUBJECT")
	for _, row := range rows {
		bidderLabel := firstNonEmpty(row.BidderName, row.BidderID)
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			formatPredictionSubjectBidAmount(row.Amount),
			bidderLabel,
			row.LowestLosingBidPredictionSubjectDetailID,
			row.PredictionSubjectID,
		)
	}
	return writer.Flush()
}

func predictionSubjectBidRowFromSingle(resp jsonAPISingleResponse) predictionSubjectBidRow {
	row := predictionSubjectBidRow{ID: resp.Data.ID}
	if amount, ok := floatAttrValue(resp.Data.Attributes, "amount"); ok {
		row.Amount = &amount
	}
	if rel, ok := resp.Data.Relationships["bidder"]; ok && rel.Data != nil {
		row.BidderID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["lowest-losing-bid-prediction-subject-detail"]; ok && rel.Data != nil {
		row.LowestLosingBidPredictionSubjectDetailID = rel.Data.ID
	}
	return row
}

func formatPredictionSubjectBidAmount(amount *float64) string {
	if amount == nil {
		return ""
	}
	return strconv.FormatFloat(*amount, 'f', -1, 64)
}
