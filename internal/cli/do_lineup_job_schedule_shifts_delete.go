package cli

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type doLineupJobScheduleShiftsDeleteOptions struct {
	BaseURL string
	Token   string
	ID      string
	Confirm bool
	JSON    bool
}

func newDoLineupJobScheduleShiftsDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a lineup job schedule shift",
		Long: `Delete a lineup job schedule shift.

Required flags:
  --confirm    Confirm deletion`,
		Example: `  # Delete a lineup job schedule shift
  xbe do lineup-job-schedule-shifts delete 123 --confirm`,
		Args: cobra.ExactArgs(1),
		RunE: runDoLineupJobScheduleShiftsDelete,
	}
	initDoLineupJobScheduleShiftsDeleteFlags(cmd)
	return cmd
}

func init() {
	doLineupJobScheduleShiftsCmd.AddCommand(newDoLineupJobScheduleShiftsDeleteCmd())
}

func initDoLineupJobScheduleShiftsDeleteFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("confirm", false, "Confirm deletion")
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoLineupJobScheduleShiftsDelete(cmd *cobra.Command, args []string) error {
	opts, err := parseDoLineupJobScheduleShiftsDeleteOptions(cmd, args)
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

	body, _, err := client.Delete(cmd.Context(), "/v1/lineup-job-schedule-shifts/"+opts.ID)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Deleted lineup job schedule shift %s\n", opts.ID)
	return nil
}

func parseDoLineupJobScheduleShiftsDeleteOptions(cmd *cobra.Command, args []string) (doLineupJobScheduleShiftsDeleteOptions, error) {
	confirm, _ := cmd.Flags().GetBool("confirm")
	jsonOut, _ := cmd.Flags().GetBool("json")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doLineupJobScheduleShiftsDeleteOptions{
		BaseURL: baseURL,
		Token:   token,
		ID:      args[0],
		Confirm: confirm,
		JSON:    jsonOut,
	}, nil
}
