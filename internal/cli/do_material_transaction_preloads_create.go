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

type doMaterialTransactionPreloadsCreateOptions struct {
	BaseURL             string
	Token               string
	JSON                bool
	MaterialTransaction string
	Trailer             string
}

func newDoMaterialTransactionPreloadsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a material transaction preload",
		Long: `Create a material transaction preload.

Required flags:
  --material-transaction   Material transaction ID
  --trailer                Trailer ID

The preload timestamp is derived from the material transaction's transaction time.`,
		Example: `  # Create a preload for a trailer and material transaction
  xbe do material-transaction-preloads create \
    --material-transaction 123 \
    --trailer 456`,
		Args: cobra.NoArgs,
		RunE: runDoMaterialTransactionPreloadsCreate,
	}
	initDoMaterialTransactionPreloadsCreateFlags(cmd)
	return cmd
}

func init() {
	doMaterialTransactionPreloadsCmd.AddCommand(newDoMaterialTransactionPreloadsCreateCmd())
}

func initDoMaterialTransactionPreloadsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("material-transaction", "", "Material transaction ID (required)")
	cmd.Flags().String("trailer", "", "Trailer ID (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoMaterialTransactionPreloadsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoMaterialTransactionPreloadsCreateOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if strings.TrimSpace(opts.Token) == "" {
		if token, _, err := auth.ResolveToken(opts.BaseURL, ""); err == nil {
			opts.Token = token
		} else if errors.Is(err, auth.ErrNotFound) {
			fmt.Fprintln(cmd.ErrOrStderr(), "Authentication required. Run xbe auth login first.")
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

	trailerID := strings.TrimSpace(opts.Trailer)
	if trailerID == "" {
		err := fmt.Errorf("--trailer is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	relationships := map[string]any{
		"material-transaction": map[string]any{
			"data": map[string]any{
				"type": "material-transactions",
				"id":   materialTransactionID,
			},
		},
		"trailer": map[string]any{
			"data": map[string]any{
				"type": "trailers",
				"id":   trailerID,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "material-transaction-preloads",
			"relationships": relationships,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/material-transaction-preloads", jsonBody)
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

	details := buildMaterialTransactionPreloadDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created material transaction preload %s\n", details.ID)
	return nil
}

func parseDoMaterialTransactionPreloadsCreateOptions(cmd *cobra.Command) (doMaterialTransactionPreloadsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	materialTransaction, _ := cmd.Flags().GetString("material-transaction")
	trailer, _ := cmd.Flags().GetString("trailer")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doMaterialTransactionPreloadsCreateOptions{
		BaseURL:             baseURL,
		Token:               token,
		JSON:                jsonOut,
		MaterialTransaction: materialTransaction,
		Trailer:             trailer,
	}, nil
}
