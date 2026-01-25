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

type transportOrderProjectTransportPlanStrategySetPredictionsListOptions struct {
	BaseURL        string
	Token          string
	JSON           bool
	NoAuth         bool
	Limit          int
	Offset         int
	Sort           string
	TransportOrder string
}

type transportOrderProjectTransportPlanStrategySetPredictionRow struct {
	ID                   string  `json:"id"`
	TransportOrderID     string  `json:"transport_order_id,omitempty"`
	TransportOrderNumber string  `json:"transport_order_number,omitempty"`
	PredictionsCount     int     `json:"predictions_count,omitempty"`
	TopStrategySetID     string  `json:"top_strategy_set_id,omitempty"`
	TopProbability       float64 `json:"top_probability,omitempty"`
}

func newTransportOrderProjectTransportPlanStrategySetPredictionsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List transport order strategy set predictions",
		Long: `List transport order strategy set predictions.

Output Columns:
  ID              Prediction identifier
  TRANSPORT_ORDER Transport order number or ID
  TOP_STRATEGY_SET Highest scoring strategy set ID
  TOP_PROB        Highest score from the prediction set
  PREDICTIONS     Number of predicted strategy sets

Filters:
  --transport-order  Filter by transport order ID

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List transport order strategy set predictions
  xbe view transport-order-project-transport-plan-strategy-set-predictions list

  # Filter by transport order
  xbe view transport-order-project-transport-plan-strategy-set-predictions list --transport-order 123

  # Paginate results
  xbe view transport-order-project-transport-plan-strategy-set-predictions list --limit 20 --offset 40

  # Output as JSON
  xbe view transport-order-project-transport-plan-strategy-set-predictions list --json`,
		Args: cobra.NoArgs,
		RunE: runTransportOrderProjectTransportPlanStrategySetPredictionsList,
	}
	initTransportOrderProjectTransportPlanStrategySetPredictionsListFlags(cmd)
	return cmd
}

func init() {
	transportOrderProjectTransportPlanStrategySetPredictionsCmd.AddCommand(newTransportOrderProjectTransportPlanStrategySetPredictionsListCmd())
}

func initTransportOrderProjectTransportPlanStrategySetPredictionsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("transport-order", "", "Filter by transport order ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runTransportOrderProjectTransportPlanStrategySetPredictionsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseTransportOrderProjectTransportPlanStrategySetPredictionsListOptions(cmd)
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
	query.Set("fields[transport-order-project-transport-plan-strategy-set-predictions]", "predictions,transport-order")
	query.Set("fields[transport-orders]", "external-order-number")
	query.Set("include", "transport-order")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[transport-order]", opts.TransportOrder)

	body, _, err := client.Get(cmd.Context(), "/v1/transport-order-project-transport-plan-strategy-set-predictions", query)
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

	rows := buildTransportOrderProjectTransportPlanStrategySetPredictionRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderTransportOrderProjectTransportPlanStrategySetPredictionsTable(cmd, rows)
}

func parseTransportOrderProjectTransportPlanStrategySetPredictionsListOptions(cmd *cobra.Command) (transportOrderProjectTransportPlanStrategySetPredictionsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	transportOrder, _ := cmd.Flags().GetString("transport-order")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return transportOrderProjectTransportPlanStrategySetPredictionsListOptions{
		BaseURL:        baseURL,
		Token:          token,
		JSON:           jsonOut,
		NoAuth:         noAuth,
		Limit:          limit,
		Offset:         offset,
		Sort:           sort,
		TransportOrder: transportOrder,
	}, nil
}

func buildTransportOrderProjectTransportPlanStrategySetPredictionRows(resp jsonAPIResponse) []transportOrderProjectTransportPlanStrategySetPredictionRow {
	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	rows := make([]transportOrderProjectTransportPlanStrategySetPredictionRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		predictions := parseStrategySetPredictions(resource.Attributes["predictions"])
		topPrediction, hasTop := topStrategySetPrediction(predictions)

		row := transportOrderProjectTransportPlanStrategySetPredictionRow{
			ID:               resource.ID,
			PredictionsCount: len(predictions),
		}

		if hasTop {
			row.TopStrategySetID = topPrediction.StrategySetID
			row.TopProbability = topPrediction.Probability
		}

		if rel, ok := resource.Relationships["transport-order"]; ok && rel.Data != nil {
			row.TransportOrderID = rel.Data.ID
			if order, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
				row.TransportOrderNumber = firstNonEmpty(
					stringAttr(order.Attributes, "external-order-number"),
					stringAttr(order.Attributes, "order-number"),
				)
			}
		}

		rows = append(rows, row)
	}

	return rows
}

func renderTransportOrderProjectTransportPlanStrategySetPredictionsTable(cmd *cobra.Command, rows []transportOrderProjectTransportPlanStrategySetPredictionRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No transport order strategy set predictions found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tTRANSPORT_ORDER\tTOP_STRATEGY_SET\tTOP_PROB\tPREDICTIONS")

	for _, row := range rows {
		transportOrder := row.TransportOrderID
		if row.TransportOrderNumber != "" {
			transportOrder = row.TransportOrderNumber
		}

		topProbability := ""
		if row.PredictionsCount > 0 {
			topProbability = fmt.Sprintf("%.4f", row.TopProbability)
		}

		fmt.Fprintf(
			writer,
			"%s\t%s\t%s\t%s\t%d\n",
			row.ID,
			transportOrder,
			row.TopStrategySetID,
			topProbability,
			row.PredictionsCount,
		)
	}

	return writer.Flush()
}
