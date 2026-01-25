package cli

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type doMaterialPurchaseOrderReleaseRedemptionsDeleteOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	ID      string
	Confirm bool
}

func newDoMaterialPurchaseOrderReleaseRedemptionsDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a material purchase order release redemption",
		Long: `Delete a material purchase order release redemption.

Requires the --confirm flag to prevent accidental deletion.`,
		Example: `  # Delete a release redemption
  xbe do material-purchase-order-release-redemptions delete 123 --confirm`,
		Args: cobra.ExactArgs(1),
		RunE: runDoMaterialPurchaseOrderReleaseRedemptionsDelete,
	}
	initDoMaterialPurchaseOrderReleaseRedemptionsDeleteFlags(cmd)
	return cmd
}

func init() {
	doMaterialPurchaseOrderReleaseRedemptionsCmd.AddCommand(newDoMaterialPurchaseOrderReleaseRedemptionsDeleteCmd())
}

func initDoMaterialPurchaseOrderReleaseRedemptionsDeleteFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("confirm", false, "Confirm deletion (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	cmd.MarkFlagRequired("confirm")
}

func runDoMaterialPurchaseOrderReleaseRedemptionsDelete(cmd *cobra.Command, args []string) error {
	opts, err := parseDoMaterialPurchaseOrderReleaseRedemptionsDeleteOptions(cmd, args)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if !opts.Confirm {
		err := errors.New("deletion requires --confirm flag")
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

	client := api.NewClient(opts.BaseURL, opts.Token)

	path := fmt.Sprintf("/v1/material-purchase-order-release-redemptions/%s", opts.ID)
	body, _, err := client.Delete(cmd.Context(), path)
	if err != nil {
		if len(body) > 0 {
			fmt.Fprintln(cmd.ErrOrStderr(), string(body))
		}
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), map[string]any{
			"id":      opts.ID,
			"deleted": true,
		})
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Deleted material purchase order release redemption %s\n", opts.ID)
	return nil
}

func parseDoMaterialPurchaseOrderReleaseRedemptionsDeleteOptions(cmd *cobra.Command, args []string) (doMaterialPurchaseOrderReleaseRedemptionsDeleteOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	confirm, _ := cmd.Flags().GetBool("confirm")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doMaterialPurchaseOrderReleaseRedemptionsDeleteOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		ID:      args[0],
		Confirm: confirm,
	}, nil
}
