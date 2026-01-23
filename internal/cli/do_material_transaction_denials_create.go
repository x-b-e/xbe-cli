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

type doMaterialTransactionDenialsCreateOptions struct {
	BaseURL               string
	Token                 string
	JSON                  bool
	MaterialTransactionID string
	Comment               string
}

type materialTransactionDenialRow struct {
	ID                    string `json:"id"`
	MaterialTransactionID string `json:"material_transaction_id,omitempty"`
	Comment               string `json:"comment,omitempty"`
}

func newDoMaterialTransactionDenialsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Deny a material transaction",
		Long: `Deny a material transaction.

Required:
  --material-transaction  Material transaction ID

Optional:
  --comment  Denial comment`,
		Example: `  # Deny a material transaction
  xbe do material-transaction-denials create \
    --material-transaction 123 \
    --comment "Load contaminated"`,
		Args: cobra.NoArgs,
		RunE: runDoMaterialTransactionDenialsCreate,
	}
	initDoMaterialTransactionDenialsCreateFlags(cmd)
	return cmd
}

func init() {
	doMaterialTransactionDenialsCmd.AddCommand(newDoMaterialTransactionDenialsCreateCmd())
}

func initDoMaterialTransactionDenialsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("material-transaction", "", "Material transaction ID")
	cmd.Flags().String("comment", "", "Denial comment")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoMaterialTransactionDenialsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoMaterialTransactionDenialsCreateOptions(cmd)
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

	if opts.MaterialTransactionID == "" {
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
				"id":   opts.MaterialTransactionID,
			},
		},
	}

	data := map[string]any{
		"type":          "material-transaction-denials",
		"relationships": relationships,
	}
	if len(attributes) > 0 {
		data["attributes"] = attributes
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

	body, _, err := client.Post(cmd.Context(), "/v1/material-transaction-denials", jsonBody)
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

	row := buildMaterialTransactionDenialRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created material transaction denial %s\n", row.ID)
	return nil
}

func parseDoMaterialTransactionDenialsCreateOptions(cmd *cobra.Command) (doMaterialTransactionDenialsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	materialTransactionID, _ := cmd.Flags().GetString("material-transaction")
	comment, _ := cmd.Flags().GetString("comment")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doMaterialTransactionDenialsCreateOptions{
		BaseURL:               baseURL,
		Token:                 token,
		JSON:                  jsonOut,
		MaterialTransactionID: materialTransactionID,
		Comment:               comment,
	}, nil
}

func buildMaterialTransactionDenialRowFromSingle(resp jsonAPISingleResponse) materialTransactionDenialRow {
	resource := resp.Data
	attrs := resource.Attributes
	row := materialTransactionDenialRow{
		ID:      resource.ID,
		Comment: stringAttr(attrs, "comment"),
	}
	if rel, ok := resource.Relationships["material-transaction"]; ok && rel.Data != nil {
		row.MaterialTransactionID = rel.Data.ID
	}
	return row
}
