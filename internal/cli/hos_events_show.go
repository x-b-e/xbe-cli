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

type hosEventsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type hosEventDetails struct {
	ID                     string `json:"id"`
	RegulationSetCode      string `json:"regulation_set_code,omitempty"`
	EventType              string `json:"event_type,omitempty"`
	WorkStatus             string `json:"work_status,omitempty"`
	StartAt                string `json:"start_at,omitempty"`
	EndAt                  string `json:"end_at,omitempty"`
	OccurredAt             string `json:"occurred_at,omitempty"`
	DurationSeconds        *int   `json:"duration_seconds,omitempty"`
	ConsecutiveRestSeconds *int   `json:"consecutive_rest_seconds,omitempty"`
	VendorStatus           any    `json:"vendor_status,omitempty"`
	Odometer               any    `json:"odometer,omitempty"`
	EngineHours            any    `json:"engine_hours,omitempty"`
	Metadata               any    `json:"metadata,omitempty"`
	BrokerID               string `json:"broker_id,omitempty"`
	HosDayID               string `json:"hos_day_id,omitempty"`
	UserID                 string `json:"user_id,omitempty"`
}

func newHosEventsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show HOS event details",
		Long: `Show the full details of a specific HOS event.

Output Fields:
  ID               HOS event identifier
  Regulation Set   Regulation set code
  Event Type       Event type
  Work Status      Work status
  Start At         Start timestamp
  End At           End timestamp
  Occurred At      Event timestamp
  Duration         Duration in seconds
  Consecutive Rest Consecutive rest seconds
  Vendor Status    Vendor-provided status metadata
  Odometer         Odometer metadata (if available)
  Engine Hours     Engine hour metadata (if available)
  Metadata         Additional metadata payload
  Broker ID        Broker relationship ID
  HOS Day ID       HOS day relationship ID
  User ID          User (driver) relationship ID`,
		Example: `  # Show a HOS event by ID
  xbe view hos-events show 123

  # Show a HOS event as JSON
  xbe view hos-events show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runHosEventsShow,
	}
	initHosEventsShowFlags(cmd)
	return cmd
}

func init() {
	hosEventsCmd.AddCommand(newHosEventsShowCmd())
}

func initHosEventsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runHosEventsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseHosEventsShowOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.NoAuth {
		opts.Token = ""
	} else if strings.TrimSpace(opts.Token) == "" {
		if token, _, err := auth.ResolveToken(opts.BaseURL, ""); err == nil {
			opts.Token = token
		} else if !errors.Is(err, auth.ErrNotFound) {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	id := strings.TrimSpace(args[0])
	if id == "" {
		return fmt.Errorf("hos event id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[hos-events]", "regulation-set-code,event-type,work-status,start-at,end-at,occurred-at,duration-seconds,consecutive-rest-seconds,metadata,vendor-status,odometer,engine-hours")

	body, _, err := client.Get(cmd.Context(), "/v1/hos-events/"+id, query)
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

	details := buildHosEventDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderHosEventDetails(cmd, details)
}

func parseHosEventsShowOptions(cmd *cobra.Command) (hosEventsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return hosEventsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildHosEventDetails(resp jsonAPISingleResponse) hosEventDetails {
	resource := resp.Data
	attrs := resource.Attributes

	details := hosEventDetails{
		ID:                resource.ID,
		RegulationSetCode: stringAttr(attrs, "regulation-set-code"),
		EventType:         stringAttr(attrs, "event-type"),
		WorkStatus:        stringAttr(attrs, "work-status"),
		StartAt:           formatDateTime(stringAttr(attrs, "start-at")),
		EndAt:             formatDateTime(stringAttr(attrs, "end-at")),
		OccurredAt:        formatDateTime(stringAttr(attrs, "occurred-at")),
	}

	if duration, ok := intAttrValue(attrs, "duration-seconds"); ok {
		details.DurationSeconds = &duration
	}
	if rest, ok := intAttrValue(attrs, "consecutive-rest-seconds"); ok {
		details.ConsecutiveRestSeconds = &rest
	}

	if attrs != nil {
		details.Metadata = attrs["metadata"]
		details.VendorStatus = attrs["vendor-status"]
		details.Odometer = attrs["odometer"]
		details.EngineHours = attrs["engine-hours"]
	}

	if rel, ok := resource.Relationships["broker"]; ok && rel.Data != nil {
		details.BrokerID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["hos-day"]; ok && rel.Data != nil {
		details.HosDayID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["user"]; ok && rel.Data != nil {
		details.UserID = rel.Data.ID
	}

	return details
}

func renderHosEventDetails(cmd *cobra.Command, details hosEventDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.RegulationSetCode != "" {
		fmt.Fprintf(out, "Regulation Set: %s\n", details.RegulationSetCode)
	}
	if details.EventType != "" {
		fmt.Fprintf(out, "Event Type: %s\n", details.EventType)
	}
	if details.WorkStatus != "" {
		fmt.Fprintf(out, "Work Status: %s\n", details.WorkStatus)
	}
	if details.StartAt != "" {
		fmt.Fprintf(out, "Start At: %s\n", details.StartAt)
	}
	if details.EndAt != "" {
		fmt.Fprintf(out, "End At: %s\n", details.EndAt)
	}
	if details.OccurredAt != "" {
		fmt.Fprintf(out, "Occurred At: %s\n", details.OccurredAt)
	}
	if details.DurationSeconds != nil {
		fmt.Fprintf(out, "Duration (seconds): %d\n", *details.DurationSeconds)
	}
	if details.ConsecutiveRestSeconds != nil {
		fmt.Fprintf(out, "Consecutive Rest (seconds): %d\n", *details.ConsecutiveRestSeconds)
	}
	if details.BrokerID != "" {
		fmt.Fprintf(out, "Broker ID: %s\n", details.BrokerID)
	}
	if details.HosDayID != "" {
		fmt.Fprintf(out, "HOS Day ID: %s\n", details.HosDayID)
	}
	if details.UserID != "" {
		fmt.Fprintf(out, "User ID: %s\n", details.UserID)
	}
	if formatted := formatJSONValue(details.VendorStatus); formatted != "" {
		fmt.Fprintln(out, "Vendor Status:")
		fmt.Fprintln(out, indentLines(formatted, "  "))
	}
	if formatted := formatJSONValue(details.Odometer); formatted != "" {
		fmt.Fprintln(out, "Odometer:")
		fmt.Fprintln(out, indentLines(formatted, "  "))
	}
	if formatted := formatJSONValue(details.EngineHours); formatted != "" {
		fmt.Fprintln(out, "Engine Hours:")
		fmt.Fprintln(out, indentLines(formatted, "  "))
	}
	if formatted := formatJSONValue(details.Metadata); formatted != "" {
		fmt.Fprintln(out, "Metadata:")
		fmt.Fprintln(out, indentLines(formatted, "  "))
	}

	return nil
}
