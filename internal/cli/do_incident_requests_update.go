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

type doIncidentRequestsUpdateOptions struct {
	BaseURL       string
	Token         string
	JSON          bool
	ID            string
	StartAt       string
	EndAt         string
	Description   string
	TimeValueType string
	IsDownTime    bool
	Assignee      string
	CreatedBy     string
}

func newDoIncidentRequestsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update an incident request",
		Long: `Update an incident request.

All flags are optional. Only provided flags will update the incident request.

Optional flags:
  --start-at        Update start timestamp (ISO 8601)
  --end-at          Update end timestamp (ISO 8601)
  --description     Update description text
  --time-value-type Update time value type (deducted_time or credited_time)
  --is-down-time    Update down time flag (true/false)

Relationships:
  --assignee        Assignee user ID
  --created-by      Created by user ID`,
		Example: `  # Update description
  xbe do incident-requests update 123 --description "Updated description"

  # Update time value type and down time flag
  xbe do incident-requests update 123 --time-value-type deducted_time --is-down-time false`,
		Args: cobra.ExactArgs(1),
		RunE: runDoIncidentRequestsUpdate,
	}
	initDoIncidentRequestsUpdateFlags(cmd)
	return cmd
}

func init() {
	doIncidentRequestsCmd.AddCommand(newDoIncidentRequestsUpdateCmd())
}

func initDoIncidentRequestsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("start-at", "", "Update start timestamp (ISO 8601)")
	cmd.Flags().String("end-at", "", "Update end timestamp (ISO 8601)")
	cmd.Flags().String("description", "", "Update description text")
	cmd.Flags().String("time-value-type", "", "Update time value type (deducted_time or credited_time)")
	cmd.Flags().Bool("is-down-time", false, "Update down time flag")
	cmd.Flags().String("assignee", "", "Assignee user ID")
	cmd.Flags().String("created-by", "", "Created by user ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoIncidentRequestsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoIncidentRequestsUpdateOptions(cmd, args)
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
	relationships := map[string]any{}

	if cmd.Flags().Changed("start-at") {
		attributes["start-at"] = opts.StartAt
	}
	if cmd.Flags().Changed("end-at") {
		attributes["end-at"] = opts.EndAt
	}
	if cmd.Flags().Changed("description") {
		attributes["description"] = opts.Description
	}
	if cmd.Flags().Changed("time-value-type") {
		attributes["time-value-type"] = opts.TimeValueType
	}
	if cmd.Flags().Changed("is-down-time") {
		attributes["is-down-time"] = opts.IsDownTime
	}

	if cmd.Flags().Changed("assignee") {
		if strings.TrimSpace(opts.Assignee) == "" {
			err := fmt.Errorf("--assignee cannot be empty")
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		relationships["assignee"] = map[string]any{
			"data": map[string]any{
				"type": "users",
				"id":   opts.Assignee,
			},
		}
	}

	if cmd.Flags().Changed("created-by") {
		if strings.TrimSpace(opts.CreatedBy) == "" {
			err := fmt.Errorf("--created-by cannot be empty")
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		relationships["created-by"] = map[string]any{
			"data": map[string]any{
				"type": "users",
				"id":   opts.CreatedBy,
			},
		}
	}

	if len(attributes) == 0 && len(relationships) == 0 {
		err := fmt.Errorf("at least one field must be specified for update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	data := map[string]any{
		"type":       "incident-requests",
		"id":         opts.ID,
		"attributes": attributes,
	}
	if len(relationships) > 0 {
		data["relationships"] = relationships
	}

	requestBody := map[string]any{
		"data": data,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Patch(cmd.Context(), "/v1/incident-requests/"+opts.ID, jsonBody)
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

	details := buildIncidentRequestDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated incident request %s\n", details.ID)
	return renderIncidentRequestDetails(cmd, details)
}

func parseDoIncidentRequestsUpdateOptions(cmd *cobra.Command, args []string) (doIncidentRequestsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	startAt, _ := cmd.Flags().GetString("start-at")
	endAt, _ := cmd.Flags().GetString("end-at")
	description, _ := cmd.Flags().GetString("description")
	timeValueType, _ := cmd.Flags().GetString("time-value-type")
	isDownTime, _ := cmd.Flags().GetBool("is-down-time")
	assignee, _ := cmd.Flags().GetString("assignee")
	createdBy, _ := cmd.Flags().GetString("created-by")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doIncidentRequestsUpdateOptions{
		BaseURL:       baseURL,
		Token:         token,
		JSON:          jsonOut,
		ID:            args[0],
		StartAt:       startAt,
		EndAt:         endAt,
		Description:   description,
		TimeValueType: timeValueType,
		IsDownTime:    isDownTime,
		Assignee:      assignee,
		CreatedBy:     createdBy,
	}, nil
}
