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

type jobProductionPlanSubscriptionsListOptions struct {
	BaseURL           string
	Token             string
	JSON              bool
	NoAuth            bool
	Limit             int
	Offset            int
	Sort              string
	JobProductionPlan string
	User              string
	ContactMethod     string
}

type jobProductionPlanSubscriptionRow struct {
	ID                              string `json:"id"`
	JobProductionPlanID             string `json:"job_production_plan_id,omitempty"`
	JobProductionPlanJobNumber      string `json:"job_number,omitempty"`
	JobProductionPlanJobName        string `json:"job_name,omitempty"`
	UserID                          string `json:"user_id,omitempty"`
	UserName                        string `json:"user_name,omitempty"`
	UserEmail                       string `json:"user_email,omitempty"`
	ContactMethod                   string `json:"contact_method,omitempty"`
	CalculatedContactMethod         string `json:"calculated_contact_method,omitempty"`
	IsCreatedViaProjectSubscription bool   `json:"is_created_via_project_subscription,omitempty"`
	CreatedViaType                  string `json:"created_via_type,omitempty"`
	CreatedViaID                    string `json:"created_via_id,omitempty"`
	ProjectSubscriptionID           string `json:"project_subscription_id,omitempty"`
}

func newJobProductionPlanSubscriptionsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List job production plan subscriptions",
		Long: `List job production plan subscriptions with filtering and pagination.

Output Columns:
  ID       Subscription identifier
  PLAN     Job production plan (job number/name)
  USER     Subscriber name/email
  CONTACT  Contact method

Filters:
  --job-production-plan  Filter by job production plan ID
  --user                 Filter by user ID
  --contact-method       Filter by contact method (email_address, mobile_number)

Sorting:
  Use --sort to specify sort order. Prefix with - for descending.

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List subscriptions
  xbe view job-production-plan-subscriptions list

  # Filter by job production plan
  xbe view job-production-plan-subscriptions list --job-production-plan 123

  # Filter by user
  xbe view job-production-plan-subscriptions list --user 456

  # Filter by contact method
  xbe view job-production-plan-subscriptions list --contact-method email_address

  # Output as JSON
  xbe view job-production-plan-subscriptions list --json`,
		Args: cobra.NoArgs,
		RunE: runJobProductionPlanSubscriptionsList,
	}
	initJobProductionPlanSubscriptionsListFlags(cmd)
	return cmd
}

func init() {
	jobProductionPlanSubscriptionsCmd.AddCommand(newJobProductionPlanSubscriptionsListCmd())
}

func initJobProductionPlanSubscriptionsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("job-production-plan", "", "Filter by job production plan ID")
	cmd.Flags().String("user", "", "Filter by user ID")
	cmd.Flags().String("contact-method", "", "Filter by contact method (email_address, mobile_number)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runJobProductionPlanSubscriptionsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseJobProductionPlanSubscriptionsListOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.NoAuth {
		opts.Token = ""
	} else if strings.TrimSpace(opts.Token) == "" {
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

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[job-production-plan-subscriptions]", "job-production-plan,user,contact-method,calculated-contact-method,is-created-via-project-subscription,created-via,project-subscription")
	query.Set("include", "job-production-plan,user")
	query.Set("fields[job-production-plans]", "job-number,job-name")
	query.Set("fields[users]", "name,email-address")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if strings.TrimSpace(opts.Sort) != "" {
		query.Set("sort", opts.Sort)
	}
	setFilterIfPresent(query, "filter[job-production-plan]", opts.JobProductionPlan)
	setFilterIfPresent(query, "filter[user]", opts.User)
	setFilterIfPresent(query, "filter[contact-method]", opts.ContactMethod)

	body, _, err := client.Get(cmd.Context(), "/v1/job-production-plan-subscriptions", query)
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

	rows := buildJobProductionPlanSubscriptionRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderJobProductionPlanSubscriptionsTable(cmd, rows)
}

func parseJobProductionPlanSubscriptionsListOptions(cmd *cobra.Command) (jobProductionPlanSubscriptionsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	jobProductionPlan, _ := cmd.Flags().GetString("job-production-plan")
	user, _ := cmd.Flags().GetString("user")
	contactMethod, _ := cmd.Flags().GetString("contact-method")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return jobProductionPlanSubscriptionsListOptions{
		BaseURL:           baseURL,
		Token:             token,
		JSON:              jsonOut,
		NoAuth:            noAuth,
		Limit:             limit,
		Offset:            offset,
		Sort:              sort,
		JobProductionPlan: jobProductionPlan,
		User:              user,
		ContactMethod:     contactMethod,
	}, nil
}

func buildJobProductionPlanSubscriptionRows(resp jsonAPIResponse) []jobProductionPlanSubscriptionRow {
	included := make(map[string]jsonAPIResource, len(resp.Included))
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	rows := make([]jobProductionPlanSubscriptionRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, buildJobProductionPlanSubscriptionRow(resource, included))
	}
	return rows
}

func jobProductionPlanSubscriptionRowFromSingle(resp jsonAPISingleResponse) jobProductionPlanSubscriptionRow {
	included := make(map[string]jsonAPIResource, len(resp.Included))
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}
	return buildJobProductionPlanSubscriptionRow(resp.Data, included)
}

func buildJobProductionPlanSubscriptionRow(resource jsonAPIResource, included map[string]jsonAPIResource) jobProductionPlanSubscriptionRow {
	row := jobProductionPlanSubscriptionRow{
		ID: resource.ID,
	}

	row.ContactMethod = stringAttr(resource.Attributes, "contact-method")
	row.CalculatedContactMethod = stringAttr(resource.Attributes, "calculated-contact-method")
	row.IsCreatedViaProjectSubscription = boolAttr(resource.Attributes, "is-created-via-project-subscription")

	if rel, ok := resource.Relationships["job-production-plan"]; ok && rel.Data != nil {
		row.JobProductionPlanID = rel.Data.ID
		if plan, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			row.JobProductionPlanJobNumber = stringAttr(plan.Attributes, "job-number")
			row.JobProductionPlanJobName = stringAttr(plan.Attributes, "job-name")
		}
	}

	if rel, ok := resource.Relationships["user"]; ok && rel.Data != nil {
		row.UserID = rel.Data.ID
		if user, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			row.UserName = stringAttr(user.Attributes, "name")
			row.UserEmail = stringAttr(user.Attributes, "email-address")
		}
	}

	if rel, ok := resource.Relationships["created-via"]; ok && rel.Data != nil {
		row.CreatedViaType = rel.Data.Type
		row.CreatedViaID = rel.Data.ID
	}

	if rel, ok := resource.Relationships["project-subscription"]; ok && rel.Data != nil {
		row.ProjectSubscriptionID = rel.Data.ID
	}

	return row
}

func renderJobProductionPlanSubscriptionsTable(cmd *cobra.Command, rows []jobProductionPlanSubscriptionRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No job production plan subscriptions found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tPLAN\tUSER\tCONTACT")
	for _, row := range rows {
		plan := formatJobProductionPlanSubscriptionPlanLabel(row)
		user := firstNonEmpty(row.UserName, row.UserEmail, row.UserID)
		contact := firstNonEmpty(row.ContactMethod, row.CalculatedContactMethod)
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(plan, 28),
			truncateString(user, 24),
			truncateString(contact, 18),
		)
	}
	return writer.Flush()
}

func formatJobProductionPlanSubscriptionPlanLabel(row jobProductionPlanSubscriptionRow) string {
	if row.JobProductionPlanJobNumber != "" && row.JobProductionPlanJobName != "" {
		return fmt.Sprintf("%s %s", row.JobProductionPlanJobNumber, row.JobProductionPlanJobName)
	}
	if row.JobProductionPlanJobNumber != "" {
		return row.JobProductionPlanJobNumber
	}
	if row.JobProductionPlanJobName != "" {
		return row.JobProductionPlanJobName
	}
	return row.JobProductionPlanID
}
