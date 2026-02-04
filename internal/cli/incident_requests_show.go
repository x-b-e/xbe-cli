package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type incidentRequestsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type incidentRequestDetails struct {
	ID                       string   `json:"id"`
	Status                   string   `json:"status,omitempty"`
	StartAt                  string   `json:"start_at,omitempty"`
	EndAt                    string   `json:"end_at,omitempty"`
	TimeValueType            string   `json:"time_value_type,omitempty"`
	IsDownTime               bool     `json:"is_down_time"`
	Description              string   `json:"description,omitempty"`
	TenderJobScheduleShiftID string   `json:"tender_job_schedule_shift_id,omitempty"`
	AssigneeID               string   `json:"assignee_id,omitempty"`
	CreatedByID              string   `json:"created_by_id,omitempty"`
	CustomerID               string   `json:"customer_id,omitempty"`
	BrokerID                 string   `json:"broker_id,omitempty"`
	IncidentID               string   `json:"incident_id,omitempty"`
	CommentIDs               []string `json:"comment_ids,omitempty"`
	AttachmentIDs            []string `json:"attachment_ids,omitempty"`
}

func newIncidentRequestsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show incident request details",
		Long: `Show the full details of an incident request.

Output Fields:
  ID, status, time value type, downtime flag
  Start/end timestamps and description
  Related resource IDs (shift, assignee, created by, customer, broker, incident)
  Comment and attachment IDs

Arguments:
  <id>    The incident request ID (required). Use the list command to find IDs.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show an incident request
  xbe view incident-requests show 123

  # JSON output
  xbe view incident-requests show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runIncidentRequestsShow,
	}
	initIncidentRequestsShowFlags(cmd)
	return cmd
}

func init() {
	incidentRequestsCmd.AddCommand(newIncidentRequestsShowCmd())
}

func initIncidentRequestsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runIncidentRequestsShow(cmd *cobra.Command, args []string) error {
	if handled, err := maybeHandleClientURLShow(cmd, args); err != nil {
		return err
	} else if handled {
		return nil
	}

	opts, err := parseIncidentRequestsShowOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.NoAuth {
		opts.Token = ""
	} else if strings.TrimSpace(opts.Token) == "" {
		if token, _, err := auth.ResolveToken(opts.BaseURL, ""); err == nil {
			opts.Token = token
		} else if errors.Is(err, auth.ErrNotFound) {
			fmt.Fprintln(cmd.ErrOrStderr(), "Authentication required. Run \"xbe auth login\" first.")
			return err
		} else {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	id := strings.TrimSpace(args[0])
	if id == "" {
		return fmt.Errorf("incident request id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[incident-requests]", "status,start-at,end-at,description,is-down-time,time-value-type,tender-job-schedule-shift,assignee,created-by,customer,broker,incident,comments,file-attachments")
	query.Set("include", "tender-job-schedule-shift,assignee,created-by,customer,broker,incident,comments,file-attachments")

	body, _, err := client.Get(cmd.Context(), "/v1/incident-requests/"+id, query)
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

	handled, err := renderSparseShowIfRequested(cmd, resp)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if handled {
		return nil
	}

	details := buildIncidentRequestDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderIncidentRequestDetails(cmd, details)
}

func parseIncidentRequestsShowOptions(cmd *cobra.Command) (incidentRequestsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return incidentRequestsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildIncidentRequestDetails(resp jsonAPISingleResponse) incidentRequestDetails {
	attrs := resp.Data.Attributes
	details := incidentRequestDetails{
		ID:            resp.Data.ID,
		Status:        stringAttr(attrs, "status"),
		StartAt:       formatDateTime(stringAttr(attrs, "start-at")),
		EndAt:         formatDateTime(stringAttr(attrs, "end-at")),
		TimeValueType: stringAttr(attrs, "time-value-type"),
		IsDownTime:    boolAttr(attrs, "is-down-time"),
		Description:   stringAttr(attrs, "description"),
	}

	details.TenderJobScheduleShiftID = relationshipIDFromMap(resp.Data.Relationships, "tender-job-schedule-shift")
	details.AssigneeID = relationshipIDFromMap(resp.Data.Relationships, "assignee")
	details.CreatedByID = relationshipIDFromMap(resp.Data.Relationships, "created-by")
	details.CustomerID = relationshipIDFromMap(resp.Data.Relationships, "customer")
	details.BrokerID = relationshipIDFromMap(resp.Data.Relationships, "broker")
	details.IncidentID = relationshipIDFromMap(resp.Data.Relationships, "incident")
	details.CommentIDs = relationshipIDsFromMap(resp.Data.Relationships, "comments")
	details.AttachmentIDs = relationshipIDsFromMap(resp.Data.Relationships, "file-attachments")

	return details
}

func renderIncidentRequestDetails(cmd *cobra.Command, details incidentRequestDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.Status != "" {
		fmt.Fprintf(out, "Status: %s\n", details.Status)
	}
	if details.TimeValueType != "" {
		fmt.Fprintf(out, "Time Value Type: %s\n", details.TimeValueType)
	}
	fmt.Fprintf(out, "Down Time: %s\n", formatYesNo(details.IsDownTime))
	if details.StartAt != "" {
		fmt.Fprintf(out, "Start At: %s\n", details.StartAt)
	}
	if details.EndAt != "" {
		fmt.Fprintf(out, "End At: %s\n", details.EndAt)
	}
	if details.Description != "" {
		fmt.Fprintf(out, "Description: %s\n", details.Description)
	}
	if details.TenderJobScheduleShiftID != "" {
		fmt.Fprintf(out, "Tender Job Schedule Shift ID: %s\n", details.TenderJobScheduleShiftID)
	}
	if details.AssigneeID != "" {
		fmt.Fprintf(out, "Assignee ID: %s\n", details.AssigneeID)
	}
	if details.CreatedByID != "" {
		fmt.Fprintf(out, "Created By ID: %s\n", details.CreatedByID)
	}
	if details.CustomerID != "" {
		fmt.Fprintf(out, "Customer ID: %s\n", details.CustomerID)
	}
	if details.BrokerID != "" {
		fmt.Fprintf(out, "Broker ID: %s\n", details.BrokerID)
	}
	if details.IncidentID != "" {
		fmt.Fprintf(out, "Incident ID: %s\n", details.IncidentID)
	}
	if len(details.CommentIDs) > 0 {
		fmt.Fprintf(out, "Comment IDs: %s\n", strings.Join(details.CommentIDs, ", "))
	}
	if len(details.AttachmentIDs) > 0 {
		fmt.Fprintf(out, "Attachment IDs: %s\n", strings.Join(details.AttachmentIDs, ", "))
	}

	return nil
}
