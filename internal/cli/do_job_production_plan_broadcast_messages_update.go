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

type doJobProductionPlanBroadcastMessagesUpdateOptions struct {
	BaseURL    string
	Token      string
	JSON       bool
	ID         string
	IsHidden   bool
	NoIsHidden bool
}

func newDoJobProductionPlanBroadcastMessagesUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a job production plan broadcast message",
		Long: `Update a job production plan broadcast message.

Arguments:
  <id>  The broadcast message ID (required)

Optional flags:
  --is-hidden      Hide the message
  --no-is-hidden   Unhide the message

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Hide a broadcast message
  xbe do job-production-plan-broadcast-messages update 123 --is-hidden

  # Unhide a broadcast message
  xbe do job-production-plan-broadcast-messages update 123 --no-is-hidden`,
		Args: cobra.ExactArgs(1),
		RunE: runDoJobProductionPlanBroadcastMessagesUpdate,
	}
	initDoJobProductionPlanBroadcastMessagesUpdateFlags(cmd)
	return cmd
}

func init() {
	doJobProductionPlanBroadcastMessagesCmd.AddCommand(newDoJobProductionPlanBroadcastMessagesUpdateCmd())
}

func initDoJobProductionPlanBroadcastMessagesUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("is-hidden", false, "Hide the message")
	cmd.Flags().Bool("no-is-hidden", false, "Unhide the message")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoJobProductionPlanBroadcastMessagesUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoJobProductionPlanBroadcastMessagesUpdateOptions(cmd, args)
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
	if cmd.Flags().Changed("is-hidden") {
		attributes["is-hidden"] = true
	}
	if cmd.Flags().Changed("no-is-hidden") {
		attributes["is-hidden"] = false
	}

	if len(attributes) == 0 {
		err := fmt.Errorf("no attributes to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "job-production-plan-broadcast-messages",
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

	body, _, err := client.Patch(cmd.Context(), "/v1/job-production-plan-broadcast-messages/"+opts.ID, jsonBody)
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

	row := buildJobProductionPlanBroadcastMessageRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated job production plan broadcast message %s\n", row.ID)
	return nil
}

func parseDoJobProductionPlanBroadcastMessagesUpdateOptions(cmd *cobra.Command, args []string) (doJobProductionPlanBroadcastMessagesUpdateOptions, error) {
	id := strings.TrimSpace(args[0])
	if id == "" {
		return doJobProductionPlanBroadcastMessagesUpdateOptions{}, fmt.Errorf("broadcast message id is required")
	}

	jsonOut, _ := cmd.Flags().GetBool("json")
	isHidden, _ := cmd.Flags().GetBool("is-hidden")
	noIsHidden, _ := cmd.Flags().GetBool("no-is-hidden")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doJobProductionPlanBroadcastMessagesUpdateOptions{
		BaseURL:    baseURL,
		Token:      token,
		JSON:       jsonOut,
		ID:         id,
		IsHidden:   isHidden,
		NoIsHidden: noIsHidden,
	}, nil
}
