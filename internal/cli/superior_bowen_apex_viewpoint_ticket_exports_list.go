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

type superiorBowenApexViewpointTicketExportsListOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
	Limit   int
	Offset  int
	Sort    string
}

type superiorBowenApexViewpointTicketExportRow struct {
	ID          string   `json:"id"`
	SaleDateMin string   `json:"sale_date_min,omitempty"`
	SaleDateMax string   `json:"sale_date_max,omitempty"`
	LocationIDs []string `json:"location_ids,omitempty"`
}

func newSuperiorBowenApexViewpointTicketExportsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Superior Bowen Apex Viewpoint ticket exports",
		Long: `List Superior Bowen Apex Viewpoint ticket exports.

Output Columns:
  ID            Export identifier
  SALE DATE MIN Earliest sale date (YYYY-MM-DD)
  SALE DATE MAX Latest sale date (YYYY-MM-DD)
  LOCATION IDS  Location IDs included in the export

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List exports
  xbe view superior-bowen-apex-viewpoint-ticket-exports list

  # JSON output
  xbe view superior-bowen-apex-viewpoint-ticket-exports list --json`,
		Args: cobra.NoArgs,
		RunE: runSuperiorBowenApexViewpointTicketExportsList,
	}
	initSuperiorBowenApexViewpointTicketExportsListFlags(cmd)
	return cmd
}

func init() {
	superiorBowenApexViewpointTicketExportsCmd.AddCommand(newSuperiorBowenApexViewpointTicketExportsListCmd())
}

func initSuperiorBowenApexViewpointTicketExportsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runSuperiorBowenApexViewpointTicketExportsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseSuperiorBowenApexViewpointTicketExportsListOptions(cmd)
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
	query.Set("fields[superior-bowen-apex-viewpoint-ticket-exports]", "sale-date-min,sale-date-max,location-ids")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	body, _, err := client.Get(cmd.Context(), "/v1/superior-bowen-apex-viewpoint-ticket-exports", query)
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

	rows := buildSuperiorBowenApexViewpointTicketExportRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderSuperiorBowenApexViewpointTicketExportsTable(cmd, rows)
}

func parseSuperiorBowenApexViewpointTicketExportsListOptions(cmd *cobra.Command) (superiorBowenApexViewpointTicketExportsListOptions, error) {
	jsonOut, err := cmd.Flags().GetBool("json")
	if err != nil {
		return superiorBowenApexViewpointTicketExportsListOptions{}, err
	}
	noAuth, err := cmd.Flags().GetBool("no-auth")
	if err != nil {
		return superiorBowenApexViewpointTicketExportsListOptions{}, err
	}
	limit, err := cmd.Flags().GetInt("limit")
	if err != nil {
		return superiorBowenApexViewpointTicketExportsListOptions{}, err
	}
	offset, err := cmd.Flags().GetInt("offset")
	if err != nil {
		return superiorBowenApexViewpointTicketExportsListOptions{}, err
	}
	sort, err := cmd.Flags().GetString("sort")
	if err != nil {
		return superiorBowenApexViewpointTicketExportsListOptions{}, err
	}
	baseURL, err := cmd.Flags().GetString("base-url")
	if err != nil {
		return superiorBowenApexViewpointTicketExportsListOptions{}, err
	}
	token, err := cmd.Flags().GetString("token")
	if err != nil {
		return superiorBowenApexViewpointTicketExportsListOptions{}, err
	}

	return superiorBowenApexViewpointTicketExportsListOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
		Limit:   limit,
		Offset:  offset,
		Sort:    sort,
	}, nil
}

func buildSuperiorBowenApexViewpointTicketExportRows(resp jsonAPIResponse) []superiorBowenApexViewpointTicketExportRow {
	rows := make([]superiorBowenApexViewpointTicketExportRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, superiorBowenApexViewpointTicketExportRowFromResource(resource))
	}
	return rows
}

func superiorBowenApexViewpointTicketExportRowFromResource(resource jsonAPIResource) superiorBowenApexViewpointTicketExportRow {
	attrs := resource.Attributes
	return superiorBowenApexViewpointTicketExportRow{
		ID:          resource.ID,
		SaleDateMin: formatDate(stringAttr(attrs, "sale-date-min")),
		SaleDateMax: formatDate(stringAttr(attrs, "sale-date-max")),
		LocationIDs: stringSliceAttr(attrs, "location-ids"),
	}
}

func renderSuperiorBowenApexViewpointTicketExportsTable(cmd *cobra.Command, rows []superiorBowenApexViewpointTicketExportRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No Superior Bowen Apex Viewpoint ticket exports found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, 32, 0)
	fmt.Fprintln(writer, "ID\tSALE DATE MIN\tSALE DATE MAX\tLOCATION IDS")

	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\n",
			row.ID,
			row.SaleDateMin,
			row.SaleDateMax,
			strings.Join(row.LocationIDs, ","),
		)
	}

	return writer.Flush()
}
