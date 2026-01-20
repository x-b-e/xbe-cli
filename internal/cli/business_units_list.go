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

type businessUnitsListOptions struct {
	BaseURL         string
	Token           string
	JSON            bool
	NoAuth          bool
	Limit           int
	Offset          int
	Name            string
	Broker          string
	Parent          string
	WithChildren    string
	WithoutChildren string
}

type businessUnitRow struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Broker   string `json:"broker,omitempty"`
	BrokerID string `json:"broker_id,omitempty"`
}

func newBusinessUnitsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List business units",
		Long: `List business units with filtering and pagination.

Returns a list of business units. Use this to look up business unit IDs
for filtering job production plans.

Output Columns:
  ID      Business unit identifier (use this for --business-unit filter)
  NAME    Business unit name
  BROKER  Broker name`,
		Example: `  # List business units
  xbe view business-units list

  # Search by name
  xbe view business-units list --name "Paving"

  # Output as JSON
  xbe view business-units list --json`,
		RunE: runBusinessUnitsList,
	}
	initBusinessUnitsListFlags(cmd)
	return cmd
}

func init() {
	businessUnitsCmd.AddCommand(newBusinessUnitsListCmd())
}

func initBusinessUnitsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("name", "", "Filter by name (partial match)")
	cmd.Flags().String("broker", "", "Filter by broker ID (comma-separated for multiple)")
	cmd.Flags().String("parent", "", "Filter by parent business unit ID (comma-separated for multiple)")
	cmd.Flags().String("with-children", "", "Filter by whether unit has children (true/false)")
	cmd.Flags().String("without-children", "", "Filter by whether unit has no children (true/false)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runBusinessUnitsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseBusinessUnitsListOptions(cmd)
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
	query.Set("sort", "company-name")
	query.Set("fields[business-units]", "company-name,broker")
	query.Set("fields[brokers]", "company-name")
	query.Set("include", "broker")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Name != "" {
		query.Set("filter[company-name]", opts.Name)
	}
	setFilterIfPresent(query, "filter[broker]", opts.Broker)
	setFilterIfPresent(query, "filter[parent]", opts.Parent)
	setFilterIfPresent(query, "filter[with-children]", opts.WithChildren)
	setFilterIfPresent(query, "filter[without-children]", opts.WithoutChildren)

	body, _, err := client.Get(cmd.Context(), "/v1/business-units", query)
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

	rows := buildBusinessUnitRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderBusinessUnitsTable(cmd, rows)
}

func parseBusinessUnitsListOptions(cmd *cobra.Command) (businessUnitsListOptions, error) {
	jsonOut, err := cmd.Flags().GetBool("json")
	if err != nil {
		return businessUnitsListOptions{}, err
	}
	noAuth, err := cmd.Flags().GetBool("no-auth")
	if err != nil {
		return businessUnitsListOptions{}, err
	}
	limit, err := cmd.Flags().GetInt("limit")
	if err != nil {
		return businessUnitsListOptions{}, err
	}
	offset, err := cmd.Flags().GetInt("offset")
	if err != nil {
		return businessUnitsListOptions{}, err
	}
	name, err := cmd.Flags().GetString("name")
	if err != nil {
		return businessUnitsListOptions{}, err
	}
	broker, err := cmd.Flags().GetString("broker")
	if err != nil {
		return businessUnitsListOptions{}, err
	}
	parent, err := cmd.Flags().GetString("parent")
	if err != nil {
		return businessUnitsListOptions{}, err
	}
	withChildren, err := cmd.Flags().GetString("with-children")
	if err != nil {
		return businessUnitsListOptions{}, err
	}
	withoutChildren, err := cmd.Flags().GetString("without-children")
	if err != nil {
		return businessUnitsListOptions{}, err
	}
	baseURL, err := cmd.Flags().GetString("base-url")
	if err != nil {
		return businessUnitsListOptions{}, err
	}
	token, err := cmd.Flags().GetString("token")
	if err != nil {
		return businessUnitsListOptions{}, err
	}

	return businessUnitsListOptions{
		BaseURL:         baseURL,
		Token:           token,
		JSON:            jsonOut,
		NoAuth:          noAuth,
		Limit:           limit,
		Offset:          offset,
		Name:            name,
		Broker:          broker,
		Parent:          parent,
		WithChildren:    withChildren,
		WithoutChildren: withoutChildren,
	}, nil
}

func buildBusinessUnitRows(resp jsonAPIResponse) []businessUnitRow {
	// Build included lookup
	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		key := resourceKey(inc.Type, inc.ID)
		included[key] = inc
	}

	rows := make([]businessUnitRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := businessUnitRow{
			ID:   resource.ID,
			Name: stringAttr(resource.Attributes, "company-name"),
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

func renderBusinessUnitsTable(cmd *cobra.Command, rows []businessUnitRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No business units found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tNAME\tBROKER")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\n",
			row.ID,
			truncateString(row.Name, 40),
			truncateString(row.Broker, 25),
		)
	}
	return writer.Flush()
}
