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

type doMaterialTransactionInspectionsCreateOptions struct {
	BaseURL             string
	Token               string
	JSON                bool
	Note                string
	Status              string
	Strategy            string
	MaterialTransaction string
}

type materialTransactionInspectionCreateRow struct {
	ID                    string `json:"id"`
	Status                string `json:"status,omitempty"`
	Strategy              string `json:"strategy,omitempty"`
	Note                  string `json:"note,omitempty"`
	MaterialTransactionID string `json:"material_transaction_id,omitempty"`
}

func newDoMaterialTransactionInspectionsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a material transaction inspection",
		Long: `Create a material transaction inspection.

Required flags:
  --material-transaction   Material transaction ID

Optional attributes:
  --status                 Inspection status (open,closed)
  --strategy               Inspection strategy (delivery_site_personnel)
  --note                   Inspection note`,
		Example: `  # Create an inspection for a material transaction
  xbe do material-transaction-inspections create \\
    --material-transaction 123 \\
    --status open \\
    --strategy delivery_site_personnel \\
    --note "Checked at gate"`,
		RunE: runDoMaterialTransactionInspectionsCreate,
	}
	initDoMaterialTransactionInspectionsCreateFlags(cmd)
	return cmd
}

func init() {
	doMaterialTransactionInspectionsCmd.AddCommand(newDoMaterialTransactionInspectionsCreateCmd())
}

func initDoMaterialTransactionInspectionsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("material-transaction", "", "Material transaction ID (required)")
	cmd.Flags().String("status", "", "Inspection status (open,closed)")
	cmd.Flags().String("strategy", "", "Inspection strategy (delivery_site_personnel)")
	cmd.Flags().String("note", "", "Inspection note")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	cmd.MarkFlagRequired("material-transaction")
}

func runDoMaterialTransactionInspectionsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoMaterialTransactionInspectionsCreateOptions(cmd)
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
	if opts.Note != "" {
		attributes["note"] = opts.Note
	}
	if opts.Status != "" {
		attributes["status"] = opts.Status
	}
	if opts.Strategy != "" {
		attributes["strategy"] = opts.Strategy
	}

	relationships := map[string]any{}
	if opts.MaterialTransaction != "" {
		relationships["material-transaction"] = map[string]any{
			"data": map[string]any{
				"type": "material-transactions",
				"id":   opts.MaterialTransaction,
			},
		}
	}

	data := map[string]any{
		"type":       "material-transaction-inspections",
		"attributes": attributes,
	}

	if len(relationships) > 0 {
		data["relationships"] = relationships
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

	body, _, err := client.Post(cmd.Context(), "/v1/material-transaction-inspections", jsonBody)
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

	row := materialTransactionInspectionRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created material transaction inspection %s\n", row.ID)
	return nil
}

func parseDoMaterialTransactionInspectionsCreateOptions(cmd *cobra.Command) (doMaterialTransactionInspectionsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	materialTransaction, _ := cmd.Flags().GetString("material-transaction")
	status, _ := cmd.Flags().GetString("status")
	strategy, _ := cmd.Flags().GetString("strategy")
	note, _ := cmd.Flags().GetString("note")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doMaterialTransactionInspectionsCreateOptions{
		BaseURL:             baseURL,
		Token:               token,
		JSON:                jsonOut,
		Note:                note,
		Status:              status,
		Strategy:            strategy,
		MaterialTransaction: materialTransaction,
	}, nil
}

func materialTransactionInspectionRowFromSingle(resp jsonAPISingleResponse) materialTransactionInspectionCreateRow {
	row := materialTransactionInspectionCreateRow{
		ID:                    resp.Data.ID,
		Status:                stringAttr(resp.Data.Attributes, "status"),
		Strategy:              stringAttr(resp.Data.Attributes, "strategy"),
		Note:                  stringAttr(resp.Data.Attributes, "note"),
		MaterialTransactionID: stringAttr(resp.Data.Attributes, "material-transaction-id"),
	}

	if rel, ok := resp.Data.Relationships["material-transaction"]; ok && rel.Data != nil {
		row.MaterialTransactionID = rel.Data.ID
	}

	return row
}
