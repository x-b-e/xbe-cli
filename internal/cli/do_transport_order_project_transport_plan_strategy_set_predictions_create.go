package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type doTransportOrderProjectTransportPlanStrategySetPredictionsCreateOptions struct {
	BaseURL        string
	Token          string
	JSON           bool
	TransportOrder string
}

func newDoTransportOrderProjectTransportPlanStrategySetPredictionsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Generate transport order strategy set predictions",
		Long: `Generate a transport order strategy set prediction record.

Required flags:
  --transport-order  Transport order ID

Predictions are generated automatically by the server and returned in the response.

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Generate predictions for a transport order
  xbe do transport-order-project-transport-plan-strategy-set-predictions create --transport-order 123`,
		RunE: runDoTransportOrderProjectTransportPlanStrategySetPredictionsCreate,
	}
	initDoTransportOrderProjectTransportPlanStrategySetPredictionsCreateFlags(cmd)
	return cmd
}

func init() {
	doTransportOrderProjectTransportPlanStrategySetPredictionsCmd.AddCommand(newDoTransportOrderProjectTransportPlanStrategySetPredictionsCreateCmd())
}

func initDoTransportOrderProjectTransportPlanStrategySetPredictionsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("transport-order", "", "Transport order ID (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	cmd.MarkFlagRequired("transport-order")
}

func runDoTransportOrderProjectTransportPlanStrategySetPredictionsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoTransportOrderProjectTransportPlanStrategySetPredictionsCreateOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if strings.TrimSpace(opts.Token) == "" {
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

	relationships := map[string]any{
		"transport-order": map[string]any{
			"data": map[string]any{
				"type": "transport-orders",
				"id":   opts.TransportOrder,
			},
		},
	}

	data := map[string]any{
		"type":          "transport-order-project-transport-plan-strategy-set-predictions",
		"relationships": relationships,
	}

	requestBody := map[string]any{
		"data": data,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/transport-order-project-transport-plan-strategy-set-predictions", jsonBody)
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

	predictions := parseStrategySetPredictions(resp.Data.Attributes["predictions"])

	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), map[string]any{
			"id":                 resp.Data.ID,
			"transport_order_id": opts.TransportOrder,
			"predictions_count":  len(predictions),
		})
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created transport order strategy set prediction %s\n", resp.Data.ID)
	return nil
}

func parseDoTransportOrderProjectTransportPlanStrategySetPredictionsCreateOptions(cmd *cobra.Command) (doTransportOrderProjectTransportPlanStrategySetPredictionsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	transportOrder, _ := cmd.Flags().GetString("transport-order")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doTransportOrderProjectTransportPlanStrategySetPredictionsCreateOptions{
		BaseURL:        baseURL,
		Token:          token,
		JSON:           jsonOut,
		TransportOrder: transportOrder,
	}, nil
}
