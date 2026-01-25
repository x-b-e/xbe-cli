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

type doTenderJobScheduleShiftCancelledTruckerContactNotificationsUpdateOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	ID      string
	Read    bool
}

func newDoTenderJobScheduleShiftCancelledTruckerContactNotificationsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a tender job schedule shift cancelled trucker contact notification",
		Long: `Update a tender job schedule shift cancelled trucker contact notification.

Writable fields:
  --read    Mark the notification as read/unread (true/false)

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Mark a notification as read
  xbe do tender-job-schedule-shift-cancelled-trucker-contact-notifications update 123 --read

  # Mark a notification as unread
  xbe do tender-job-schedule-shift-cancelled-trucker-contact-notifications update 123 --read=false`,
		Args: cobra.ExactArgs(1),
		RunE: runDoTenderJobScheduleShiftCancelledTruckerContactNotificationsUpdate,
	}
	initDoTenderJobScheduleShiftCancelledTruckerContactNotificationsUpdateFlags(cmd)
	return cmd
}

func init() {
	doTenderJobScheduleShiftCancelledTruckerContactNotificationsCmd.AddCommand(newDoTenderJobScheduleShiftCancelledTruckerContactNotificationsUpdateCmd())
}

func initDoTenderJobScheduleShiftCancelledTruckerContactNotificationsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("read", false, "Mark notification as read (true/false)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoTenderJobScheduleShiftCancelledTruckerContactNotificationsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoTenderJobScheduleShiftCancelledTruckerContactNotificationsUpdateOptions(cmd, args)
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
		err := errors.New("no fields to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "tender-job-schedule-shift-cancelled-trucker-contact-notifications",
			"id":         opts.ID,
			"attributes": attributes,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Patch(cmd.Context(), "/v1/tender-job-schedule-shift-cancelled-trucker-contact-notifications/"+opts.ID, jsonBody)
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

	details := buildTenderJobScheduleShiftCancelledTruckerContactNotificationDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated tender job schedule shift cancelled trucker contact notification %s\n", details.ID)
	return nil
}

func parseDoTenderJobScheduleShiftCancelledTruckerContactNotificationsUpdateOptions(cmd *cobra.Command, args []string) (doTenderJobScheduleShiftCancelledTruckerContactNotificationsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	read, _ := cmd.Flags().GetBool("read")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doTenderJobScheduleShiftCancelledTruckerContactNotificationsUpdateOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		ID:      args[0],
		Read:    read,
	}, nil
}
