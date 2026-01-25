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

type rawTransportProjectsListOptions struct {
	BaseURL             string
	Token               string
	JSON                bool
	NoAuth              bool
	Limit               int
	Offset              int
	Sort                string
	Broker              string
	TablesRowversionMin string
	TablesRowversionMax string
}

type rawTransportProjectRow struct {
	ID                    string `json:"id"`
	ExternalProjectNumber string `json:"external_project_number,omitempty"`
	Importer              string `json:"importer,omitempty"`
	ImportStatus          string `json:"import_status,omitempty"`
	IsManaged             bool   `json:"is_managed"`
	BrokerID              string `json:"broker_id,omitempty"`
	ProjectID             string `json:"project_id,omitempty"`
}

func newRawTransportProjectsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List raw transport projects",
		Long: `List raw transport projects with filtering and pagination.

Output Columns:
  ID              Raw transport project ID
  EXTERNAL NUMBER External project number
  IMPORTER        Importer key
  STATUS          Import status
  MANAGED         Managed flag
  BROKER          Broker ID
  PROJECT         Project ID

Filters:
  --broker                Filter by broker ID
  --tables-rowversion-min Filter by tables rowversion minimum
  --tables-rowversion-max Filter by tables rowversion maximum

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List raw transport projects
  xbe view raw-transport-projects list

  # Filter by broker
  xbe view raw-transport-projects list --broker 123

  # Filter by tables rowversion minimum
  xbe view raw-transport-projects list --tables-rowversion-min 100

  # Output as JSON
  xbe view raw-transport-projects list --json`,
		Args: cobra.NoArgs,
		RunE: runRawTransportProjectsList,
	}
	initRawTransportProjectsListFlags(cmd)
	return cmd
}

func init() {
	rawTransportProjectsCmd.AddCommand(newRawTransportProjectsListCmd())
}

func initRawTransportProjectsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("broker", "", "Filter by broker ID")
	cmd.Flags().String("tables-rowversion-min", "", "Filter by tables rowversion minimum")
	cmd.Flags().String("tables-rowversion-max", "", "Filter by tables rowversion maximum")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runRawTransportProjectsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseRawTransportProjectsListOptions(cmd)
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
	query.Set("fields[raw-transport-projects]", "external-project-number,importer,import-status,is-managed")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[broker]", opts.Broker)
	setFilterIfPresent(query, "filter[tables-rowversion-min]", opts.TablesRowversionMin)
	setFilterIfPresent(query, "filter[tables-rowversion-max]", opts.TablesRowversionMax)

	body, _, err := client.Get(cmd.Context(), "/v1/raw-transport-projects", query)
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

	rows := buildRawTransportProjectRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderRawTransportProjectsTable(cmd, rows)
}

func parseRawTransportProjectsListOptions(cmd *cobra.Command) (rawTransportProjectsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	broker, _ := cmd.Flags().GetString("broker")
	rowversionMin, _ := cmd.Flags().GetString("tables-rowversion-min")
	rowversionMax, _ := cmd.Flags().GetString("tables-rowversion-max")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return rawTransportProjectsListOptions{
		BaseURL:             baseURL,
		Token:               token,
		JSON:                jsonOut,
		NoAuth:              noAuth,
		Limit:               limit,
		Offset:              offset,
		Sort:                sort,
		Broker:              broker,
		TablesRowversionMin: rowversionMin,
		TablesRowversionMax: rowversionMax,
	}, nil
}

func buildRawTransportProjectRows(resp jsonAPIResponse) []rawTransportProjectRow {
	rows := make([]rawTransportProjectRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		attrs := resource.Attributes
		row := rawTransportProjectRow{
			ID:                    resource.ID,
			ExternalProjectNumber: stringAttr(attrs, "external-project-number"),
			Importer:              stringAttr(attrs, "importer"),
			ImportStatus:          stringAttr(attrs, "import-status"),
			IsManaged:             boolAttr(attrs, "is-managed"),
			BrokerID:              relationshipIDFromMap(resource.Relationships, "broker"),
			ProjectID:             relationshipIDFromMap(resource.Relationships, "project"),
		}
		rows = append(rows, row)
	}
	return rows
}

func renderRawTransportProjectsTable(cmd *cobra.Command, rows []rawTransportProjectRow) error {
	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tEXTERNAL NUMBER\tIMPORTER\tSTATUS\tMANAGED\tBROKER\tPROJECT")
	for _, row := range rows {
		managed := "no"
		if row.IsManaged {
			managed = "yes"
		}
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(row.ExternalProjectNumber, 24),
			truncateString(row.Importer, 24),
			row.ImportStatus,
			managed,
			row.BrokerID,
			row.ProjectID,
		)
	}
	return writer.Flush()
}
