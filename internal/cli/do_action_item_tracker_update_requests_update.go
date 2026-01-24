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

type doActionItemTrackerUpdateRequestsUpdateOptions struct {
	BaseURL     string
	Token       string
	JSON        bool
	ID          string
	RequestNote string
	DueOn       string
	UpdateNote  string
}

func newDoActionItemTrackerUpdateRequestsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update an action item tracker update request",
		Long: `Update an existing action item tracker update request.

Optional flags:
  --request-note  Request note
  --due-on        Requested due date (ISO 8601, e.g. 2024-12-31)
  --update-note   Fulfillment update note`,
		Example: `  # Update the request note
  xbe do action-item-tracker-update-requests update 123 --request-note "Updated request"

  # Fulfill with an update note
  xbe do action-item-tracker-update-requests update 123 --update-note "Work complete"`,
		Args: cobra.ExactArgs(1),
		RunE: runDoActionItemTrackerUpdateRequestsUpdate,
	}
	initDoActionItemTrackerUpdateRequestsUpdateFlags(cmd)
	return cmd
}

func init() {
	doActionItemTrackerUpdateRequestsCmd.AddCommand(newDoActionItemTrackerUpdateRequestsUpdateCmd())
}

func initDoActionItemTrackerUpdateRequestsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("request-note", "", "Request note")
	cmd.Flags().String("due-on", "", "Requested due date (ISO 8601, e.g. 2024-12-31)")
	cmd.Flags().String("update-note", "", "Fulfillment update note")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoActionItemTrackerUpdateRequestsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoActionItemTrackerUpdateRequestsUpdateOptions(cmd, args)
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
	if cmd.Flags().Changed("request-note") {
		attributes["request-note"] = opts.RequestNote
	}
	if opts.DueOn != "" {
		attributes["due-on"] = opts.DueOn
	}
	if cmd.Flags().Changed("update-note") {
		attributes["update-note"] = opts.UpdateNote
	}

	if len(attributes) == 0 {
		err := fmt.Errorf("no attributes to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "action-item-tracker-update-requests",
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

	body, _, err := client.Patch(cmd.Context(), "/v1/action-item-tracker-update-requests/"+opts.ID, jsonBody)
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

	row := buildActionItemTrackerUpdateRequestRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated action item tracker update request %s\n", row.ID)
	return nil
}

func parseDoActionItemTrackerUpdateRequestsUpdateOptions(cmd *cobra.Command, args []string) (doActionItemTrackerUpdateRequestsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	requestNote, _ := cmd.Flags().GetString("request-note")
	dueOn, _ := cmd.Flags().GetString("due-on")
	updateNote, _ := cmd.Flags().GetString("update-note")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doActionItemTrackerUpdateRequestsUpdateOptions{
		BaseURL:     baseURL,
		Token:       token,
		JSON:        jsonOut,
		ID:          args[0],
		RequestNote: requestNote,
		DueOn:       dueOn,
		UpdateNote:  updateNote,
	}, nil
}
