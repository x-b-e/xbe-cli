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

type projectOfficesListOptions struct {
	BaseURL      string
	Token        string
	JSON         bool
	NoAuth       bool
	Limit        int
	Offset       int
	Broker       string
	Name         string
	Abbreviation string
	IsActive     string
	Q            string
}

type projectOfficeRow struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Abbreviation string `json:"abbreviation,omitempty"`
	IsActive     bool   `json:"is_active"`
	Broker       string `json:"broker,omitempty"`
	BrokerID     string `json:"broker_id,omitempty"`
}

func newProjectOfficesListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List project offices",
		Long: `List project offices with filtering and pagination.

Project offices are geographic or organizational divisions used to group
projects and transport orders by region or office location.

Output Columns:
  ID            Project office identifier
  NAME          Project office name
  ABBREVIATION  Short code for the project office
  ACTIVE        Whether the project office is active
  BROKER        Broker name

Filters:
  --broker         Filter by broker ID
  --name           Filter by name (partial match, case-insensitive)
  --abbreviation   Filter by abbreviation (partial match, case-insensitive)
  --is-active      Filter by active status (true/false)
  --q              General search across name and abbreviation`,
		Example: `  # List all project offices
  xbe view project-offices list

  # Filter by broker
  xbe view project-offices list --broker 123

  # Search by name
  xbe view project-offices list --name "Chicago"

  # Search by abbreviation
  xbe view project-offices list --abbreviation "CHI"

  # Show only active project offices
  xbe view project-offices list --is-active true

  # General search
  xbe view project-offices list --q "west"

  # Output as JSON
  xbe view project-offices list --json`,
		RunE: runProjectOfficesList,
	}
	initProjectOfficesListFlags(cmd)
	return cmd
}

func init() {
	projectOfficesCmd.AddCommand(newProjectOfficesListCmd())
}

func initProjectOfficesListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("broker", "", "Filter by broker ID")
	cmd.Flags().String("name", "", "Filter by name (partial match)")
	cmd.Flags().String("abbreviation", "", "Filter by abbreviation (partial match)")
	cmd.Flags().String("is-active", "", "Filter by active status (true/false)")
	cmd.Flags().String("q", "", "General search (name or abbreviation)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runProjectOfficesList(cmd *cobra.Command, _ []string) error {
	opts, err := parseProjectOfficesListOptions(cmd)
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
	query.Set("fields[project-offices]", "name,abbreviation,is-active,broker")
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
	setFilterIfPresent(query, "filter[is-active]", opts.IsActive)
	setFilterIfPresent(query, "filter[q]", opts.Q)

	body, _, err := client.Get(cmd.Context(), "/v1/project-offices", query)
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

	rows := buildProjectOfficeRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderProjectOfficesTable(cmd, rows)
}

func parseProjectOfficesListOptions(cmd *cobra.Command) (projectOfficesListOptions, error) {
	jsonOut, err := cmd.Flags().GetBool("json")
	if err != nil {
		return projectOfficesListOptions{}, err
	}
	noAuth, err := cmd.Flags().GetBool("no-auth")
	if err != nil {
		return projectOfficesListOptions{}, err
	}
	limit, err := cmd.Flags().GetInt("limit")
	if err != nil {
		return projectOfficesListOptions{}, err
	}
	offset, err := cmd.Flags().GetInt("offset")
	if err != nil {
		return projectOfficesListOptions{}, err
	}
	broker, err := cmd.Flags().GetString("broker")
	if err != nil {
		return projectOfficesListOptions{}, err
	}
	name, err := cmd.Flags().GetString("name")
	if err != nil {
		return projectOfficesListOptions{}, err
	}
	abbreviation, err := cmd.Flags().GetString("abbreviation")
	if err != nil {
		return projectOfficesListOptions{}, err
	}
	isActive, err := cmd.Flags().GetString("is-active")
	if err != nil {
		return projectOfficesListOptions{}, err
	}
	q, err := cmd.Flags().GetString("q")
	if err != nil {
		return projectOfficesListOptions{}, err
	}
	baseURL, err := cmd.Flags().GetString("base-url")
	if err != nil {
		return projectOfficesListOptions{}, err
	}
	token, err := cmd.Flags().GetString("token")
	if err != nil {
		return projectOfficesListOptions{}, err
	}

	return projectOfficesListOptions{
		BaseURL:      baseURL,
		Token:        token,
		JSON:         jsonOut,
		NoAuth:       noAuth,
		Limit:        limit,
		Offset:       offset,
		Broker:       broker,
		Name:         name,
		Abbreviation: abbreviation,
		IsActive:     isActive,
		Q:            q,
	}, nil
}

func buildProjectOfficeRows(resp jsonAPIResponse) []projectOfficeRow {
	// Build included lookup
	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		key := resourceKey(inc.Type, inc.ID)
		included[key] = inc
	}

	rows := make([]projectOfficeRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := projectOfficeRow{
			ID:           resource.ID,
			Name:         stringAttr(resource.Attributes, "name"),
			Abbreviation: stringAttr(resource.Attributes, "abbreviation"),
			IsActive:     boolAttr(resource.Attributes, "is-active"),
		}

		// Resolve broker
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

func renderProjectOfficesTable(cmd *cobra.Command, rows []projectOfficeRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No project offices found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tNAME\tABBREVIATION\tACTIVE\tBROKER")
	for _, row := range rows {
		active := "no"
		if row.IsActive {
			active = "yes"
		}
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(row.Name, 35),
			truncateString(row.Abbreviation, 10),
			active,
			truncateString(row.Broker, 25),
		)
	}
	return writer.Flush()
}
