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

type doActionItemTrackerUpdateRequestsCreateOptions struct {
	BaseURL           string
	Token             string
	JSON              bool
	ActionItemTracker string
	RequestedBy       string
	RequestedFrom     string
	RequestNote       string
	DueOn             string
	UpdateNote        string
}

func newDoActionItemTrackerUpdateRequestsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create an action item tracker update request",
		Long: `Create a new action item tracker update request.

Required flags:
  --action-item-tracker  Action item tracker ID (required)
  --requested-by         User ID requesting the update (required)
  --requested-from       User ID expected to provide the update (required)

Optional flags:
  --request-note         Request note
  --due-on               Requested due date (ISO 8601, e.g. 2024-12-31)
  --update-note          Fulfillment update note`,
		Example: `  # Create an update request
  xbe do action-item-tracker-update-requests create \
    --action-item-tracker 123 \
    --requested-by 456 \
    --requested-from 789 \
    --request-note "Please send an update" \
    --due-on 2024-12-31

  # Create and include an update note
  xbe do action-item-tracker-update-requests create \
    --action-item-tracker 123 \
    --requested-by 456 \
    --requested-from 789 \
    --update-note "Completed as of today"`,
		Args: cobra.NoArgs,
		RunE: runDoActionItemTrackerUpdateRequestsCreate,
	}
	initDoActionItemTrackerUpdateRequestsCreateFlags(cmd)
	return cmd
}

func init() {
	doActionItemTrackerUpdateRequestsCmd.AddCommand(newDoActionItemTrackerUpdateRequestsCreateCmd())
}

func initDoActionItemTrackerUpdateRequestsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("action-item-tracker", "", "Action item tracker ID (required)")
	cmd.Flags().String("requested-by", "", "User ID requesting the update (required)")
	cmd.Flags().String("requested-from", "", "User ID expected to provide the update (required)")
	cmd.Flags().String("request-note", "", "Request note")
	cmd.Flags().String("due-on", "", "Requested due date (ISO 8601, e.g. 2024-12-31)")
	cmd.Flags().String("update-note", "", "Fulfillment update note")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoActionItemTrackerUpdateRequestsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoActionItemTrackerUpdateRequestsCreateOptions(cmd)
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

	if opts.ActionItemTracker == "" {
		err := fmt.Errorf("--action-item-tracker is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.RequestedBy == "" {
		err := fmt.Errorf("--requested-by is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.RequestedFrom == "" {
		err := fmt.Errorf("--requested-from is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{}
	if opts.RequestNote != "" {
		attributes["request-note"] = opts.RequestNote
	}
	if opts.DueOn != "" {
		attributes["due-on"] = opts.DueOn
	}
	if opts.UpdateNote != "" {
		attributes["update-note"] = opts.UpdateNote
	}

	relationships := map[string]any{
		"action-item-tracker": map[string]any{
			"data": map[string]any{
				"type": "action-item-trackers",
				"id":   opts.ActionItemTracker,
			},
		},
		"requested-by": map[string]any{
			"data": map[string]any{
				"type": "users",
				"id":   opts.RequestedBy,
			},
		},
		"requested-from": map[string]any{
			"data": map[string]any{
				"type": "users",
				"id":   opts.RequestedFrom,
			},
		},
	}

	data := map[string]any{
		"type":          "action-item-tracker-update-requests",
		"relationships": relationships,
	}
	if len(attributes) > 0 {
		data["attributes"] = attributes
	}

	requestBody := map[string]any{"data": data}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/action-item-tracker-update-requests", jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Created action item tracker update request %s\n", row.ID)
	return nil
}

func parseDoActionItemTrackerUpdateRequestsCreateOptions(cmd *cobra.Command) (doActionItemTrackerUpdateRequestsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	actionItemTracker, _ := cmd.Flags().GetString("action-item-tracker")
	requestedBy, _ := cmd.Flags().GetString("requested-by")
	requestedFrom, _ := cmd.Flags().GetString("requested-from")
	requestNote, _ := cmd.Flags().GetString("request-note")
	dueOn, _ := cmd.Flags().GetString("due-on")
	updateNote, _ := cmd.Flags().GetString("update-note")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doActionItemTrackerUpdateRequestsCreateOptions{
		BaseURL:           baseURL,
		Token:             token,
		JSON:              jsonOut,
		ActionItemTracker: actionItemTracker,
		RequestedBy:       requestedBy,
		RequestedFrom:     requestedFrom,
		RequestNote:       requestNote,
		DueOn:             dueOn,
		UpdateNote:        updateNote,
	}, nil
}
