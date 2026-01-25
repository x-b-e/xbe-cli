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

type projectDivisionsListOptions struct {
	BaseURL      string
	Token        string
	JSON         bool
	NoAuth       bool
	Limit        int
	Offset       int
	Broker       string
	Name         string
	Abbreviation string
	Q            string
}

type projectDivisionRow struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Abbreviation string `json:"abbreviation,omitempty"`
	Broker       string `json:"broker,omitempty"`
	BrokerID     string `json:"broker_id,omitempty"`
}

func newProjectDivisionsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List project divisions",
		Long: `List project divisions with filtering and pagination.

Project divisions are organizational units that group projects by business
division or department for reporting and access control.

Output Columns:
  ID            Project division identifier
  NAME          Division name
  ABBREVIATION  Short code
  BROKER        Broker name

Filters:
  --broker         Filter by broker ID
  --name           Filter by name (partial match, case-insensitive)
  --abbreviation   Filter by abbreviation (partial match, case-insensitive)
  --q              General search across name and abbreviation`,
		Example: `  # List all project divisions
  xbe view project-divisions list

  # Filter by broker
  xbe view project-divisions list --broker 123

  # Search by name
  xbe view project-divisions list --name "North"

  # General search
  xbe view project-divisions list --q "south"

  # Output as JSON
  xbe view project-divisions list --json`,
		RunE: runProjectDivisionsList,
	}
	initProjectDivisionsListFlags(cmd)
	return cmd
}

func init() {
	projectDivisionsCmd.AddCommand(newProjectDivisionsListCmd())
}

func initProjectDivisionsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("broker", "", "Filter by broker ID")
	cmd.Flags().String("name", "", "Filter by name (partial match)")
	cmd.Flags().String("abbreviation", "", "Filter by abbreviation (partial match)")
	cmd.Flags().String("q", "", "General search (name or abbreviation)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runProjectDivisionsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseProjectDivisionsListOptions(cmd)
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
	query.Set("fields[project-divisions]", "name,abbreviation,broker")
	query.Set("fields[brokers]", "company-name")
	query.Set("include", "broker")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	setFilterIfPresent(query, "filter[broker]", opts.Broker)
	setFilterIfPresent(query, "filter[name]", opts.Name)
	setFilterIfPresent(query, "filter[abbreviation]", opts.Abbreviation)
	setFilterIfPresent(query, "filter[q]", opts.Q)

	body, _, err := client.Get(cmd.Context(), "/v1/project-divisions", query)
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

	rows := buildProjectDivisionRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderProjectDivisionsTable(cmd, rows)
}

func parseProjectDivisionsListOptions(cmd *cobra.Command) (projectDivisionsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	broker, _ := cmd.Flags().GetString("broker")
	name, _ := cmd.Flags().GetString("name")
	abbreviation, _ := cmd.Flags().GetString("abbreviation")
	q, _ := cmd.Flags().GetString("q")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return projectDivisionsListOptions{
		BaseURL:      baseURL,
		Token:        token,
		JSON:         jsonOut,
		NoAuth:       noAuth,
		Limit:        limit,
		Offset:       offset,
		Broker:       broker,
		Name:         name,
		Abbreviation: abbreviation,
		Q:            q,
	}, nil
}

func buildProjectDivisionRows(resp jsonAPIResponse) []projectDivisionRow {
	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		key := resourceKey(inc.Type, inc.ID)
		included[key] = inc
	}

	rows := make([]projectDivisionRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := projectDivisionRow{
			ID:           resource.ID,
			Name:         stringAttr(resource.Attributes, "name"),
			Abbreviation: stringAttr(resource.Attributes, "abbreviation"),
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

func renderProjectDivisionsTable(cmd *cobra.Command, rows []projectDivisionRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No project divisions found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tNAME\tABBREVIATION\tBROKER")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(row.Name, 30),
			truncateString(row.Abbreviation, 10),
			truncateString(row.Broker, 25),
		)
	}
	return writer.Flush()
}
