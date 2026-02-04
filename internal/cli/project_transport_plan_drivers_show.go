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

type projectTransportPlanDriversShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type projectTransportPlanDriverDetails struct {
	ID                                        string   `json:"id"`
	Status                                    string   `json:"status,omitempty"`
	ConfirmNote                               string   `json:"confirm_note,omitempty"`
	ConfirmAtMax                              string   `json:"confirm_at_max,omitempty"`
	WindowStartAtCached                       string   `json:"window_start_at_cached,omitempty"`
	WindowEndAtCached                         string   `json:"window_end_at_cached,omitempty"`
	SkipAssignmentRulesValidation             bool     `json:"skip_assignment_rules_validation"`
	AssignmentRuleOverrideReason              string   `json:"assignment_rule_override_reason,omitempty"`
	ProjectTransportPlanID                    string   `json:"project_transport_plan_id,omitempty"`
	ProjectTransportPlanStatus                string   `json:"project_transport_plan_status,omitempty"`
	BrokerID                                  string   `json:"broker_id,omitempty"`
	BrokerName                                string   `json:"broker_name,omitempty"`
	SegmentStartID                            string   `json:"segment_start_id,omitempty"`
	SegmentStartPosition                      string   `json:"segment_start_position,omitempty"`
	SegmentEndID                              string   `json:"segment_end_id,omitempty"`
	SegmentEndPosition                        string   `json:"segment_end_position,omitempty"`
	DriverID                                  string   `json:"driver_id,omitempty"`
	DriverName                                string   `json:"driver_name,omitempty"`
	LastUpdatedByID                           string   `json:"last_updated_by_id,omitempty"`
	LastUpdatedByName                         string   `json:"last_updated_by_name,omitempty"`
	InboundProjectOfficeID                    string   `json:"inbound_project_office_id,omitempty"`
	InboundProjectOfficeName                  string   `json:"inbound_project_office_name,omitempty"`
	InboundProjectOfficeAbbreviation          string   `json:"inbound_project_office_abbreviation,omitempty"`
	InboundProjectOfficeExplicitID            string   `json:"inbound_project_office_explicit_id,omitempty"`
	InboundProjectOfficeExplicitName          string   `json:"inbound_project_office_explicit_name,omitempty"`
	InboundProjectOfficeExplicitAbbreviation  string   `json:"inbound_project_office_explicit_abbreviation,omitempty"`
	InboundProjectOfficeImplicitCachedID      string   `json:"inbound_project_office_implicit_cached_id,omitempty"`
	InboundProjectOfficeImplicitCachedName    string   `json:"inbound_project_office_implicit_cached_name,omitempty"`
	InboundProjectOfficeImplicitCachedAbbrev  string   `json:"inbound_project_office_implicit_cached_abbreviation,omitempty"`
	ProjectTransportPlanSegmentDriverIDs      []string `json:"project_transport_plan_segment_driver_ids,omitempty"`
	SegmentIDs                                []string `json:"segment_ids,omitempty"`
	ProjectTransportPlanDriverConfirmationIDs []string `json:"project_transport_plan_driver_confirmation_ids,omitempty"`
}

func newProjectTransportPlanDriversShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show project transport plan driver details",
		Long: `Show the full details of a project transport plan driver assignment.

Output Fields:
  ID, status, confirmation note and max confirmation time
  Window start/end (cached)
  Assignment rule overrides
  Project transport plan and broker
  Segment range (start/end)
  Driver and last updated by
  Inbound project office (resolved/explicit/implicit)
  Related segment drivers, segments, and confirmations

Arguments:
  <id>  The project transport plan driver ID (required).`,
		Example: `  # Show project transport plan driver details
  xbe view project-transport-plan-drivers show 123

  # Output as JSON
  xbe view project-transport-plan-drivers show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runProjectTransportPlanDriversShow,
	}
	initProjectTransportPlanDriversShowFlags(cmd)
	return cmd
}

func init() {
	projectTransportPlanDriversCmd.AddCommand(newProjectTransportPlanDriversShowCmd())
}

func initProjectTransportPlanDriversShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runProjectTransportPlanDriversShow(cmd *cobra.Command, args []string) error {
	if handled, err := maybeHandleClientURLShow(cmd, args); err != nil {
		return err
	} else if handled {
		return nil
	}

	opts, err := parseProjectTransportPlanDriversShowOptions(cmd)
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
		return fmt.Errorf("project transport plan driver id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[project-transport-plan-drivers]", "status,confirm-note,confirm-at-max,window-start-at-cached,window-end-at-cached,skip-assignment-rules-validation,assignment-rule-override-reason,project-transport-plan,segment-start,segment-end,driver,last-updated-by,inbound-project-office-implicit-cached,inbound-project-office-explicit,inbound-project-office,project-transport-plan-segment-drivers,segments,project-transport-plan-driver-confirmations")
	query.Set("include", "project-transport-plan,project-transport-plan.broker,segment-start,segment-end,driver,last-updated-by,inbound-project-office-implicit-cached,inbound-project-office-explicit,inbound-project-office")
	query.Set("fields[project-transport-plans]", "status,broker")
	query.Set("fields[project-transport-plan-segments]", "position")
	query.Set("fields[users]", "name")
	query.Set("fields[project-offices]", "name,abbreviation")
	query.Set("fields[brokers]", "company-name")

	body, _, err := client.Get(cmd.Context(), "/v1/project-transport-plan-drivers/"+id, query)
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

	details := buildProjectTransportPlanDriverDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderProjectTransportPlanDriverDetails(cmd, details)
}

func parseProjectTransportPlanDriversShowOptions(cmd *cobra.Command) (projectTransportPlanDriversShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return projectTransportPlanDriversShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildProjectTransportPlanDriverDetails(resp jsonAPISingleResponse) projectTransportPlanDriverDetails {
	included := make(map[string]jsonAPIResource, len(resp.Included))
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	attrs := resp.Data.Attributes
	details := projectTransportPlanDriverDetails{
		ID:                            resp.Data.ID,
		Status:                        stringAttr(attrs, "status"),
		ConfirmNote:                   stringAttr(attrs, "confirm-note"),
		ConfirmAtMax:                  formatDateTime(stringAttr(attrs, "confirm-at-max")),
		WindowStartAtCached:           formatDateTime(stringAttr(attrs, "window-start-at-cached")),
		WindowEndAtCached:             formatDateTime(stringAttr(attrs, "window-end-at-cached")),
		SkipAssignmentRulesValidation: boolAttr(attrs, "skip-assignment-rules-validation"),
		AssignmentRuleOverrideReason:  stringAttr(attrs, "assignment-rule-override-reason"),
	}

	if rel, ok := resp.Data.Relationships["project-transport-plan"]; ok && rel.Data != nil {
		details.ProjectTransportPlanID = rel.Data.ID
		if plan, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.ProjectTransportPlanStatus = stringAttr(plan.Attributes, "status")
			if brokerRel, ok := plan.Relationships["broker"]; ok && brokerRel.Data != nil {
				details.BrokerID = brokerRel.Data.ID
				if broker, ok := included[resourceKey(brokerRel.Data.Type, brokerRel.Data.ID)]; ok {
					details.BrokerName = stringAttr(broker.Attributes, "company-name")
				}
			}
		}
	}

	if rel, ok := resp.Data.Relationships["segment-start"]; ok && rel.Data != nil {
		details.SegmentStartID = rel.Data.ID
		if segment, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.SegmentStartPosition = stringAttr(segment.Attributes, "position")
		}
	}

	if rel, ok := resp.Data.Relationships["segment-end"]; ok && rel.Data != nil {
		details.SegmentEndID = rel.Data.ID
		if segment, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.SegmentEndPosition = stringAttr(segment.Attributes, "position")
		}
	}

	if rel, ok := resp.Data.Relationships["driver"]; ok && rel.Data != nil {
		details.DriverID = rel.Data.ID
		if driver, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.DriverName = stringAttr(driver.Attributes, "name")
		}
	}

	if rel, ok := resp.Data.Relationships["last-updated-by"]; ok && rel.Data != nil {
		details.LastUpdatedByID = rel.Data.ID
		if user, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.LastUpdatedByName = stringAttr(user.Attributes, "name")
		}
	}

	if rel, ok := resp.Data.Relationships["inbound-project-office"]; ok && rel.Data != nil {
		details.InboundProjectOfficeID = rel.Data.ID
		if office, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.InboundProjectOfficeName = stringAttr(office.Attributes, "name")
			details.InboundProjectOfficeAbbreviation = stringAttr(office.Attributes, "abbreviation")
		}
	}

	if rel, ok := resp.Data.Relationships["inbound-project-office-explicit"]; ok && rel.Data != nil {
		details.InboundProjectOfficeExplicitID = rel.Data.ID
		if office, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.InboundProjectOfficeExplicitName = stringAttr(office.Attributes, "name")
			details.InboundProjectOfficeExplicitAbbreviation = stringAttr(office.Attributes, "abbreviation")
		}
	}

	if rel, ok := resp.Data.Relationships["inbound-project-office-implicit-cached"]; ok && rel.Data != nil {
		details.InboundProjectOfficeImplicitCachedID = rel.Data.ID
		if office, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.InboundProjectOfficeImplicitCachedName = stringAttr(office.Attributes, "name")
			details.InboundProjectOfficeImplicitCachedAbbrev = stringAttr(office.Attributes, "abbreviation")
		}
	}

	if rel, ok := resp.Data.Relationships["project-transport-plan-segment-drivers"]; ok && rel.raw != nil {
		details.ProjectTransportPlanSegmentDriverIDs = extractRelationshipIDs(rel)
	}

	if rel, ok := resp.Data.Relationships["segments"]; ok && rel.raw != nil {
		details.SegmentIDs = extractRelationshipIDs(rel)
	}

	if rel, ok := resp.Data.Relationships["project-transport-plan-driver-confirmations"]; ok && rel.raw != nil {
		details.ProjectTransportPlanDriverConfirmationIDs = extractRelationshipIDs(rel)
	}

	return details
}

func renderProjectTransportPlanDriverDetails(cmd *cobra.Command, d projectTransportPlanDriverDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", d.ID)
	if d.Status != "" {
		fmt.Fprintf(out, "Status: %s\n", d.Status)
	}
	if d.WindowStartAtCached != "" {
		fmt.Fprintf(out, "Window Start (cached): %s\n", d.WindowStartAtCached)
	}
	if d.WindowEndAtCached != "" {
		fmt.Fprintf(out, "Window End (cached): %s\n", d.WindowEndAtCached)
	}
	if d.ConfirmNote != "" {
		fmt.Fprintf(out, "Confirm Note: %s\n", d.ConfirmNote)
	}
	if d.ConfirmAtMax != "" {
		fmt.Fprintf(out, "Confirm At Max: %s\n", d.ConfirmAtMax)
	}
	fmt.Fprintf(out, "Skip Assignment Rules Validation: %t\n", d.SkipAssignmentRulesValidation)
	if d.AssignmentRuleOverrideReason != "" {
		fmt.Fprintf(out, "Assignment Rule Override Reason: %s\n", d.AssignmentRuleOverrideReason)
	}

	if d.SegmentStartID != "" || d.SegmentStartPosition != "" {
		if d.SegmentStartID != "" && d.SegmentStartPosition != "" {
			fmt.Fprintf(out, "Segment Start: %s (ID: %s)\n", d.SegmentStartPosition, d.SegmentStartID)
		} else {
			fmt.Fprintf(out, "Segment Start: %s\n", firstNonEmpty(d.SegmentStartPosition, d.SegmentStartID))
		}
	}
	if d.SegmentEndID != "" || d.SegmentEndPosition != "" {
		if d.SegmentEndID != "" && d.SegmentEndPosition != "" {
			fmt.Fprintf(out, "Segment End: %s (ID: %s)\n", d.SegmentEndPosition, d.SegmentEndID)
		} else {
			fmt.Fprintf(out, "Segment End: %s\n", firstNonEmpty(d.SegmentEndPosition, d.SegmentEndID))
		}
	}

	if d.ProjectTransportPlanID != "" || d.DriverID != "" || d.LastUpdatedByID != "" || d.InboundProjectOfficeID != "" || d.InboundProjectOfficeExplicitID != "" || d.InboundProjectOfficeImplicitCachedID != "" || d.BrokerID != "" {
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "Relationships:")
		fmt.Fprintln(out, strings.Repeat("-", 40))
		if d.ProjectTransportPlanStatus != "" {
			fmt.Fprintf(out, "  Project Transport Plan: %s (ID: %s)\n", d.ProjectTransportPlanStatus, d.ProjectTransportPlanID)
		} else if d.ProjectTransportPlanID != "" {
			fmt.Fprintf(out, "  Project Transport Plan ID: %s\n", d.ProjectTransportPlanID)
		}
		if d.BrokerName != "" {
			fmt.Fprintf(out, "  Broker: %s (ID: %s)\n", d.BrokerName, d.BrokerID)
		} else if d.BrokerID != "" {
			fmt.Fprintf(out, "  Broker ID: %s\n", d.BrokerID)
		}
		if d.DriverName != "" {
			fmt.Fprintf(out, "  Driver: %s (ID: %s)\n", d.DriverName, d.DriverID)
		} else if d.DriverID != "" {
			fmt.Fprintf(out, "  Driver ID: %s\n", d.DriverID)
		}
		if d.LastUpdatedByName != "" {
			fmt.Fprintf(out, "  Last Updated By: %s (ID: %s)\n", d.LastUpdatedByName, d.LastUpdatedByID)
		} else if d.LastUpdatedByID != "" {
			fmt.Fprintf(out, "  Last Updated By ID: %s\n", d.LastUpdatedByID)
		}
		if d.InboundProjectOfficeName != "" {
			fmt.Fprintf(out, "  Inbound Project Office: %s (ID: %s)\n", d.InboundProjectOfficeName, d.InboundProjectOfficeID)
		} else if d.InboundProjectOfficeID != "" {
			fmt.Fprintf(out, "  Inbound Project Office ID: %s\n", d.InboundProjectOfficeID)
		}
		if d.InboundProjectOfficeExplicitName != "" {
			fmt.Fprintf(out, "  Inbound Project Office (Explicit): %s (ID: %s)\n", d.InboundProjectOfficeExplicitName, d.InboundProjectOfficeExplicitID)
		} else if d.InboundProjectOfficeExplicitID != "" {
			fmt.Fprintf(out, "  Inbound Project Office Explicit ID: %s\n", d.InboundProjectOfficeExplicitID)
		}
		if d.InboundProjectOfficeImplicitCachedName != "" {
			fmt.Fprintf(out, "  Inbound Project Office (Implicit): %s (ID: %s)\n", d.InboundProjectOfficeImplicitCachedName, d.InboundProjectOfficeImplicitCachedID)
		} else if d.InboundProjectOfficeImplicitCachedID != "" {
			fmt.Fprintf(out, "  Inbound Project Office Implicit ID: %s\n", d.InboundProjectOfficeImplicitCachedID)
		}
	}

	if len(d.ProjectTransportPlanSegmentDriverIDs) > 0 || len(d.SegmentIDs) > 0 || len(d.ProjectTransportPlanDriverConfirmationIDs) > 0 {
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "Related IDs:")
		fmt.Fprintln(out, strings.Repeat("-", 40))
		if len(d.ProjectTransportPlanSegmentDriverIDs) > 0 {
			fmt.Fprintf(out, "  Segment Driver IDs: %s\n", strings.Join(d.ProjectTransportPlanSegmentDriverIDs, ", "))
		}
		if len(d.SegmentIDs) > 0 {
			fmt.Fprintf(out, "  Segment IDs: %s\n", strings.Join(d.SegmentIDs, ", "))
		}
		if len(d.ProjectTransportPlanDriverConfirmationIDs) > 0 {
			fmt.Fprintf(out, "  Driver Confirmation IDs: %s\n", strings.Join(d.ProjectTransportPlanDriverConfirmationIDs, ", "))
		}
	}

	return nil
}
