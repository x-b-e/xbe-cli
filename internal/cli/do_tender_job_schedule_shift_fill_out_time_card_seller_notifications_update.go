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

type doTenderJobScheduleShiftFillOutTimeCardSellerNotificationsUpdateOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	ID      string
	Read    bool
}

func newDoTenderJobScheduleShiftFillOutTimeCardSellerNotificationsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a tender job schedule shift fill out time card seller notification",
		Long: `Update an existing tender job schedule shift fill out time card seller notification.

Only the read status can be updated.

Optional flags:
  --read  Mark the notification as read (true/false)`,
		Example: `  # Mark as read
  xbe do tender-job-schedule-shift-fill-out-time-card-seller-notifications update 123 --read

  # Mark as unread
  xbe do tender-job-schedule-shift-fill-out-time-card-seller-notifications update 123 --read=false`,
		Args: cobra.ExactArgs(1),
		RunE: runDoTenderJobScheduleShiftFillOutTimeCardSellerNotificationsUpdate,
	}
	initDoTenderJobScheduleShiftFillOutTimeCardSellerNotificationsUpdateFlags(cmd)
	return cmd
}

func init() {
	doTenderJobScheduleShiftFillOutTimeCardSellerNotificationsCmd.AddCommand(newDoTenderJobScheduleShiftFillOutTimeCardSellerNotificationsUpdateCmd())
}

func initDoTenderJobScheduleShiftFillOutTimeCardSellerNotificationsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("read", false, "Mark the notification as read (true/false)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoTenderJobScheduleShiftFillOutTimeCardSellerNotificationsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoTenderJobScheduleShiftFillOutTimeCardSellerNotificationsUpdateOptions(cmd, args)
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

	attributes := map[string]any{}
	if cmd.Flags().Changed("read") {
		attributes["read"] = opts.Read
	}

	if len(attributes) == 0 {
		err := fmt.Errorf("no fields to update; specify at least one field flag")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	data := map[string]any{
		"type": "tender-job-schedule-shift-fill-out-time-card-seller-notifications",
		"id":   opts.ID,
	}
	data["attributes"] = attributes

	requestBody := map[string]any{
		"data": data,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Patch(cmd.Context(), "/v1/tender-job-schedule-shift-fill-out-time-card-seller-notifications/"+opts.ID, jsonBody)
	if err != nil {
		if len(body) > 0 {
			fmt.Fprintln(cmd.ErrOrStderr(), string(body))
		}
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	var resp jsonAPISingleResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	row := tenderJobScheduleShiftFillOutTimeCardSellerNotificationRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated tender job schedule shift fill out time card seller notification %s\n", row.ID)
	return nil
}

func parseDoTenderJobScheduleShiftFillOutTimeCardSellerNotificationsUpdateOptions(cmd *cobra.Command, args []string) (doTenderJobScheduleShiftFillOutTimeCardSellerNotificationsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	read, _ := cmd.Flags().GetBool("read")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doTenderJobScheduleShiftFillOutTimeCardSellerNotificationsUpdateOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		ID:      args[0],
		Read:    read,
	}, nil
}
