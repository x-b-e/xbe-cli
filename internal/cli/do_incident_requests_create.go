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

type doIncidentRequestsCreateOptions struct {
	BaseURL                string
	Token                  string
	JSON                   bool
	StartAt                string
	EndAt                  string
	Description            string
	TimeValueType          string
	IsDownTime             bool
	TenderJobScheduleShift string
	Assignee               string
	CreatedBy              string
}

func newDoIncidentRequestsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create an incident request",
		Long: `Create a new incident request.

Required flags:
  --start-at                 Start timestamp (ISO 8601, required)
  --tender-job-schedule-shift Tender job schedule shift ID (required)

Optional flags:
  --end-at           End timestamp (ISO 8601)
  --description      Description text
  --time-value-type  Time value type (deducted_time or credited_time)
  --is-down-time     Mark as down time (true/false)
  --assignee         Assignee user ID
  --created-by       Created by user ID

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Create an incident request
  xbe do incident-requests create --start-at 2025-01-01T08:00:00Z --tender-job-schedule-shift 123

  # Create with time value type and assignee
  xbe do incident-requests create \
    --start-at 2025-01-01T08:00:00Z \
    --end-at 2025-01-01T09:00:00Z \
    --time-value-type credited_time \
    --is-down-time \
    --assignee 456 \
    --tender-job-schedule-shift 123`,
		Args: cobra.NoArgs,
		RunE: runDoIncidentRequestsCreate,
	}
	initDoIncidentRequestsCreateFlags(cmd)
	return cmd
}

func init() {
	doIncidentRequestsCmd.AddCommand(newDoIncidentRequestsCreateCmd())
}

func initDoIncidentRequestsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("start-at", "", "Start timestamp (ISO 8601, required)")
	cmd.Flags().String("end-at", "", "End timestamp (ISO 8601)")
	cmd.Flags().String("description", "", "Description text")
	cmd.Flags().String("time-value-type", "", "Time value type (deducted_time or credited_time)")
	cmd.Flags().Bool("is-down-time", false, "Mark as down time")
	cmd.Flags().String("tender-job-schedule-shift", "", "Tender job schedule shift ID (required)")
	cmd.Flags().String("assignee", "", "Assignee user ID")
	cmd.Flags().String("created-by", "", "Created by user ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoIncidentRequestsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoIncidentRequestsCreateOptions(cmd)
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

	if strings.TrimSpace(opts.StartAt) == "" {
		err := fmt.Errorf("--start-at is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if strings.TrimSpace(opts.TenderJobScheduleShift) == "" {
		err := fmt.Errorf("--tender-job-schedule-shift is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{
		"start-at": opts.StartAt,
	}
	setStringAttrIfPresent(attributes, "end-at", opts.EndAt)
	setStringAttrIfPresent(attributes, "description", opts.Description)
	setStringAttrIfPresent(attributes, "time-value-type", opts.TimeValueType)
	if cmd.Flags().Changed("is-down-time") {
		attributes["is-down-time"] = opts.IsDownTime
	}

	relationships := map[string]any{
		"tender-job-schedule-shift": map[string]any{
			"data": map[string]any{
				"type": "tender-job-schedule-shifts",
				"id":   opts.TenderJobScheduleShift,
			},
		},
	}

	if opts.Assignee != "" {
		relationships["assignee"] = map[string]any{
			"data": map[string]any{
				"type": "users",
				"id":   opts.Assignee,
			},
		}
	}
	if opts.CreatedBy != "" {
		relationships["created-by"] = map[string]any{
			"data": map[string]any{
				"type": "users",
				"id":   opts.CreatedBy,
			},
		}
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "incident-requests",
			"attributes":    attributes,
			"relationships": relationships,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/incident-requests", jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Created incident request %s\n", details.ID)
	return renderIncidentRequestDetails(cmd, details)
}

func parseDoIncidentRequestsCreateOptions(cmd *cobra.Command) (doIncidentRequestsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	startAt, _ := cmd.Flags().GetString("start-at")
	endAt, _ := cmd.Flags().GetString("end-at")
	description, _ := cmd.Flags().GetString("description")
	timeValueType, _ := cmd.Flags().GetString("time-value-type")
	isDownTime, _ := cmd.Flags().GetBool("is-down-time")
	tenderJobScheduleShift, _ := cmd.Flags().GetString("tender-job-schedule-shift")
	assignee, _ := cmd.Flags().GetString("assignee")
	createdBy, _ := cmd.Flags().GetString("created-by")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doIncidentRequestsCreateOptions{
		BaseURL:                baseURL,
		Token:                  token,
		JSON:                   jsonOut,
		StartAt:                startAt,
		EndAt:                  endAt,
		Description:            description,
		TimeValueType:          timeValueType,
		IsDownTime:             isDownTime,
		TenderJobScheduleShift: tenderJobScheduleShift,
		Assignee:               assignee,
		CreatedBy:              createdBy,
	}, nil
}
