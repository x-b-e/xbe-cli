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

type doMaterialTransactionCostCodeAllocationsUpdateOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	ID      string
	Details string
}

func newDoMaterialTransactionCostCodeAllocationsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a material transaction cost code allocation",
		Long: `Update a material transaction cost code allocation.

Writable attributes:
  --details    JSON array of cost code allocations

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Update allocation details
  xbe do material-transaction-cost-code-allocations update 123 \
    --details '[{"cost_code_id":"456","percentage":1}]'`,
		Args: cobra.ExactArgs(1),
		RunE: runDoMaterialTransactionCostCodeAllocationsUpdate,
	}
	initDoMaterialTransactionCostCodeAllocationsUpdateFlags(cmd)
	return cmd
}

func init() {
	doMaterialTransactionCostCodeAllocationsCmd.AddCommand(newDoMaterialTransactionCostCodeAllocationsUpdateCmd())
}

func initDoMaterialTransactionCostCodeAllocationsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("details", "", "Cost code allocation details JSON array")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoMaterialTransactionCostCodeAllocationsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoMaterialTransactionCostCodeAllocationsUpdateOptions(cmd, args)
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

	id := strings.TrimSpace(opts.ID)
	if id == "" {
		return fmt.Errorf("material transaction cost code allocation id is required")
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
			err := fmt.Errorf("invalid details JSON: %w", err)
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		if len(details) == 0 {
			err := fmt.Errorf("--details must include at least one allocation")
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
			"type":       "material-transaction-cost-code-allocations",
			"id":         id,
			"attributes": attributes,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Patch(cmd.Context(), "/v1/material-transaction-cost-code-allocations/"+id, jsonBody)
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
		row := buildMaterialTransactionCostCodeAllocationDetails(resp)
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated material transaction cost code allocation %s\n", resp.Data.ID)
	return nil
}

func parseDoMaterialTransactionCostCodeAllocationsUpdateOptions(cmd *cobra.Command, args []string) (doMaterialTransactionCostCodeAllocationsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	details, _ := cmd.Flags().GetString("details")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doMaterialTransactionCostCodeAllocationsUpdateOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		ID:      args[0],
		Details: details,
	}, nil
}
