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

type doTimeCardCostCodeAllocationsUpdateOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	ID      string
	Details string
}

func newDoTimeCardCostCodeAllocationsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a time card cost code allocation",
		Long: `Update a time card cost code allocation.

Optional flags:
  --details  Allocation details as JSON array

Details JSON format:
  [
    {"cost_code_id":123,"percentage":0.5},
    {"cost_code_id":456,"percentage":0.5,"project_cost_classification_id":789}
  ]

Percentages must sum to 1.0 (100%).

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Update allocation details
  xbe do time-card-cost-code-allocations update 123 \\
    --details '[{"cost_code_id":1,"percentage":1.0}]'`,
		Args: cobra.ExactArgs(1),
		RunE: runDoTimeCardCostCodeAllocationsUpdate,
	}
	initDoTimeCardCostCodeAllocationsUpdateFlags(cmd)
	return cmd
}

func init() {
	doTimeCardCostCodeAllocationsCmd.AddCommand(newDoTimeCardCostCodeAllocationsUpdateCmd())
}

func initDoTimeCardCostCodeAllocationsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("details", "", "Allocation details JSON array")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoTimeCardCostCodeAllocationsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoTimeCardCostCodeAllocationsUpdateOptions(cmd, args)
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
		if strings.TrimSpace(opts.Details) == "" {
			err := fmt.Errorf("--details cannot be empty")
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		var details []map[string]any
		if err := json.Unmarshal([]byte(opts.Details), &details); err != nil {
			err = fmt.Errorf("invalid details JSON: %w", err)
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
			"type":       "time-card-cost-code-allocations",
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

	body, _, err := client.Patch(cmd.Context(), "/v1/time-card-cost-code-allocations/"+opts.ID, jsonBody)
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

	row := timeCardCostCodeAllocationRow{ID: resp.Data.ID}
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated time card cost code allocation %s\n", row.ID)
	return nil
}

func parseDoTimeCardCostCodeAllocationsUpdateOptions(cmd *cobra.Command, args []string) (doTimeCardCostCodeAllocationsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	details, _ := cmd.Flags().GetString("details")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doTimeCardCostCodeAllocationsUpdateOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		ID:      args[0],
		Details: details,
	}, nil
}
