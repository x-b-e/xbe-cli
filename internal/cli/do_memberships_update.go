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

type doMembershipsUpdateOptions struct {
	BaseURL                              string
	Token                                string
	JSON                                 bool
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
	// Track which flags were explicitly set
	FlagsSet map[string]bool
}

func newDoMembershipsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a membership",
		Long: `Update an existing membership.

Only the fields you specify will be updated. Fields not provided remain unchanged.
Note: User and organization cannot be changed after creation.

Arguments:
  <id>    The membership ID (required)

Flags:
  --kind                     Role: operations or manager
  --is-admin                 Admin status (true/false)
  --title                    Title within the organization
  --color-hex                Display color (e.g. #FF0000)
  --external-employee-id     External system employee ID
  --explicit-sort-order      Manual sort order
  --start-at                 Membership start date (ISO 8601)
  --end-at                   Membership end date (ISO 8601)
  --drives-shift-type        Shift type: any, day, or night
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
		Example: `  # Update kind to operations
  xbe do memberships update 686 --kind operations

  # Update multiple fields
  xbe do memberships update 686 --kind manager --is-admin true --title "Manager"

  # Update permissions
  xbe do memberships update 686 --can-see-rates-as-manager true --is-rate-editor true

  # Get JSON output
  xbe do memberships update 686 --kind manager --json`,
		Args: cobra.ExactArgs(1),
		RunE: runDoMembershipsUpdate,
	}
	initDoMembershipsUpdateFlags(cmd)
	return cmd
}

func init() {
	doMembershipsCmd.AddCommand(newDoMembershipsUpdateCmd())
}

func initDoMembershipsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
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

func runDoMembershipsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoMembershipsUpdateOptions(cmd)
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

	id := strings.TrimSpace(args[0])
	if id == "" {
		return fmt.Errorf("membership id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	// First, fetch the membership to determine its type
	getBody, _, err := client.Get(cmd.Context(), "/v1/memberships/"+id, nil)
	if err != nil {
		if len(getBody) > 0 {
			fmt.Fprintln(cmd.ErrOrStderr(), string(getBody))
		}
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	var getResp jsonAPISingleResponse
	if err := json.Unmarshal(getBody, &getResp); err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	// Use the membership type to determine the correct endpoint
	membershipType := getResp.Data.Type
	endpoint := "/v1/" + membershipType + "/" + id

	// Check if at least one field is being updated
	if !hasAnyFlagSet(opts.FlagsSet) {
		err := fmt.Errorf("at least one field to update is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	// Build attributes - only include flags that were explicitly set
	attributes := map[string]any{}
	if opts.FlagsSet["kind"] {
		attributes["kind"] = opts.Kind
	}
	if opts.FlagsSet["is-admin"] {
		attributes["is-admin"] = opts.IsAdmin == "true"
	}
	if opts.FlagsSet["title"] {
		attributes["title"] = opts.Title
	}
	if opts.FlagsSet["color-hex"] {
		attributes["color-hex"] = opts.ColorHex
	}
	if opts.FlagsSet["external-employee-id"] {
		attributes["external-employee-id"] = opts.ExternalEmployeeID
	}
	if opts.FlagsSet["explicit-sort-order"] {
		var i int
		if _, err := fmt.Sscanf(opts.ExplicitSortOrder, "%d", &i); err == nil {
			attributes["explicit-sort-order"] = i
		}
	}
	if opts.FlagsSet["start-at"] {
		attributes["start-at"] = opts.StartAt
	}
	if opts.FlagsSet["end-at"] {
		attributes["end-at"] = opts.EndAt
	}
	if opts.FlagsSet["drives-shift-type"] {
		attributes["drives-shift-type"] = opts.DrivesShiftType
	}
	if opts.FlagsSet["can-see-rates-as-driver"] {
		attributes["can-see-rates-as-driver"] = opts.CanSeeRatesAsDriver == "true"
	}
	if opts.FlagsSet["can-see-rates-as-manager"] {
		attributes["can-see-rates-as-manager"] = opts.CanSeeRatesAsManager == "true"
	}
	if opts.FlagsSet["can-validate-profit-improvements"] {
		attributes["can-validate-profit-improvements"] = opts.CanValidateProfitImprovements == "true"
	}
	if opts.FlagsSet["is-rate-editor"] {
		attributes["is-rate-editor"] = opts.IsRateEditor == "true"
	}
	if opts.FlagsSet["is-time-card-auditor"] {
		attributes["is-time-card-auditor"] = opts.IsTimeCardAuditor == "true"
	}
	if opts.FlagsSet["is-equipment-rental-team-member"] {
		attributes["is-equipment-rental-team-member"] = opts.IsEquipmentRentalTeamMember == "true"
	}
	if opts.FlagsSet["is-geofence-violation-team-member"] {
		attributes["is-geofence-violation-team-member"] = opts.IsGeofenceViolationTeamMember == "true"
	}
	if opts.FlagsSet["is-unapproved-time-card-subscriber"] {
		attributes["is-unapproved-time-card-subscriber"] = opts.IsUnapprovedTimeCardSubscriber == "true"
	}
	if opts.FlagsSet["is-default-job-production-plan-subscriber"] {
		attributes["is-default-job-production-plan-subscriber"] = opts.IsDefaultJobProductionPlanSubscriber == "true"
	}
	if opts.FlagsSet["enable-recap-notifications"] {
		attributes["enable-recap-notifications"] = opts.EnableRecapNotifications == "true"
	}
	if opts.FlagsSet["enable-inventory-capacity-notifications"] {
		attributes["enable-inventory-capacity-notifications"] = opts.EnableInventoryCapacityNotifications == "true"
	}

	// Build request body
	data := map[string]any{
		"id":         id,
		"type":       membershipType,
		"attributes": attributes,
	}

	// Handle project-office relationship update
	if opts.FlagsSet["project-office"] {
		if opts.ProjectOffice == "" {
			// Setting to null
			data["relationships"] = map[string]any{
				"project-office": map[string]any{
					"data": nil,
				},
			}
		} else {
			data["relationships"] = map[string]any{
				"project-office": map[string]any{
					"data": map[string]string{
						"type": "project-offices",
						"id":   opts.ProjectOffice,
					},
				},
			}
		}
	}

	requestBody := map[string]any{"data": data}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	body, _, err := client.Patch(cmd.Context(), endpoint, jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Updated membership %s\n\n", details.ID)
	return renderMembershipDetails(cmd, details)
}

func parseDoMembershipsUpdateOptions(cmd *cobra.Command) (doMembershipsUpdateOptions, error) {
	flagsSet := make(map[string]bool)
	flagNames := []string{
		"kind", "is-admin", "title", "color-hex", "external-employee-id",
		"explicit-sort-order", "start-at", "end-at", "drives-shift-type",
		"project-office", "can-see-rates-as-driver", "can-see-rates-as-manager",
		"can-validate-profit-improvements", "is-rate-editor", "is-time-card-auditor",
		"is-equipment-rental-team-member", "is-geofence-violation-team-member",
		"is-unapproved-time-card-subscriber", "is-default-job-production-plan-subscriber",
		"enable-recap-notifications", "enable-inventory-capacity-notifications",
	}
	for _, name := range flagNames {
		flagsSet[name] = cmd.Flags().Changed(name)
	}

	jsonOut, _ := cmd.Flags().GetBool("json")
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

	return doMembershipsUpdateOptions{
		BaseURL:                              baseURL,
		Token:                                token,
		JSON:                                 jsonOut,
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
		FlagsSet:                             flagsSet,
	}, nil
}

func hasAnyFlagSet(flagsSet map[string]bool) bool {
	for _, set := range flagsSet {
		if set {
			return true
		}
	}
	return false
}
