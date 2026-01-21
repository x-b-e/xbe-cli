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

type projectCostClassificationsListOptions struct {
	BaseURL   string
	Token     string
	JSON      bool
	NoAuth    bool
	Limit     int
	Offset    int
	Name      string
	Broker    string
	HasBroker string
	Parent    string
}

func newProjectCostClassificationsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List project cost classifications",
		Long: `List project cost classifications with filtering and pagination.

Project cost classifications define the hierarchy of cost categories for projects.
They are broker-scoped and can have parent-child relationships.

Output Columns:
  ID      Classification identifier
  NAME    Classification name
  BROKER  Broker organization
  PARENT  Parent classification name (if hierarchical)

Filters:
  --name        Filter by name
  --broker      Filter by broker ID
  --has-broker  Filter by broker presence (true/false)
  --parent      Filter by parent classification ID`,
		Example: `  # List all project cost classifications
  xbe view project-cost-classifications list

  # Filter by broker
  xbe view project-cost-classifications list --broker 123

  # Filter by name
  xbe view project-cost-classifications list --name "Labor"

  # Show only root classifications (no parent)
  xbe view project-cost-classifications list --parent null

  # Output as JSON
  xbe view project-cost-classifications list --json`,
		RunE: runProjectCostClassificationsList,
	}
	initProjectCostClassificationsListFlags(cmd)
	return cmd
}

func init() {
	projectCostClassificationsCmd.AddCommand(newProjectCostClassificationsListCmd())
}

func initProjectCostClassificationsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("name", "", "Filter by name")
	cmd.Flags().String("broker", "", "Filter by broker ID")
	cmd.Flags().String("has-broker", "", "Filter by broker presence (true/false)")
	cmd.Flags().String("parent", "", "Filter by parent classification ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runProjectCostClassificationsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseProjectCostClassificationsListOptions(cmd)
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
	query.Set("fields[project-cost-classifications]", "name,broker,parent")
	query.Set("include", "broker,parent")
	query.Set("fields[brokers]", "company-name")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	setFilterIfPresent(query, "filter[name]", opts.Name)
	setFilterIfPresent(query, "filter[broker]", opts.Broker)
	setFilterIfPresent(query, "filter[has_broker]", opts.HasBroker)
	setFilterIfPresent(query, "filter[parent]", opts.Parent)

	body, _, err := client.Get(cmd.Context(), "/v1/project-cost-classifications", query)
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

	rows := buildProjectCostClassificationRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderProjectCostClassificationsTable(cmd, rows)
}

func parseProjectCostClassificationsListOptions(cmd *cobra.Command) (projectCostClassificationsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	name, _ := cmd.Flags().GetString("name")
	broker, _ := cmd.Flags().GetString("broker")
	hasBroker, _ := cmd.Flags().GetString("has-broker")
	parent, _ := cmd.Flags().GetString("parent")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return projectCostClassificationsListOptions{
		BaseURL:   baseURL,
		Token:     token,
		JSON:      jsonOut,
		NoAuth:    noAuth,
		Limit:     limit,
		Offset:    offset,
		Name:      name,
		Broker:    broker,
		HasBroker: hasBroker,
		Parent:    parent,
	}, nil
}

type projectCostClassificationRow struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	BrokerID   string `json:"broker_id,omitempty"`
	BrokerName string `json:"broker_name,omitempty"`
	ParentID   string `json:"parent_id,omitempty"`
	ParentName string `json:"parent_name,omitempty"`
}

func buildProjectCostClassificationRows(resp jsonAPIResponse) []projectCostClassificationRow {
	// Build included lookup
	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		key := resourceKey(inc.Type, inc.ID)
		included[key] = inc
	}

	rows := make([]projectCostClassificationRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := projectCostClassificationRow{
			ID:   resource.ID,
			Name: stringAttr(resource.Attributes, "name"),
		}

		if rel, ok := resource.Relationships["broker"]; ok && rel.Data != nil {
			row.BrokerID = rel.Data.ID
			if broker, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
				row.BrokerName = stringAttr(broker.Attributes, "company-name")
			}
		}

		if rel, ok := resource.Relationships["parent"]; ok && rel.Data != nil {
			row.ParentID = rel.Data.ID
			if parent, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
				row.ParentName = stringAttr(parent.Attributes, "name")
			}
		}

		rows = append(rows, row)
	}
	return rows
}

func renderProjectCostClassificationsTable(cmd *cobra.Command, rows []projectCostClassificationRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No project cost classifications found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tNAME\tBROKER\tPARENT")
	for _, row := range rows {
		broker := row.BrokerName
		if broker == "" && row.BrokerID != "" {
			broker = row.BrokerID
		}
		parent := row.ParentName
		if parent == "" && row.ParentID != "" {
			parent = row.ParentID
		}
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(row.Name, 30),
			truncateString(broker, 30),
			truncateString(parent, 30),
		)
	}
	return writer.Flush()
}
