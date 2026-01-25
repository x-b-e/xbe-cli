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

type doMaterialTransactionsExportsCreateOptions struct {
	BaseURL                 string
	Token                   string
	JSON                    bool
	OrganizationFormatterID string
	MaterialTransactionIDs  []string
}

func newDoMaterialTransactionsExportsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a material transaction export",
		Long: `Create a material transaction export.

Required flags:
  --organization-formatter   Organization formatter ID (required)
  --material-transaction-ids Material transaction IDs (comma-separated or repeated) (required)

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Create an export with a single transaction
  xbe do material-transactions-exports create \
    --organization-formatter 123 \
    --material-transaction-ids 456

  # Create an export with multiple transactions
  xbe do material-transactions-exports create \
    --organization-formatter 123 \
    --material-transaction-ids 456,789

  # JSON output
  xbe do material-transactions-exports create \
    --organization-formatter 123 \
    --material-transaction-ids 456 --json`,
		Args: cobra.NoArgs,
		RunE: runDoMaterialTransactionsExportsCreate,
	}
	initDoMaterialTransactionsExportsCreateFlags(cmd)
	return cmd
}

func init() {
	doMaterialTransactionsExportsCmd.AddCommand(newDoMaterialTransactionsExportsCreateCmd())
}

func initDoMaterialTransactionsExportsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("organization-formatter", "", "Organization formatter ID (required)")
	cmd.Flags().StringSlice("material-transaction-ids", nil, "Material transaction IDs (comma-separated or repeated) (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	cmd.MarkFlagRequired("organization-formatter")
	cmd.MarkFlagRequired("material-transaction-ids")
}

func runDoMaterialTransactionsExportsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoMaterialTransactionsExportsCreateOptions(cmd)
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

	materialTransactionIDs := compactStringSlice(opts.MaterialTransactionIDs)
	if len(materialTransactionIDs) == 0 {
		err := fmt.Errorf("--material-transaction-ids is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if strings.TrimSpace(opts.OrganizationFormatterID) == "" {
		err := fmt.Errorf("--organization-formatter is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	relationships := map[string]any{
		"organization-formatter": map[string]any{
			"data": map[string]any{
				"type": "organization-formatters",
				"id":   opts.OrganizationFormatterID,
			},
		},
		"material-transactions": map[string]any{
			"data": buildRelationshipData("material-transactions", materialTransactionIDs),
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "material-transactions-exports",
			"relationships": relationships,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/material-transactions-exports", jsonBody)
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

	row := buildMaterialTransactionsExportRow(resp.Data)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created material transaction export %s\n", row.ID)
	return nil
}

func parseDoMaterialTransactionsExportsCreateOptions(cmd *cobra.Command) (doMaterialTransactionsExportsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	organizationFormatterID, _ := cmd.Flags().GetString("organization-formatter")
	materialTransactionIDs, _ := cmd.Flags().GetStringSlice("material-transaction-ids")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doMaterialTransactionsExportsCreateOptions{
		BaseURL:                 baseURL,
		Token:                   token,
		JSON:                    jsonOut,
		OrganizationFormatterID: organizationFormatterID,
		MaterialTransactionIDs:  materialTransactionIDs,
	}, nil
}
