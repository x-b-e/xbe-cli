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

type doMaterialTransactionRejectionsCreateOptions struct {
	BaseURL             string
	Token               string
	JSON                bool
	MaterialTransaction string
	Comment             string
}

func newDoMaterialTransactionRejectionsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Reject a material transaction",
		Long: `Reject a material transaction.

Material transactions must be in one of these statuses before rejection:
  invalidated, denied, submitted, accepted, editing, unmatched

Required flags:
  --material-transaction   Material transaction ID

Optional flags:
  --comment                Rejection comment

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Reject a material transaction with a comment
  xbe do material-transaction-rejections create \
    --material-transaction 123 \
    --comment "Missing ticket details"`,
		Args: cobra.NoArgs,
		RunE: runDoMaterialTransactionRejectionsCreate,
	}
	initDoMaterialTransactionRejectionsCreateFlags(cmd)
	return cmd
}

func init() {
	doMaterialTransactionRejectionsCmd.AddCommand(newDoMaterialTransactionRejectionsCreateCmd())
}

func initDoMaterialTransactionRejectionsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("material-transaction", "", "Material transaction ID")
	cmd.Flags().String("comment", "", "Rejection comment")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoMaterialTransactionRejectionsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoMaterialTransactionRejectionsCreateOptions(cmd)
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
	if opts.Comment != "" {
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
			"type":          "material-transaction-rejections",
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

	body, _, err := client.Post(cmd.Context(), "/v1/material-transaction-rejections", jsonBody)
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

	row := buildMaterialTransactionRejectionRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created material transaction rejection %s\n", row.ID)
	return nil
}

func parseDoMaterialTransactionRejectionsCreateOptions(cmd *cobra.Command) (doMaterialTransactionRejectionsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	materialTransaction, _ := cmd.Flags().GetString("material-transaction")
	comment, _ := cmd.Flags().GetString("comment")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doMaterialTransactionRejectionsCreateOptions{
		BaseURL:             baseURL,
		Token:               token,
		JSON:                jsonOut,
		MaterialTransaction: materialTransaction,
		Comment:             comment,
	}, nil
}
