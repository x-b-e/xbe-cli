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

type doNotificationsUpdateOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	ID      string
	Read    bool
}

func newDoNotificationsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a notification",
		Long: `Update an existing notification.

Only the read status can be updated.

Optional flags:
  --read  Mark the notification as read (true/false)`,
		Example: `  # Mark as read
  xbe do notifications update 123 --read

  # Mark as unread
  xbe do notifications update 123 --read=false`,
		Args: cobra.ExactArgs(1),
		RunE: runDoNotificationsUpdate,
	}
	initDoNotificationsUpdateFlags(cmd)
	return cmd
}

func init() {
	doNotificationsCmd.AddCommand(newDoNotificationsUpdateCmd())
}

func initDoNotificationsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("read", false, "Mark the notification as read (true/false)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoNotificationsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoNotificationsUpdateOptions(cmd, args)
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
		"type": "notifications",
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

	body, _, err := client.Patch(cmd.Context(), "/v1/notifications/"+opts.ID, jsonBody)
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

	row := notificationRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated notification %s\n", row.ID)
	return nil
}

func parseDoNotificationsUpdateOptions(cmd *cobra.Command, args []string) (doNotificationsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	read, _ := cmd.Flags().GetBool("read")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doNotificationsUpdateOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		ID:      args[0],
		Read:    read,
	}, nil
}
