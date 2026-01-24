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

type doCostCodeTruckingCostSummariesCreateOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	Broker  string
	StartOn string
	EndOn   string
}

func newDoCostCodeTruckingCostSummariesCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a cost code trucking cost summary",
		Long: `Create a cost code trucking cost summary.

Required flags:
  --broker    Broker ID
  --start-on  Start date (YYYY-MM-DD)
  --end-on    End date (YYYY-MM-DD)`,
		Example: `  # Create a summary
  xbe do cost-code-trucking-cost-summaries create --broker 123 --start-on 2025-01-01 --end-on 2025-01-31

  # Output JSON
  xbe do cost-code-trucking-cost-summaries create --broker 123 --start-on 2025-01-01 --end-on 2025-01-31 --json`,
		Args: cobra.NoArgs,
		RunE: runDoCostCodeTruckingCostSummariesCreate,
	}
	initDoCostCodeTruckingCostSummariesCreateFlags(cmd)
	return cmd
}

func init() {
	doCostCodeTruckingCostSummariesCmd.AddCommand(newDoCostCodeTruckingCostSummariesCreateCmd())
}

func initDoCostCodeTruckingCostSummariesCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("broker", "", "Broker ID (required)")
	cmd.Flags().String("start-on", "", "Start date (YYYY-MM-DD) (required)")
	cmd.Flags().String("end-on", "", "End date (YYYY-MM-DD) (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	_ = cmd.MarkFlagRequired("broker")
	_ = cmd.MarkFlagRequired("start-on")
	_ = cmd.MarkFlagRequired("end-on")
}

func runDoCostCodeTruckingCostSummariesCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoCostCodeTruckingCostSummariesCreateOptions(cmd)
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

	attributes := map[string]any{
		"start-on": opts.StartOn,
		"end-on":   opts.EndOn,
	}

	relationships := map[string]any{
		"broker": map[string]any{
			"data": map[string]any{
				"type": "brokers",
				"id":   opts.Broker,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "cost-code-trucking-cost-summaries",
			"attributes":    attributes,
			"relationships": relationships,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/cost-code-trucking-cost-summaries", jsonBody)
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

	if opts.JSON {
		row := buildCostCodeTruckingCostSummaryRow(resp.Data)
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created cost code trucking cost summary %s\n", resp.Data.ID)
	return nil
}

func parseDoCostCodeTruckingCostSummariesCreateOptions(cmd *cobra.Command) (doCostCodeTruckingCostSummariesCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	broker, _ := cmd.Flags().GetString("broker")
	startOn, _ := cmd.Flags().GetString("start-on")
	endOn, _ := cmd.Flags().GetString("end-on")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doCostCodeTruckingCostSummariesCreateOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		Broker:  broker,
		StartOn: startOn,
		EndOn:   endOn,
	}, nil
}
