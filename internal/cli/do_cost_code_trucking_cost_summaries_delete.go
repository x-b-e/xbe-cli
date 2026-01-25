package cli

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type doCostCodeTruckingCostSummariesDeleteOptions struct {
	BaseURL string
	Token   string
	ID      string
	Confirm bool
}

func newDoCostCodeTruckingCostSummariesDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a cost code trucking cost summary",
		Long: `Delete a cost code trucking cost summary.

This permanently removes a summary and its computed results.

Arguments:
  <id>  Summary ID (required). Find IDs using the list command.`,
		Example: `  # Delete a summary
  xbe do cost-code-trucking-cost-summaries delete 123 --confirm`,
		Args: cobra.ExactArgs(1),
		RunE: runDoCostCodeTruckingCostSummariesDelete,
	}
	initDoCostCodeTruckingCostSummariesDeleteFlags(cmd)
	return cmd
}

func init() {
	doCostCodeTruckingCostSummariesCmd.AddCommand(newDoCostCodeTruckingCostSummariesDeleteCmd())
}

func initDoCostCodeTruckingCostSummariesDeleteFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("confirm", false, "Confirm deletion")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoCostCodeTruckingCostSummariesDelete(cmd *cobra.Command, args []string) error {
	opts, err := parseDoCostCodeTruckingCostSummariesDeleteOptions(cmd, args)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if !opts.Confirm {
		err := errors.New("deletion requires --confirm")
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

	body, _, err := client.Delete(cmd.Context(), "/v1/cost-code-trucking-cost-summaries/"+opts.ID)
	if err != nil {
		if len(body) > 0 {
			fmt.Fprintln(cmd.ErrOrStderr(), string(body))
		}
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Deleted cost code trucking cost summary %s\n", opts.ID)
	return nil
}

func parseDoCostCodeTruckingCostSummariesDeleteOptions(cmd *cobra.Command, args []string) (doCostCodeTruckingCostSummariesDeleteOptions, error) {
	confirm, _ := cmd.Flags().GetBool("confirm")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doCostCodeTruckingCostSummariesDeleteOptions{
		BaseURL: baseURL,
		Token:   token,
		ID:      args[0],
		Confirm: confirm,
	}, nil
}
