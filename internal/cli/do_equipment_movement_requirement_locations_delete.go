package cli

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type doEquipmentMovementRequirementLocationsDeleteOptions struct {
	BaseURL string
	Token   string
	ID      string
	Confirm bool
}

func newDoEquipmentMovementRequirementLocationsDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete an equipment movement requirement location",
		Long: `Delete an equipment movement requirement location.

Required flags:
  --confirm    Confirm deletion`,
		Example: `  # Delete a location
  xbe do equipment-movement-requirement-locations delete 123 --confirm`,
		Args: cobra.ExactArgs(1),
		RunE: runDoEquipmentMovementRequirementLocationsDelete,
	}
	initDoEquipmentMovementRequirementLocationsDeleteFlags(cmd)
	return cmd
}

func init() {
	doEquipmentMovementRequirementLocationsCmd.AddCommand(newDoEquipmentMovementRequirementLocationsDeleteCmd())
}

func initDoEquipmentMovementRequirementLocationsDeleteFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("confirm", false, "Confirm deletion")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoEquipmentMovementRequirementLocationsDelete(cmd *cobra.Command, args []string) error {
	opts, err := parseDoEquipmentMovementRequirementLocationsDeleteOptions(cmd, args)
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

	body, _, err := client.Delete(cmd.Context(), "/v1/equipment-movement-requirement-locations/"+opts.ID)
	if err != nil {
		if len(body) > 0 {
			fmt.Fprintln(cmd.ErrOrStderr(), string(body))
		}
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Deleted equipment movement requirement location %s\n", opts.ID)
	return nil
}

func parseDoEquipmentMovementRequirementLocationsDeleteOptions(cmd *cobra.Command, args []string) (doEquipmentMovementRequirementLocationsDeleteOptions, error) {
	confirm, _ := cmd.Flags().GetBool("confirm")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doEquipmentMovementRequirementLocationsDeleteOptions{
		BaseURL: baseURL,
		Token:   token,
		ID:      args[0],
		Confirm: confirm,
	}, nil
}
