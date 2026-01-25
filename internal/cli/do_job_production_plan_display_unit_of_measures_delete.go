package cli

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type doJobProductionPlanDisplayUnitOfMeasuresDeleteOptions struct {
	BaseURL string
	Token   string
	ID      string
	Confirm bool
}

func newDoJobProductionPlanDisplayUnitOfMeasuresDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a job production plan display unit of measure",
		Long: `Delete a job production plan display unit of measure.

Provide the display unit of measure ID as an argument. The --confirm flag is required
to prevent accidental deletions.`,
		Example: `  # Delete a display unit of measure
  xbe do job-production-plan-display-unit-of-measures delete 123 --confirm`,
		Args: cobra.ExactArgs(1),
		RunE: runDoJobProductionPlanDisplayUnitOfMeasuresDelete,
	}
	initDoJobProductionPlanDisplayUnitOfMeasuresDeleteFlags(cmd)
	return cmd
}

func init() {
	doJobProductionPlanDisplayUnitOfMeasuresCmd.AddCommand(newDoJobProductionPlanDisplayUnitOfMeasuresDeleteCmd())
}

func initDoJobProductionPlanDisplayUnitOfMeasuresDeleteFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("confirm", false, "Confirm deletion (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoJobProductionPlanDisplayUnitOfMeasuresDelete(cmd *cobra.Command, args []string) error {
	opts, err := parseDoJobProductionPlanDisplayUnitOfMeasuresDeleteOptions(cmd, args)
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
		err := fmt.Errorf("--confirm flag is required to delete a job production plan display unit of measure")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Delete(cmd.Context(), "/v1/job-production-plan-display-unit-of-measures/"+opts.ID)
	if err != nil {
		if len(body) > 0 {
			fmt.Fprintln(cmd.ErrOrStderr(), string(body))
		}
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Deleted job production plan display unit of measure %s\n", opts.ID)
	return nil
}

func parseDoJobProductionPlanDisplayUnitOfMeasuresDeleteOptions(cmd *cobra.Command, args []string) (doJobProductionPlanDisplayUnitOfMeasuresDeleteOptions, error) {
	confirm, _ := cmd.Flags().GetBool("confirm")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doJobProductionPlanDisplayUnitOfMeasuresDeleteOptions{
		BaseURL: baseURL,
		Token:   token,
		ID:      args[0],
		Confirm: confirm,
	}, nil
}
