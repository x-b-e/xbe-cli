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

type integrationExportsListOptions struct {
	BaseURL   string
	Token     string
	JSON      bool
	NoAuth    bool
	Limit     int
	Offset    int
	Sort      string
	CreatedBy string
}

type integrationExportRow struct {
	ID                    string `json:"id"`
	Summary               string `json:"summary,omitempty"`
	JID                   string `json:"jid,omitempty"`
	ExportedAt            string `json:"exported_at,omitempty"`
	IntegrationConfigID   string `json:"integration_config_id,omitempty"`
	IntegrationConfigName string `json:"integration_config_name,omitempty"`
	BrokerID              string `json:"broker_id,omitempty"`
	BrokerName            string `json:"broker_name,omitempty"`
	CreatedByID           string `json:"created_by_id,omitempty"`
	CreatedByName         string `json:"created_by_name,omitempty"`
}

func newIntegrationExportsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List integration exports",
		Long: `List integration exports with filtering and pagination.

Output Columns:
  ID          Integration export identifier
  SUMMARY     Summary
  JID         Job ID
  EXPORTED AT Export timestamp
  CONFIG      Integration config
  BROKER      Broker
  CREATED BY  Creator

Filters:
  --created-by  Filter by creator user ID

Sorting:
  Use --sort to specify sort order. Prefix with - for descending.

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List integration exports
  xbe view integration-exports list

  # Filter by creator
  xbe view integration-exports list --created-by 123

  # Output as JSON
  xbe view integration-exports list --json`,
		Args: cobra.NoArgs,
		RunE: runIntegrationExportsList,
	}
	initIntegrationExportsListFlags(cmd)
	return cmd
}

func init() {
	integrationExportsCmd.AddCommand(newIntegrationExportsListCmd())
}

func initIntegrationExportsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("created-by", "", "Filter by creator user ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runIntegrationExportsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseIntegrationExportsListOptions(cmd)
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
	query.Set("fields[integration-exports]", strings.Join([]string{
		"summary",
		"jid",
		"exported-at",
		"integration-config",
		"broker",
		"created-by",
	}, ","))
	query.Set("include", "integration-config,broker,created-by")
	query.Set("fields[integration-configs]", "friendly-name")
	query.Set("fields[brokers]", "company-name")
	query.Set("fields[users]", "name")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if strings.TrimSpace(opts.Sort) != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[created-by]", opts.CreatedBy)

	body, _, err := client.Get(cmd.Context(), "/v1/integration-exports", query)
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

	rows := buildIntegrationExportRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderIntegrationExportsTable(cmd, rows)
}

func parseIntegrationExportsListOptions(cmd *cobra.Command) (integrationExportsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	createdBy, _ := cmd.Flags().GetString("created-by")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return integrationExportsListOptions{
		BaseURL:   baseURL,
		Token:     token,
		JSON:      jsonOut,
		NoAuth:    noAuth,
		Limit:     limit,
		Offset:    offset,
		Sort:      sort,
		CreatedBy: createdBy,
	}, nil
}

func buildIntegrationExportRows(resp jsonAPIResponse) []integrationExportRow {
	rows := make([]integrationExportRow, 0, len(resp.Data))
	included := make(map[string]jsonAPIResource, len(resp.Included))
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	for _, resource := range resp.Data {
		attrs := resource.Attributes
		row := integrationExportRow{
			ID:         resource.ID,
			Summary:    stringAttr(attrs, "summary"),
			JID:        stringAttr(attrs, "jid"),
			ExportedAt: formatDateTime(stringAttr(attrs, "exported-at")),
		}

		if rel, ok := resource.Relationships["integration-config"]; ok && rel.Data != nil {
			row.IntegrationConfigID = rel.Data.ID
			if config, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
				row.IntegrationConfigName = strings.TrimSpace(stringAttr(config.Attributes, "friendly-name"))
			}
		}

		if rel, ok := resource.Relationships["broker"]; ok && rel.Data != nil {
			row.BrokerID = rel.Data.ID
			if broker, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
				row.BrokerName = strings.TrimSpace(stringAttr(broker.Attributes, "company-name"))
			}
		}

		if rel, ok := resource.Relationships["created-by"]; ok && rel.Data != nil {
			row.CreatedByID = rel.Data.ID
			if user, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
				row.CreatedByName = strings.TrimSpace(stringAttr(user.Attributes, "name"))
			}
		}

		rows = append(rows, row)
	}

	return rows
}

func renderIntegrationExportsTable(cmd *cobra.Command, rows []integrationExportRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No integration exports found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tSUMMARY\tJID\tEXPORTED AT\tCONFIG\tBROKER\tCREATED BY")
	for _, row := range rows {
		config := formatRelated(row.IntegrationConfigName, row.IntegrationConfigID)
		broker := formatRelated(row.BrokerName, row.BrokerID)
		createdBy := formatRelated(row.CreatedByName, row.CreatedByID)
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(row.Summary, 30),
			truncateString(row.JID, 24),
			row.ExportedAt,
			truncateString(config, 28),
			truncateString(broker, 24),
			truncateString(createdBy, 24),
		)
	}
	return writer.Flush()
}
