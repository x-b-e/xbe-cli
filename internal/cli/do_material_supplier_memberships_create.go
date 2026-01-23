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

type doMaterialSupplierMembershipsCreateOptions struct {
	BaseURL                              string
	Token                                string
	JSON                                 bool
	User                                 string
	MaterialSupplier                     string
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
	CanSeeRatesAsManager                 string
	CanValidateProfitImprovements        string
	IsGeofenceViolationTeamMember        string
	IsUnapprovedTimeCardSubscriber       string
	IsDefaultJobProductionPlanSubscriber string
	EnableRecapNotifications             string
	EnableInventoryCapacityNotifications string
}

func newDoMaterialSupplierMembershipsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new material supplier membership",
		Long: `Create a new material supplier membership.

Required flags:
  --user               User ID (required)
  --material-supplier  Material supplier ID (required)

Optional flags:
  --kind                          Role: operations (default) or manager
  --is-admin                      Admin status (true/false)
  --title                         Title within the organization
  --color-hex                     Display color (e.g. #FF0000)
  --external-employee-id          External system employee ID
  --explicit-sort-order           Manual sort order
  --start-at                      Membership start date (ISO 8601)
  --end-at                        Membership end date (ISO 8601)
  --drives-shift-type             Shift type: any (default), day, or night
  --project-office                Project office ID

Permission flags (true/false):
  --can-see-rates-as-manager
  --can-validate-profit-improvements
  --is-geofence-violation-team-member

Notification flags (true/false):
  --is-unapproved-time-card-subscriber
  --is-default-job-production-plan-subscriber (must be false)
  --enable-recap-notifications
  --enable-inventory-capacity-notifications

Notes:
  Material supplier memberships cannot enable rate visibility or profit
  improvement validation. Setting those flags to true is rejected.`,
		Example: `  # Create a material supplier membership
  xbe do material-supplier-memberships create --user 123 --material-supplier 456

  # Create with role and title
  xbe do material-supplier-memberships create --user 123 --material-supplier 456 --kind manager --title "Operations Manager"

  # Get JSON output
  xbe do material-supplier-memberships create --user 123 --material-supplier 456 --json`,
		Args: cobra.NoArgs,
		RunE: runDoMaterialSupplierMembershipsCreate,
	}
	initDoMaterialSupplierMembershipsCreateFlags(cmd)
	return cmd
}

func init() {
	doMaterialSupplierMembershipsCmd.AddCommand(newDoMaterialSupplierMembershipsCreateCmd())
}

func initDoMaterialSupplierMembershipsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("user", "", "User ID (required)")
	cmd.Flags().String("material-supplier", "", "Material supplier ID (required)")
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
	cmd.Flags().String("can-see-rates-as-manager", "", "Can see rates as manager (true/false)")
	cmd.Flags().String("can-validate-profit-improvements", "", "Can validate profit improvements (true/false)")
	cmd.Flags().String("is-geofence-violation-team-member", "", "Is geofence violation team member (true/false)")
	cmd.Flags().String("is-unapproved-time-card-subscriber", "", "Is unapproved time card subscriber (true/false)")
	cmd.Flags().String("is-default-job-production-plan-subscriber", "", "Is default job production plan subscriber (true/false; must be false)")
	cmd.Flags().String("enable-recap-notifications", "", "Enable recap notifications (true/false)")
	cmd.Flags().String("enable-inventory-capacity-notifications", "", "Enable inventory capacity notifications (true/false)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoMaterialSupplierMembershipsCreate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoMaterialSupplierMembershipsCreateOptions(cmd)
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
	if opts.MaterialSupplier == "" {
		err := fmt.Errorf("--material-supplier is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if err := validateMaterialSupplierMembershipFlags(opts.CanSeeRatesAsManager, opts.CanValidateProfitImprovements, opts.IsDefaultJobProductionPlanSubscriber); err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

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
	setBoolAttrIfPresent(attributes, "can-see-rates-as-manager", opts.CanSeeRatesAsManager)
	setBoolAttrIfPresent(attributes, "can-validate-profit-improvements", opts.CanValidateProfitImprovements)
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
				"type": "material-suppliers",
				"id":   opts.MaterialSupplier,
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
			"type":          "material-supplier-memberships",
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

	body, _, err := client.Post(cmd.Context(), "/v1/material-supplier-memberships", jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Created material supplier membership %s\n\n", details.ID)
	return renderMembershipDetails(cmd, details)
}

func parseDoMaterialSupplierMembershipsCreateOptions(cmd *cobra.Command) (doMaterialSupplierMembershipsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	user, _ := cmd.Flags().GetString("user")
	materialSupplier, _ := cmd.Flags().GetString("material-supplier")
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
	canSeeRatesAsManager, _ := cmd.Flags().GetString("can-see-rates-as-manager")
	canValidateProfitImprovements, _ := cmd.Flags().GetString("can-validate-profit-improvements")
	isGeofenceViolationTeamMember, _ := cmd.Flags().GetString("is-geofence-violation-team-member")
	isUnapprovedTimeCardSubscriber, _ := cmd.Flags().GetString("is-unapproved-time-card-subscriber")
	isDefaultJobProductionPlanSubscriber, _ := cmd.Flags().GetString("is-default-job-production-plan-subscriber")
	enableRecapNotifications, _ := cmd.Flags().GetString("enable-recap-notifications")
	enableInventoryCapacityNotifications, _ := cmd.Flags().GetString("enable-inventory-capacity-notifications")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doMaterialSupplierMembershipsCreateOptions{
		BaseURL:                              baseURL,
		Token:                                token,
		JSON:                                 jsonOut,
		User:                                 user,
		MaterialSupplier:                     materialSupplier,
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
		CanSeeRatesAsManager:                 canSeeRatesAsManager,
		CanValidateProfitImprovements:        canValidateProfitImprovements,
		IsGeofenceViolationTeamMember:        isGeofenceViolationTeamMember,
		IsUnapprovedTimeCardSubscriber:       isUnapprovedTimeCardSubscriber,
		IsDefaultJobProductionPlanSubscriber: isDefaultJobProductionPlanSubscriber,
		EnableRecapNotifications:             enableRecapNotifications,
		EnableInventoryCapacityNotifications: enableInventoryCapacityNotifications,
	}, nil
}

func validateMaterialSupplierMembershipFlags(canSeeRatesAsManager, canValidateProfitImprovements, isDefaultJobProductionPlanSubscriber string) error {
	if canSeeRatesAsManager == "true" {
		return fmt.Errorf("--can-see-rates-as-manager is not supported for material supplier memberships")
	}
	if canValidateProfitImprovements == "true" {
		return fmt.Errorf("--can-validate-profit-improvements is not supported for material supplier memberships")
	}
	if isDefaultJobProductionPlanSubscriber == "true" {
		return fmt.Errorf("--is-default-job-production-plan-subscriber must be false for material supplier memberships")
	}
	return nil
}
