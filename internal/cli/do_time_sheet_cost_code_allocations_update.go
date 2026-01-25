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

type doTimeSheetCostCodeAllocationsUpdateOptions struct {
	BaseURL    string
	Token      string
	JSON       bool
	ID         string
	DetailsRaw string
}

func newDoTimeSheetCostCodeAllocationsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a time sheet cost code allocation",
		Long: `Update a time sheet cost code allocation.

Optional flags:
  --details  Allocation details as JSON array`,
		Example: `  # Update allocations
  xbe do time-sheet-cost-code-allocations update 123 \
    --details '[{"cost_code_id":"456","percentage":1}]'`,
		Args: cobra.ExactArgs(1),
		RunE: runDoTimeSheetCostCodeAllocationsUpdate,
	}
	initDoTimeSheetCostCodeAllocationsUpdateFlags(cmd)
	return cmd
}

func init() {
	doTimeSheetCostCodeAllocationsCmd.AddCommand(newDoTimeSheetCostCodeAllocationsUpdateCmd())
}

func initDoTimeSheetCostCodeAllocationsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("details", "", "Allocation details JSON array")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoTimeSheetCostCodeAllocationsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoTimeSheetCostCodeAllocationsUpdateOptions(cmd, args)
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
	if cmd.Flags().Changed("details") {
		details, err := parseCostCodeAllocationDetailsInput(opts.DetailsRaw)
		if err != nil {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		attributes["details"] = details
	}

	if len(attributes) == 0 {
		err := fmt.Errorf("no attributes to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "time-sheet-cost-code-allocations",
			"id":         opts.ID,
			"attributes": attributes,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Patch(cmd.Context(), "/v1/time-sheet-cost-code-allocations/"+opts.ID, jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Updated time sheet cost code allocation %s\n", row.ID)
	return nil
}

func parseDoTimeSheetCostCodeAllocationsUpdateOptions(cmd *cobra.Command, args []string) (doTimeSheetCostCodeAllocationsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	detailsRaw, _ := cmd.Flags().GetString("details")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doTimeSheetCostCodeAllocationsUpdateOptions{
		BaseURL:    baseURL,
		Token:      token,
		JSON:       jsonOut,
		ID:         args[0],
		DetailsRaw: detailsRaw,
	}, nil
}
