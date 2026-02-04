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

type timeSheetsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type timeSheetDetails struct {
	ID              string `json:"id"`
	Status          string `json:"status,omitempty"`
	EffectiveStatus string `json:"effective_status,omitempty"`
	StartAt         string `json:"start_at,omitempty"`
	EndAt           string `json:"end_at,omitempty"`
	BreakMinutes    string `json:"break_minutes,omitempty"`
	DurationMinutes string `json:"duration_minutes,omitempty"`
	TimeZoneID      string `json:"time_zone_id,omitempty"`
	ProductivePct   string `json:"productive_pct,omitempty"`
	JobSiteHours    string `json:"job_site_hours,omitempty"`
	Notes           string `json:"notes,omitempty"`
	SubmittedTerms  string `json:"submitted_terms,omitempty"`

	SkipValidateOverlap bool `json:"skip_validate_overlap,omitempty"`

	SubjectType string `json:"subject_type,omitempty"`
	SubjectID   string `json:"subject_id,omitempty"`

	TimeSheetCostCodeAllocationID string `json:"time_sheet_cost_code_allocation_id,omitempty"`
	TruckerShiftSetID             string `json:"trucker_shift_set_id,omitempty"`
	CrewRequirementID             string `json:"crew_requirement_id,omitempty"`
	LaborRequirementID            string `json:"labor_requirement_id,omitempty"`
	EquipmentRequirementID        string `json:"equipment_requirement_id,omitempty"`
	WorkOrderID                   string `json:"work_order_id,omitempty"`
	DriverID                      string `json:"driver_id,omitempty"`

	TimeSheetStatusChangeCount int `json:"time_sheet_status_change_count,omitempty"`
	TimeSheetLineItemCount     int `json:"time_sheet_line_item_count,omitempty"`
	TimeSheetNoShowCount       int `json:"time_sheet_no_show_count,omitempty"`
	FileAttachmentCount        int `json:"file_attachment_count,omitempty"`

	HasTimeSheetStatusChanges bool `json:"-"`
	HasTimeSheetLineItems     bool `json:"-"`
	HasTimeSheetNoShows       bool `json:"-"`
	HasFileAttachments        bool `json:"-"`
}

func newTimeSheetsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show time sheet details",
		Long: `Show the full details of a specific time sheet.

Output Fields:
  ID
  Status
  Effective Status
  Start At
  End At
  Break Minutes
  Duration Minutes
  Time Zone
  Productive Percent
  Job Site Hours
  Submitted Terms
  Notes
  Skip Validate Overlap
  Subject
  Driver
  Trucker Shift Set
  Crew Requirement
  Labor Requirement
  Equipment Requirement
  Work Order
  Time Sheet Cost Code Allocation
  Time Sheet Status Changes (count)
  Time Sheet Line Items (count)
  Time Sheet No Shows (count)
  File Attachments (count)

Arguments:
  <id>    Time sheet ID (required).

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a time sheet
  xbe view time-sheets show 123

  # JSON output
  xbe view time-sheets show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runTimeSheetsShow,
	}
	initTimeSheetsShowFlags(cmd)
	return cmd
}

func init() {
	timeSheetsCmd.AddCommand(newTimeSheetsShowCmd())
}

func initTimeSheetsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runTimeSheetsShow(cmd *cobra.Command, args []string) error {
	if handled, err := maybeHandleClientURLShow(cmd, args); err != nil {
		return err
	} else if handled {
		return nil
	}

	opts, err := parseTimeSheetsShowOptions(cmd)
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
		return fmt.Errorf("time sheet id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[time-sheets]", "start-at,end-at,break-minutes,notes,skip-validate-overlap,status,effective-status,duration-minutes,time-zone-id,productive-pct,job-site-hours,submitted-terms,subject,time-sheet-cost-code-allocation,trucker-shift-set,crew-requirement,labor-requirement,equipment-requirement,work-order,driver,time-sheet-status-changes,time-sheet-line-items,time-sheet-no-shows,file-attachments")

	body, _, err := client.Get(cmd.Context(), "/v1/time-sheets/"+id, query)
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

	details := buildTimeSheetDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderTimeSheetDetails(cmd, details)
}

func parseTimeSheetsShowOptions(cmd *cobra.Command) (timeSheetsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return timeSheetsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildTimeSheetDetails(resp jsonAPISingleResponse) timeSheetDetails {
	attrs := resp.Data.Attributes
	details := timeSheetDetails{
		ID:                  resp.Data.ID,
		Status:              stringAttr(attrs, "status"),
		EffectiveStatus:     stringAttr(attrs, "effective-status"),
		StartAt:             formatDateTime(stringAttr(attrs, "start-at")),
		EndAt:               formatDateTime(stringAttr(attrs, "end-at")),
		BreakMinutes:        stringAttr(attrs, "break-minutes"),
		DurationMinutes:     stringAttr(attrs, "duration-minutes"),
		TimeZoneID:          stringAttr(attrs, "time-zone-id"),
		ProductivePct:       stringAttr(attrs, "productive-pct"),
		JobSiteHours:        stringAttr(attrs, "job-site-hours"),
		Notes:               strings.TrimSpace(stringAttr(attrs, "notes")),
		SubmittedTerms:      strings.TrimSpace(stringAttr(attrs, "submitted-terms")),
		SkipValidateOverlap: boolAttr(attrs, "skip-validate-overlap"),
	}

	if rel, ok := resp.Data.Relationships["subject"]; ok && rel.Data != nil {
		details.SubjectType = rel.Data.Type
		details.SubjectID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["time-sheet-cost-code-allocation"]; ok && rel.Data != nil {
		details.TimeSheetCostCodeAllocationID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["trucker-shift-set"]; ok && rel.Data != nil {
		details.TruckerShiftSetID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["crew-requirement"]; ok && rel.Data != nil {
		details.CrewRequirementID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["labor-requirement"]; ok && rel.Data != nil {
		details.LaborRequirementID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["equipment-requirement"]; ok && rel.Data != nil {
		details.EquipmentRequirementID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["work-order"]; ok && rel.Data != nil {
		details.WorkOrderID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["driver"]; ok && rel.Data != nil {
		details.DriverID = rel.Data.ID
	}

	if rel, ok := resp.Data.Relationships["time-sheet-status-changes"]; ok {
		details.HasTimeSheetStatusChanges = true
		details.TimeSheetStatusChangeCount = relationshipCount(rel)
	}
	if rel, ok := resp.Data.Relationships["time-sheet-line-items"]; ok {
		details.HasTimeSheetLineItems = true
		details.TimeSheetLineItemCount = relationshipCount(rel)
	}
	if rel, ok := resp.Data.Relationships["time-sheet-no-shows"]; ok {
		details.HasTimeSheetNoShows = true
		details.TimeSheetNoShowCount = relationshipCount(rel)
	}
	if rel, ok := resp.Data.Relationships["file-attachments"]; ok {
		details.HasFileAttachments = true
		details.FileAttachmentCount = relationshipCount(rel)
	}

	return details
}

func renderTimeSheetDetails(cmd *cobra.Command, details timeSheetDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.Status != "" {
		fmt.Fprintf(out, "Status: %s\n", details.Status)
	}
	if details.EffectiveStatus != "" {
		fmt.Fprintf(out, "Effective Status: %s\n", details.EffectiveStatus)
	}
	if details.SubjectType != "" || details.SubjectID != "" {
		fmt.Fprintf(out, "Subject: %s\n", formatTimeSheetSubject(details.SubjectType, details.SubjectID))
	}
	if details.StartAt != "" {
		fmt.Fprintf(out, "Start At: %s\n", details.StartAt)
	}
	if details.EndAt != "" {
		fmt.Fprintf(out, "End At: %s\n", details.EndAt)
	}
	if details.BreakMinutes != "" {
		fmt.Fprintf(out, "Break Minutes: %s\n", details.BreakMinutes)
	}
	if details.DurationMinutes != "" {
		fmt.Fprintf(out, "Duration Minutes: %s\n", details.DurationMinutes)
	}
	if details.TimeZoneID != "" {
		fmt.Fprintf(out, "Time Zone: %s\n", details.TimeZoneID)
	}
	if details.ProductivePct != "" {
		fmt.Fprintf(out, "Productive Percent: %s\n", details.ProductivePct)
	}
	if details.JobSiteHours != "" {
		fmt.Fprintf(out, "Job Site Hours: %s\n", details.JobSiteHours)
	}
	if details.SubmittedTerms != "" {
		fmt.Fprintf(out, "Submitted Terms: %s\n", details.SubmittedTerms)
	}
	if details.Notes != "" {
		fmt.Fprintf(out, "Notes: %s\n", details.Notes)
	}
	fmt.Fprintf(out, "Skip Validate Overlap: %t\n", details.SkipValidateOverlap)

	if details.DriverID != "" {
		fmt.Fprintf(out, "Driver: %s\n", details.DriverID)
	}
	if details.TruckerShiftSetID != "" {
		fmt.Fprintf(out, "Trucker Shift Set: %s\n", details.TruckerShiftSetID)
	}
	if details.CrewRequirementID != "" {
		fmt.Fprintf(out, "Crew Requirement: %s\n", details.CrewRequirementID)
	}
	if details.LaborRequirementID != "" {
		fmt.Fprintf(out, "Labor Requirement: %s\n", details.LaborRequirementID)
	}
	if details.EquipmentRequirementID != "" {
		fmt.Fprintf(out, "Equipment Requirement: %s\n", details.EquipmentRequirementID)
	}
	if details.WorkOrderID != "" {
		fmt.Fprintf(out, "Work Order: %s\n", details.WorkOrderID)
	}
	if details.TimeSheetCostCodeAllocationID != "" {
		fmt.Fprintf(out, "Time Sheet Cost Code Allocation: %s\n", details.TimeSheetCostCodeAllocationID)
	}

	if details.HasTimeSheetStatusChanges {
		fmt.Fprintf(out, "Time Sheet Status Changes: %d\n", details.TimeSheetStatusChangeCount)
	}
	if details.HasTimeSheetLineItems {
		fmt.Fprintf(out, "Time Sheet Line Items: %d\n", details.TimeSheetLineItemCount)
	}
	if details.HasTimeSheetNoShows {
		fmt.Fprintf(out, "Time Sheet No Shows: %d\n", details.TimeSheetNoShowCount)
	}
	if details.HasFileAttachments {
		fmt.Fprintf(out, "File Attachments: %d\n", details.FileAttachmentCount)
	}

	return nil
}
