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

type brokerProjectTransportEventTypesListOptions struct {
	BaseURL                   string
	Token                     string
	JSON                      bool
	NoAuth                    bool
	Limit                     int
	Offset                    int
	Sort                      string
	Broker                    string
	ProjectTransportEventType string
}

type brokerProjectTransportEventTypeRow struct {
	ID                            string `json:"id"`
	Code                          string `json:"code"`
	BrokerID                      string `json:"broker_id,omitempty"`
	BrokerName                    string `json:"broker_name,omitempty"`
	ProjectTransportEventTypeID   string `json:"project_transport_event_type_id,omitempty"`
	ProjectTransportEventType     string `json:"project_transport_event_type,omitempty"`
	ProjectTransportEventTypeCode string `json:"project_transport_event_type_code,omitempty"`
}

func newBrokerProjectTransportEventTypesListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List broker project transport event types",
		Long: `List broker project transport event types with filtering and pagination.

Output Columns:
  ID        Broker project transport event type identifier
  CODE      Broker-specific event type code
  EVENT     Project transport event type (code and name)
  BROKER    Broker name

Filters:
  --broker                    Filter by broker ID
  --project-transport-event-type Filter by project transport event type ID

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List broker project transport event types
  xbe view broker-project-transport-event-types list

  # Filter by broker
  xbe view broker-project-transport-event-types list --broker 123

  # Filter by project transport event type
  xbe view broker-project-transport-event-types list --project-transport-event-type 456

  # Output as JSON
  xbe view broker-project-transport-event-types list --json`,
		Args: cobra.NoArgs,
		RunE: runBrokerProjectTransportEventTypesList,
	}
	initBrokerProjectTransportEventTypesListFlags(cmd)
	return cmd
}

func init() {
	brokerProjectTransportEventTypesCmd.AddCommand(newBrokerProjectTransportEventTypesListCmd())
}

func initBrokerProjectTransportEventTypesListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("broker", "", "Filter by broker ID")
	cmd.Flags().String("project-transport-event-type", "", "Filter by project transport event type ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runBrokerProjectTransportEventTypesList(cmd *cobra.Command, _ []string) error {
	opts, err := parseBrokerProjectTransportEventTypesListOptions(cmd)
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
	query.Set("fields[broker-project-transport-event-types]", "code,broker,project-transport-event-type")
	query.Set("fields[brokers]", "company-name")
	query.Set("fields[project-transport-event-types]", "code,name")
	query.Set("include", "broker,project-transport-event-type")

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
	setFilterIfPresent(query, "filter[project-transport-event-type]", opts.ProjectTransportEventType)

	body, _, err := client.Get(cmd.Context(), "/v1/broker-project-transport-event-types", query)
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

	rows := buildBrokerProjectTransportEventTypeRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderBrokerProjectTransportEventTypesTable(cmd, rows)
}

func parseBrokerProjectTransportEventTypesListOptions(cmd *cobra.Command) (brokerProjectTransportEventTypesListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	broker, _ := cmd.Flags().GetString("broker")
	eventType, _ := cmd.Flags().GetString("project-transport-event-type")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return brokerProjectTransportEventTypesListOptions{
		BaseURL:                   baseURL,
		Token:                     token,
		JSON:                      jsonOut,
		NoAuth:                    noAuth,
		Limit:                     limit,
		Offset:                    offset,
		Sort:                      sort,
		Broker:                    broker,
		ProjectTransportEventType: eventType,
	}, nil
}

func buildBrokerProjectTransportEventTypeRows(resp jsonAPIResponse) []brokerProjectTransportEventTypeRow {
	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	rows := make([]brokerProjectTransportEventTypeRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := brokerProjectTransportEventTypeRow{
			ID:   resource.ID,
			Code: stringAttr(resource.Attributes, "code"),
		}

		if rel, ok := resource.Relationships["broker"]; ok && rel.Data != nil {
			row.BrokerID = rel.Data.ID
			if broker, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
				row.BrokerName = strings.TrimSpace(stringAttr(broker.Attributes, "company-name"))
			}
		}

		if rel, ok := resource.Relationships["project-transport-event-type"]; ok && rel.Data != nil {
			row.ProjectTransportEventTypeID = rel.Data.ID
			if eventType, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
				row.ProjectTransportEventType = strings.TrimSpace(stringAttr(eventType.Attributes, "name"))
				row.ProjectTransportEventTypeCode = strings.TrimSpace(stringAttr(eventType.Attributes, "code"))
			}
		}

		rows = append(rows, row)
	}

	return rows
}

func renderBrokerProjectTransportEventTypesTable(cmd *cobra.Command, rows []brokerProjectTransportEventTypeRow) error {
	out := cmd.OutOrStdout()
	if len(rows) == 0 {
		fmt.Fprintln(out, "No broker project transport event types found.")
		return nil
	}

	writer := tabwriter.NewWriter(out, 0, 0, 3, ' ', 0)
	fmt.Fprintln(writer, "ID\tCODE\tEVENT TYPE\tBROKER")
	for _, row := range rows {
		eventTypeDisplay := formatProjectTransportEventTypeDisplay(row.ProjectTransportEventTypeCode, row.ProjectTransportEventType)
		if eventTypeDisplay == "" {
			eventTypeDisplay = row.ProjectTransportEventTypeID
		}
		brokerDisplay := firstNonEmpty(row.BrokerName, row.BrokerID)
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(row.Code, 12),
			truncateString(eventTypeDisplay, 30),
			truncateString(brokerDisplay, 20),
		)
	}

	return writer.Flush()
}
