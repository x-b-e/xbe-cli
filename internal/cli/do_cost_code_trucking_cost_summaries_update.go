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

type doCostCodeTruckingCostSummariesUpdateOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	ID      string
	Broker  string
	StartOn string
	EndOn   string
}

func newDoCostCodeTruckingCostSummariesUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a cost code trucking cost summary",
		Long: `Update a cost code trucking cost summary.

Summaries are immutable once created and update attempts will return an error.

Optional flags:
  --start-on  Start date (YYYY-MM-DD)
  --end-on    End date (YYYY-MM-DD)
  --broker    Broker ID`,
		Example: `  # Attempt to update a summary
  xbe do cost-code-trucking-cost-summaries update 123 --start-on 2025-02-01 --end-on 2025-02-28`,
		Args: cobra.ExactArgs(1),
		RunE: runDoCostCodeTruckingCostSummariesUpdate,
	}
	initDoCostCodeTruckingCostSummariesUpdateFlags(cmd)
	return cmd
}

func init() {
	doCostCodeTruckingCostSummariesCmd.AddCommand(newDoCostCodeTruckingCostSummariesUpdateCmd())
}

func initDoCostCodeTruckingCostSummariesUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("start-on", "", "Start date (YYYY-MM-DD)")
	cmd.Flags().String("end-on", "", "End date (YYYY-MM-DD)")
	cmd.Flags().String("broker", "", "Broker ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoCostCodeTruckingCostSummariesUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoCostCodeTruckingCostSummariesUpdateOptions(cmd, args)
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

	attributes := map[string]any{}
	relationships := map[string]any{}

	if cmd.Flags().Changed("start-on") {
		attributes["start-on"] = opts.StartOn
	}
	if cmd.Flags().Changed("end-on") {
		attributes["end-on"] = opts.EndOn
	}
	if cmd.Flags().Changed("broker") {
		relationships["broker"] = map[string]any{
			"data": map[string]any{
				"type": "brokers",
				"id":   opts.Broker,
			},
		}
	}

	if len(attributes) == 0 && len(relationships) == 0 {
		err := fmt.Errorf("no attributes or relationships to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	data := map[string]any{
		"type": "cost-code-trucking-cost-summaries",
		"id":   opts.ID,
	}
	if len(attributes) > 0 {
		data["attributes"] = attributes
	}
	if len(relationships) > 0 {
		data["relationships"] = relationships
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

	body, _, err := client.Patch(cmd.Context(), "/v1/cost-code-trucking-cost-summaries/"+opts.ID, jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Updated cost code trucking cost summary %s\n", resp.Data.ID)
	return nil
}

func parseDoCostCodeTruckingCostSummariesUpdateOptions(cmd *cobra.Command, args []string) (doCostCodeTruckingCostSummariesUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	startOn, _ := cmd.Flags().GetString("start-on")
	endOn, _ := cmd.Flags().GetString("end-on")
	broker, _ := cmd.Flags().GetString("broker")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doCostCodeTruckingCostSummariesUpdateOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		ID:      args[0],
		Broker:  broker,
		StartOn: startOn,
		EndOn:   endOn,
	}, nil
}
