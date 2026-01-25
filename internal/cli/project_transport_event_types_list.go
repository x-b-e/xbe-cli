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

type projectTransportEventTypesListOptions struct {
	BaseURL                string
	Token                  string
	JSON                   bool
	NoAuth                 bool
	Limit                  int
	Offset                 int
	Broker                 string
	DwellMinutesMinDefault string
	Q                      string
}

func newProjectTransportEventTypesListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List project transport event types",
		Long: `List project transport event types with filtering and pagination.

Project transport event types define the types of events that can occur at transport stops.

Output Columns:
  ID                       Event type identifier
  CODE                     Event type code
  NAME                     Event type name
  DWELL MIN                Default minimum dwell minutes
  STOP ROLE                Transport order stop role
  BROKER                   Broker name

Filters:
  --broker                    Filter by broker ID
  --dwell-minutes-min-default Filter by default minimum dwell minutes
  --q                         General search`,
		Example: `  # List all project transport event types
  xbe view project-transport-event-types list

  # Filter by broker
  xbe view project-transport-event-types list --broker 123

  # Search
  xbe view project-transport-event-types list --q "pickup"

  # Output as JSON
  xbe view project-transport-event-types list --json`,
		RunE: runProjectTransportEventTypesList,
	}
	initProjectTransportEventTypesListFlags(cmd)
	return cmd
}

func init() {
	projectTransportEventTypesCmd.AddCommand(newProjectTransportEventTypesListCmd())
}

func initProjectTransportEventTypesListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("broker", "", "Filter by broker ID")
	cmd.Flags().String("dwell-minutes-min-default", "", "Filter by default minimum dwell minutes")
	cmd.Flags().String("q", "", "General search")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runProjectTransportEventTypesList(cmd *cobra.Command, _ []string) error {
	opts, err := parseProjectTransportEventTypesListOptions(cmd)
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
	query.Set("fields[project-transport-event-types]", "code,name,dwell-minutes-min-default,transport-order-stop-role,broker")
	query.Set("fields[brokers]", "company-name")
	query.Set("include", "broker")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	setFilterIfPresent(query, "filter[broker]", opts.Broker)
	setFilterIfPresent(query, "filter[dwell-minutes-min-default]", opts.DwellMinutesMinDefault)
	setFilterIfPresent(query, "filter[q]", opts.Q)

	body, _, err := client.Get(cmd.Context(), "/v1/project-transport-event-types", query)
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

	rows := buildProjectTransportEventTypeRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderProjectTransportEventTypesTable(cmd, rows)
}

func parseProjectTransportEventTypesListOptions(cmd *cobra.Command) (projectTransportEventTypesListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	broker, _ := cmd.Flags().GetString("broker")
	dwellMinutesMinDefault, _ := cmd.Flags().GetString("dwell-minutes-min-default")
	q, _ := cmd.Flags().GetString("q")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return projectTransportEventTypesListOptions{
		BaseURL:                baseURL,
		Token:                  token,
		JSON:                   jsonOut,
		NoAuth:                 noAuth,
		Limit:                  limit,
		Offset:                 offset,
		Broker:                 broker,
		DwellMinutesMinDefault: dwellMinutesMinDefault,
		Q:                      q,
	}, nil
}

type projectTransportEventTypeRow struct {
	ID                     string `json:"id"`
	Code                   string `json:"code,omitempty"`
	Name                   string `json:"name"`
	DwellMinutesMinDefault any    `json:"dwell_minutes_min_default,omitempty"`
	TransportOrderStopRole string `json:"transport_order_stop_role,omitempty"`
	BrokerID               string `json:"broker_id,omitempty"`
	BrokerName             string `json:"broker_name,omitempty"`
}

func buildProjectTransportEventTypeRows(resp jsonAPIResponse) []projectTransportEventTypeRow {
	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		key := resourceKey(inc.Type, inc.ID)
		included[key] = inc
	}

	rows := make([]projectTransportEventTypeRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := projectTransportEventTypeRow{
			ID:                     resource.ID,
			Code:                   stringAttr(resource.Attributes, "code"),
			Name:                   stringAttr(resource.Attributes, "name"),
			DwellMinutesMinDefault: resource.Attributes["dwell-minutes-min-default"],
			TransportOrderStopRole: stringAttr(resource.Attributes, "transport-order-stop-role"),
		}

		if rel, ok := resource.Relationships["broker"]; ok && rel.Data != nil {
			row.BrokerID = rel.Data.ID
			if broker, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
				row.BrokerName = stringAttr(broker.Attributes, "company-name")
			}
		}

		rows = append(rows, row)
	}
	return rows
}

func renderProjectTransportEventTypesTable(cmd *cobra.Command, rows []projectTransportEventTypeRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No project transport event types found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tCODE\tNAME\tDWELL MIN\tSTOP ROLE\tBROKER")
	for _, row := range rows {
		dwellMin := ""
		if row.DwellMinutesMinDefault != nil {
			dwellMin = fmt.Sprintf("%v", row.DwellMinutesMinDefault)
		}
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(row.Code, 15),
			truncateString(row.Name, 25),
			dwellMin,
			truncateString(row.TransportOrderStopRole, 15),
			truncateString(row.BrokerName, 20),
		)
	}
	return writer.Flush()
}
