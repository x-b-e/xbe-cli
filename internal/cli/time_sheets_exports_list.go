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

type timeSheetsExportsListOptions struct {
	BaseURL               string
	Token                 string
	JSON                  bool
	NoAuth                bool
	Limit                 int
	Offset                int
	Sort                  string
	OrganizationFormatter string
	Status                string
	Broker                string
	CreatedBy             string
	Organization          string
	OrganizationType      string
	OrganizationID        string
	TimeSheets            string
}

type timeSheetsExportRow struct {
	ID                      string `json:"id"`
	Status                  string `json:"status,omitempty"`
	FileName                string `json:"file_name,omitempty"`
	MimeType                string `json:"mime_type,omitempty"`
	OrganizationFormatterID string `json:"organization_formatter_id,omitempty"`
	OrganizationType        string `json:"organization_type,omitempty"`
	OrganizationID          string `json:"organization_id,omitempty"`
	BrokerID                string `json:"broker_id,omitempty"`
	CreatedByID             string `json:"created_by_id,omitempty"`
	CreatedAt               string `json:"created_at,omitempty"`
}

func newTimeSheetsExportsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List time sheets exports",
		Long: `List time sheets exports with filtering and pagination.

Output Columns:
  ID          Time sheets export identifier
  STATUS      Export status
  FILE        Export file name
  FORMATTER   Organization formatter ID
  ORGANIZATION Organization (Type/ID)
  BROKER      Broker ID
  CREATED BY  Creator user ID
  CREATED AT  Created timestamp

Filters:
  --organization-formatter  Filter by organization formatter ID
  --status                  Filter by status (processing, processed, failed)
  --broker                  Filter by broker ID
  --created-by              Filter by created-by user ID
  --organization            Filter by organization (format: Type|ID, e.g., Broker|123)
  --organization-id         Filter by organization ID (requires --organization-type or Type|ID)
  --organization-type       Filter by organization type (e.g., Broker, Customer, Trucker)
  --time-sheets             Filter by time sheet IDs (comma-separated)

Sorting:
  Use --sort to specify sort order. Prefix with - for descending.

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List time sheets exports
  xbe view time-sheets-exports list

  # Filter by status
  xbe view time-sheets-exports list --status processed

  # Filter by formatter and time sheets
  xbe view time-sheets-exports list --organization-formatter 123 --time-sheets 456,789

  # Output as JSON
  xbe view time-sheets-exports list --json`,
		Args: cobra.NoArgs,
		RunE: runTimeSheetsExportsList,
	}
	initTimeSheetsExportsListFlags(cmd)
	return cmd
}

func init() {
	timeSheetsExportsCmd.AddCommand(newTimeSheetsExportsListCmd())
}

func initTimeSheetsExportsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("organization-formatter", "", "Filter by organization formatter ID")
	cmd.Flags().String("status", "", "Filter by status (processing, processed, failed)")
	cmd.Flags().String("broker", "", "Filter by broker ID")
	cmd.Flags().String("created-by", "", "Filter by created-by user ID")
	cmd.Flags().String("organization", "", "Filter by organization (format: Type|ID, e.g., Broker|123)")
	cmd.Flags().String("organization-id", "", "Filter by organization ID (requires --organization-type or Type|ID)")
	cmd.Flags().String("organization-type", "", "Filter by organization type (e.g., Broker, Customer, Trucker)")
	cmd.Flags().String("time-sheets", "", "Filter by time sheet IDs (comma-separated)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runTimeSheetsExportsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseTimeSheetsExportsListOptions(cmd)
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
	query.Set("fields[time-sheets-exports]", strings.Join([]string{
		"status",
		"file-name",
		"mime-type",
		"organization-formatter",
		"organization",
		"broker",
		"created-by",
		"created-at",
	}, ","))

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if strings.TrimSpace(opts.Sort) != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[organization-formatter]", opts.OrganizationFormatter)
	setFilterIfPresent(query, "filter[status]", opts.Status)
	setFilterIfPresent(query, "filter[broker]", opts.Broker)
	setFilterIfPresent(query, "filter[created-by]", opts.CreatedBy)
	setFilterIfPresent(query, "filter[organization]", opts.Organization)
	organizationIDFilter, err := buildOrganizationIDFilter(opts.OrganizationType, opts.OrganizationID)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if organizationIDFilter != "" {
		query.Set("filter[organization-id]", organizationIDFilter)
	} else {
		setFilterIfPresent(query, "filter[organization-type]", opts.OrganizationType)
	}
	setFilterIfPresent(query, "filter[time-sheets]", opts.TimeSheets)

	body, _, err := client.Get(cmd.Context(), "/v1/time-sheets-exports", query)
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

	rows := buildTimeSheetsExportRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderTimeSheetsExportsTable(cmd, rows)
}

func parseTimeSheetsExportsListOptions(cmd *cobra.Command) (timeSheetsExportsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	organizationFormatter, _ := cmd.Flags().GetString("organization-formatter")
	status, _ := cmd.Flags().GetString("status")
	broker, _ := cmd.Flags().GetString("broker")
	createdBy, _ := cmd.Flags().GetString("created-by")
	organization, _ := cmd.Flags().GetString("organization")
	organizationType, _ := cmd.Flags().GetString("organization-type")
	organizationID, _ := cmd.Flags().GetString("organization-id")
	timeSheets, _ := cmd.Flags().GetString("time-sheets")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return timeSheetsExportsListOptions{
		BaseURL:               baseURL,
		Token:                 token,
		JSON:                  jsonOut,
		NoAuth:                noAuth,
		Limit:                 limit,
		Offset:                offset,
		Sort:                  sort,
		OrganizationFormatter: organizationFormatter,
		Status:                status,
		Broker:                broker,
		CreatedBy:             createdBy,
		Organization:          organization,
		OrganizationType:      organizationType,
		OrganizationID:        organizationID,
		TimeSheets:            timeSheets,
	}, nil
}

func buildTimeSheetsExportRows(resp jsonAPIResponse) []timeSheetsExportRow {
	rows := make([]timeSheetsExportRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, buildTimeSheetsExportRow(resource))
	}
	return rows
}

func buildTimeSheetsExportRow(resource jsonAPIResource) timeSheetsExportRow {
	attrs := resource.Attributes
	row := timeSheetsExportRow{
		ID:        resource.ID,
		Status:    stringAttr(attrs, "status"),
		FileName:  stringAttr(attrs, "file-name"),
		MimeType:  stringAttr(attrs, "mime-type"),
		CreatedAt: formatDateTime(stringAttr(attrs, "created-at")),
	}

	if rel, ok := resource.Relationships["organization-formatter"]; ok && rel.Data != nil {
		row.OrganizationFormatterID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["organization"]; ok && rel.Data != nil {
		row.OrganizationType = rel.Data.Type
		row.OrganizationID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["broker"]; ok && rel.Data != nil {
		row.BrokerID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["created-by"]; ok && rel.Data != nil {
		row.CreatedByID = rel.Data.ID
	}

	return row
}

func renderTimeSheetsExportsTable(cmd *cobra.Command, rows []timeSheetsExportRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No time sheets exports found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tSTATUS\tFILE\tFORMATTER\tORGANIZATION\tBROKER\tCREATED BY\tCREATED AT")
	for _, row := range rows {
		fileName := truncateString(row.FileName, 40)
		if fileName == "" {
			fileName = "-"
		}
		status := row.Status
		if status == "" {
			status = "-"
		}
		formatter := row.OrganizationFormatterID
		if formatter == "" {
			formatter = "-"
		}
		organization := formatPolymorphic(row.OrganizationType, row.OrganizationID)
		if organization == "" {
			organization = "-"
		}
		broker := row.BrokerID
		if broker == "" {
			broker = "-"
		}
		createdBy := row.CreatedByID
		if createdBy == "" {
			createdBy = "-"
		}
		createdAt := row.CreatedAt
		if createdAt == "" {
			createdAt = "-"
		}

		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			status,
			fileName,
			formatter,
			organization,
			broker,
			createdBy,
			createdAt,
		)
	}
	return writer.Flush()
}
