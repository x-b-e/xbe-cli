package cli

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type doEquipmentUtilizationReadingsDeleteOptions struct {
	BaseURL string
	Token   string
	ID      string
	Confirm bool
}

func newDoEquipmentUtilizationReadingsDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete an equipment utilization reading",
		Long: `Delete an equipment utilization reading.

Required flags:
  --confirm    Confirm deletion`,
		Example: `  # Delete a reading
  xbe do equipment-utilization-readings delete 123 --confirm`,
		Args: cobra.ExactArgs(1),
		RunE: runDoEquipmentUtilizationReadingsDelete,
	}
	initDoEquipmentUtilizationReadingsDeleteFlags(cmd)
	return cmd
}

func init() {
	doEquipmentUtilizationReadingsCmd.AddCommand(newDoEquipmentUtilizationReadingsDeleteCmd())
}

func initDoEquipmentUtilizationReadingsDeleteFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("confirm", false, "Confirm deletion")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoEquipmentUtilizationReadingsDelete(cmd *cobra.Command, args []string) error {
	opts, err := parseDoEquipmentUtilizationReadingsDeleteOptions(cmd, args)
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

	body, _, err := client.Delete(cmd.Context(), "/v1/equipment-utilization-readings/"+opts.ID)
	if err != nil {
		if len(body) > 0 {
			fmt.Fprintln(cmd.ErrOrStderr(), string(body))
		}
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Deleted equipment utilization reading %s\n", opts.ID)
	return nil
}

func parseDoEquipmentUtilizationReadingsDeleteOptions(cmd *cobra.Command, args []string) (doEquipmentUtilizationReadingsDeleteOptions, error) {
	confirm, _ := cmd.Flags().GetBool("confirm")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doEquipmentUtilizationReadingsDeleteOptions{
		BaseURL: baseURL,
		Token:   token,
		ID:      args[0],
		Confirm: confirm,
	}, nil
}
