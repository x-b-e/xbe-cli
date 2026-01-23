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

type doJobProductionPlanServiceTypeUnitOfMeasuresDeleteOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	ID      string
	Confirm bool
}

func newDoJobProductionPlanServiceTypeUnitOfMeasuresDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Remove a service type unit of measure from a job production plan",
		Long: `Remove a service type unit of measure from a job production plan.

This action cannot be undone.

Arguments:
  <id>    The job production plan service type unit of measure ID (required)

Flags:
  --confirm    Confirm deletion

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Delete a job production plan service type unit of measure
  xbe do job-production-plan-service-type-unit-of-measures delete 123 --confirm

  # Output as JSON
  xbe do job-production-plan-service-type-unit-of-measures delete 123 --confirm --json`,
		Args: cobra.ExactArgs(1),
		RunE: runDoJobProductionPlanServiceTypeUnitOfMeasuresDelete,
	}
	initDoJobProductionPlanServiceTypeUnitOfMeasuresDeleteFlags(cmd)
	return cmd
}

func init() {
	doJobProductionPlanServiceTypeUnitOfMeasuresCmd.AddCommand(newDoJobProductionPlanServiceTypeUnitOfMeasuresDeleteCmd())
}

func initDoJobProductionPlanServiceTypeUnitOfMeasuresDeleteFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("confirm", false, "Confirm deletion")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoJobProductionPlanServiceTypeUnitOfMeasuresDelete(cmd *cobra.Command, args []string) error {
	opts, err := parseDoJobProductionPlanServiceTypeUnitOfMeasuresDeleteOptions(cmd, args)
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
		err := fmt.Errorf("--confirm is required to delete a job production plan service type unit of measure")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Delete(cmd.Context(), "/v1/job-production-plan-service-type-unit-of-measures/"+opts.ID)
	if err != nil {
		if len(body) > 0 {
			fmt.Fprintln(cmd.ErrOrStderr(), string(body))
		}
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if opts.JSON {
		resp := map[string]any{
			"deleted": true,
			"id":      opts.ID,
		}
		jsonBytes, err := json.Marshal(resp)
		if err != nil {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		fmt.Fprintln(cmd.OutOrStdout(), string(jsonBytes))
		return nil
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Deleted job production plan service type unit of measure %s\n", opts.ID)
	return nil
}

func parseDoJobProductionPlanServiceTypeUnitOfMeasuresDeleteOptions(cmd *cobra.Command, args []string) (doJobProductionPlanServiceTypeUnitOfMeasuresDeleteOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	confirm, _ := cmd.Flags().GetBool("confirm")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doJobProductionPlanServiceTypeUnitOfMeasuresDeleteOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		ID:      args[0],
		Confirm: confirm,
	}, nil
}
