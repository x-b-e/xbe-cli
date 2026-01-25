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

type developerMembershipsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type developerMembershipDetails struct {
	ID                                   string `json:"id"`
	UserID                               string `json:"user_id"`
	UserName                             string `json:"user_name"`
	UserEmail                            string `json:"user_email,omitempty"`
	UserMobile                           string `json:"user_mobile,omitempty"`
	DeveloperID                          string `json:"developer_id"`
	DeveloperName                        string `json:"developer_name"`
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
	CanSeeRatesAsManager                 bool   `json:"can_see_rates_as_manager"`
	CanValidateProfitImprovements        bool   `json:"can_validate_profit_improvements"`
	IsGeofenceViolationTeamMember        bool   `json:"is_geofence_violation_team_member"`
	IsUnapprovedTimeCardSubscriber       bool   `json:"is_unapproved_time_card_subscriber"`
	IsDefaultJobProductionPlanSubscriber bool   `json:"is_default_job_production_plan_subscriber"`
	EnableInventoryCapacityNotifications bool   `json:"enable_inventory_capacity_notifications"`
}

func newDeveloperMembershipsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show developer membership details",
		Long: `Show the full details of a developer membership.

Includes user information, developer relationship, role settings, and
notification preferences for the membership.

Output Fields:
  ID
  User
  Developer
  Broker
  Project Office
  Role and permissions
  Notification preferences

Arguments:
  <id>    The developer membership ID (required). You can find IDs using the list command.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # View a developer membership
  xbe view developer-memberships show 686

  # Get developer membership as JSON
  xbe view developer-memberships show 686 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runDeveloperMembershipsShow,
	}
	initDeveloperMembershipsShowFlags(cmd)
	return cmd
}

func init() {
	developerMembershipsCmd.AddCommand(newDeveloperMembershipsShowCmd())
}

func initDeveloperMembershipsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDeveloperMembershipsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseDeveloperMembershipsShowOptions(cmd)
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
		return fmt.Errorf("developer membership id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("include", "user,organization,broker,project-office")
	query.Set("fields[users]", "name,email-address,mobile-number")
	query.Set("fields[developers]", "name")
	query.Set("fields[brokers]", "company-name")
	query.Set("fields[project-offices]", "name")

	body, _, err := client.Get(cmd.Context(), "/v1/developer-memberships/"+id, query)
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

	details := buildDeveloperMembershipDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderDeveloperMembershipDetails(cmd, details)
}

func parseDeveloperMembershipsShowOptions(cmd *cobra.Command) (developerMembershipsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return developerMembershipsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildDeveloperMembershipDetails(resp jsonAPISingleResponse) developerMembershipDetails {
	resource := resp.Data
	attrs := resource.Attributes

	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		key := resourceKey(inc.Type, inc.ID)
		included[key] = inc
	}

	details := developerMembershipDetails{
		ID:                                   resource.ID,
		Kind:                                 stringAttr(attrs, "kind"),
		IsAdmin:                              boolAttr(attrs, "is-admin"),
		Title:                                strings.TrimSpace(stringAttr(attrs, "title")),
		ColorHex:                             strings.TrimSpace(stringAttr(attrs, "color-hex")),
		ExternalEmployeeID:                   strings.TrimSpace(stringAttr(attrs, "external-employee-id")),
		StartAt:                              stringAttr(attrs, "start-at"),
		EndAt:                                stringAttr(attrs, "end-at"),
		DrivesShiftType:                      stringAttr(attrs, "drives-shift-type"),
		CanSeeRatesAsManager:                 boolAttr(attrs, "can-see-rates-as-manager"),
		CanValidateProfitImprovements:        boolAttr(attrs, "can-validate-profit-improvements"),
		IsGeofenceViolationTeamMember:        boolAttr(attrs, "is-geofence-violation-team-member"),
		IsUnapprovedTimeCardSubscriber:       boolAttr(attrs, "is-unapproved-time-card-subscriber"),
		IsDefaultJobProductionPlanSubscriber: boolAttr(attrs, "is-default-job-production-plan-subscriber"),
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

	if rel, ok := resource.Relationships["organization"]; ok && rel.Data != nil {
		details.DeveloperID = rel.Data.ID
		key := resourceKey(rel.Data.Type, rel.Data.ID)
		if inc, ok := included[key]; ok {
			details.DeveloperName = firstNonEmpty(
				stringAttr(inc.Attributes, "name"),
				stringAttr(inc.Attributes, "company-name"),
			)
		}
	}

	if rel, ok := resource.Relationships["broker"]; ok && rel.Data != nil {
		details.BrokerID = rel.Data.ID
		key := resourceKey(rel.Data.Type, rel.Data.ID)
		if inc, ok := included[key]; ok {
			details.BrokerName = strings.TrimSpace(stringAttr(inc.Attributes, "company-name"))
		}
	}

	if rel, ok := resource.Relationships["project-office"]; ok && rel.Data != nil {
		details.ProjectOfficeID = rel.Data.ID
		key := resourceKey(rel.Data.Type, rel.Data.ID)
		if inc, ok := included[key]; ok {
			details.ProjectOfficeName = strings.TrimSpace(stringAttr(inc.Attributes, "name"))
		}
	}

	return details
}

func renderDeveloperMembershipDetails(cmd *cobra.Command, d developerMembershipDetails) error {
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

	fmt.Fprintln(out, "Developer:")
	fmt.Fprintf(out, "  ID: %s\n", d.DeveloperID)
	fmt.Fprintf(out, "  Name: %s\n", d.DeveloperName)
	fmt.Fprintln(out, "")

	if d.BrokerID != "" {
		fmt.Fprintln(out, "Broker:")
		fmt.Fprintf(out, "  ID: %s\n", d.BrokerID)
		fmt.Fprintf(out, "  Name: %s\n", d.BrokerName)
		fmt.Fprintln(out, "")
	}

	if d.ProjectOfficeID != "" {
		fmt.Fprintln(out, "Project Office:")
		fmt.Fprintf(out, "  ID: %s\n", d.ProjectOfficeID)
		fmt.Fprintf(out, "  Name: %s\n", d.ProjectOfficeName)
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
	fmt.Fprintf(out, "  Can See Rates As Manager: %s\n", formatBool(d.CanSeeRatesAsManager))
	fmt.Fprintf(out, "  Can Validate Profit Improvements: %s\n", formatBool(d.CanValidateProfitImprovements))
	fmt.Fprintf(out, "  Is Geofence Violation Team Member: %s\n", formatBool(d.IsGeofenceViolationTeamMember))
	fmt.Fprintln(out, "")

	fmt.Fprintln(out, "Notifications:")
	fmt.Fprintf(out, "  Unapproved Time Card Subscriber: %s\n", formatBool(d.IsUnapprovedTimeCardSubscriber))
	fmt.Fprintf(out, "  Default Job Production Plan Subscriber: %s\n", formatBool(d.IsDefaultJobProductionPlanSubscriber))
	fmt.Fprintf(out, "  Inventory Capacity Notifications: %s\n", formatBool(d.EnableInventoryCapacityNotifications))

	return nil
}
