package cli

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type doCustomerTendersDeleteOptions struct {
	BaseURL string
	Token   string
	ID      string
	Confirm bool
}

func newDoCustomerTendersDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a customer tender",
		Long: `Delete a customer tender.

Provide the customer tender ID as an argument. The --confirm flag is required
to prevent accidental deletions.

Global flags (see xbe --help): --base-url, --token`,
		Example: `  # Delete a customer tender
  xbe do customer-tenders delete 123 --confirm`,
		Args: cobra.ExactArgs(1),
		RunE: runDoCustomerTendersDelete,
	}
	initDoCustomerTendersDeleteFlags(cmd)
	return cmd
}

func init() {
	doCustomerTendersCmd.AddCommand(newDoCustomerTendersDeleteCmd())
}

func initDoCustomerTendersDeleteFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("confirm", false, "Confirm deletion (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoCustomerTendersDelete(cmd *cobra.Command, args []string) error {
	opts, err := parseDoCustomerTendersDeleteOptions(cmd, args)
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
		err := fmt.Errorf("--confirm flag is required to delete a customer tender")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Delete(cmd.Context(), "/v1/customer-tenders/"+opts.ID)
	if err != nil {
		if len(body) > 0 {
			fmt.Fprintln(cmd.ErrOrStderr(), string(body))
		}
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Deleted customer tender %s\n", opts.ID)
	return nil
}

func parseDoCustomerTendersDeleteOptions(cmd *cobra.Command, args []string) (doCustomerTendersDeleteOptions, error) {
	confirm, _ := cmd.Flags().GetBool("confirm")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doCustomerTendersDeleteOptions{
		BaseURL: baseURL,
		Token:   token,
		ID:      args[0],
		Confirm: confirm,
	}, nil
}
