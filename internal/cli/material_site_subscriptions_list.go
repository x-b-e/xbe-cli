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

type materialSiteSubscriptionsListOptions struct {
	BaseURL       string
	Token         string
	JSON          bool
	NoAuth        bool
	Limit         int
	Offset        int
	Sort          string
	ContactMethod string
	User          string
	MaterialSite  string
}

type materialSiteSubscriptionRow struct {
	ID                      string `json:"id"`
	ContactMethod           string `json:"contact_method,omitempty"`
	CalculatedContactMethod string `json:"calculated_contact_method,omitempty"`
	UserID                  string `json:"user_id,omitempty"`
	User                    string `json:"user,omitempty"`
	UserEmail               string `json:"user_email,omitempty"`
	UserMobile              string `json:"user_mobile,omitempty"`
	MaterialSiteID          string `json:"material_site_id,omitempty"`
	MaterialSite            string `json:"material_site,omitempty"`
}

func newMaterialSiteSubscriptionsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List material site subscriptions",
		Long: `List material site subscriptions with filtering and pagination.

Output Columns:
  ID             Subscription identifier
  USER           Subscribed user
  MATERIAL SITE  Material site name
  CONTACT        Contact method used for notifications

Filters:
  --contact-method  Filter by contact method (email_address, mobile_number)
  --user            Filter by user ID (comma-separated for multiple)
  --material-site   Filter by material site ID (comma-separated for multiple)

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List subscriptions
  xbe view material-site-subscriptions list

  # Filter by material site
  xbe view material-site-subscriptions list --material-site 123

  # Filter by user
  xbe view material-site-subscriptions list --user 456

  # Filter by contact method
  xbe view material-site-subscriptions list --contact-method email_address

  # Output as JSON
  xbe view material-site-subscriptions list --json`,
		Args: cobra.NoArgs,
		RunE: runMaterialSiteSubscriptionsList,
	}
	initMaterialSiteSubscriptionsListFlags(cmd)
	return cmd
}

func init() {
	materialSiteSubscriptionsCmd.AddCommand(newMaterialSiteSubscriptionsListCmd())
}

func initMaterialSiteSubscriptionsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("contact-method", "", "Filter by contact method (email_address, mobile_number)")
	cmd.Flags().String("user", "", "Filter by user ID (comma-separated for multiple)")
	cmd.Flags().String("material-site", "", "Filter by material site ID (comma-separated for multiple)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runMaterialSiteSubscriptionsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseMaterialSiteSubscriptionsListOptions(cmd)
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
	query.Set("fields[material-site-subscriptions]", "contact-method,calculated-contact-method,user,material-site")
	query.Set("fields[users]", "name,email-address,mobile-number")
	query.Set("fields[material-sites]", "name")
	query.Set("include", "user,material-site")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[contact_method]", opts.ContactMethod)
	setFilterIfPresent(query, "filter[user]", opts.User)
	setFilterIfPresent(query, "filter[material_site]", opts.MaterialSite)

	body, _, err := client.Get(cmd.Context(), "/v1/material-site-subscriptions", query)
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

	rows := buildMaterialSiteSubscriptionRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderMaterialSiteSubscriptionsTable(cmd, rows)
}

func parseMaterialSiteSubscriptionsListOptions(cmd *cobra.Command) (materialSiteSubscriptionsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	contactMethod, _ := cmd.Flags().GetString("contact-method")
	user, _ := cmd.Flags().GetString("user")
	materialSite, _ := cmd.Flags().GetString("material-site")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return materialSiteSubscriptionsListOptions{
		BaseURL:       baseURL,
		Token:         token,
		JSON:          jsonOut,
		NoAuth:        noAuth,
		Limit:         limit,
		Offset:        offset,
		Sort:          sort,
		ContactMethod: contactMethod,
		User:          user,
		MaterialSite:  materialSite,
	}, nil
}

func buildMaterialSiteSubscriptionRows(resp jsonAPIResponse) []materialSiteSubscriptionRow {
	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	rows := make([]materialSiteSubscriptionRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, buildMaterialSiteSubscriptionRowFromResource(resource, included))
	}

	return rows
}

func buildMaterialSiteSubscriptionRowFromResource(resource jsonAPIResource, included map[string]jsonAPIResource) materialSiteSubscriptionRow {
	attrs := resource.Attributes
	row := materialSiteSubscriptionRow{
		ID:                      resource.ID,
		ContactMethod:           stringAttr(attrs, "contact-method"),
		CalculatedContactMethod: stringAttr(attrs, "calculated-contact-method"),
	}

	if rel, ok := resource.Relationships["user"]; ok && rel.Data != nil {
		row.UserID = rel.Data.ID
		if user, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			row.User = stringAttr(user.Attributes, "name")
			row.UserEmail = stringAttr(user.Attributes, "email-address")
			row.UserMobile = stringAttr(user.Attributes, "mobile-number")
		}
	}

	if rel, ok := resource.Relationships["material-site"]; ok && rel.Data != nil {
		row.MaterialSiteID = rel.Data.ID
		if site, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			row.MaterialSite = stringAttr(site.Attributes, "name")
		}
	}

	return row
}

func renderMaterialSiteSubscriptionsTable(cmd *cobra.Command, rows []materialSiteSubscriptionRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No material site subscriptions found.")
		return nil
	}

	const (
		maxUser = 26
		maxSite = 28
		maxMeth = 16
	)

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tUSER\tMATERIAL SITE\tCONTACT")
	for _, row := range rows {
		userLabel := firstNonEmpty(row.User, row.UserEmail, row.UserID)
		siteLabel := firstNonEmpty(row.MaterialSite, row.MaterialSiteID)
		contactLabel := firstNonEmpty(row.CalculatedContactMethod, row.ContactMethod)
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(userLabel, maxUser),
			truncateString(siteLabel, maxSite),
			truncateString(contactLabel, maxMeth),
		)
	}
	return writer.Flush()
}
