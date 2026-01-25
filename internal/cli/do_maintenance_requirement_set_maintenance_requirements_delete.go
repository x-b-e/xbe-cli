package cli

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type doMaintenanceRequirementSetMaintenanceRequirementsDeleteOptions struct {
	BaseURL string
	Token   string
	ID      string
	Confirm bool
	JSON    bool
}

func newDoMaintenanceRequirementSetMaintenanceRequirementsDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a maintenance requirement set maintenance requirement",
		Long: `Delete a maintenance requirement set maintenance requirement.

Required flags:
  --confirm    Confirm deletion`,
		Example: `  # Delete a record
  xbe do maintenance-requirement-set-maintenance-requirements delete 123 --confirm`,
		Args: cobra.ExactArgs(1),
		RunE: runDoMaintenanceRequirementSetMaintenanceRequirementsDelete,
	}
	initDoMaintenanceRequirementSetMaintenanceRequirementsDeleteFlags(cmd)
	return cmd
}

func init() {
	doMaintenanceRequirementSetMaintenanceRequirementsCmd.AddCommand(newDoMaintenanceRequirementSetMaintenanceRequirementsDeleteCmd())
}

func initDoMaintenanceRequirementSetMaintenanceRequirementsDeleteFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("confirm", false, "Confirm deletion")
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoMaintenanceRequirementSetMaintenanceRequirementsDelete(cmd *cobra.Command, args []string) error {
	opts, err := parseDoMaintenanceRequirementSetMaintenanceRequirementsDeleteOptions(cmd, args)
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

	body, _, err := client.Delete(cmd.Context(), "/v1/maintenance-requirement-set-maintenance-requirements/"+opts.ID)
	if err != nil {
		if len(body) > 0 {
			fmt.Fprintln(cmd.ErrOrStderr(), string(body))
		}
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), map[string]any{"deleted": true, "id": opts.ID})
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Deleted maintenance requirement set maintenance requirement %s\n", opts.ID)
	return nil
}

func parseDoMaintenanceRequirementSetMaintenanceRequirementsDeleteOptions(cmd *cobra.Command, args []string) (doMaintenanceRequirementSetMaintenanceRequirementsDeleteOptions, error) {
	confirm, _ := cmd.Flags().GetBool("confirm")
	jsonOut, _ := cmd.Flags().GetBool("json")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doMaintenanceRequirementSetMaintenanceRequirementsDeleteOptions{
		BaseURL: baseURL,
		Token:   token,
		ID:      args[0],
		Confirm: confirm,
		JSON:    jsonOut,
	}, nil
}
