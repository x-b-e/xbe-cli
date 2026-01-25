package cli

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type doRawMaterialTransactionSalesCustomersDeleteOptions struct {
	BaseURL string
	Token   string
	ID      string
	Confirm bool
}

func newDoRawMaterialTransactionSalesCustomersDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a raw material transaction sales customer",
		Long: `Delete a raw material transaction sales customer.

Provide the raw material transaction sales customer ID as an argument. The --confirm flag is required
to prevent accidental deletions.`,
		Example: `  # Delete a raw material transaction sales customer
  xbe do raw-material-transaction-sales-customers delete 123 --confirm`,
		Args: cobra.ExactArgs(1),
		RunE: runDoRawMaterialTransactionSalesCustomersDelete,
	}
	initDoRawMaterialTransactionSalesCustomersDeleteFlags(cmd)
	return cmd
}

func init() {
	doRawMaterialTransactionSalesCustomersCmd.AddCommand(newDoRawMaterialTransactionSalesCustomersDeleteCmd())
}

func initDoRawMaterialTransactionSalesCustomersDeleteFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("confirm", false, "Confirm deletion (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoRawMaterialTransactionSalesCustomersDelete(cmd *cobra.Command, args []string) error {
	opts, err := parseDoRawMaterialTransactionSalesCustomersDeleteOptions(cmd, args)
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

	if !opts.Confirm {
		err := fmt.Errorf("--confirm flag is required to delete a raw material transaction sales customer")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Delete(cmd.Context(), "/v1/raw-material-transaction-sales-customers/"+opts.ID)
	if err != nil {
		if len(body) > 0 {
			fmt.Fprintln(cmd.ErrOrStderr(), string(body))
		}
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Deleted raw material transaction sales customer %s\n", opts.ID)
	return nil
}

func parseDoRawMaterialTransactionSalesCustomersDeleteOptions(cmd *cobra.Command, args []string) (doRawMaterialTransactionSalesCustomersDeleteOptions, error) {
	confirm, _ := cmd.Flags().GetBool("confirm")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doRawMaterialTransactionSalesCustomersDeleteOptions{
		BaseURL: baseURL,
		Token:   token,
		ID:      args[0],
		Confirm: confirm,
	}, nil
}
