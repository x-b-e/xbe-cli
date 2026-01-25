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

type customerMembershipsListOptions struct {
	BaseURL                              string
	Token                                string
	JSON                                 bool
	NoAuth                               bool
	Limit                                int
	Offset                               int
	Broker                               string
	User                                 string
	Customer                             string
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

func newCustomerMembershipsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List customer memberships",
		Long: `List customer memberships with filtering and pagination.

Returns a list of customer memberships matching the specified criteria.
Customer memberships define relationships between users and customers.

Output Columns (table format):
  ID        Unique membership identifier
  USER      User name
  CUSTOMER  Customer name
  KIND      Role type (operations/manager)

Pagination:
  Use --limit and --offset to paginate through large result sets.

Filtering:
  Multiple filters can be combined. All filters use AND logic.`,
		Example: `  # List all customer memberships
  xbe view customer-memberships list

  # Filter by customer
  xbe view customer-memberships list --customer 123

  # Filter by broker
  xbe view customer-memberships list --broker 456

  # Filter by user
  xbe view customer-memberships list --user 789

  # Search by user name
  xbe view customer-memberships list --q "Jordan"

  # Filter by kind
  xbe view customer-memberships list --kind manager

  # Filter by drives shift type
  xbe view customer-memberships list --drives-shift-type day

  # Filter by notification subscription
  xbe view customer-memberships list --is-default-job-production-plan-subscriber true

  # Paginate results
  xbe view customer-memberships list --limit 50 --offset 100

  # Output as JSON
  xbe view customer-memberships list --json`,
		RunE: runCustomerMembershipsList,
	}
	initCustomerMembershipsListFlags(cmd)
	return cmd
}

func init() {
	customerMembershipsCmd.AddCommand(newCustomerMembershipsListCmd())
}

func initCustomerMembershipsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 0, "Page size (defaults to server default)")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("customer", "", "Filter by customer ID")
	cmd.Flags().String("organization", "", "Filter by organization (Type|ID, e.g. Customer|123)")
	cmd.Flags().String("broker", "", "Filter by broker ID")
	cmd.Flags().String("user", "", "Filter by user ID")
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

func runCustomerMembershipsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseCustomerMembershipsListOptions(cmd)
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

	if opts.Customer != "" && opts.Organization != "" {
		err := fmt.Errorf("--customer and --organization cannot be used together")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.Customer != "" {
		opts.Organization = fmt.Sprintf("Customer|%s", opts.Customer)
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("include", "user,organization,broker")
	query.Set("fields[users]", "name,email-address,mobile-number")
	query.Set("fields[customers]", "company-name")
	query.Set("fields[brokers]", "company-name")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	setFilterIfPresent(query, "filter[organization]", opts.Organization)
	setFilterIfPresent(query, "filter[broker]", opts.Broker)
	setFilterIfPresent(query, "filter[user]", opts.User)
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

	body, _, err := client.Get(cmd.Context(), "/v1/customer-memberships", query)
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

	return renderCustomerMembershipsList(cmd, resp)
}

func parseCustomerMembershipsListOptions(cmd *cobra.Command) (customerMembershipsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	customer, _ := cmd.Flags().GetString("customer")
	organization, _ := cmd.Flags().GetString("organization")
	broker, _ := cmd.Flags().GetString("broker")
	user, _ := cmd.Flags().GetString("user")
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

	return customerMembershipsListOptions{
		BaseURL:                              baseURL,
		Token:                                token,
		JSON:                                 jsonOut,
		NoAuth:                               noAuth,
		Limit:                                limit,
		Offset:                               offset,
		Customer:                             customer,
		Organization:                         organization,
		Broker:                               broker,
		User:                                 user,
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

func renderCustomerMembershipsList(cmd *cobra.Command, resp jsonAPIResponse) error {
	rows := buildMembershipRows(resp)
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No customer memberships found.")
		return nil
	}

	const userMax = 20
	const customerMax = 30

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tUSER\tCUSTOMER\tKIND")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(row.UserName, userMax),
			truncateString(row.OrganizationName, customerMax),
			row.Kind,
		)
	}
	return writer.Flush()
}
