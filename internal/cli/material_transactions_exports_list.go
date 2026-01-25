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

type materialTransactionsExportsListOptions struct {
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
	OrganizationID        string
	OrganizationType      string
	NotOrganizationType   string
	MaterialTransactions  string
}

type materialTransactionsExportRow struct {
	ID                      string `json:"id"`
	Status                  string `json:"status,omitempty"`
	FileName                string `json:"file_name,omitempty"`
	OrganizationType        string `json:"organization_type,omitempty"`
	OrganizationID          string `json:"organization_id,omitempty"`
	BrokerID                string `json:"broker_id,omitempty"`
	OrganizationFormatterID string `json:"organization_formatter_id,omitempty"`
	CreatedByID             string `json:"created_by_id,omitempty"`
}

func newMaterialTransactionsExportsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List material transaction exports",
		Long: `List material transaction exports with filtering and pagination.

Output Columns:
  ID         Export identifier
  STATUS     Processing status
  FILE NAME  Generated file name
  ORG TYPE   Organization type
  ORG ID     Organization ID
  BROKER     Broker ID
  FORMATTER  Organization formatter ID
  CREATED BY Creator user ID

Filters:
  --organization-formatter  Filter by organization formatter ID
  --status                  Filter by status (processing, processed, failed)
  --broker                  Filter by broker ID
  --created-by              Filter by created-by user ID
  --organization            Filter by organization (Type|ID, e.g. Broker|123)
  --organization-id         Filter by organization ID (requires --organization-type)
  --organization-type       Filter by organization type (e.g. Broker, Customer)
  --not-organization-type   Exclude organization type (e.g. Broker)
  --material-transactions   Filter by material transaction IDs (comma-separated)

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List exports
  xbe view material-transactions-exports list

  # Filter by status
  xbe view material-transactions-exports list --status processed

  # Filter by organization
  xbe view material-transactions-exports list --organization "Broker|123"

  # Output as JSON
  xbe view material-transactions-exports list --json`,
		Args: cobra.NoArgs,
		RunE: runMaterialTransactionsExportsList,
	}
	initMaterialTransactionsExportsListFlags(cmd)
	return cmd
}

func init() {
	materialTransactionsExportsCmd.AddCommand(newMaterialTransactionsExportsListCmd())
}

func initMaterialTransactionsExportsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("organization-formatter", "", "Filter by organization formatter ID")
	cmd.Flags().String("status", "", "Filter by status (processing, processed, failed)")
	cmd.Flags().String("broker", "", "Filter by broker ID")
	cmd.Flags().String("created-by", "", "Filter by created-by user ID")
	cmd.Flags().String("organization", "", "Filter by organization (Type|ID, e.g. Broker|123)")
	cmd.Flags().String("organization-id", "", "Filter by organization ID (requires --organization-type)")
	cmd.Flags().String("organization-type", "", "Filter by organization type (e.g. Broker)")
	cmd.Flags().String("not-organization-type", "", "Exclude organization type (e.g. Broker)")
	cmd.Flags().String("material-transactions", "", "Filter by material transaction IDs (comma-separated)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runMaterialTransactionsExportsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseMaterialTransactionsExportsListOptions(cmd)
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
	query.Set("fields[material-transactions-exports]", "status,file-name,organization,broker,organization-formatter,created-by")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[organization_formatter]", opts.OrganizationFormatter)
	setFilterIfPresent(query, "filter[status]", opts.Status)
	setFilterIfPresent(query, "filter[broker]", opts.Broker)
	setFilterIfPresent(query, "filter[created_by]", opts.CreatedBy)
	setFilterIfPresent(query, "filter[organization]", opts.Organization)
	if opts.OrganizationID != "" {
		if strings.Contains(opts.OrganizationID, "|") {
			query.Set("filter[organization_id]", opts.OrganizationID)
		} else if opts.OrganizationType != "" {
			query.Set("filter[organization_id]", opts.OrganizationType+"|"+opts.OrganizationID)
		} else {
			return fmt.Errorf("--organization-id requires --organization-type")
		}
	}
	setFilterIfPresent(query, "filter[organization_type]", opts.OrganizationType)
	setFilterIfPresent(query, "filter[not_organization_type]", opts.NotOrganizationType)
	setFilterIfPresent(query, "filter[material_transactions]", opts.MaterialTransactions)

	body, _, err := client.Get(cmd.Context(), "/v1/material-transactions-exports", query)
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

	rows := buildMaterialTransactionsExportRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderMaterialTransactionsExportsTable(cmd, rows)
}

func parseMaterialTransactionsExportsListOptions(cmd *cobra.Command) (materialTransactionsExportsListOptions, error) {
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
	organizationID, _ := cmd.Flags().GetString("organization-id")
	organizationType, _ := cmd.Flags().GetString("organization-type")
	notOrganizationType, _ := cmd.Flags().GetString("not-organization-type")
	materialTransactions, _ := cmd.Flags().GetString("material-transactions")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return materialTransactionsExportsListOptions{
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
		OrganizationID:        organizationID,
		OrganizationType:      organizationType,
		NotOrganizationType:   notOrganizationType,
		MaterialTransactions:  materialTransactions,
	}, nil
}

func buildMaterialTransactionsExportRows(resp jsonAPIResponse) []materialTransactionsExportRow {
	rows := make([]materialTransactionsExportRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, buildMaterialTransactionsExportRow(resource))
	}
	return rows
}

func buildMaterialTransactionsExportRow(resource jsonAPIResource) materialTransactionsExportRow {
	attrs := resource.Attributes
	row := materialTransactionsExportRow{
		ID:                      resource.ID,
		Status:                  stringAttr(attrs, "status"),
		FileName:                stringAttr(attrs, "file-name"),
		BrokerID:                relationshipIDFromMap(resource.Relationships, "broker"),
		OrganizationFormatterID: relationshipIDFromMap(resource.Relationships, "organization-formatter"),
		CreatedByID:             relationshipIDFromMap(resource.Relationships, "created-by"),
	}

	if rel, ok := resource.Relationships["organization"]; ok && rel.Data != nil {
		row.OrganizationType = rel.Data.Type
		row.OrganizationID = rel.Data.ID
	}

	return row
}

func renderMaterialTransactionsExportsTable(cmd *cobra.Command, rows []materialTransactionsExportRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No material transaction exports found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tSTATUS\tFILE NAME\tORG TYPE\tORG ID\tBROKER\tFORMATTER\tCREATED BY")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(row.Status, 12),
			truncateString(row.FileName, 30),
			truncateString(row.OrganizationType, 12),
			truncateString(row.OrganizationID, 14),
			truncateString(row.BrokerID, 14),
			truncateString(row.OrganizationFormatterID, 14),
			truncateString(row.CreatedByID, 14),
		)
	}

	return writer.Flush()
}
