package cli

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type doJobProductionPlanMaterialTypeQualityControlRequirementsDeleteOptions struct {
	BaseURL string
	Token   string
	ID      string
	Confirm bool
}

func newDoJobProductionPlanMaterialTypeQualityControlRequirementsDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a job production plan material type quality control requirement",
		Long: `Delete a job production plan material type quality control requirement.

This action requires confirmation.

Arguments:
  <id>    The requirement ID (required).`,
		Example: `  # Delete a requirement
  xbe do job-production-plan-material-type-quality-control-requirements delete 123 --confirm`,
		Args: cobra.ExactArgs(1),
		RunE: runDoJobProductionPlanMaterialTypeQualityControlRequirementsDelete,
	}
	initDoJobProductionPlanMaterialTypeQualityControlRequirementsDeleteFlags(cmd)
	return cmd
}

func init() {
	doJobProductionPlanMaterialTypeQualityControlRequirementsCmd.AddCommand(newDoJobProductionPlanMaterialTypeQualityControlRequirementsDeleteCmd())
}

func initDoJobProductionPlanMaterialTypeQualityControlRequirementsDeleteFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("confirm", false, "Confirm deletion")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoJobProductionPlanMaterialTypeQualityControlRequirementsDelete(cmd *cobra.Command, args []string) error {
	opts, err := parseDoJobProductionPlanMaterialTypeQualityControlRequirementsDeleteOptions(cmd, args)
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

	body, _, err := client.Delete(cmd.Context(), "/v1/job-production-plan-material-type-quality-control-requirements/"+opts.ID)
	if err != nil {
		if len(body) > 0 {
			fmt.Fprintln(cmd.ErrOrStderr(), string(body))
		}
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Deleted job production plan material type quality control requirement %s\n", opts.ID)
	return nil
}

func parseDoJobProductionPlanMaterialTypeQualityControlRequirementsDeleteOptions(cmd *cobra.Command, args []string) (doJobProductionPlanMaterialTypeQualityControlRequirementsDeleteOptions, error) {
	confirm, _ := cmd.Flags().GetBool("confirm")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doJobProductionPlanMaterialTypeQualityControlRequirementsDeleteOptions{
		BaseURL: baseURL,
		Token:   token,
		ID:      args[0],
		Confirm: confirm,
	}, nil
}
