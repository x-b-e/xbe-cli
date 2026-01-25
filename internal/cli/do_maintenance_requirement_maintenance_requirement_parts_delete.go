package cli

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type doMaintenanceRequirementMaintenanceRequirementPartsDeleteOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	ID      string
	Confirm bool
}

func newDoMaintenanceRequirementMaintenanceRequirementPartsDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a maintenance requirement part link",
		Long: `Delete a maintenance requirement part link.

Requires the --confirm flag to prevent accidental deletion.`,
		Example: `  # Delete a maintenance requirement part link
  xbe do maintenance-requirement-maintenance-requirement-parts delete 123 --confirm`,
		Args: cobra.ExactArgs(1),
		RunE: runDoMaintenanceRequirementMaintenanceRequirementPartsDelete,
	}
	initDoMaintenanceRequirementMaintenanceRequirementPartsDeleteFlags(cmd)
	return cmd
}

func init() {
	doMaintenanceRequirementMaintenanceRequirementPartsCmd.AddCommand(newDoMaintenanceRequirementMaintenanceRequirementPartsDeleteCmd())
}

func initDoMaintenanceRequirementMaintenanceRequirementPartsDeleteFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("confirm", false, "Confirm deletion (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoMaintenanceRequirementMaintenanceRequirementPartsDelete(cmd *cobra.Command, args []string) error {
	opts, err := parseDoMaintenanceRequirementMaintenanceRequirementPartsDeleteOptions(cmd, args)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if !opts.Confirm {
		err := fmt.Errorf("--confirm flag is required to delete a maintenance requirement part link")
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

	body, _, err := client.Delete(cmd.Context(), "/v1/maintenance-requirement-maintenance-requirement-parts/"+opts.ID)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Deleted maintenance requirement part link %s\n", opts.ID)
	return nil
}

func parseDoMaintenanceRequirementMaintenanceRequirementPartsDeleteOptions(cmd *cobra.Command, args []string) (doMaintenanceRequirementMaintenanceRequirementPartsDeleteOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	confirm, _ := cmd.Flags().GetBool("confirm")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doMaintenanceRequirementMaintenanceRequirementPartsDeleteOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		ID:      args[0],
		Confirm: confirm,
	}, nil
}
