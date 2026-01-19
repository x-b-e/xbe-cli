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

type doMembershipsCreateOptions struct {
	BaseURL                              string
	Token                                string
	JSON                                 bool
	User                                 string
	Organization                         string
	Kind                                 string
	IsAdmin                              string
	Title                                string
	ColorHex                             string
	ExternalEmployeeID                   string
	ExplicitSortOrder                    string
	StartAt                              string
	EndAt                                string
	DrivesShiftType                      string
	ProjectOffice                        string
	CanSeeRatesAsDriver                  string
	CanSeeRatesAsManager                 string
	CanValidateProfitImprovements        string
	IsRateEditor                         string
	IsTimeCardAuditor                    string
	IsEquipmentRentalTeamMember          string
	IsGeofenceViolationTeamMember        string
	IsUnapprovedTimeCardSubscriber       string
	IsDefaultJobProductionPlanSubscriber string
	EnableRecapNotifications             string
	EnableInventoryCapacityNotifications string
}

func newDoMembershipsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new membership",
		Long: `Create a new membership.

Required flags:
  --user           User ID (required)
  --organization   Organization in Type|ID format (required)
                   Types: Broker, Customer, Trucker, MaterialSupplier, Developer

Optional flags:
  --kind                     Role: operations (default) or manager
  --is-admin                 Admin status (true/false)
  --title                    Title within the organization
  --color-hex                Display color (e.g. #FF0000)
  --external-employee-id     External system employee ID
  --explicit-sort-order      Manual sort order
  --start-at                 Membership start date (ISO 8601)
  --end-at                   Membership end date (ISO 8601)
  --drives-shift-type        Shift type: any (default), day, or night
  --project-office           Project office ID

Permission flags (true/false):
  --can-see-rates-as-driver
  --can-see-rates-as-manager
  --can-validate-profit-improvements
  --is-rate-editor
  --is-time-card-auditor
  --is-equipment-rental-team-member
  --is-geofence-violation-team-member

Notification flags (true/false):
  --is-unapproved-time-card-subscriber
  --is-default-job-production-plan-subscriber
  --enable-recap-notifications
  --enable-inventory-capacity-notifications`,
		Example: `  # Create a broker membership
  xbe do memberships create --user 123 --organization Broker|4 --kind manager

  # Create a trucker membership with title
  xbe do memberships create --user 456 --organization Trucker|789 --kind operations --title "Driver"

  # Create with admin privileges
  xbe do memberships create --user 123 --organization Broker|4 --kind manager --is-admin true

  # Get JSON output
  xbe do memberships create --user 123 --organization Broker|4 --json`,
		Args: cobra.NoArgs,
		RunE: runDoMembershipsCreate,
	}
	initDoMembershipsCreateFlags(cmd)
	return cmd
}

func init() {
	doMembershipsCmd.AddCommand(newDoMembershipsCreateCmd())
}

func initDoMembershipsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("user", "", "User ID (required)")
	cmd.Flags().String("organization", "", "Organization in Type|ID format, e.g. Broker|123 (required)")
	cmd.Flags().String("kind", "", "Role: operations or manager")
	cmd.Flags().String("is-admin", "", "Admin status (true/false)")
	cmd.Flags().String("title", "", "Title within the organization")
	cmd.Flags().String("color-hex", "", "Display color (e.g. #FF0000)")
	cmd.Flags().String("external-employee-id", "", "External system employee ID")
	cmd.Flags().String("explicit-sort-order", "", "Manual sort order")
	cmd.Flags().String("start-at", "", "Membership start date (ISO 8601)")
	cmd.Flags().String("end-at", "", "Membership end date (ISO 8601)")
	cmd.Flags().String("drives-shift-type", "", "Shift type: any, day, or night")
	cmd.Flags().String("project-office", "", "Project office ID")
	cmd.Flags().String("can-see-rates-as-driver", "", "Can see rates as driver (true/false)")
	cmd.Flags().String("can-see-rates-as-manager", "", "Can see rates as manager (true/false)")
	cmd.Flags().String("can-validate-profit-improvements", "", "Can validate profit improvements (true/false)")
	cmd.Flags().String("is-rate-editor", "", "Is rate editor (true/false)")
	cmd.Flags().String("is-time-card-auditor", "", "Is time card auditor (true/false)")
	cmd.Flags().String("is-equipment-rental-team-member", "", "Is equipment rental team member (true/false)")
	cmd.Flags().String("is-geofence-violation-team-member", "", "Is geofence violation team member (true/false)")
	cmd.Flags().String("is-unapproved-time-card-subscriber", "", "Is unapproved time card subscriber (true/false)")
	cmd.Flags().String("is-default-job-production-plan-subscriber", "", "Is default job production plan subscriber (true/false)")
	cmd.Flags().String("enable-recap-notifications", "", "Enable recap notifications (true/false)")
	cmd.Flags().String("enable-inventory-capacity-notifications", "", "Enable inventory capacity notifications (true/false)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoMembershipsCreate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoMembershipsCreateOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	// Require authentication
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

	// Validate required fields
	if opts.User == "" {
		err := fmt.Errorf("--user is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.Organization == "" {
		err := fmt.Errorf("--organization is required (format: Type|ID, e.g. Broker|123)")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	// Parse organization
	orgType, orgID, err := parseOrganization(opts.Organization)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	// Derive membership type and endpoint from organization type
	// e.g., "brokers" -> "broker-memberships", "/v1/broker-memberships"
	membershipType, membershipEndpoint := deriveMembershipTypeAndEndpoint(orgType)

	// Build attributes
	attributes := map[string]any{}
	setStringAttrIfPresent(attributes, "kind", opts.Kind)
	setBoolAttrIfPresent(attributes, "is-admin", opts.IsAdmin)
	setStringAttrIfPresent(attributes, "title", opts.Title)
	setStringAttrIfPresent(attributes, "color-hex", opts.ColorHex)
	setStringAttrIfPresent(attributes, "external-employee-id", opts.ExternalEmployeeID)
	setIntAttrIfPresent(attributes, "explicit-sort-order", opts.ExplicitSortOrder)
	setStringAttrIfPresent(attributes, "start-at", opts.StartAt)
	setStringAttrIfPresent(attributes, "end-at", opts.EndAt)
	setStringAttrIfPresent(attributes, "drives-shift-type", opts.DrivesShiftType)
	setBoolAttrIfPresent(attributes, "can-see-rates-as-driver", opts.CanSeeRatesAsDriver)
	setBoolAttrIfPresent(attributes, "can-see-rates-as-manager", opts.CanSeeRatesAsManager)
	setBoolAttrIfPresent(attributes, "can-validate-profit-improvements", opts.CanValidateProfitImprovements)
	setBoolAttrIfPresent(attributes, "is-rate-editor", opts.IsRateEditor)
	setBoolAttrIfPresent(attributes, "is-time-card-auditor", opts.IsTimeCardAuditor)
	setBoolAttrIfPresent(attributes, "is-equipment-rental-team-member", opts.IsEquipmentRentalTeamMember)
	setBoolAttrIfPresent(attributes, "is-geofence-violation-team-member", opts.IsGeofenceViolationTeamMember)
	setBoolAttrIfPresent(attributes, "is-unapproved-time-card-subscriber", opts.IsUnapprovedTimeCardSubscriber)
	setBoolAttrIfPresent(attributes, "is-default-job-production-plan-subscriber", opts.IsDefaultJobProductionPlanSubscriber)
	setBoolAttrIfPresent(attributes, "enable-recap-notifications", opts.EnableRecapNotifications)
	setBoolAttrIfPresent(attributes, "enable-inventory-capacity-notifications", opts.EnableInventoryCapacityNotifications)

	// Build relationships
	relationships := map[string]any{
		"user": map[string]any{
			"data": map[string]string{
				"type": "users",
				"id":   opts.User,
			},
		},
		"organization": map[string]any{
			"data": map[string]string{
				"type": orgType,
				"id":   orgID,
			},
		},
	}
	if opts.ProjectOffice != "" {
		relationships["project-office"] = map[string]any{
			"data": map[string]string{
				"type": "project-offices",
				"id":   opts.ProjectOffice,
			},
		}
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          membershipType,
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

	body, _, err := client.Post(cmd.Context(), membershipEndpoint, jsonBody)
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

	details := buildMembershipDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created membership %s\n\n", details.ID)
	return renderMembershipDetails(cmd, details)
}

func parseDoMembershipsCreateOptions(cmd *cobra.Command) (doMembershipsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	user, _ := cmd.Flags().GetString("user")
	organization, _ := cmd.Flags().GetString("organization")
	kind, _ := cmd.Flags().GetString("kind")
	isAdmin, _ := cmd.Flags().GetString("is-admin")
	title, _ := cmd.Flags().GetString("title")
	colorHex, _ := cmd.Flags().GetString("color-hex")
	externalEmployeeID, _ := cmd.Flags().GetString("external-employee-id")
	explicitSortOrder, _ := cmd.Flags().GetString("explicit-sort-order")
	startAt, _ := cmd.Flags().GetString("start-at")
	endAt, _ := cmd.Flags().GetString("end-at")
	drivesShiftType, _ := cmd.Flags().GetString("drives-shift-type")
	projectOffice, _ := cmd.Flags().GetString("project-office")
	canSeeRatesAsDriver, _ := cmd.Flags().GetString("can-see-rates-as-driver")
	canSeeRatesAsManager, _ := cmd.Flags().GetString("can-see-rates-as-manager")
	canValidateProfitImprovements, _ := cmd.Flags().GetString("can-validate-profit-improvements")
	isRateEditor, _ := cmd.Flags().GetString("is-rate-editor")
	isTimeCardAuditor, _ := cmd.Flags().GetString("is-time-card-auditor")
	isEquipmentRentalTeamMember, _ := cmd.Flags().GetString("is-equipment-rental-team-member")
	isGeofenceViolationTeamMember, _ := cmd.Flags().GetString("is-geofence-violation-team-member")
	isUnapprovedTimeCardSubscriber, _ := cmd.Flags().GetString("is-unapproved-time-card-subscriber")
	isDefaultJobProductionPlanSubscriber, _ := cmd.Flags().GetString("is-default-job-production-plan-subscriber")
	enableRecapNotifications, _ := cmd.Flags().GetString("enable-recap-notifications")
	enableInventoryCapacityNotifications, _ := cmd.Flags().GetString("enable-inventory-capacity-notifications")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doMembershipsCreateOptions{
		BaseURL:                              baseURL,
		Token:                                token,
		JSON:                                 jsonOut,
		User:                                 user,
		Organization:                         organization,
		Kind:                                 kind,
		IsAdmin:                              isAdmin,
		Title:                                title,
		ColorHex:                             colorHex,
		ExternalEmployeeID:                   externalEmployeeID,
		ExplicitSortOrder:                    explicitSortOrder,
		StartAt:                              startAt,
		EndAt:                                endAt,
		DrivesShiftType:                      drivesShiftType,
		ProjectOffice:                        projectOffice,
		CanSeeRatesAsDriver:                  canSeeRatesAsDriver,
		CanSeeRatesAsManager:                 canSeeRatesAsManager,
		CanValidateProfitImprovements:        canValidateProfitImprovements,
		IsRateEditor:                         isRateEditor,
		IsTimeCardAuditor:                    isTimeCardAuditor,
		IsEquipmentRentalTeamMember:          isEquipmentRentalTeamMember,
		IsGeofenceViolationTeamMember:        isGeofenceViolationTeamMember,
		IsUnapprovedTimeCardSubscriber:       isUnapprovedTimeCardSubscriber,
		IsDefaultJobProductionPlanSubscriber: isDefaultJobProductionPlanSubscriber,
		EnableRecapNotifications:             enableRecapNotifications,
		EnableInventoryCapacityNotifications: enableInventoryCapacityNotifications,
	}, nil
}

// parseOrganization parses "Type|ID" format into JSON:API type and ID
func parseOrganization(org string) (string, string, error) {
	parts := strings.SplitN(org, "|", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid organization format: %q (expected Type|ID, e.g. Broker|123)", org)
	}
	orgType := strings.TrimSpace(parts[0])
	orgID := strings.TrimSpace(parts[1])
	if orgType == "" || orgID == "" {
		return "", "", fmt.Errorf("invalid organization format: %q (expected Type|ID, e.g. Broker|123)", org)
	}

	// Convert to JSON:API type (lowercase plural)
	jsonAPIType := strings.ToLower(orgType)
	switch jsonAPIType {
	case "broker":
		jsonAPIType = "brokers"
	case "customer":
		jsonAPIType = "customers"
	case "trucker":
		jsonAPIType = "truckers"
	case "materialsupplier", "material-supplier", "material_supplier":
		jsonAPIType = "material-suppliers"
	case "developer":
		jsonAPIType = "developers"
	default:
		// Allow already-plural forms
		if !strings.HasSuffix(jsonAPIType, "s") {
			jsonAPIType = jsonAPIType + "s"
		}
	}

	return jsonAPIType, orgID, nil
}

func setStringAttrIfPresent(attrs map[string]any, key, value string) {
	if value != "" {
		attrs[key] = value
	}
}

func setBoolAttrIfPresent(attrs map[string]any, key, value string) {
	if value != "" {
		attrs[key] = value == "true"
	}
}

func setIntAttrIfPresent(attrs map[string]any, key, value string) {
	if value != "" {
		var i int
		if _, err := fmt.Sscanf(value, "%d", &i); err == nil {
			attrs[key] = i
		}
	}
}

// deriveMembershipTypeAndEndpoint converts an organization type to the
// corresponding membership type and API endpoint.
// e.g., "brokers" -> ("broker-memberships", "/v1/broker-memberships")
func deriveMembershipTypeAndEndpoint(orgType string) (string, string) {
	// Remove trailing 's' to get singular form, then add -memberships
	singular := strings.TrimSuffix(orgType, "s")
	membershipType := singular + "-memberships"
	endpoint := "/v1/" + membershipType
	return membershipType, endpoint
}
