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

type lineupsListOptions struct {
	BaseURL    string
	Token      string
	JSON       bool
	NoAuth     bool
	Limit      int
	Offset     int
	Sort       string
	Customer   string
	Broker     string
	NameLike   string
	StartAtMin string
	StartAtMax string
}

type lineupRow struct {
	ID           string `json:"id"`
	Name         string `json:"name,omitempty"`
	StartAtMin   string `json:"start_at_min,omitempty"`
	StartAtMax   string `json:"start_at_max,omitempty"`
	CustomerID   string `json:"customer_id,omitempty"`
	CustomerName string `json:"customer_name,omitempty"`
}

func newLineupsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List lineups",
		Long: `List lineups with filtering and pagination.

Output Columns:
  ID          Lineup identifier
  NAME        Lineup name (if set)
  START MIN   Earliest start time (ISO 8601)
  START MAX   Latest start time (ISO 8601)
  CUSTOMER    Customer name or ID

Filters:
  --customer     Filter by customer ID
  --broker       Filter by broker ID
  --name-like    Filter by lineup name (partial match)
  --start-at-min Filter by start time on/after (ISO 8601)
  --start-at-max Filter by start time on/before (ISO 8601)

Sorting:
  Use --sort to specify sort order (e.g. start-at-min,-start-at-max).`,
		Example: `  # List lineups
  xbe view lineups list

  # Filter by customer
  xbe view lineups list --customer 123

  # Filter by broker
  xbe view lineups list --broker 456

  # Filter by name
  xbe view lineups list --name-like "Morning"

  # Filter by time window
  xbe view lineups list --start-at-min 2026-01-01T00:00:00Z --start-at-max 2026-01-02T00:00:00Z

  # JSON output
  xbe view lineups list --json`,
		RunE: runLineupsList,
	}
	initLineupsListFlags(cmd)
	return cmd
}

func init() {
	lineupsCmd.AddCommand(newLineupsListCmd())
}

func initLineupsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 0, "Page size (defaults to server default)")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort order (e.g. start-at-min,-start-at-max)")
	cmd.Flags().String("customer", "", "Filter by customer ID")
	cmd.Flags().String("broker", "", "Filter by broker ID")
	cmd.Flags().String("name-like", "", "Filter by lineup name (partial match)")
	cmd.Flags().String("start-at-min", "", "Filter by start time on/after (ISO 8601)")
	cmd.Flags().String("start-at-max", "", "Filter by start time on/before (ISO 8601)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runLineupsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseLineupsListOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.NoAuth {
		opts.Token = ""
	} else if strings.TrimSpace(opts.Token) == "" {
		if token, _, err := auth.ResolveToken(opts.BaseURL, ""); err == nil {
			opts.Token = token
		} else if !errors.Is(err, auth.ErrNotFound) {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[lineups]", "name,start-at-min,start-at-max,customer")
	query.Set("fields[customers]", "company-name")
	query.Set("include", "customer")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if strings.TrimSpace(opts.Sort) != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[customer]", opts.Customer)
	setFilterIfPresent(query, "filter[broker]", opts.Broker)
	setFilterIfPresent(query, "filter[name_like]", opts.NameLike)
	setFilterIfPresent(query, "filter[start_at_min]", opts.StartAtMin)
	setFilterIfPresent(query, "filter[start_at_max]", opts.StartAtMax)

	body, _, err := client.Get(cmd.Context(), "/v1/lineups", query)
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

	rows := buildLineupRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderLineupsTable(cmd, rows)
}

func parseLineupsListOptions(cmd *cobra.Command) (lineupsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	customer, _ := cmd.Flags().GetString("customer")
	broker, _ := cmd.Flags().GetString("broker")
	nameLike, _ := cmd.Flags().GetString("name-like")
	startAtMin, _ := cmd.Flags().GetString("start-at-min")
	startAtMax, _ := cmd.Flags().GetString("start-at-max")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return lineupsListOptions{
		BaseURL:    baseURL,
		Token:      token,
		JSON:       jsonOut,
		NoAuth:     noAuth,
		Limit:      limit,
		Offset:     offset,
		Sort:       sort,
		Customer:   customer,
		Broker:     broker,
		NameLike:   nameLike,
		StartAtMin: startAtMin,
		StartAtMax: startAtMax,
	}, nil
}

func buildLineupRows(resp jsonAPIResponse) []lineupRow {
	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		key := resourceKey(inc.Type, inc.ID)
		included[key] = inc
	}

	rows := make([]lineupRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := lineupRow{
			ID:         resource.ID,
			Name:       stringAttr(resource.Attributes, "name"),
			StartAtMin: formatDateTime(stringAttr(resource.Attributes, "start-at-min")),
			StartAtMax: formatDateTime(stringAttr(resource.Attributes, "start-at-max")),
		}

		if rel, ok := resource.Relationships["customer"]; ok && rel.Data != nil {
			row.CustomerID = rel.Data.ID
			if customer, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
				row.CustomerName = stringAttr(customer.Attributes, "company-name")
			}
		}

		rows = append(rows, row)
	}

	return rows
}

func renderLineupsTable(cmd *cobra.Command, rows []lineupRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No lineups found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tNAME\tSTART MIN\tSTART MAX\tCUSTOMER")
	for _, row := range rows {
		customer := firstNonEmpty(row.CustomerName, row.CustomerID)
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(row.Name, 30),
			row.StartAtMin,
			row.StartAtMax,
			truncateString(customer, 30),
		)
	}
	return writer.Flush()
}
