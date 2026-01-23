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

type doTimeSheetCostCodeAllocationsCreateOptions struct {
	BaseURL    string
	Token      string
	JSON       bool
	TimeSheet  string
	DetailsRaw string
}

func newDoTimeSheetCostCodeAllocationsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a time sheet cost code allocation",
		Long: `Create a time sheet cost code allocation.

Required flags:
  --time-sheet  Time sheet ID (required)
  --details     Allocation details as JSON array (required)

Details format:
  [{"cost_code_id":"123","percentage":0.5},{"cost_code_id":"456","percentage":0.5}]

Optional detail fields:
  project_cost_classification_id`,
		Example: `  # Allocate 100% to a single cost code
  xbe do time-sheet-cost-code-allocations create \
    --time-sheet 123 \
    --details '[{"cost_code_id":"456","percentage":1}]'

  # Allocate across multiple cost codes
  xbe do time-sheet-cost-code-allocations create \
    --time-sheet 123 \
    --details '[{"cost_code_id":"456","percentage":0.6},{"cost_code_id":"789","percentage":0.4}]'

  # JSON output
  xbe do time-sheet-cost-code-allocations create \
    --time-sheet 123 \
    --details '[{"cost_code_id":"456","percentage":1}]' \
    --json`,
		Args: cobra.NoArgs,
		RunE: runDoTimeSheetCostCodeAllocationsCreate,
	}
	initDoTimeSheetCostCodeAllocationsCreateFlags(cmd)
	return cmd
}

func init() {
	doTimeSheetCostCodeAllocationsCmd.AddCommand(newDoTimeSheetCostCodeAllocationsCreateCmd())
}

func initDoTimeSheetCostCodeAllocationsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("time-sheet", "", "Time sheet ID (required)")
	cmd.Flags().String("details", "", "Allocation details JSON array (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	_ = cmd.MarkFlagRequired("time-sheet")
	_ = cmd.MarkFlagRequired("details")
}

func runDoTimeSheetCostCodeAllocationsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoTimeSheetCostCodeAllocationsCreateOptions(cmd)
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

	if strings.TrimSpace(opts.TimeSheet) == "" {
		err := fmt.Errorf("--time-sheet is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	details, err := parseCostCodeAllocationDetailsInput(opts.DetailsRaw)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{
		"details": details,
	}

	relationships := map[string]any{
		"time-sheet": map[string]any{
			"data": map[string]any{
				"type": "time-sheets",
				"id":   opts.TimeSheet,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "time-sheet-cost-code-allocations",
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

	body, _, err := client.Post(cmd.Context(), "/v1/time-sheet-cost-code-allocations", jsonBody)
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

	row := buildTimeSheetCostCodeAllocationRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created time sheet cost code allocation %s\n", row.ID)
	return nil
}

func parseDoTimeSheetCostCodeAllocationsCreateOptions(cmd *cobra.Command) (doTimeSheetCostCodeAllocationsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	timeSheet, _ := cmd.Flags().GetString("time-sheet")
	detailsRaw, _ := cmd.Flags().GetString("details")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doTimeSheetCostCodeAllocationsCreateOptions{
		BaseURL:    baseURL,
		Token:      token,
		JSON:       jsonOut,
		TimeSheet:  timeSheet,
		DetailsRaw: detailsRaw,
	}, nil
}

func parseCostCodeAllocationDetailsInput(raw string) ([]map[string]any, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, fmt.Errorf("--details is required")
	}

	var details []map[string]any
	if err := json.Unmarshal([]byte(raw), &details); err != nil {
		return nil, fmt.Errorf("invalid --details JSON (expected array of objects): %w", err)
	}
	if len(details) == 0 {
		return nil, fmt.Errorf("--details must include at least one allocation")
	}
	return details, nil
}
