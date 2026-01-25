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

type membershipsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type membershipDetails struct {
	ID                                   string `json:"id"`
	Type                                 string `json:"type"`
	UserID                               string `json:"user_id"`
	UserName                             string `json:"user_name"`
	UserEmail                            string `json:"user_email,omitempty"`
	UserMobile                           string `json:"user_mobile,omitempty"`
	OrganizationType                     string `json:"organization_type"`
	OrganizationID                       string `json:"organization_id"`
	OrganizationName                     string `json:"organization_name"`
	BrokerID                             string `json:"broker_id,omitempty"`
	BrokerName                           string `json:"broker_name,omitempty"`
	ProjectOfficeID                      string `json:"project_office_id,omitempty"`
	ProjectOfficeName                    string `json:"project_office_name,omitempty"`
	Kind                                 string `json:"kind"`
	IsAdmin                              bool   `json:"is_admin"`
	Title                                string `json:"title,omitempty"`
	ColorHex                             string `json:"color_hex,omitempty"`
	ExternalEmployeeID                   string `json:"external_employee_id,omitempty"`
	ExplicitSortOrder                    *int   `json:"explicit_sort_order,omitempty"`
	StartAt                              string `json:"start_at,omitempty"`
	EndAt                                string `json:"end_at,omitempty"`
	DrivesShiftType                      string `json:"drives_shift_type,omitempty"`
	TrailerCoassignmentsResetOn          string `json:"trailer_coassignments_reset_on,omitempty"`
	CanSeeRatesAsDriver                  bool   `json:"can_see_rates_as_driver"`
	CanSeeRatesAsManager                 bool   `json:"can_see_rates_as_manager"`
	CanValidateProfitImprovements        bool   `json:"can_validate_profit_improvements"`
	IsRateEditor                         bool   `json:"is_rate_editor"`
	IsTimeCardAuditor                    bool   `json:"is_time_card_auditor"`
	IsEquipmentRentalTeamMember          bool   `json:"is_equipment_rental_team_member"`
	IsGeofenceViolationTeamMember        bool   `json:"is_geofence_violation_team_member"`
	IsUnapprovedTimeCardSubscriber       bool   `json:"is_unapproved_time_card_subscriber"`
	IsDefaultJobProductionPlanSubscriber bool   `json:"is_default_job_production_plan_subscriber"`
	EnableRecapNotifications             bool   `json:"enable_recap_notifications"`
	EnableInventoryCapacityNotifications bool   `json:"enable_inventory_capacity_notifications"`
}

func newMembershipsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show membership details",
		Long: `Show the full details of a specific membership.

Retrieves and displays comprehensive information about a membership including
user information, organization, role settings, and all configuration options.

Output Fields:
  ID                        Unique membership identifier
  Type                      Membership type (BrokerMembership, TruckerMembership, etc.)
  User                      User name, email, and mobile
  Organization              Organization name and type
  Broker                    Associated broker/branch
  Project Office            Associated project office (if any)
  Kind                      Role type (operations/manager)
  Is Admin                  Whether user is an admin in this organization
  Title                     User's title within the organization
  Color Hex                 Display color for the user
  External Employee ID      External system employee identifier
  Explicit Sort Order       Manual sort order override
  Start At / End At         Membership effective dates
  Drives Shift Type         Shift type preference (any/day/night)

  Permission Flags:
    Can See Rates As Driver
    Can See Rates As Manager
    Can Validate Profit Improvements
    Is Rate Editor
    Is Time Card Auditor
    Is Equipment Rental Team Member
    Is Geofence Violation Team Member

  Notification Subscriptions:
    Is Unapproved Time Card Subscriber
    Is Default Job Production Plan Subscriber
    Enable Recap Notifications
    Enable Inventory Capacity Notifications

Arguments:
  <id>    The membership ID (required). You can find IDs using the list command.`,
		Example: `  # View a membership by ID
  xbe view memberships show 686

  # Get membership as JSON
  xbe view memberships show 686 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runMembershipsShow,
	}
	initMembershipsShowFlags(cmd)
	return cmd
}

func init() {
	membershipsCmd.AddCommand(newMembershipsShowCmd())
}

func initMembershipsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runMembershipsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseMembershipsShowOptions(cmd)
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
		return fmt.Errorf("membership id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("include", "user,organization,broker,project-office")
	query.Set("fields[users]", "name,email-address,mobile-number")
	query.Set("fields[brokers]", "company-name")
	query.Set("fields[customers]", "company-name")
	query.Set("fields[truckers]", "company-name")
	query.Set("fields[material-suppliers]", "name")
	query.Set("fields[developers]", "name")
	query.Set("fields[project-offices]", "name")

	body, _, err := client.Get(cmd.Context(), "/v1/memberships/"+id, query)
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

	details := buildMembershipDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderMembershipDetails(cmd, details)
}

func parseMembershipsShowOptions(cmd *cobra.Command) (membershipsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return membershipsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildMembershipDetails(resp jsonAPISingleResponse) membershipDetails {
	attrs := resp.Data.Attributes

	// Build included map for lookups
	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		key := resourceKey(inc.Type, inc.ID)
		included[key] = inc
	}

	details := membershipDetails{
		ID:                                   resp.Data.ID,
		Type:                                 resp.Data.Type,
		Kind:                                 stringAttr(attrs, "kind"),
		IsAdmin:                              boolAttr(attrs, "is-admin"),
		Title:                                strings.TrimSpace(stringAttr(attrs, "title")),
		ColorHex:                             strings.TrimSpace(stringAttr(attrs, "color-hex")),
		ExternalEmployeeID:                   strings.TrimSpace(stringAttr(attrs, "external-employee-id")),
		StartAt:                              stringAttr(attrs, "start-at"),
		EndAt:                                stringAttr(attrs, "end-at"),
		DrivesShiftType:                      stringAttr(attrs, "drives-shift-type"),
		TrailerCoassignmentsResetOn:          stringAttr(attrs, "trailer-coassignments-reset-on"),
		CanSeeRatesAsDriver:                  boolAttr(attrs, "can-see-rates-as-driver"),
		CanSeeRatesAsManager:                 boolAttr(attrs, "can-see-rates-as-manager"),
		CanValidateProfitImprovements:        boolAttr(attrs, "can-validate-profit-improvements"),
		IsRateEditor:                         boolAttr(attrs, "is-rate-editor"),
		IsTimeCardAuditor:                    boolAttr(attrs, "is-time-card-auditor"),
		IsEquipmentRentalTeamMember:          boolAttr(attrs, "is-equipment-rental-team-member"),
		IsGeofenceViolationTeamMember:        boolAttr(attrs, "is-geofence-violation-team-member"),
		IsUnapprovedTimeCardSubscriber:       boolAttr(attrs, "is-unapproved-time-card-subscriber"),
		IsDefaultJobProductionPlanSubscriber: boolAttr(attrs, "is-default-job-production-plan-subscriber"),
		EnableRecapNotifications:             boolAttr(attrs, "enable-recap-notifications"),
		EnableInventoryCapacityNotifications: boolAttr(attrs, "enable-inventory-capacity-notifications"),
	}

	// Handle explicit_sort_order (nullable int)
	if v, ok := attrs["explicit-sort-order"]; ok && v != nil {
		if f, ok := v.(float64); ok {
			i := int(f)
			details.ExplicitSortOrder = &i
		}
	}

	// Get user info
	if rel, ok := resp.Data.Relationships["user"]; ok && rel.Data != nil {
		details.UserID = rel.Data.ID
		key := resourceKey(rel.Data.Type, rel.Data.ID)
		if inc, ok := included[key]; ok {
			details.UserName = strings.TrimSpace(stringAttr(inc.Attributes, "name"))
			details.UserEmail = strings.TrimSpace(stringAttr(inc.Attributes, "email-address"))
			details.UserMobile = strings.TrimSpace(stringAttr(inc.Attributes, "mobile-number"))
		}
	}

	// Get organization info (polymorphic)
	if rel, ok := resp.Data.Relationships["organization"]; ok && rel.Data != nil {
		details.OrganizationType = rel.Data.Type
		details.OrganizationID = rel.Data.ID
		key := resourceKey(rel.Data.Type, rel.Data.ID)
		if inc, ok := included[key]; ok {
			details.OrganizationName = firstNonEmpty(
				stringAttr(inc.Attributes, "company-name"),
				stringAttr(inc.Attributes, "name"),
			)
		}
	}

	// Get broker info
	if rel, ok := resp.Data.Relationships["broker"]; ok && rel.Data != nil {
		details.BrokerID = rel.Data.ID
		key := resourceKey(rel.Data.Type, rel.Data.ID)
		if inc, ok := included[key]; ok {
			details.BrokerName = strings.TrimSpace(stringAttr(inc.Attributes, "company-name"))
		}
	}

	// Get project office info
	if rel, ok := resp.Data.Relationships["project-office"]; ok && rel.Data != nil {
		details.ProjectOfficeID = rel.Data.ID
		key := resourceKey(rel.Data.Type, rel.Data.ID)
		if inc, ok := included[key]; ok {
			details.ProjectOfficeName = strings.TrimSpace(stringAttr(inc.Attributes, "name"))
		}
	}

	return details
}

func renderMembershipDetails(cmd *cobra.Command, d membershipDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", d.ID)
	fmt.Fprintf(out, "Type: %s\n", d.Type)
	fmt.Fprintln(out, "")

	// User section
	fmt.Fprintln(out, "User:")
	fmt.Fprintf(out, "  ID: %s\n", d.UserID)
	fmt.Fprintf(out, "  Name: %s\n", d.UserName)
	if d.UserEmail != "" {
		fmt.Fprintf(out, "  Email: %s\n", d.UserEmail)
	}
	if d.UserMobile != "" {
		fmt.Fprintf(out, "  Mobile: %s\n", d.UserMobile)
	}
	fmt.Fprintln(out, "")

	// Organization section
	fmt.Fprintln(out, "Organization:")
	fmt.Fprintf(out, "  Type: %s\n", d.OrganizationType)
	fmt.Fprintf(out, "  ID: %s\n", d.OrganizationID)
	fmt.Fprintf(out, "  Name: %s\n", d.OrganizationName)
	fmt.Fprintln(out, "")

	// Broker section
	if d.BrokerID != "" {
		fmt.Fprintln(out, "Broker:")
		fmt.Fprintf(out, "  ID: %s\n", d.BrokerID)
		fmt.Fprintf(out, "  Name: %s\n", d.BrokerName)
		fmt.Fprintln(out, "")
	}

	// Project Office section
	if d.ProjectOfficeID != "" {
		fmt.Fprintln(out, "Project Office:")
		fmt.Fprintf(out, "  ID: %s\n", d.ProjectOfficeID)
		fmt.Fprintf(out, "  Name: %s\n", d.ProjectOfficeName)
		fmt.Fprintln(out, "")
	}

	// Role section
	fmt.Fprintln(out, "Role:")
	fmt.Fprintf(out, "  Kind: %s\n", d.Kind)
	fmt.Fprintf(out, "  Is Admin: %s\n", formatBool(d.IsAdmin))
	if d.Title != "" {
		fmt.Fprintf(out, "  Title: %s\n", d.Title)
	}
	if d.ColorHex != "" {
		fmt.Fprintf(out, "  Color: %s\n", d.ColorHex)
	}
	if d.ExternalEmployeeID != "" {
		fmt.Fprintf(out, "  External Employee ID: %s\n", d.ExternalEmployeeID)
	}
	if d.ExplicitSortOrder != nil {
		fmt.Fprintf(out, "  Explicit Sort Order: %d\n", *d.ExplicitSortOrder)
	}
	if d.DrivesShiftType != "" {
		fmt.Fprintf(out, "  Drives Shift Type: %s\n", d.DrivesShiftType)
	}
	if d.TrailerCoassignmentsResetOn != "" {
		fmt.Fprintf(out, "  Trailer Coassignments Reset On: %s\n", formatDate(d.TrailerCoassignmentsResetOn))
	}
	fmt.Fprintln(out, "")

	// Effective dates
	if d.StartAt != "" || d.EndAt != "" {
		fmt.Fprintln(out, "Effective Period:")
		if d.StartAt != "" {
			fmt.Fprintf(out, "  Start: %s\n", formatDate(d.StartAt))
		}
		if d.EndAt != "" {
			fmt.Fprintf(out, "  End: %s\n", formatDate(d.EndAt))
		}
		fmt.Fprintln(out, "")
	}

	// Permissions section
	fmt.Fprintln(out, "Permissions:")
	fmt.Fprintf(out, "  Can See Rates As Driver: %s\n", formatBool(d.CanSeeRatesAsDriver))
	fmt.Fprintf(out, "  Can See Rates As Manager: %s\n", formatBool(d.CanSeeRatesAsManager))
	fmt.Fprintf(out, "  Can Validate Profit Improvements: %s\n", formatBool(d.CanValidateProfitImprovements))
	fmt.Fprintf(out, "  Is Rate Editor: %s\n", formatBool(d.IsRateEditor))
	fmt.Fprintf(out, "  Is Time Card Auditor: %s\n", formatBool(d.IsTimeCardAuditor))
	fmt.Fprintf(out, "  Is Equipment Rental Team Member: %s\n", formatBool(d.IsEquipmentRentalTeamMember))
	fmt.Fprintf(out, "  Is Geofence Violation Team Member: %s\n", formatBool(d.IsGeofenceViolationTeamMember))
	fmt.Fprintln(out, "")

	// Notifications section
	fmt.Fprintln(out, "Notifications:")
	fmt.Fprintf(out, "  Unapproved Time Card Subscriber: %s\n", formatBool(d.IsUnapprovedTimeCardSubscriber))
	fmt.Fprintf(out, "  Default Job Production Plan Subscriber: %s\n", formatBool(d.IsDefaultJobProductionPlanSubscriber))
	fmt.Fprintf(out, "  Recap Notifications: %s\n", formatBool(d.EnableRecapNotifications))
	fmt.Fprintf(out, "  Inventory Capacity Notifications: %s\n", formatBool(d.EnableInventoryCapacityNotifications))

	return nil
}

func formatBool(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}
