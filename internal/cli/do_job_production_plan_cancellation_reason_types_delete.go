package cli

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type doJobProductionPlanCancellationReasonTypesDeleteOptions struct {
	BaseURL string
	Token   string
	ID      string
	Confirm bool
}

func newDoJobProductionPlanCancellationReasonTypesDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a cancellation reason type",
		Long: `Delete a job production plan cancellation reason type.

Provide the type ID as an argument. The --confirm flag is required
to prevent accidental deletions.

Note: Only admin users can delete cancellation reason types.`,
		Example: `  # Delete a cancellation reason type
  xbe do job-production-plan-cancellation-reason-types delete 123 --confirm`,
		Args: cobra.ExactArgs(1),
		RunE: runDoJobProductionPlanCancellationReasonTypesDelete,
	}
	initDoJobProductionPlanCancellationReasonTypesDeleteFlags(cmd)
	return cmd
}

func init() {
	doJobProductionPlanCancellationReasonTypesCmd.AddCommand(newDoJobProductionPlanCancellationReasonTypesDeleteCmd())
}

func initDoJobProductionPlanCancellationReasonTypesDeleteFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("confirm", false, "Confirm deletion (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoJobProductionPlanCancellationReasonTypesDelete(cmd *cobra.Command, args []string) error {
	opts, err := parseDoJobProductionPlanCancellationReasonTypesDeleteOptions(cmd, args)
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
		err := fmt.Errorf("--confirm flag is required to delete a cancellation reason type")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Delete(cmd.Context(), "/v1/job-production-plan-cancellation-reason-types/"+opts.ID)
	if err != nil {
		if len(body) > 0 {
			fmt.Fprintln(cmd.ErrOrStderr(), string(body))
		}
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Deleted cancellation reason type %s\n", opts.ID)
	return nil
}

func parseDoJobProductionPlanCancellationReasonTypesDeleteOptions(cmd *cobra.Command, args []string) (doJobProductionPlanCancellationReasonTypesDeleteOptions, error) {
	confirm, _ := cmd.Flags().GetBool("confirm")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doJobProductionPlanCancellationReasonTypesDeleteOptions{
		BaseURL: baseURL,
		Token:   token,
		ID:      args[0],
		Confirm: confirm,
	}, nil
}
