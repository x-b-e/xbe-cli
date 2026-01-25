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

type projectTransportOrganizationsListOptions struct {
	BaseURL                     string
	Token                       string
	JSON                        bool
	NoAuth                      bool
	Limit                       int
	Offset                      int
	Sort                        string
	Broker                      string
	Q                           string
	ExternalTmsMasterCompanyID  string
	ExternalIdentificationValue string
}

type projectTransportOrganizationRow struct {
	ID                         string `json:"id"`
	Name                       string `json:"name"`
	ExternalTmsMasterCompanyID string `json:"external_tms_master_company_id,omitempty"`
	Broker                     string `json:"broker,omitempty"`
	BrokerID                   string `json:"broker_id,omitempty"`
}

func newProjectTransportOrganizationsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List project transport organizations",
		Long: `List project transport organizations with filtering and pagination.

Output Columns:
  ID                         Project transport organization identifier
  NAME                       Organization name
  TMS MASTER COMPANY ID      External TMS master company identifier
  BROKER                     Broker name

Filters:
  --broker                         Filter by broker ID
  --q                              Search by name (partial match)
  --external-tms-master-company-id Filter by external TMS master company ID
  --external-identification-value  Filter by any external identification value

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List project transport organizations
  xbe view project-transport-organizations list

  # Filter by broker
  xbe view project-transport-organizations list --broker 123

  # Search by name
  xbe view project-transport-organizations list --q "Acme"

  # Filter by external TMS master company ID
  xbe view project-transport-organizations list --external-tms-master-company-id "TMS-001"

  # Filter by any external identification value
  xbe view project-transport-organizations list --external-identification-value "TMS-001"

  # Output as JSON
  xbe view project-transport-organizations list --json`,
		Args: cobra.NoArgs,
		RunE: runProjectTransportOrganizationsList,
	}
	initProjectTransportOrganizationsListFlags(cmd)
	return cmd
}

func init() {
	projectTransportOrganizationsCmd.AddCommand(newProjectTransportOrganizationsListCmd())
}

func initProjectTransportOrganizationsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("broker", "", "Filter by broker ID")
	cmd.Flags().String("q", "", "Search by name (partial match)")
	cmd.Flags().String("external-tms-master-company-id", "", "Filter by external TMS master company ID")
	cmd.Flags().String("external-identification-value", "", "Filter by any external identification value")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runProjectTransportOrganizationsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseProjectTransportOrganizationsListOptions(cmd)
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
	query.Set("fields[project-transport-organizations]", "name,external-tms-master-company-id,broker")
	query.Set("fields[brokers]", "company-name")
	query.Set("include", "broker")
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	} else {
		query.Set("sort", "name")
	}

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	setFilterIfPresent(query, "filter[broker]", opts.Broker)
	setFilterIfPresent(query, "filter[q]", opts.Q)
	setFilterIfPresent(query, "filter[external-tms-master-company-id]", opts.ExternalTmsMasterCompanyID)
	setFilterIfPresent(query, "filter[external-identification-value]", opts.ExternalIdentificationValue)

	body, _, err := client.Get(cmd.Context(), "/v1/project-transport-organizations", query)
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

	rows := buildProjectTransportOrganizationRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderProjectTransportOrganizationsTable(cmd, rows)
}

func parseProjectTransportOrganizationsListOptions(cmd *cobra.Command) (projectTransportOrganizationsListOptions, error) {
	jsonOut, err := cmd.Flags().GetBool("json")
	if err != nil {
		return projectTransportOrganizationsListOptions{}, err
	}
	noAuth, err := cmd.Flags().GetBool("no-auth")
	if err != nil {
		return projectTransportOrganizationsListOptions{}, err
	}
	limit, err := cmd.Flags().GetInt("limit")
	if err != nil {
		return projectTransportOrganizationsListOptions{}, err
	}
	offset, err := cmd.Flags().GetInt("offset")
	if err != nil {
		return projectTransportOrganizationsListOptions{}, err
	}
	sort, err := cmd.Flags().GetString("sort")
	if err != nil {
		return projectTransportOrganizationsListOptions{}, err
	}
	broker, err := cmd.Flags().GetString("broker")
	if err != nil {
		return projectTransportOrganizationsListOptions{}, err
	}
	q, err := cmd.Flags().GetString("q")
	if err != nil {
		return projectTransportOrganizationsListOptions{}, err
	}
	externalTmsMasterCompanyID, err := cmd.Flags().GetString("external-tms-master-company-id")
	if err != nil {
		return projectTransportOrganizationsListOptions{}, err
	}
	externalIdentificationValue, err := cmd.Flags().GetString("external-identification-value")
	if err != nil {
		return projectTransportOrganizationsListOptions{}, err
	}
	baseURL, err := cmd.Flags().GetString("base-url")
	if err != nil {
		return projectTransportOrganizationsListOptions{}, err
	}
	token, err := cmd.Flags().GetString("token")
	if err != nil {
		return projectTransportOrganizationsListOptions{}, err
	}

	return projectTransportOrganizationsListOptions{
		BaseURL:                     baseURL,
		Token:                       token,
		JSON:                        jsonOut,
		NoAuth:                      noAuth,
		Limit:                       limit,
		Offset:                      offset,
		Sort:                        sort,
		Broker:                      broker,
		Q:                           q,
		ExternalTmsMasterCompanyID:  externalTmsMasterCompanyID,
		ExternalIdentificationValue: externalIdentificationValue,
	}, nil
}

func buildProjectTransportOrganizationRows(resp jsonAPIResponse) []projectTransportOrganizationRow {
	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		key := resourceKey(inc.Type, inc.ID)
		included[key] = inc
	}

	rows := make([]projectTransportOrganizationRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := projectTransportOrganizationRow{
			ID:                         resource.ID,
			Name:                       stringAttr(resource.Attributes, "name"),
			ExternalTmsMasterCompanyID: stringAttr(resource.Attributes, "external-tms-master-company-id"),
		}

		if rel, ok := resource.Relationships["broker"]; ok && rel.Data != nil {
			row.BrokerID = rel.Data.ID
			if broker, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
				row.Broker = stringAttr(broker.Attributes, "company-name")
			}
		}

		rows = append(rows, row)
	}
	return rows
}

func renderProjectTransportOrganizationsTable(cmd *cobra.Command, rows []projectTransportOrganizationRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No project transport organizations found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tNAME\tTMS MASTER COMPANY ID\tBROKER")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(row.Name, 35),
			truncateString(row.ExternalTmsMasterCompanyID, 24),
			truncateString(row.Broker, 25),
		)
	}
	return writer.Flush()
}
