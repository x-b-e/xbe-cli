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

type materialTransactionCostCodeAllocationsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type materialTransactionCostCodeAllocationDetails struct {
	ID                    string   `json:"id"`
	MaterialTransactionID string   `json:"material_transaction_id,omitempty"`
	CostCodeIDs           []string `json:"cost_code_ids,omitempty"`
	CostCodes             []string `json:"cost_codes,omitempty"`
	Details               any      `json:"details,omitempty"`
	CreatedAt             string   `json:"created_at,omitempty"`
	UpdatedAt             string   `json:"updated_at,omitempty"`
}

func newMaterialTransactionCostCodeAllocationsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show material transaction cost code allocation details",
		Long: `Show the full details of a material transaction cost code allocation.

Output Fields:
  ID
  Material Transaction ID
  Cost Code IDs
  Cost Codes
  Details
  Created At
  Updated At

Arguments:
  <id>    The cost code allocation ID (required). You can find IDs using the list command.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show allocation details
  xbe view material-transaction-cost-code-allocations show 123

  # Get JSON output
  xbe view material-transaction-cost-code-allocations show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runMaterialTransactionCostCodeAllocationsShow,
	}
	initMaterialTransactionCostCodeAllocationsShowFlags(cmd)
	return cmd
}

func init() {
	materialTransactionCostCodeAllocationsCmd.AddCommand(newMaterialTransactionCostCodeAllocationsShowCmd())
}

func initMaterialTransactionCostCodeAllocationsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runMaterialTransactionCostCodeAllocationsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseMaterialTransactionCostCodeAllocationsShowOptions(cmd)
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
		return fmt.Errorf("material transaction cost code allocation id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[material-transaction-cost-code-allocations]", "details,created-at,updated-at,material-transaction,cost-codes")
	query.Set("fields[cost-codes]", "code,description")
	query.Set("include", "cost-codes")

	body, _, err := client.Get(cmd.Context(), "/v1/material-transaction-cost-code-allocations/"+id, query)
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

	details := buildMaterialTransactionCostCodeAllocationDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderMaterialTransactionCostCodeAllocationDetails(cmd, details)
}

func parseMaterialTransactionCostCodeAllocationsShowOptions(cmd *cobra.Command) (materialTransactionCostCodeAllocationsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return materialTransactionCostCodeAllocationsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildMaterialTransactionCostCodeAllocationDetails(resp jsonAPISingleResponse) materialTransactionCostCodeAllocationDetails {
	resource := resp.Data
	attrs := resource.Attributes
	included := map[string]map[string]any{}
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc.Attributes
	}

	costCodeIDs := relationshipIDsFromMap(resource.Relationships, "cost-codes")

	return materialTransactionCostCodeAllocationDetails{
		ID:                    resource.ID,
		MaterialTransactionID: relationshipIDFromMap(resource.Relationships, "material-transaction"),
		CostCodeIDs:           costCodeIDs,
		CostCodes:             resolveCostCodeCodes(costCodeIDs, included),
		Details:               anyAttr(attrs, "details"),
		CreatedAt:             formatDateTime(stringAttr(attrs, "created-at")),
		UpdatedAt:             formatDateTime(stringAttr(attrs, "updated-at")),
	}
}

func renderMaterialTransactionCostCodeAllocationDetails(cmd *cobra.Command, details materialTransactionCostCodeAllocationDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.MaterialTransactionID != "" {
		fmt.Fprintf(out, "Material Transaction ID: %s\n", details.MaterialTransactionID)
	}
	if len(details.CostCodes) > 0 {
		fmt.Fprintf(out, "Cost Codes: %s\n", strings.Join(details.CostCodes, ", "))
	} else if len(details.CostCodeIDs) > 0 {
		fmt.Fprintf(out, "Cost Code IDs: %s\n", strings.Join(details.CostCodeIDs, ", "))
	}
	if details.CreatedAt != "" {
		fmt.Fprintf(out, "Created At: %s\n", details.CreatedAt)
	}
	if details.UpdatedAt != "" {
		fmt.Fprintf(out, "Updated At: %s\n", details.UpdatedAt)
	}
	if details.Details != nil {
		fmt.Fprintln(out, "Details:")
		if err := writeJSON(out, details.Details); err != nil {
			return err
		}
	}

	return nil
}
