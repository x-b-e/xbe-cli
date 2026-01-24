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

type organizationFormattersListOptions struct {
	BaseURL       string
	Token         string
	JSON          bool
	NoAuth        bool
	Limit         int
	Offset        int
	Sort          string
	FormatterType string
	Organization  string
	Status        string
}

type organizationFormatterRow struct {
	ID               string   `json:"id"`
	Description      string   `json:"description,omitempty"`
	FormatterType    string   `json:"formatter_type,omitempty"`
	Status           string   `json:"status,omitempty"`
	IsLibrary        bool     `json:"is_library"`
	MimeTypes        []string `json:"mime_types,omitempty"`
	Organization     string   `json:"organization,omitempty"`
	OrganizationID   string   `json:"organization_id,omitempty"`
	OrganizationType string   `json:"organization_type,omitempty"`
}

func newOrganizationFormattersListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List organization formatters",
		Long: `List organization formatters with filtering and pagination.

Output Columns:
  ID            Formatter identifier
  DESCRIPTION   Formatter description
  TYPE          Formatter type (STI class name)
  STATUS        Formatter status (active/inactive)
  LIBRARY       Whether the formatter is a shared library
  MIME TYPES    Supported MIME types (comma-separated)
  ORGANIZATION  Organization name or Type/ID

Filters:
  --formatter-type  Filter by formatter type (e.g., TimeSheetsExportFormatter)
  --organization    Filter by organization (format: Type|ID, e.g., Broker|123)
  --status          Filter by status (active/inactive)

Sorting:
  Use --sort to specify sort order. Prefix with - for descending. Defaults to description.

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List organization formatters
  xbe view organization-formatters list

  # Filter by organization
  xbe view organization-formatters list --organization "Broker|123"

  # Filter by formatter type
  xbe view organization-formatters list --formatter-type TimeSheetsExportFormatter

  # Show inactive formatters
  xbe view organization-formatters list --status inactive

  # Output as JSON
  xbe view organization-formatters list --json`,
		Args: cobra.NoArgs,
		RunE: runOrganizationFormattersList,
	}
	initOrganizationFormattersListFlags(cmd)
	return cmd
}

func init() {
	organizationFormattersCmd.AddCommand(newOrganizationFormattersListCmd())
}

func initOrganizationFormattersListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("formatter-type", "", "Filter by formatter type")
	cmd.Flags().String("organization", "", "Filter by organization (Type|ID)")
	cmd.Flags().String("status", "", "Filter by status (active/inactive)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runOrganizationFormattersList(cmd *cobra.Command, _ []string) error {
	opts, err := parseOrganizationFormattersListOptions(cmd)
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
	query.Set("fields[organization-formatters]", "description,formatter-type,status,is-library,mime-types,organization")
	query.Set("include", "organization")
	query.Set("fields[brokers]", "company-name")
	query.Set("fields[customers]", "company-name")
	query.Set("fields[truckers]", "company-name")
	query.Set("fields[developers]", "name")
	query.Set("fields[material-suppliers]", "name")

	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	} else {
		query.Set("sort", "description")
	}

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	setFilterIfPresent(query, "filter[formatter-type]", opts.FormatterType)
	setFilterIfPresent(query, "filter[organization]", opts.Organization)
	setFilterIfPresent(query, "filter[status]", opts.Status)

	body, _, err := client.Get(cmd.Context(), "/v1/organization-formatters", query)
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

	rows := buildOrganizationFormatterRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderOrganizationFormattersTable(cmd, rows)
}

func parseOrganizationFormattersListOptions(cmd *cobra.Command) (organizationFormattersListOptions, error) {
	jsonOut, err := cmd.Flags().GetBool("json")
	if err != nil {
		return organizationFormattersListOptions{}, err
	}
	noAuth, err := cmd.Flags().GetBool("no-auth")
	if err != nil {
		return organizationFormattersListOptions{}, err
	}
	limit, err := cmd.Flags().GetInt("limit")
	if err != nil {
		return organizationFormattersListOptions{}, err
	}
	offset, err := cmd.Flags().GetInt("offset")
	if err != nil {
		return organizationFormattersListOptions{}, err
	}
	sort, err := cmd.Flags().GetString("sort")
	if err != nil {
		return organizationFormattersListOptions{}, err
	}
	formatterType, err := cmd.Flags().GetString("formatter-type")
	if err != nil {
		return organizationFormattersListOptions{}, err
	}
	organization, err := cmd.Flags().GetString("organization")
	if err != nil {
		return organizationFormattersListOptions{}, err
	}
	status, err := cmd.Flags().GetString("status")
	if err != nil {
		return organizationFormattersListOptions{}, err
	}
	baseURL, err := cmd.Flags().GetString("base-url")
	if err != nil {
		return organizationFormattersListOptions{}, err
	}
	token, err := cmd.Flags().GetString("token")
	if err != nil {
		return organizationFormattersListOptions{}, err
	}

	return organizationFormattersListOptions{
		BaseURL:       baseURL,
		Token:         token,
		JSON:          jsonOut,
		NoAuth:        noAuth,
		Limit:         limit,
		Offset:        offset,
		Sort:          sort,
		FormatterType: formatterType,
		Organization:  organization,
		Status:        status,
	}, nil
}

func buildOrganizationFormatterRows(resp jsonAPIResponse) []organizationFormatterRow {
	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	rows := make([]organizationFormatterRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, buildOrganizationFormatterRow(resource, included))
	}
	return rows
}

func buildOrganizationFormatterRow(resource jsonAPIResource, included map[string]jsonAPIResource) organizationFormatterRow {
	row := organizationFormatterRow{
		ID:            resource.ID,
		Description:   stringAttr(resource.Attributes, "description"),
		FormatterType: stringAttr(resource.Attributes, "formatter-type"),
		Status:        stringAttr(resource.Attributes, "status"),
		IsLibrary:     boolAttr(resource.Attributes, "is-library"),
		MimeTypes:     stringSliceAttr(resource.Attributes, "mime-types"),
	}

	if rel, ok := resource.Relationships["organization"]; ok && rel.Data != nil {
		row.OrganizationID = rel.Data.ID
		row.OrganizationType = rel.Data.Type
		if inc, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			row.Organization = organizationNameFromIncluded(inc)
		}
	}

	return row
}

func buildOrganizationFormatterRowFromSingle(resp jsonAPISingleResponse) organizationFormatterRow {
	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}
	return buildOrganizationFormatterRow(resp.Data, included)
}

func organizationNameFromIncluded(resource jsonAPIResource) string {
	name := stringAttr(resource.Attributes, "company-name")
	if name == "" {
		name = stringAttr(resource.Attributes, "name")
	}
	return name
}

func renderOrganizationFormattersTable(cmd *cobra.Command, rows []organizationFormatterRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No organization formatters found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tDESCRIPTION\tTYPE\tSTATUS\tLIBRARY\tMIME TYPES\tORGANIZATION")
	for _, row := range rows {
		organization := formatRelated(row.Organization, formatPolymorphic(row.OrganizationType, row.OrganizationID))
		mimeTypes := strings.Join(row.MimeTypes, ", ")
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%t\t%s\t%s\n",
			row.ID,
			row.Description,
			row.FormatterType,
			row.Status,
			row.IsLibrary,
			mimeTypes,
			organization,
		)
	}
	return writer.Flush()
}
