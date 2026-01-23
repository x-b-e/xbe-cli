package cli

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type doMaterialTransactionCostCodeAllocationsDeleteOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	ID      string
	Confirm bool
}

func newDoMaterialTransactionCostCodeAllocationsDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a material transaction cost code allocation",
		Long: `Delete a material transaction cost code allocation.

Required flags:
  --confirm    Confirm deletion`,
		Example: `  # Delete a material transaction cost code allocation
  xbe do material-transaction-cost-code-allocations delete 123 --confirm

  # Get JSON output
  xbe do material-transaction-cost-code-allocations delete 123 --confirm --json`,
		Args: cobra.ExactArgs(1),
		RunE: runDoMaterialTransactionCostCodeAllocationsDelete,
	}
	initDoMaterialTransactionCostCodeAllocationsDeleteFlags(cmd)
	return cmd
}

func init() {
	doMaterialTransactionCostCodeAllocationsCmd.AddCommand(newDoMaterialTransactionCostCodeAllocationsDeleteCmd())
}

func initDoMaterialTransactionCostCodeAllocationsDeleteFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("confirm", false, "Confirm deletion")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoMaterialTransactionCostCodeAllocationsDelete(cmd *cobra.Command, args []string) error {
	opts, err := parseDoMaterialTransactionCostCodeAllocationsDeleteOptions(cmd, args)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if !opts.Confirm {
		err := fmt.Errorf("--confirm flag is required to delete")
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

	body, _, err := client.Delete(cmd.Context(), "/v1/material-transaction-cost-code-allocations/"+opts.ID)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Deleted material transaction cost code allocation %s\n", opts.ID)
	return nil
}

func parseDoMaterialTransactionCostCodeAllocationsDeleteOptions(cmd *cobra.Command, args []string) (doMaterialTransactionCostCodeAllocationsDeleteOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	confirm, _ := cmd.Flags().GetBool("confirm")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doMaterialTransactionCostCodeAllocationsDeleteOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		ID:      args[0],
		Confirm: confirm,
	}, nil
}
