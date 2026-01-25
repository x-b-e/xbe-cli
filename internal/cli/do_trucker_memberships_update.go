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

type doTruckerMembershipsUpdateOptions struct {
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
	TrailerCoassignmentsResetOn          string
	ProjectOffice                        string
	CanSeeRatesAsDriver                  string
	CanSeeRatesAsManager                 string
	CanValidateProfitImprovements        string
	IsGeofenceViolationTeamMember        string
	IsUnapprovedTimeCardSubscriber       string
	IsDefaultJobProductionPlanSubscriber string
	EnableRecapNotifications             string
	EnableInventoryCapacityNotifications string
	FlagsSet                             map[string]bool
}

func newDoTruckerMembershipsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a trucker membership",
		Long: `Update an existing trucker membership.

Only the fields you specify will be updated. Fields not provided remain unchanged.
Note: User and trucker cannot be changed after creation.

Arguments:
  <id>    The trucker membership ID (required)

Flags:
  --kind                           Role: operations or manager
  --is-admin                       Admin status (true/false)
  --title                          Title within the organization
  --color-hex                      Display color (e.g. #FF0000)
  --external-employee-id           External system employee ID
  --explicit-sort-order            Manual sort order
  --start-at                       Membership start date (ISO 8601)
  --end-at                         Membership end date (ISO 8601)
  --drives-shift-type              Shift type: any, day, or night
  --trailer-coassignments-reset-on Trailer coassignment reset date (YYYY-MM-DD)
  --project-office                 Project office ID

Permission flags (true/false):
  --can-see-rates-as-driver
  --can-see-rates-as-manager
  --can-validate-profit-improvements
  --is-geofence-violation-team-member

Notification flags (true/false):
  --is-unapproved-time-card-subscriber (must be false)
  --is-default-job-production-plan-subscriber (must be false)
  --enable-recap-notifications
  --enable-inventory-capacity-notifications

Notes:
  Trucker memberships do not support rate editor, time card auditor, or
  equipment rental team member flags. Profit improvement validation is only
  for broker managers; setting it to true is rejected.`,
		Example: `  # Update kind to operations
  xbe do trucker-memberships update 686 --kind operations

  # Update multiple fields
  xbe do trucker-memberships update 686 --kind manager --is-admin true --title "Manager"

  # Update trailer coassignment reset date
  xbe do trucker-memberships update 686 --trailer-coassignments-reset-on 2025-01-15

  # Get JSON output
  xbe do trucker-memberships update 686 --kind manager --json`,
		Args: cobra.ExactArgs(1),
		RunE: runDoTruckerMembershipsUpdate,
	}
	initDoTruckerMembershipsUpdateFlags(cmd)
	return cmd
}

func init() {
	doTruckerMembershipsCmd.AddCommand(newDoTruckerMembershipsUpdateCmd())
}

func initDoTruckerMembershipsUpdateFlags(cmd *cobra.Command) {
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
	cmd.Flags().String("trailer-coassignments-reset-on", "", "Trailer coassignment reset date (YYYY-MM-DD)")
	cmd.Flags().String("project-office", "", "Project office ID")
	cmd.Flags().String("can-see-rates-as-driver", "", "Can see rates as driver (true/false)")
	cmd.Flags().String("can-see-rates-as-manager", "", "Can see rates as manager (true/false)")
	cmd.Flags().String("can-validate-profit-improvements", "", "Can validate profit improvements (true/false)")
	cmd.Flags().String("is-geofence-violation-team-member", "", "Is geofence violation team member (true/false)")
	cmd.Flags().String("is-unapproved-time-card-subscriber", "", "Is unapproved time card subscriber (true/false; must be false)")
	cmd.Flags().String("is-default-job-production-plan-subscriber", "", "Is default job production plan subscriber (true/false; must be false)")
	cmd.Flags().String("enable-recap-notifications", "", "Enable recap notifications (true/false)")
	cmd.Flags().String("enable-inventory-capacity-notifications", "", "Enable inventory capacity notifications (true/false)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoTruckerMembershipsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoTruckerMembershipsUpdateOptions(cmd)
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
		return fmt.Errorf("trucker membership id is required")
	}

	if !hasAnyFlagSet(opts.FlagsSet) {
		err := fmt.Errorf("at least one field to update is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if err := validateTruckerMembershipFlags(opts.CanValidateProfitImprovements, opts.IsUnapprovedTimeCardSubscriber, opts.IsDefaultJobProductionPlanSubscriber); err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

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
	if opts.FlagsSet["trailer-coassignments-reset-on"] {
		attributes["trailer-coassignments-reset-on"] = opts.TrailerCoassignmentsResetOn
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

	data := map[string]any{
		"id":         id,
		"type":       "trucker-memberships",
		"attributes": attributes,
	}

	if opts.FlagsSet["project-office"] {
		if opts.ProjectOffice == "" {
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

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Patch(cmd.Context(), "/v1/trucker-memberships/"+id, jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Updated trucker membership %s\n\n", details.ID)
	return renderMembershipDetails(cmd, details)
}

func parseDoTruckerMembershipsUpdateOptions(cmd *cobra.Command) (doTruckerMembershipsUpdateOptions, error) {
	flagsSet := make(map[string]bool)
	flagNames := []string{
		"kind", "is-admin", "title", "color-hex", "external-employee-id",
		"explicit-sort-order", "start-at", "end-at", "drives-shift-type",
		"trailer-coassignments-reset-on", "project-office", "can-see-rates-as-driver",
		"can-see-rates-as-manager", "can-validate-profit-improvements",
		"is-geofence-violation-team-member", "is-unapproved-time-card-subscriber",
		"is-default-job-production-plan-subscriber", "enable-recap-notifications",
		"enable-inventory-capacity-notifications",
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
	trailerCoassignmentsResetOn, _ := cmd.Flags().GetString("trailer-coassignments-reset-on")
	projectOffice, _ := cmd.Flags().GetString("project-office")
	canSeeRatesAsDriver, _ := cmd.Flags().GetString("can-see-rates-as-driver")
	canSeeRatesAsManager, _ := cmd.Flags().GetString("can-see-rates-as-manager")
	canValidateProfitImprovements, _ := cmd.Flags().GetString("can-validate-profit-improvements")
	isGeofenceViolationTeamMember, _ := cmd.Flags().GetString("is-geofence-violation-team-member")
	isUnapprovedTimeCardSubscriber, _ := cmd.Flags().GetString("is-unapproved-time-card-subscriber")
	isDefaultJobProductionPlanSubscriber, _ := cmd.Flags().GetString("is-default-job-production-plan-subscriber")
	enableRecapNotifications, _ := cmd.Flags().GetString("enable-recap-notifications")
	enableInventoryCapacityNotifications, _ := cmd.Flags().GetString("enable-inventory-capacity-notifications")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doTruckerMembershipsUpdateOptions{
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
		TrailerCoassignmentsResetOn:          trailerCoassignmentsResetOn,
		ProjectOffice:                        projectOffice,
		CanSeeRatesAsDriver:                  canSeeRatesAsDriver,
		CanSeeRatesAsManager:                 canSeeRatesAsManager,
		CanValidateProfitImprovements:        canValidateProfitImprovements,
		IsGeofenceViolationTeamMember:        isGeofenceViolationTeamMember,
		IsUnapprovedTimeCardSubscriber:       isUnapprovedTimeCardSubscriber,
		IsDefaultJobProductionPlanSubscriber: isDefaultJobProductionPlanSubscriber,
		EnableRecapNotifications:             enableRecapNotifications,
		EnableInventoryCapacityNotifications: enableInventoryCapacityNotifications,
		FlagsSet:                             flagsSet,
	}, nil
}
