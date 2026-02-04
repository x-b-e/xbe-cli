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

type transportOrderProjectTransportPlanStrategySetPredictionsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type transportOrderProjectTransportPlanStrategySetPredictionDetails struct {
	ID                   string                  `json:"id"`
	TransportOrderID     string                  `json:"transport_order_id,omitempty"`
	TransportOrderNumber string                  `json:"transport_order_number,omitempty"`
	Predictions          []strategySetPrediction `json:"predictions,omitempty"`
}

func newTransportOrderProjectTransportPlanStrategySetPredictionsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show transport order strategy set prediction details",
		Long: `Show the full details of a transport order strategy set prediction.

Output Fields:
  ID               Prediction identifier
  Transport Order  Transport order number or ID
  Predictions      Array of strategy set probability entries

Arguments:
  <id>    The prediction ID (required). You can find IDs using the list command.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show prediction details
  xbe view transport-order-project-transport-plan-strategy-set-predictions show 123

  # Get JSON output
  xbe view transport-order-project-transport-plan-strategy-set-predictions show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runTransportOrderProjectTransportPlanStrategySetPredictionsShow,
	}
	initTransportOrderProjectTransportPlanStrategySetPredictionsShowFlags(cmd)
	return cmd
}

func init() {
	transportOrderProjectTransportPlanStrategySetPredictionsCmd.AddCommand(newTransportOrderProjectTransportPlanStrategySetPredictionsShowCmd())
}

func initTransportOrderProjectTransportPlanStrategySetPredictionsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runTransportOrderProjectTransportPlanStrategySetPredictionsShow(cmd *cobra.Command, args []string) error {
	if handled, err := maybeHandleClientURLShow(cmd, args); err != nil {
		return err
	} else if handled {
		return nil
	}

	opts, err := parseTransportOrderProjectTransportPlanStrategySetPredictionsShowOptions(cmd)
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
		return fmt.Errorf("prediction id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[transport-order-project-transport-plan-strategy-set-predictions]", "predictions,transport-order")
	query.Set("fields[transport-orders]", "external-order-number")
	query.Set("include", "transport-order")

	body, _, err := client.Get(cmd.Context(), "/v1/transport-order-project-transport-plan-strategy-set-predictions/"+id, query)
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

	details := buildTransportOrderProjectTransportPlanStrategySetPredictionDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderTransportOrderProjectTransportPlanStrategySetPredictionDetails(cmd, details)
}

func parseTransportOrderProjectTransportPlanStrategySetPredictionsShowOptions(cmd *cobra.Command) (transportOrderProjectTransportPlanStrategySetPredictionsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return transportOrderProjectTransportPlanStrategySetPredictionsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildTransportOrderProjectTransportPlanStrategySetPredictionDetails(resp jsonAPISingleResponse) transportOrderProjectTransportPlanStrategySetPredictionDetails {
	resource := resp.Data
	attrs := resource.Attributes

	details := transportOrderProjectTransportPlanStrategySetPredictionDetails{
		ID:          resource.ID,
		Predictions: parseStrategySetPredictions(attrs["predictions"]),
	}

	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	if rel, ok := resource.Relationships["transport-order"]; ok && rel.Data != nil {
		details.TransportOrderID = rel.Data.ID
		if order, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.TransportOrderNumber = firstNonEmpty(
				stringAttr(order.Attributes, "external-order-number"),
				stringAttr(order.Attributes, "order-number"),
			)
		}
	}

	return details
}

func renderTransportOrderProjectTransportPlanStrategySetPredictionDetails(cmd *cobra.Command, details transportOrderProjectTransportPlanStrategySetPredictionDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.TransportOrderID != "" {
		transportOrder := details.TransportOrderID
		if details.TransportOrderNumber != "" {
			transportOrder = fmt.Sprintf("%s (%s)", details.TransportOrderNumber, details.TransportOrderID)
		}
		fmt.Fprintf(out, "Transport Order: %s\n", transportOrder)
	}

	if len(details.Predictions) > 0 {
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "Predictions:")
		fmt.Fprintln(out, strings.Repeat("-", 40))
		fmt.Fprintln(out, formatPredictionJSON(details.Predictions))
	}

	return nil
}

func formatPredictionJSON(value any) string {
	encoded, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return fmt.Sprintf("%v", value)
	}
	return string(encoded)
}
