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

type doMaterialTransactionAcceptancesCreateOptions struct {
	BaseURL                      string
	Token                        string
	JSON                         bool
	MaterialTransaction          string
	Comment                      string
	SkipNotOverlappingValidation bool
}

type materialTransactionAcceptanceRow struct {
	ID                           string `json:"id"`
	MaterialTransactionID        string `json:"material_transaction_id,omitempty"`
	Comment                      string `json:"comment,omitempty"`
	SkipNotOverlappingValidation bool   `json:"skip_not_overlapping_validation"`
}

func newDoMaterialTransactionAcceptancesCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Accept a material transaction",
		Long: `Accept a material transaction.

Required flags:
  --material-transaction  Material transaction ID (required)

Optional flags:
  --comment                         Acceptance comment
  --skip-not-overlapping-validation Skip ticket overlap validation`,
		Example: `  # Accept a material transaction
  xbe do material-transaction-acceptances create --material-transaction 123 --comment "Reviewed"

  # Accept while skipping overlap validation
  xbe do material-transaction-acceptances create --material-transaction 123 --skip-not-overlapping-validation`,
		Args: cobra.NoArgs,
		RunE: runDoMaterialTransactionAcceptancesCreate,
	}
	initDoMaterialTransactionAcceptancesCreateFlags(cmd)
	return cmd
}

func init() {
	doMaterialTransactionAcceptancesCmd.AddCommand(newDoMaterialTransactionAcceptancesCreateCmd())
}

func initDoMaterialTransactionAcceptancesCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("material-transaction", "", "Material transaction ID (required)")
	cmd.Flags().String("comment", "", "Acceptance comment")
	cmd.Flags().Bool("skip-not-overlapping-validation", false, "Skip ticket overlap validation")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoMaterialTransactionAcceptancesCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoMaterialTransactionAcceptancesCreateOptions(cmd)
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
	if cmd.Flags().Changed("skip-not-overlapping-validation") {
		attributes["skip-not-overlapping-validation"] = opts.SkipNotOverlappingValidation
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
			"type":          "material-transaction-acceptances",
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

	body, _, err := client.Post(cmd.Context(), "/v1/material-transaction-acceptances", jsonBody)
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

	row := materialTransactionAcceptanceRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created material transaction acceptance %s\n", row.ID)
	return nil
}

func materialTransactionAcceptanceRowFromSingle(resp jsonAPISingleResponse) materialTransactionAcceptanceRow {
	attrs := resp.Data.Attributes
	row := materialTransactionAcceptanceRow{
		ID:                           resp.Data.ID,
		Comment:                      stringAttr(attrs, "comment"),
		SkipNotOverlappingValidation: boolAttr(attrs, "skip-not-overlapping-validation"),
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

func parseDoMaterialTransactionAcceptancesCreateOptions(cmd *cobra.Command) (doMaterialTransactionAcceptancesCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	materialTransaction, _ := cmd.Flags().GetString("material-transaction")
	comment, _ := cmd.Flags().GetString("comment")
	skipNotOverlappingValidation, _ := cmd.Flags().GetBool("skip-not-overlapping-validation")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doMaterialTransactionAcceptancesCreateOptions{
		BaseURL:                      baseURL,
		Token:                        token,
		JSON:                         jsonOut,
		MaterialTransaction:          materialTransaction,
		Comment:                      comment,
		SkipNotOverlappingValidation: skipNotOverlappingValidation,
	}, nil
}
