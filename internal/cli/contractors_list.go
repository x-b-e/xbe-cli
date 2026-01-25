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

type contractorsListOptions struct {
	BaseURL   string
	Token     string
	JSON      bool
	NoAuth    bool
	Limit     int
	Offset    int
	Broker    string
	Name      string
	Incidents string
}

type contractorRow struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	BrokerID   string `json:"broker_id,omitempty"`
	BrokerName string `json:"broker_name,omitempty"`
}

func newContractorsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List contractors",
		Long: `List contractors with filtering and pagination.

Contractors are broker-associated organizations used in job production plans
and incident tracking.

Output Columns:
  ID       Contractor identifier
  NAME     Contractor name
  BROKER   Broker organization

Filters:
  --broker     Filter by broker ID
  --name       Filter by name (partial match)
  --incidents  Filter by incident IDs (comma-separated)`,
		Example: `  # List contractors
  xbe view contractors list

  # Filter by broker
  xbe view contractors list --broker 123

  # Search by name
  xbe view contractors list --name "Acme"

  # Filter by incident IDs
  xbe view contractors list --incidents 10,12

  # JSON output
  xbe view contractors list --json`,
		RunE: runContractorsList,
	}
	initContractorsListFlags(cmd)
	return cmd
}

func init() {
	contractorsCmd.AddCommand(newContractorsListCmd())
}

func initContractorsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("broker", "", "Filter by broker ID")
	cmd.Flags().String("name", "", "Filter by name (partial match)")
	cmd.Flags().String("incidents", "", "Filter by incident IDs (comma-separated)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runContractorsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseContractorsListOptions(cmd)
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
	query.Set("sort", "name")
	query.Set("fields[contractors]", "name,broker")
	query.Set("include", "broker")
	query.Set("fields[brokers]", "company-name")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	setFilterIfPresent(query, "filter[broker]", opts.Broker)
	setFilterIfPresent(query, "filter[name]", opts.Name)
	setFilterIfPresent(query, "filter[incidents]", opts.Incidents)

	body, _, err := client.Get(cmd.Context(), "/v1/contractors", query)
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

	rows := buildContractorRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderContractorsTable(cmd, rows)
}

func parseContractorsListOptions(cmd *cobra.Command) (contractorsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	broker, _ := cmd.Flags().GetString("broker")
	name, _ := cmd.Flags().GetString("name")
	incidents, _ := cmd.Flags().GetString("incidents")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return contractorsListOptions{
		BaseURL:   baseURL,
		Token:     token,
		JSON:      jsonOut,
		NoAuth:    noAuth,
		Limit:     limit,
		Offset:    offset,
		Broker:    broker,
		Name:      name,
		Incidents: incidents,
	}, nil
}

func buildContractorRows(resp jsonAPIResponse) []contractorRow {
	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		key := resourceKey(inc.Type, inc.ID)
		included[key] = inc
	}

	rows := make([]contractorRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := contractorRow{
			ID:   resource.ID,
			Name: strings.TrimSpace(stringAttr(resource.Attributes, "name")),
		}

		if rel, ok := resource.Relationships["broker"]; ok && rel.Data != nil {
			row.BrokerID = rel.Data.ID
			if broker, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
				row.BrokerName = strings.TrimSpace(stringAttr(broker.Attributes, "company-name"))
			}
		}

		rows = append(rows, row)
	}
	return rows
}

func renderContractorsTable(cmd *cobra.Command, rows []contractorRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No contractors found.")
		return nil
	}

	const nameMax = 40
	const brokerMax = 30

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tNAME\tBROKER")
	for _, row := range rows {
		broker := row.BrokerName
		if broker == "" && row.BrokerID != "" {
			broker = row.BrokerID
		}

		fmt.Fprintf(writer, "%s\t%s\t%s\n",
			row.ID,
			truncateString(row.Name, nameMax),
			truncateString(broker, brokerMax),
		)
	}
	return writer.Flush()
}
