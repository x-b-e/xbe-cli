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

type qualityControlClassificationsListOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
	Limit   int
	Offset  int
	Broker  string
	Name    string
}

func newQualityControlClassificationsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List quality control classifications",
		Long: `List quality control classifications with filtering and pagination.

Quality control classifications define types of quality inspections
and checks that can be performed.

Output Columns:
  ID           Classification identifier
  NAME         Classification name
  DESCRIPTION  Description
  BROKER       Broker organization

Filters:
  --broker  Filter by broker ID
  --name    Filter by name (partial match, case-insensitive)`,
		Example: `  # List all quality control classifications
  xbe view quality-control-classifications list

  # Filter by broker
  xbe view quality-control-classifications list --broker 123

  # Filter by name
  xbe view quality-control-classifications list --name "temperature"

  # Output as JSON
  xbe view quality-control-classifications list --json`,
		RunE: runQualityControlClassificationsList,
	}
	initQualityControlClassificationsListFlags(cmd)
	return cmd
}

func init() {
	qualityControlClassificationsCmd.AddCommand(newQualityControlClassificationsListCmd())
}

func initQualityControlClassificationsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("broker", "", "Filter by broker ID")
	cmd.Flags().String("name", "", "Filter by name (partial match)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runQualityControlClassificationsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseQualityControlClassificationsListOptions(cmd)
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
	query.Set("fields[quality-control-classifications]", "name,description,broker")
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

	body, _, err := client.Get(cmd.Context(), "/v1/quality-control-classifications", query)
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

	rows := buildQualityControlClassificationRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderQualityControlClassificationsTable(cmd, rows)
}

func parseQualityControlClassificationsListOptions(cmd *cobra.Command) (qualityControlClassificationsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	broker, _ := cmd.Flags().GetString("broker")
	name, _ := cmd.Flags().GetString("name")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return qualityControlClassificationsListOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
		Limit:   limit,
		Offset:  offset,
		Broker:  broker,
		Name:    name,
	}, nil
}

type qualityControlClassificationRow struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	BrokerID    string `json:"broker_id,omitempty"`
	BrokerName  string `json:"broker_name,omitempty"`
}

func buildQualityControlClassificationRows(resp jsonAPIResponse) []qualityControlClassificationRow {
	// Build included lookup
	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		key := resourceKey(inc.Type, inc.ID)
		included[key] = inc
	}

	rows := make([]qualityControlClassificationRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := qualityControlClassificationRow{
			ID:          resource.ID,
			Name:        stringAttr(resource.Attributes, "name"),
			Description: stringAttr(resource.Attributes, "description"),
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

func renderQualityControlClassificationsTable(cmd *cobra.Command, rows []qualityControlClassificationRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No quality control classifications found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tNAME\tDESCRIPTION\tBROKER")
	for _, row := range rows {
		broker := row.BrokerName
		if broker == "" && row.BrokerID != "" {
			broker = row.BrokerID
		}
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(row.Name, 25),
			truncateString(row.Description, 30),
			truncateString(broker, 25),
		)
	}
	return writer.Flush()
}
