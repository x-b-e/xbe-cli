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

type membershipsListOptions struct {
	BaseURL                              string
	Token                                string
	JSON                                 bool
	NoAuth                               bool
	Limit                                int
	Offset                               int
	Broker                               string
	User                                 string
	Organization                         string
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

func newMembershipsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List memberships",
		Long: `List memberships with filtering and pagination.

Returns a list of memberships matching the specified criteria. Memberships
define relationships between users and organizations.

Output Columns (table format):
  ID            Unique membership identifier
  USER          User name
  TYPE          Organization type (Broker, Customer, etc.)
  NAME          Organization name
  KIND          Role type (operations/manager)

Pagination:
  Use --limit and --offset to paginate through large result sets.

Filtering:
  Multiple filters can be combined. All filters use AND logic.

Organization Types:
  Memberships can be for different organization types:
    - Broker (branch)
    - Customer
    - Trucker
    - MaterialSupplier
    - Developer`,
		Example: `  # List all memberships
  xbe view memberships list

  # Filter by broker
  xbe view memberships list --broker 123

  # Filter by user
  xbe view memberships list --user 456

  # Search by user name
  xbe view memberships list --q "John"

  # Filter by kind
  xbe view memberships list --kind manager
  xbe view memberships list --kind operations

  # Filter by drives shift type
  xbe view memberships list --drives-shift-type day
  xbe view memberships list --drives-shift-type night

  # Filter by rate editor status
  xbe view memberships list --is-rate-editor true

  # Filter by time card auditor
  xbe view memberships list --is-time-card-auditor true

  # Paginate results
  xbe view memberships list --limit 50 --offset 100

  # Output as JSON
  xbe view memberships list --json`,
		RunE: runMembershipsList,
	}
	initMembershipsListFlags(cmd)
	return cmd
}

func init() {
	membershipsCmd.AddCommand(newMembershipsListCmd())
}

func initMembershipsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 0, "Page size (defaults to server default)")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("broker", "", "Filter by broker ID")
	cmd.Flags().String("user", "", "Filter by user ID")
	cmd.Flags().String("organization", "", "Filter by organization (Type|ID, e.g. Broker|123)")
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

func runMembershipsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseMembershipsListOptions(cmd)
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
	// Note: Sparse fieldsets for the primary resource can interfere with
	// relationship inclusion. Only specify fields for included resources.
	query.Set("fields[users]", "name,email-address,mobile-number")
	query.Set("fields[brokers]", "company-name")
	query.Set("fields[customers]", "company-name")
	query.Set("fields[truckers]", "company-name")
	query.Set("fields[material-suppliers]", "name")
	query.Set("fields[developers]", "name")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	setFilterIfPresent(query, "filter[broker]", opts.Broker)
	setFilterIfPresent(query, "filter[user]", opts.User)
	setFilterIfPresent(query, "filter[organization]", opts.Organization)
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

	body, _, err := client.Get(cmd.Context(), "/v1/memberships", query)
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

	if opts.JSON {
		rows := buildMembershipRows(resp)
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderMembershipsList(cmd, resp)
}

func parseMembershipsListOptions(cmd *cobra.Command) (membershipsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	broker, _ := cmd.Flags().GetString("broker")
	user, _ := cmd.Flags().GetString("user")
	organization, _ := cmd.Flags().GetString("organization")
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

	return membershipsListOptions{
		BaseURL:                              baseURL,
		Token:                                token,
		JSON:                                 jsonOut,
		NoAuth:                               noAuth,
		Limit:                                limit,
		Offset:                               offset,
		Broker:                               broker,
		User:                                 user,
		Organization:                         organization,
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

type membershipRow struct {
	ID               string `json:"id"`
	UserID           string `json:"user_id"`
	UserName         string `json:"user_name"`
	UserEmail        string `json:"user_email,omitempty"`
	UserMobile       string `json:"user_mobile,omitempty"`
	OrganizationType string `json:"organization_type"`
	OrganizationID   string `json:"organization_id"`
	OrganizationName string `json:"organization_name"`
	BrokerID         string `json:"broker_id,omitempty"`
	BrokerName       string `json:"broker_name,omitempty"`
	Kind             string `json:"kind"`
	IsAdmin          bool   `json:"is_admin"`
	Title            string `json:"title,omitempty"`
	ExternalID       string `json:"external_employee_id,omitempty"`
	ColorHex         string `json:"color_hex,omitempty"`
}

func buildMembershipRows(resp jsonAPIResponse) []membershipRow {
	// Build included map for lookups
	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		key := resourceKey(inc.Type, inc.ID)
		included[key] = inc
	}

	rows := make([]membershipRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := membershipRow{
			ID:         resource.ID,
			Kind:       stringAttr(resource.Attributes, "kind"),
			IsAdmin:    boolAttr(resource.Attributes, "is-admin"),
			Title:      strings.TrimSpace(stringAttr(resource.Attributes, "title")),
			ExternalID: strings.TrimSpace(stringAttr(resource.Attributes, "external-employee-id")),
			ColorHex:   strings.TrimSpace(stringAttr(resource.Attributes, "color-hex")),
		}

		// Get user info
		if rel, ok := resource.Relationships["user"]; ok && rel.Data != nil {
			row.UserID = rel.Data.ID
			key := resourceKey(rel.Data.Type, rel.Data.ID)
			if inc, ok := included[key]; ok {
				row.UserName = strings.TrimSpace(stringAttr(inc.Attributes, "name"))
				row.UserEmail = strings.TrimSpace(stringAttr(inc.Attributes, "email-address"))
				row.UserMobile = strings.TrimSpace(stringAttr(inc.Attributes, "mobile-number"))
			}
		}

		// Get organization info (polymorphic)
		if rel, ok := resource.Relationships["organization"]; ok && rel.Data != nil {
			row.OrganizationType = rel.Data.Type
			row.OrganizationID = rel.Data.ID
			key := resourceKey(rel.Data.Type, rel.Data.ID)
			if inc, ok := included[key]; ok {
				// Different organization types use different name fields
				row.OrganizationName = firstNonEmpty(
					stringAttr(inc.Attributes, "company-name"),
					stringAttr(inc.Attributes, "name"),
				)
			}
		}

		// Get broker info
		if rel, ok := resource.Relationships["broker"]; ok && rel.Data != nil {
			row.BrokerID = rel.Data.ID
			key := resourceKey(rel.Data.Type, rel.Data.ID)
			if inc, ok := included[key]; ok {
				row.BrokerName = strings.TrimSpace(stringAttr(inc.Attributes, "company-name"))
			}
		}

		rows = append(rows, row)
	}
	return rows
}

func renderMembershipsList(cmd *cobra.Command, resp jsonAPIResponse) error {
	rows := buildMembershipRows(resp)
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No memberships found.")
		return nil
	}

	const userMax = 20
	const orgMax = 25

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "\t\tORGANIZATION\t\t")
	fmt.Fprintln(writer, "ID\tUSER\tTYPE\tNAME\tKIND")
	for _, row := range rows {
		// Format org type (e.g., "brokers" -> "Broker")
		orgType := strings.TrimSuffix(row.OrganizationType, "s")
		if orgType != "" {
			orgType = strings.ToUpper(orgType[:1]) + orgType[1:]
		}

		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(row.UserName, userMax),
			orgType,
			truncateString(row.OrganizationName, orgMax),
			row.Kind,
		)
	}
	return writer.Flush()
}
