package cli

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type doEquipmentMovementTripJobProductionPlansDeleteOptions struct {
	BaseURL string
	Token   string
	ID      string
	Confirm bool
}

func newDoEquipmentMovementTripJobProductionPlansDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete an equipment movement trip job production plan link",
		Long: `Delete an equipment movement trip job production plan link.

Provide the link ID as an argument. The --confirm flag is required
to prevent accidental deletions.

Global flags (see xbe --help): --base-url, --token`,
		Example: `  # Delete a link
  xbe do equipment-movement-trip-job-production-plans delete 123 --confirm`,
		Args: cobra.ExactArgs(1),
		RunE: runDoEquipmentMovementTripJobProductionPlansDelete,
	}
	initDoEquipmentMovementTripJobProductionPlansDeleteFlags(cmd)
	return cmd
}

func init() {
	doEquipmentMovementTripJobProductionPlansCmd.AddCommand(newDoEquipmentMovementTripJobProductionPlansDeleteCmd())
}

func initDoEquipmentMovementTripJobProductionPlansDeleteFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("confirm", false, "Confirm deletion (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoEquipmentMovementTripJobProductionPlansDelete(cmd *cobra.Command, args []string) error {
	opts, err := parseDoEquipmentMovementTripJobProductionPlansDeleteOptions(cmd, args)
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
		err := fmt.Errorf("--confirm flag is required to delete an equipment movement trip job production plan")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Delete(cmd.Context(), "/v1/equipment-movement-trip-job-production-plans/"+opts.ID)
	if err != nil {
		if len(body) > 0 {
			fmt.Fprintln(cmd.ErrOrStderr(), string(body))
		}
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Deleted equipment movement trip job production plan %s\n", opts.ID)
	return nil
}

func parseDoEquipmentMovementTripJobProductionPlansDeleteOptions(cmd *cobra.Command, args []string) (doEquipmentMovementTripJobProductionPlansDeleteOptions, error) {
	confirm, _ := cmd.Flags().GetBool("confirm")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doEquipmentMovementTripJobProductionPlansDeleteOptions{
		BaseURL: baseURL,
		Token:   token,
		ID:      args[0],
		Confirm: confirm,
	}, nil
}
