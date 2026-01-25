package cli

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type doJobProductionPlanProjectPhaseRevenueItemsDeleteOptions struct {
	BaseURL string
	Token   string
	ID      string
	Confirm bool
}

func newDoJobProductionPlanProjectPhaseRevenueItemsDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a job production plan project phase revenue item",
		Long: `Delete a job production plan project phase revenue item.

Requires --confirm flag to prevent accidental deletion.`,
		Example: `  # Delete an item
  xbe do job-production-plan-project-phase-revenue-items delete 123 --confirm`,
		Args: cobra.ExactArgs(1),
		RunE: runDoJobProductionPlanProjectPhaseRevenueItemsDelete,
	}
	initDoJobProductionPlanProjectPhaseRevenueItemsDeleteFlags(cmd)
	return cmd
}

func init() {
	doJobProductionPlanProjectPhaseRevenueItemsCmd.AddCommand(newDoJobProductionPlanProjectPhaseRevenueItemsDeleteCmd())
}

func initDoJobProductionPlanProjectPhaseRevenueItemsDeleteFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("confirm", false, "Confirm deletion")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	_ = cmd.MarkFlagRequired("confirm")
}

func runDoJobProductionPlanProjectPhaseRevenueItemsDelete(cmd *cobra.Command, args []string) error {
	opts, err := parseDoJobProductionPlanProjectPhaseRevenueItemsDeleteOptions(cmd, args)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if !opts.Confirm {
		err := fmt.Errorf("deletion requires --confirm flag")
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

	body, _, err := client.Delete(cmd.Context(), "/v1/job-production-plan-project-phase-revenue-items/"+opts.ID)
	if err != nil {
		if len(body) > 0 {
			fmt.Fprintln(cmd.ErrOrStderr(), string(body))
		}
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Deleted job production plan project phase revenue item %s\n", opts.ID)
	return nil
}

func parseDoJobProductionPlanProjectPhaseRevenueItemsDeleteOptions(cmd *cobra.Command, args []string) (doJobProductionPlanProjectPhaseRevenueItemsDeleteOptions, error) {
	confirm, _ := cmd.Flags().GetBool("confirm")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doJobProductionPlanProjectPhaseRevenueItemsDeleteOptions{
		BaseURL: baseURL,
		Token:   token,
		ID:      args[0],
		Confirm: confirm,
	}, nil
}
