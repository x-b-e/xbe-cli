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

type doMaterialTransactionInvalidationsCreateOptions struct {
	BaseURL             string
	Token               string
	JSON                bool
	MaterialTransaction string
	Comment             string
}

type materialTransactionInvalidationRow struct {
	ID                    string `json:"id"`
	MaterialTransactionID string `json:"material_transaction_id,omitempty"`
	Comment               string `json:"comment,omitempty"`
}

func newDoMaterialTransactionInvalidationsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Invalidate a material transaction",
		Long: `Invalidate a material transaction.

Required flags:
  --material-transaction  Material transaction ID (required)

Optional flags:
  --comment  Invalidation comment`,
		Example: `  # Invalidate a material transaction
  xbe do material-transaction-invalidations create --material-transaction 123 --comment "Duplicate ticket"`,
		Args: cobra.NoArgs,
		RunE: runDoMaterialTransactionInvalidationsCreate,
	}
	initDoMaterialTransactionInvalidationsCreateFlags(cmd)
	return cmd
}

func init() {
	doMaterialTransactionInvalidationsCmd.AddCommand(newDoMaterialTransactionInvalidationsCreateCmd())
}

func initDoMaterialTransactionInvalidationsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("material-transaction", "", "Material transaction ID (required)")
	cmd.Flags().String("comment", "", "Invalidation comment")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoMaterialTransactionInvalidationsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoMaterialTransactionInvalidationsCreateOptions(cmd)
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

	if strings.TrimSpace(opts.MaterialTransaction) == "" {
		err := fmt.Errorf("--material-transaction is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{}
	if strings.TrimSpace(opts.Comment) != "" {
		attributes["comment"] = opts.Comment
	}

	relationships := map[string]any{
		"material-transaction": map[string]any{
			"data": map[string]any{
				"type": "material-transactions",
				"id":   opts.MaterialTransaction,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "material-transaction-invalidations",
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

	body, _, err := client.Post(cmd.Context(), "/v1/material-transaction-invalidations", jsonBody)
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

	row := materialTransactionInvalidationRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created material transaction invalidation %s\n", row.ID)
	return nil
}

func materialTransactionInvalidationRowFromSingle(resp jsonAPISingleResponse) materialTransactionInvalidationRow {
	attrs := resp.Data.Attributes
	row := materialTransactionInvalidationRow{
		ID:      resp.Data.ID,
		Comment: stringAttr(attrs, "comment"),
	}

	if rel, ok := resp.Data.Relationships["material-transaction"]; ok && rel.Data != nil {
		row.MaterialTransactionID = rel.Data.ID
	}

	row.MaterialTransactionID = firstNonEmpty(
		row.MaterialTransactionID,
		stringAttr(attrs, "material-transaction-id"),
		stringAttr(attrs, "material_transaction_id"),
	)

	return row
}

func parseDoMaterialTransactionInvalidationsCreateOptions(cmd *cobra.Command) (doMaterialTransactionInvalidationsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	materialTransaction, _ := cmd.Flags().GetString("material-transaction")
	comment, _ := cmd.Flags().GetString("comment")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doMaterialTransactionInvalidationsCreateOptions{
		BaseURL:             baseURL,
		Token:               token,
		JSON:                jsonOut,
		MaterialTransaction: materialTransaction,
		Comment:             comment,
	}, nil
}
