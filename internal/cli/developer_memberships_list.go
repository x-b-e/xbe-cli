package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type developerMembershipsListOptions struct {
	BaseURL                              string
	Token                                string
	JSON                                 bool
	NoAuth                               bool
	Limit                                int
	Offset                               int
	Sort                                 string
	Broker                               string
	Organization                         string
	ProjectOffice                        string
	User                                 string
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

type developerMembershipRow struct {
	ID          string `json:"id"`
	UserID      string `json:"user_id"`
	UserName    string `json:"user_name"`
	UserEmail   string `json:"user_email,omitempty"`
	UserMobile  string `json:"user_mobile,omitempty"`
	DeveloperID string `json:"developer_id"`
	Developer   string `json:"developer_name"`
	Kind        string `json:"kind"`
	IsAdmin     bool   `json:"is_admin"`
	Title       string `json:"title,omitempty"`
	ExternalID  string `json:"external_employee_id,omitempty"`
	ColorHex    string `json:"color_hex,omitempty"`
}

func newDeveloperMembershipsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List developer memberships",
		Long: `List developer memberships with filtering and pagination.

Developer memberships define which users belong to a developer organization.

Output Columns:
  ID          Membership identifier
  USER        User name
  DEVELOPER   Developer name
  KIND        Role (operations/manager)

Filters:
  --broker                             Filter by broker ID
  --organization                       Filter by organization (Developer|ID)
  --project-office                     Filter by project office ID
  --user                               Filter by user ID
  --kind                               Filter by role (operations/manager)
  --q                                  Search by user name
  --drives-shift-type                  Filter by shift type (any/day/night)
  --external-employee-id               Filter by external employee ID
  --is-rate-editor                     Filter by rate editor status (true/false)
  --is-time-card-auditor               Filter by time card auditor status (true/false)
  --is-equipment-rental-team-member    Filter by equipment rental team member (true/false)
  --is-geofence-violation-team-member  Filter by geofence violation team member (true/false)
  --is-unapproved-time-card-subscriber Filter by unapproved time card subscriber (true/false)
  --is-default-job-production-plan-subscriber Filter by default job production plan subscriber (true/false)

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List developer memberships
  xbe view developer-memberships list

  # Filter by broker
  xbe view developer-memberships list --broker 123

  # Filter by user
  xbe view developer-memberships list --user 456

  # Filter by project office
  xbe view developer-memberships list --project-office 789

  # Search by user name
  xbe view developer-memberships list --q "Jane"

  # Output JSON
  xbe view developer-memberships list --json`,
		RunE: runDeveloperMembershipsList,
	}
	initDeveloperMembershipsListFlags(cmd)
	return cmd
}

func init() {
	developerMembershipsCmd.AddCommand(newDeveloperMembershipsListCmd())
}

func initDeveloperMembershipsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 0, "Page size (defaults to server default)")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort order")
	cmd.Flags().String("broker", "", "Filter by broker ID")
	cmd.Flags().String("organization", "", "Filter by organization (Developer|ID)")
	cmd.Flags().String("project-office", "", "Filter by project office ID")
	cmd.Flags().String("user", "", "Filter by user ID")
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

func runDeveloperMembershipsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseDeveloperMembershipsListOptions(cmd)
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
	query.Set("include", "user,organization")
	query.Set("fields[users]", "name,email-address,mobile-number")
	query.Set("fields[developers]", "name")

	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}
	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}

	setFilterIfPresent(query, "filter[broker]", opts.Broker)
	setFilterIfPresent(query, "filter[organization]", opts.Organization)
	setFilterIfPresent(query, "filter[project_office]", opts.ProjectOffice)
	setFilterIfPresent(query, "filter[user]", opts.User)
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

	body, _, err := client.Get(cmd.Context(), "/v1/developer-memberships", query)
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

	rows := buildDeveloperMembershipRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderDeveloperMembershipsTable(cmd, rows)
}

func parseDeveloperMembershipsListOptions(cmd *cobra.Command) (developerMembershipsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	broker, _ := cmd.Flags().GetString("broker")
	organization, _ := cmd.Flags().GetString("organization")
	projectOffice, _ := cmd.Flags().GetString("project-office")
	user, _ := cmd.Flags().GetString("user")
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

	return developerMembershipsListOptions{
		BaseURL:                              baseURL,
		Token:                                token,
		JSON:                                 jsonOut,
		NoAuth:                               noAuth,
		Limit:                                limit,
		Offset:                               offset,
		Sort:                                 sort,
		Broker:                               broker,
		Organization:                         organization,
		ProjectOffice:                        projectOffice,
		User:                                 user,
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

func buildDeveloperMembershipRows(resp jsonAPIResponse) []developerMembershipRow {
	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		key := resourceKey(inc.Type, inc.ID)
		included[key] = inc
	}

	rows := make([]developerMembershipRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := developerMembershipRow{
			ID:         resource.ID,
			Kind:       stringAttr(resource.Attributes, "kind"),
			IsAdmin:    boolAttr(resource.Attributes, "is-admin"),
			Title:      strings.TrimSpace(stringAttr(resource.Attributes, "title")),
			ExternalID: strings.TrimSpace(stringAttr(resource.Attributes, "external-employee-id")),
			ColorHex:   strings.TrimSpace(stringAttr(resource.Attributes, "color-hex")),
		}

		if rel, ok := resource.Relationships["user"]; ok && rel.Data != nil {
			row.UserID = rel.Data.ID
			key := resourceKey(rel.Data.Type, rel.Data.ID)
			if inc, ok := included[key]; ok {
				row.UserName = strings.TrimSpace(stringAttr(inc.Attributes, "name"))
				row.UserEmail = strings.TrimSpace(stringAttr(inc.Attributes, "email-address"))
				row.UserMobile = strings.TrimSpace(stringAttr(inc.Attributes, "mobile-number"))
			}
		}

		if rel, ok := resource.Relationships["organization"]; ok && rel.Data != nil {
			row.DeveloperID = rel.Data.ID
			key := resourceKey(rel.Data.Type, rel.Data.ID)
			if inc, ok := included[key]; ok {
				row.Developer = firstNonEmpty(
					stringAttr(inc.Attributes, "name"),
					stringAttr(inc.Attributes, "company-name"),
				)
			}
		}

		rows = append(rows, row)
	}

	return rows
}

func renderDeveloperMembershipsTable(cmd *cobra.Command, rows []developerMembershipRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No developer memberships found.")
		return nil
	}

	const userMax = 20
	const developerMax = 25

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tUSER\tDEVELOPER\tKIND")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(row.UserName, userMax),
			truncateString(row.Developer, developerMax),
			row.Kind,
		)
	}
	return writer.Flush()
}
