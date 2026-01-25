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

type projectSubscriptionsListOptions struct {
	BaseURL       string
	Token         string
	JSON          bool
	NoAuth        bool
	Limit         int
	Offset        int
	Sort          string
	Project       string
	User          string
	ContactMethod string
}

type projectSubscriptionRow struct {
	ID                      string `json:"id"`
	ProjectID               string `json:"project_id,omitempty"`
	ProjectName             string `json:"project_name,omitempty"`
	UserID                  string `json:"user_id,omitempty"`
	UserName                string `json:"user_name,omitempty"`
	UserEmail               string `json:"user_email,omitempty"`
	ContactMethod           string `json:"contact_method,omitempty"`
	CalculatedContactMethod string `json:"calculated_contact_method,omitempty"`
}

func newProjectSubscriptionsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List project subscriptions",
		Long: `List project subscriptions with filtering and pagination.

Output Columns:
  ID       Subscription identifier
  PROJECT  Project name
  USER     Subscriber name/email
  CONTACT  Contact method

Filters:
  --project         Filter by project ID
  --user            Filter by user ID
  --contact-method  Filter by contact method (email_address, mobile_number)

Sorting:
  Use --sort to specify sort order. Prefix with - for descending.

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List subscriptions
  xbe view project-subscriptions list

  # Filter by project
  xbe view project-subscriptions list --project 123

  # Filter by user
  xbe view project-subscriptions list --user 456

  # Filter by contact method
  xbe view project-subscriptions list --contact-method email_address

  # Output as JSON
  xbe view project-subscriptions list --json`,
		Args: cobra.NoArgs,
		RunE: runProjectSubscriptionsList,
	}
	initProjectSubscriptionsListFlags(cmd)
	return cmd
}

func init() {
	projectSubscriptionsCmd.AddCommand(newProjectSubscriptionsListCmd())
}

func initProjectSubscriptionsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("project", "", "Filter by project ID")
	cmd.Flags().String("user", "", "Filter by user ID")
	cmd.Flags().String("contact-method", "", "Filter by contact method (email_address, mobile_number)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runProjectSubscriptionsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseProjectSubscriptionsListOptions(cmd)
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
	query.Set("fields[project-subscriptions]", "project,user,contact-method,calculated-contact-method")
	query.Set("include", "project,user")
	query.Set("fields[projects]", "name")
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
	setFilterIfPresent(query, "filter[project]", opts.Project)
	setFilterIfPresent(query, "filter[user]", opts.User)
	setFilterIfPresent(query, "filter[contact-method]", opts.ContactMethod)

	body, _, err := client.Get(cmd.Context(), "/v1/project-subscriptions", query)
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

	rows := buildProjectSubscriptionRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderProjectSubscriptionsTable(cmd, rows)
}

func parseProjectSubscriptionsListOptions(cmd *cobra.Command) (projectSubscriptionsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	project, _ := cmd.Flags().GetString("project")
	user, _ := cmd.Flags().GetString("user")
	contactMethod, _ := cmd.Flags().GetString("contact-method")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return projectSubscriptionsListOptions{
		BaseURL:       baseURL,
		Token:         token,
		JSON:          jsonOut,
		NoAuth:        noAuth,
		Limit:         limit,
		Offset:        offset,
		Sort:          sort,
		Project:       project,
		User:          user,
		ContactMethod: contactMethod,
	}, nil
}

func buildProjectSubscriptionRows(resp jsonAPIResponse) []projectSubscriptionRow {
	included := make(map[string]jsonAPIResource, len(resp.Included))
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	rows := make([]projectSubscriptionRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, buildProjectSubscriptionRow(resource, included))
	}
	return rows
}

func projectSubscriptionRowFromSingle(resp jsonAPISingleResponse) projectSubscriptionRow {
	included := make(map[string]jsonAPIResource, len(resp.Included))
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}
	return buildProjectSubscriptionRow(resp.Data, included)
}

func buildProjectSubscriptionRow(resource jsonAPIResource, included map[string]jsonAPIResource) projectSubscriptionRow {
	row := projectSubscriptionRow{
		ID: resource.ID,
	}

	row.ContactMethod = stringAttr(resource.Attributes, "contact-method")
	row.CalculatedContactMethod = stringAttr(resource.Attributes, "calculated-contact-method")

	if rel, ok := resource.Relationships["project"]; ok && rel.Data != nil {
		row.ProjectID = rel.Data.ID
		if project, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			row.ProjectName = stringAttr(project.Attributes, "name")
		}
	}

	if rel, ok := resource.Relationships["user"]; ok && rel.Data != nil {
		row.UserID = rel.Data.ID
		if user, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			row.UserName = stringAttr(user.Attributes, "name")
			row.UserEmail = stringAttr(user.Attributes, "email-address")
		}
	}

	return row
}

func renderProjectSubscriptionsTable(cmd *cobra.Command, rows []projectSubscriptionRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No project subscriptions found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tPROJECT\tUSER\tCONTACT")
	for _, row := range rows {
		project := formatProjectSubscriptionProjectLabel(row)
		user := firstNonEmpty(row.UserName, row.UserEmail, row.UserID)
		contact := firstNonEmpty(row.ContactMethod, row.CalculatedContactMethod)
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(project, 28),
			truncateString(user, 24),
			truncateString(contact, 18),
		)
	}
	return writer.Flush()
}

func formatProjectSubscriptionProjectLabel(row projectSubscriptionRow) string {
	if row.ProjectName != "" {
		return row.ProjectName
	}
	return row.ProjectID
}
