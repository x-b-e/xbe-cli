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

type doMaterialTransactionCostCodeAllocationsCreateOptions struct {
	BaseURL             string
	Token               string
	JSON                bool
	MaterialTransaction string
	Details             string
}

func newDoMaterialTransactionCostCodeAllocationsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a material transaction cost code allocation",
		Long: `Create a material transaction cost code allocation.

Required flags:
  --material-transaction   Material transaction ID
  --details                JSON array of cost code allocations

Each details entry should include:
  cost_code_id                  Cost code ID (required)
  percentage                    Allocation percentage (0-1, must sum to 1)
  project_cost_classification_id Optional project cost classification ID

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Allocate 100% to one cost code
  xbe do material-transaction-cost-code-allocations create \
    --material-transaction 123 \
    --details '[{"cost_code_id":"456","percentage":1}]'

  # Split across two cost codes
  xbe do material-transaction-cost-code-allocations create \
    --material-transaction 123 \
    --details '[{"cost_code_id":"456","percentage":0.5},{"cost_code_id":"789","percentage":0.5}]'`,
		Args: cobra.NoArgs,
		RunE: runDoMaterialTransactionCostCodeAllocationsCreate,
	}
	initDoMaterialTransactionCostCodeAllocationsCreateFlags(cmd)
	return cmd
}

func init() {
	doMaterialTransactionCostCodeAllocationsCmd.AddCommand(newDoMaterialTransactionCostCodeAllocationsCreateCmd())
}

func initDoMaterialTransactionCostCodeAllocationsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("material-transaction", "", "Material transaction ID (required)")
	cmd.Flags().String("details", "", "Cost code allocation details JSON array (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoMaterialTransactionCostCodeAllocationsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoMaterialTransactionCostCodeAllocationsCreateOptions(cmd)
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

	materialTransactionID := strings.TrimSpace(opts.MaterialTransaction)
	if materialTransactionID == "" {
		err := fmt.Errorf("--material-transaction is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if strings.TrimSpace(opts.Details) == "" {
		err := fmt.Errorf("--details is required")
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

	attributes := map[string]any{
		"details": details,
	}
	relationships := map[string]any{
		"material-transaction": map[string]any{
			"data": map[string]any{
				"type": "material-transactions",
				"id":   materialTransactionID,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "material-transaction-cost-code-allocations",
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

	body, _, err := client.Post(cmd.Context(), "/v1/material-transaction-cost-code-allocations", jsonBody)
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

	row := buildMaterialTransactionCostCodeAllocationDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created material transaction cost code allocation %s\n", row.ID)
	return nil
}

func parseDoMaterialTransactionCostCodeAllocationsCreateOptions(cmd *cobra.Command) (doMaterialTransactionCostCodeAllocationsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	materialTransaction, _ := cmd.Flags().GetString("material-transaction")
	details, _ := cmd.Flags().GetString("details")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doMaterialTransactionCostCodeAllocationsCreateOptions{
		BaseURL:             baseURL,
		Token:               token,
		JSON:                jsonOut,
		MaterialTransaction: materialTransaction,
		Details:             details,
	}, nil
}
