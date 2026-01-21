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

type maintenanceRulesShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type ruleDetails struct {
	ID              string `json:"id"`
	Name            string `json:"name,omitempty"`
	MaintenanceType string `json:"maintenance_type,omitempty"`
	IsActive        bool   `json:"is_active"`
	CreatedAt       string `json:"created_at,omitempty"`
	UpdatedAt       string `json:"updated_at,omitempty"`

	// Schedule
	IntervalDays         int    `json:"interval_days,omitempty"`
	IntervalHours        int    `json:"interval_hours,omitempty"`
	IntervalMiles        int    `json:"interval_miles,omitempty"`
	WarningDays          int    `json:"warning_days,omitempty"`
	WarningHours         int    `json:"warning_hours,omitempty"`
	WarningMiles         int    `json:"warning_miles,omitempty"`
	ScheduleType         string `json:"schedule_type,omitempty"`
	LastServiceThreshold int    `json:"last_service_threshold,omitempty"`

	// Scope
	ScopeLevel                string `json:"scope_level,omitempty"`
	Scope                     string `json:"scope,omitempty"`
	BusinessUnitID            string `json:"business_unit_id,omitempty"`
	BusinessUnit              string `json:"business_unit,omitempty"`
	EquipmentID               string `json:"equipment_id,omitempty"`
	Equipment                 string `json:"equipment,omitempty"`
	EquipmentClassificationID string `json:"equipment_classification_id,omitempty"`
	EquipmentClassification   string `json:"equipment_classification,omitempty"`

	// Description
	Description string `json:"description,omitempty"`
}

func newMaintenanceRulesShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show maintenance rule details",
		Long: `Show the full details of a maintenance requirement rule.

Retrieves and displays comprehensive information about a rule including
schedule configuration, scope, and related equipment.

Output Sections (table format):
  Core Info       ID, name, maintenance type, active status
  Schedule        Interval settings (days, hours, miles)
  Warning         Warning thresholds
  Scope           Business unit, equipment classification
  Description     Full description

Arguments:
  <id>          The rule ID (required). Find IDs using the list command.`,
		Example: `  # View a rule by ID
  xbe view maintenance rules show 123

  # Get rule as JSON
  xbe view maintenance rules show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runMaintenanceRulesShow,
	}
	initMaintenanceRulesShowFlags(cmd)
	return cmd
}

func init() {
	maintenanceRulesCmd.AddCommand(newMaintenanceRulesShowCmd())
}

func initMaintenanceRulesShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runMaintenanceRulesShow(cmd *cobra.Command, args []string) error {
	opts, err := parseMaintenanceRulesShowOptions(cmd)
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
		return fmt.Errorf("rule id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("include", "equipment,equipment-classification,business-unit")

	body, _, err := client.Get(cmd.Context(), "/v1/maintenance-requirement-rules/"+id, query)
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

	details := buildRuleDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderRuleDetails(cmd, details)
}

func parseMaintenanceRulesShowOptions(cmd *cobra.Command) (maintenanceRulesShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return maintenanceRulesShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildRuleDetails(resp jsonAPISingleResponse) ruleDetails {
	attrs := resp.Data.Attributes

	// Build included lookup
	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		key := resourceKey(inc.Type, inc.ID)
		included[key] = inc
	}

	// The API uses "rule" attribute for the rule name
	name := strings.TrimSpace(stringAttr(attrs, "rule"))
	if name == "" {
		name = strings.TrimSpace(stringAttr(attrs, "name"))
	}

	details := ruleDetails{
		ID:                   resp.Data.ID,
		Name:                 name,
		MaintenanceType:      stringAttr(attrs, "maintenance-type"),
		IsActive:             boolAttr(attrs, "is-active"),
		CreatedAt:            formatDate(stringAttr(attrs, "created-at")),
		UpdatedAt:            formatDate(stringAttr(attrs, "updated-at")),
		IntervalDays:         intAttr(attrs, "interval-days"),
		IntervalHours:        intAttr(attrs, "interval-hours"),
		IntervalMiles:        intAttr(attrs, "interval-miles"),
		WarningDays:          intAttr(attrs, "warning-days"),
		WarningHours:         intAttr(attrs, "warning-hours"),
		WarningMiles:         intAttr(attrs, "warning-miles"),
		ScheduleType:         stringAttr(attrs, "schedule-type"),
		LastServiceThreshold: intAttr(attrs, "last-service-threshold"),
		Description:          strings.TrimSpace(stringAttr(attrs, "description")),
	}

	// Equipment
	if rel, ok := resp.Data.Relationships["equipment"]; ok && rel.Data != nil {
		details.EquipmentID = rel.Data.ID
		key := resourceKey(rel.Data.Type, rel.Data.ID)
		if inc, ok := included[key]; ok {
			details.Equipment = firstNonEmpty(
				stringAttr(inc.Attributes, "name"),
				stringAttr(inc.Attributes, "equipment-number"),
				rel.Data.ID,
			)
		}
	}

	// Equipment classification
	if rel, ok := resp.Data.Relationships["equipment-classification"]; ok && rel.Data != nil {
		details.EquipmentClassificationID = rel.Data.ID
		key := resourceKey(rel.Data.Type, rel.Data.ID)
		if inc, ok := included[key]; ok {
			details.EquipmentClassification = stringAttr(inc.Attributes, "name")
		}
	}

	// Business unit
	if rel, ok := resp.Data.Relationships["business-unit"]; ok && rel.Data != nil {
		details.BusinessUnitID = rel.Data.ID
		key := resourceKey(rel.Data.Type, rel.Data.ID)
		if inc, ok := included[key]; ok {
			details.BusinessUnit = firstNonEmpty(
				stringAttr(inc.Attributes, "company-name"),
				stringAttr(inc.Attributes, "name"),
				rel.Data.ID,
			)
		}
	}

	// Compute scope level and display
	details.ScopeLevel, details.Scope = getRuleScopeInfo(
		details.EquipmentID, details.Equipment,
		details.EquipmentClassificationID, details.EquipmentClassification,
		details.BusinessUnitID, details.BusinessUnit,
	)

	return details
}

func renderRuleDetails(cmd *cobra.Command, d ruleDetails) error {
	out := cmd.OutOrStdout()

	// Core info
	fmt.Fprintf(out, "ID: %s\n", d.ID)
	if d.Name != "" {
		fmt.Fprintf(out, "Name: %s\n", d.Name)
	}
	if d.MaintenanceType != "" {
		fmt.Fprintf(out, "Maintenance Type: %s\n", d.MaintenanceType)
	}
	if d.IsActive {
		fmt.Fprintln(out, "Active: true")
	} else {
		fmt.Fprintln(out, "Active: false")
	}
	if d.CreatedAt != "" {
		fmt.Fprintf(out, "Created: %s\n", d.CreatedAt)
	}
	if d.UpdatedAt != "" {
		fmt.Fprintf(out, "Updated: %s\n", d.UpdatedAt)
	}

	// Schedule
	if d.IntervalDays > 0 || d.IntervalHours > 0 || d.IntervalMiles > 0 || d.ScheduleType != "" {
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "Schedule:")
		fmt.Fprintln(out, strings.Repeat("-", 40))
		if d.ScheduleType != "" {
			fmt.Fprintf(out, "  Type: %s\n", d.ScheduleType)
		}
		if d.IntervalDays > 0 {
			fmt.Fprintf(out, "  Interval (Days): %d\n", d.IntervalDays)
		}
		if d.IntervalHours > 0 {
			fmt.Fprintf(out, "  Interval (Hours): %d\n", d.IntervalHours)
		}
		if d.IntervalMiles > 0 {
			fmt.Fprintf(out, "  Interval (Miles): %d\n", d.IntervalMiles)
		}
		if d.LastServiceThreshold > 0 {
			fmt.Fprintf(out, "  Last Service Threshold: %d\n", d.LastServiceThreshold)
		}
	}

	// Warning thresholds
	if d.WarningDays > 0 || d.WarningHours > 0 || d.WarningMiles > 0 {
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "Warning Thresholds:")
		fmt.Fprintln(out, strings.Repeat("-", 40))
		if d.WarningDays > 0 {
			fmt.Fprintf(out, "  Days: %d\n", d.WarningDays)
		}
		if d.WarningHours > 0 {
			fmt.Fprintf(out, "  Hours: %d\n", d.WarningHours)
		}
		if d.WarningMiles > 0 {
			fmt.Fprintf(out, "  Miles: %d\n", d.WarningMiles)
		}
	}

	// Scope
	fmt.Fprintln(out, "")
	fmt.Fprintln(out, "Scope:")
	fmt.Fprintln(out, strings.Repeat("-", 40))
	fmt.Fprintf(out, "  Level: %s\n", d.Scope)
	if d.Equipment != "" {
		fmt.Fprintf(out, "  Equipment: %s (ID: %s)\n", d.Equipment, d.EquipmentID)
	}
	if d.EquipmentClassification != "" {
		fmt.Fprintf(out, "  Equipment Classification: %s\n", d.EquipmentClassification)
	}
	if d.BusinessUnit != "" {
		fmt.Fprintf(out, "  Business Unit: %s (ID: %s)\n", d.BusinessUnit, d.BusinessUnitID)
	}

	// Description
	if d.Description != "" {
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "Description:")
		fmt.Fprintln(out, strings.Repeat("-", 40))
		fmt.Fprintln(out, d.Description)
	}

	return nil
}
