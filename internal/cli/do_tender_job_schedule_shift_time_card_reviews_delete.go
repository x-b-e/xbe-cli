package cli

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type doTenderJobScheduleShiftTimeCardReviewsDeleteOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	ID      string
	Confirm bool
}

func newDoTenderJobScheduleShiftTimeCardReviewsDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a tender job schedule shift time card review",
		Long: `Delete a tender job schedule shift time card review.

Requires the --confirm flag to prevent accidental deletion.`,
		Example: `  # Delete a time card review
  xbe do tender-job-schedule-shift-time-card-reviews delete 123 --confirm`,
		Args: cobra.ExactArgs(1),
		RunE: runDoTenderJobScheduleShiftTimeCardReviewsDelete,
	}
	initDoTenderJobScheduleShiftTimeCardReviewsDeleteFlags(cmd)
	return cmd
}

func init() {
	doTenderJobScheduleShiftTimeCardReviewsCmd.AddCommand(newDoTenderJobScheduleShiftTimeCardReviewsDeleteCmd())
}

func initDoTenderJobScheduleShiftTimeCardReviewsDeleteFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("confirm", false, "Confirm deletion (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoTenderJobScheduleShiftTimeCardReviewsDelete(cmd *cobra.Command, args []string) error {
	opts, err := parseDoTenderJobScheduleShiftTimeCardReviewsDeleteOptions(cmd, args)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if !opts.Confirm {
		err := fmt.Errorf("--confirm flag is required to delete a tender job schedule shift time card review")
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

	body, _, err := client.Delete(cmd.Context(), "/v1/tender-job-schedule-shift-time-card-reviews/"+opts.ID)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Deleted tender job schedule shift time card review %s\n", opts.ID)
	return nil
}

func parseDoTenderJobScheduleShiftTimeCardReviewsDeleteOptions(cmd *cobra.Command, args []string) (doTenderJobScheduleShiftTimeCardReviewsDeleteOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	confirm, _ := cmd.Flags().GetBool("confirm")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doTenderJobScheduleShiftTimeCardReviewsDeleteOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		ID:      args[0],
		Confirm: confirm,
	}, nil
}
