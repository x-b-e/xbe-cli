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

type brokerMembershipsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type brokerMembershipBusinessUnit struct {
	ID   string `json:"id"`
	Name string `json:"name,omitempty"`
}

type brokerMembershipDetails struct {
	ID                                   string                         `json:"id"`
	UserID                               string                         `json:"user_id"`
	UserName                             string                         `json:"user_name"`
	UserEmail                            string                         `json:"user_email,omitempty"`
	UserMobile                           string                         `json:"user_mobile,omitempty"`
	BrokerID                             string                         `json:"broker_id"`
	BrokerName                           string                         `json:"broker_name"`
	ProjectOfficeID                      string                         `json:"project_office_id,omitempty"`
	ProjectOfficeName                    string                         `json:"project_office_name,omitempty"`
	BusinessUnits                        []brokerMembershipBusinessUnit `json:"business_units,omitempty"`
	Kind                                 string                         `json:"kind"`
	IsAdmin                              bool                           `json:"is_admin"`
	Title                                string                         `json:"title,omitempty"`
	ColorHex                             string                         `json:"color_hex,omitempty"`
	ExternalEmployeeID                   string                         `json:"external_employee_id,omitempty"`
	ExplicitSortOrder                    *int                           `json:"explicit_sort_order,omitempty"`
	StartAt                              string                         `json:"start_at,omitempty"`
	EndAt                                string                         `json:"end_at,omitempty"`
	DrivesShiftType                      string                         `json:"drives_shift_type,omitempty"`
	CanSeeRatesAsDriver                  bool                           `json:"can_see_rates_as_driver"`
	CanSeeRatesAsManager                 bool                           `json:"can_see_rates_as_manager"`
	CanValidateProfitImprovements        bool                           `json:"can_validate_profit_improvements"`
	IsRateEditor                         bool                           `json:"is_rate_editor"`
	IsTimeCardAuditor                    bool                           `json:"is_time_card_auditor"`
	IsEquipmentRentalTeamMember          bool                           `json:"is_equipment_rental_team_member"`
	IsGeofenceViolationTeamMember        bool                           `json:"is_geofence_violation_team_member"`
	IsUnapprovedTimeCardSubscriber       bool                           `json:"is_unapproved_time_card_subscriber"`
	IsDefaultJobProductionPlanSubscriber bool                           `json:"is_default_job_production_plan_subscriber"`
	EnableRecapNotifications             bool                           `json:"enable_recap_notifications"`
	EnableInventoryCapacityNotifications bool                           `json:"enable_inventory_capacity_notifications"`
}

func newBrokerMembershipsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show broker membership details",
		Long: `Show the full details of a broker membership.

Includes user information, broker relationship, role settings, and notification
preferences for the membership.

Output Fields:
  ID
  User
  Broker
  Project Office
  Business Units
  Role and permissions
  Notification preferences

Arguments:
  <id>    The broker membership ID (required). You can find IDs using the list command.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # View a broker membership
  xbe view broker-memberships show 686

  # Get broker membership as JSON
  xbe view broker-memberships show 686 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runBrokerMembershipsShow,
	}
	initBrokerMembershipsShowFlags(cmd)
	return cmd
}

func init() {
	brokerMembershipsCmd.AddCommand(newBrokerMembershipsShowCmd())
}

func initBrokerMembershipsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runBrokerMembershipsShow(cmd *cobra.Command, args []string) error {
	if handled, err := maybeHandleClientURLShow(cmd, args); err != nil {
		return err
	} else if handled {
		return nil
	}

	opts, err := parseBrokerMembershipsShowOptions(cmd)
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
		return fmt.Errorf("broker membership id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("include", "user,organization,broker,project-office,business-units")
	query.Set("fields[users]", "name,email-address,mobile-number")
	query.Set("fields[brokers]", "company-name")
	query.Set("fields[project-offices]", "name")
	query.Set("fields[business-units]", "company-name")

	body, _, err := client.Get(cmd.Context(), "/v1/broker-memberships/"+id, query)
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

	details := buildBrokerMembershipDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderBrokerMembershipDetails(cmd, details)
}

func parseBrokerMembershipsShowOptions(cmd *cobra.Command) (brokerMembershipsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return brokerMembershipsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildBrokerMembershipDetails(resp jsonAPISingleResponse) brokerMembershipDetails {
	resource := resp.Data
	attrs := resource.Attributes

	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		key := resourceKey(inc.Type, inc.ID)
		included[key] = inc
	}

	details := brokerMembershipDetails{
		ID:                                   resource.ID,
		Kind:                                 stringAttr(attrs, "kind"),
		IsAdmin:                              boolAttr(attrs, "is-admin"),
		Title:                                strings.TrimSpace(stringAttr(attrs, "title")),
		ColorHex:                             strings.TrimSpace(stringAttr(attrs, "color-hex")),
		ExternalEmployeeID:                   strings.TrimSpace(stringAttr(attrs, "external-employee-id")),
		StartAt:                              stringAttr(attrs, "start-at"),
		EndAt:                                stringAttr(attrs, "end-at"),
		DrivesShiftType:                      stringAttr(attrs, "drives-shift-type"),
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

	if v, ok := attrs["explicit-sort-order"]; ok && v != nil {
		if f, ok := v.(float64); ok {
			i := int(f)
			details.ExplicitSortOrder = &i
		}
	}

	if rel, ok := resource.Relationships["user"]; ok && rel.Data != nil {
		details.UserID = rel.Data.ID
		key := resourceKey(rel.Data.Type, rel.Data.ID)
		if inc, ok := included[key]; ok {
			details.UserName = strings.TrimSpace(stringAttr(inc.Attributes, "name"))
			details.UserEmail = strings.TrimSpace(stringAttr(inc.Attributes, "email-address"))
			details.UserMobile = strings.TrimSpace(stringAttr(inc.Attributes, "mobile-number"))
		}
	}

	if rel, ok := resource.Relationships["broker"]; ok && rel.Data != nil {
		details.BrokerID = rel.Data.ID
		key := resourceKey(rel.Data.Type, rel.Data.ID)
		if inc, ok := included[key]; ok {
			details.BrokerName = strings.TrimSpace(stringAttr(inc.Attributes, "company-name"))
		}
	}
	if details.BrokerID == "" {
		if rel, ok := resource.Relationships["organization"]; ok && rel.Data != nil {
			details.BrokerID = rel.Data.ID
			key := resourceKey(rel.Data.Type, rel.Data.ID)
			if inc, ok := included[key]; ok {
				details.BrokerName = firstNonEmpty(
					stringAttr(inc.Attributes, "company-name"),
					stringAttr(inc.Attributes, "name"),
				)
			}
		}
	}

	if rel, ok := resource.Relationships["project-office"]; ok && rel.Data != nil {
		details.ProjectOfficeID = rel.Data.ID
		key := resourceKey(rel.Data.Type, rel.Data.ID)
		if inc, ok := included[key]; ok {
			details.ProjectOfficeName = strings.TrimSpace(stringAttr(inc.Attributes, "name"))
		}
	}

	if rel, ok := resource.Relationships["business-units"]; ok {
		for _, ref := range relationshipIDs(rel) {
			unit := brokerMembershipBusinessUnit{ID: ref.ID}
			key := resourceKey(ref.Type, ref.ID)
			if inc, ok := included[key]; ok {
				unit.Name = firstNonEmpty(
					stringAttr(inc.Attributes, "company-name"),
					stringAttr(inc.Attributes, "name"),
				)
			}
			details.BusinessUnits = append(details.BusinessUnits, unit)
		}
	}

	return details
}

func renderBrokerMembershipDetails(cmd *cobra.Command, d brokerMembershipDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", d.ID)
	fmt.Fprintln(out, "")

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

	fmt.Fprintln(out, "Broker:")
	fmt.Fprintf(out, "  ID: %s\n", d.BrokerID)
	fmt.Fprintf(out, "  Name: %s\n", d.BrokerName)
	fmt.Fprintln(out, "")

	if d.ProjectOfficeID != "" {
		fmt.Fprintln(out, "Project Office:")
		fmt.Fprintf(out, "  ID: %s\n", d.ProjectOfficeID)
		fmt.Fprintf(out, "  Name: %s\n", d.ProjectOfficeName)
		fmt.Fprintln(out, "")
	}

	if len(d.BusinessUnits) > 0 {
		fmt.Fprintln(out, "Business Units:")
		for _, unit := range d.BusinessUnits {
			label := unit.ID
			if unit.Name != "" {
				label = fmt.Sprintf("%s (%s)", unit.Name, unit.ID)
			}
			fmt.Fprintf(out, "  %s\n", label)
		}
		fmt.Fprintln(out, "")
	}

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
	fmt.Fprintln(out, "")

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

	fmt.Fprintln(out, "Permissions:")
	fmt.Fprintf(out, "  Can See Rates As Driver: %s\n", formatBool(d.CanSeeRatesAsDriver))
	fmt.Fprintf(out, "  Can See Rates As Manager: %s\n", formatBool(d.CanSeeRatesAsManager))
	fmt.Fprintf(out, "  Can Validate Profit Improvements: %s\n", formatBool(d.CanValidateProfitImprovements))
	fmt.Fprintf(out, "  Is Rate Editor: %s\n", formatBool(d.IsRateEditor))
	fmt.Fprintf(out, "  Is Time Card Auditor: %s\n", formatBool(d.IsTimeCardAuditor))
	fmt.Fprintf(out, "  Is Equipment Rental Team Member: %s\n", formatBool(d.IsEquipmentRentalTeamMember))
	fmt.Fprintf(out, "  Is Geofence Violation Team Member: %s\n", formatBool(d.IsGeofenceViolationTeamMember))
	fmt.Fprintln(out, "")

	fmt.Fprintln(out, "Notifications:")
	fmt.Fprintf(out, "  Unapproved Time Card Subscriber: %s\n", formatBool(d.IsUnapprovedTimeCardSubscriber))
	fmt.Fprintf(out, "  Default Job Production Plan Subscriber: %s\n", formatBool(d.IsDefaultJobProductionPlanSubscriber))
	fmt.Fprintf(out, "  Recap Notifications: %s\n", formatBool(d.EnableRecapNotifications))
	fmt.Fprintf(out, "  Inventory Capacity Notifications: %s\n", formatBool(d.EnableInventoryCapacityNotifications))

	return nil
}
