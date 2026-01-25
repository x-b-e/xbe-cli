package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type truckerMembershipsListOptions struct {
	BaseURL                              string
	Token                                string
	JSON                                 bool
	NoAuth                               bool
	Limit                                int
	Offset                               int
	Broker                               string
	User                                 string
	Trucker                              string
	ProjectOffice                        string
	Kind                                 string
	Query                                string
	DrivesShiftType                      string
	ExternalEmployeeID                   string
	IsRateEditor                         string
	IsTimeCardAuditor                    string
	IsEquipmentRentalTeamMember          string
	IsGeofenceViolationTeamMember        string
	IsUnapprovedTimeCardSubscriber       string
	IsDefaultJobProductionPlanSubscriber string
}

func newTruckerMembershipsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List trucker memberships",
		Long: `List trucker memberships with filtering and pagination.

Trucker memberships define relationships between users and truckers.

Output Columns (table format):
  ID            Unique membership identifier
  USER          User name
  TYPE          Organization type (Trucker)
  NAME          Trucker name
  KIND          Role type (operations/manager)

Pagination:
  Use --limit and --offset to paginate through large result sets.

Filtering:
  Multiple filters can be combined. All filters use AND logic.`,
		Example: `  # List all trucker memberships
  xbe view trucker-memberships list

  # Filter by trucker
  xbe view trucker-memberships list --trucker 123

  # Filter by broker
  xbe view trucker-memberships list --broker 456

  # Filter by user
  xbe view trucker-memberships list --user 789

  # Search by user name
  xbe view trucker-memberships list --q "Jordan"

  # Filter by kind
  xbe view trucker-memberships list --kind manager

  # Filter by drives shift type
  xbe view trucker-memberships list --drives-shift-type day

  # Output as JSON
  xbe view trucker-memberships list --json`,
		RunE: runTruckerMembershipsList,
	}
	initTruckerMembershipsListFlags(cmd)
	return cmd
}

func init() {
	truckerMembershipsCmd.AddCommand(newTruckerMembershipsListCmd())
}

func initTruckerMembershipsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 0, "Page size (defaults to server default)")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("broker", "", "Filter by broker ID")
	cmd.Flags().String("user", "", "Filter by user ID")
	cmd.Flags().String("trucker", "", "Filter by trucker ID")
	cmd.Flags().String("project-office", "", "Filter by project office ID")
	cmd.Flags().String("kind", "", "Filter by kind (operations/manager)")
	cmd.Flags().String("q", "", "Search by user name")
	cmd.Flags().String("drives-shift-type", "", "Filter by drives shift type (any/day/night)")
	cmd.Flags().String("external-employee-id", "", "Filter by external employee ID")
	cmd.Flags().String("is-rate-editor", "", "Filter by rate editor status (true/false)")
	cmd.Flags().String("is-time-card-auditor", "", "Filter by time card auditor status (true/false)")
	cmd.Flags().String("is-equipment-rental-team-member", "", "Filter by equipment rental team member (true/false)")
	cmd.Flags().String("is-geofence-violation-team-member", "", "Filter by geofence violation team member (true/false)")
	cmd.Flags().String("is-unapproved-time-card-subscriber", "", "Filter by unapproved time card subscriber (true/false)")
	cmd.Flags().String("is-default-job-production-plan-subscriber", "", "Filter by default job production plan subscriber (true/false)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runTruckerMembershipsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseTruckerMembershipsListOptions(cmd)
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

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("include", "user,organization,broker")
	query.Set("fields[users]", "name,email-address,mobile-number")
	query.Set("fields[brokers]", "company-name")
	query.Set("fields[truckers]", "company-name")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	setFilterIfPresent(query, "filter[broker]", opts.Broker)
	setFilterIfPresent(query, "filter[user]", opts.User)
	if opts.Trucker != "" {
		setFilterIfPresent(query, "filter[organization]", normalizeTruckerFilter(opts.Trucker))
	}
	setFilterIfPresent(query, "filter[project_office]", opts.ProjectOffice)
	setFilterIfPresent(query, "filter[kind]", opts.Kind)
	setFilterIfPresent(query, "filter[q]", opts.Query)
	setFilterIfPresent(query, "filter[drives_shift_type]", opts.DrivesShiftType)
	setFilterIfPresent(query, "filter[external_employee_id]", opts.ExternalEmployeeID)
	setFilterIfPresent(query, "filter[is_rate_editor]", opts.IsRateEditor)
	setFilterIfPresent(query, "filter[is_time_card_auditor]", opts.IsTimeCardAuditor)
	setFilterIfPresent(query, "filter[is_equipment_rental_team_member]", opts.IsEquipmentRentalTeamMember)
	setFilterIfPresent(query, "filter[is_geofence_violation_team_member]", opts.IsGeofenceViolationTeamMember)
	setFilterIfPresent(query, "filter[is_unapproved_time_card_subscriber]", opts.IsUnapprovedTimeCardSubscriber)
	setFilterIfPresent(query, "filter[is_default_job_production_plan_subscriber]", opts.IsDefaultJobProductionPlanSubscriber)

	body, _, err := client.Get(cmd.Context(), "/v1/trucker-memberships", query)
	if err != nil {
		if len(body) > 0 {
			fmt.Fprintln(cmd.ErrOrStderr(), string(body))
		}
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	var resp jsonAPIResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	handled, err := renderSparseListIfRequested(cmd, resp)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if handled {
		return nil
	}

	rows := buildMembershipRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderMembershipsList(cmd, resp)
}

func parseTruckerMembershipsListOptions(cmd *cobra.Command) (truckerMembershipsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	broker, _ := cmd.Flags().GetString("broker")
	user, _ := cmd.Flags().GetString("user")
	trucker, _ := cmd.Flags().GetString("trucker")
	projectOffice, _ := cmd.Flags().GetString("project-office")
	kind, _ := cmd.Flags().GetString("kind")
	query, _ := cmd.Flags().GetString("q")
	drivesShiftType, _ := cmd.Flags().GetString("drives-shift-type")
	externalEmployeeID, _ := cmd.Flags().GetString("external-employee-id")
	isRateEditor, _ := cmd.Flags().GetString("is-rate-editor")
	isTimeCardAuditor, _ := cmd.Flags().GetString("is-time-card-auditor")
	isEquipmentRentalTeamMember, _ := cmd.Flags().GetString("is-equipment-rental-team-member")
	isGeofenceViolationTeamMember, _ := cmd.Flags().GetString("is-geofence-violation-team-member")
	isUnapprovedTimeCardSubscriber, _ := cmd.Flags().GetString("is-unapproved-time-card-subscriber")
	isDefaultJobProductionPlanSubscriber, _ := cmd.Flags().GetString("is-default-job-production-plan-subscriber")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return truckerMembershipsListOptions{
		BaseURL:                              baseURL,
		Token:                                token,
		JSON:                                 jsonOut,
		NoAuth:                               noAuth,
		Limit:                                limit,
		Offset:                               offset,
		Broker:                               broker,
		User:                                 user,
		Trucker:                              trucker,
		ProjectOffice:                        projectOffice,
		Kind:                                 kind,
		Query:                                query,
		DrivesShiftType:                      drivesShiftType,
		ExternalEmployeeID:                   externalEmployeeID,
		IsRateEditor:                         isRateEditor,
		IsTimeCardAuditor:                    isTimeCardAuditor,
		IsEquipmentRentalTeamMember:          isEquipmentRentalTeamMember,
		IsGeofenceViolationTeamMember:        isGeofenceViolationTeamMember,
		IsUnapprovedTimeCardSubscriber:       isUnapprovedTimeCardSubscriber,
		IsDefaultJobProductionPlanSubscriber: isDefaultJobProductionPlanSubscriber,
	}, nil
}

func normalizeTruckerFilter(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ""
	}

	parts := strings.SplitN(trimmed, "|", 2)
	if len(parts) == 1 {
		return "Trucker|" + trimmed
	}

	return trimmed
}
