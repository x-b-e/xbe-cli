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

type lehmanRobertsApexViewpointTicketExportsListOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
	Limit   int
	Offset  int
	Sort    string
}

type lehmanRobertsApexViewpointTicketExportRow struct {
	ID            string   `json:"id"`
	TemplateName  string   `json:"template_name,omitempty"`
	SaleDateMin   string   `json:"sale_date_min,omitempty"`
	SaleDateMax   string   `json:"sale_date_max,omitempty"`
	LocationIDs   []string `json:"location_ids,omitempty"`
	OmitHeaderRow bool     `json:"omit_header_row"`
}

func newLehmanRobertsApexViewpointTicketExportsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Lehman Roberts Apex Viewpoint ticket exports",
		Long: `List Lehman Roberts Apex Viewpoint ticket exports.

Output Columns:
  ID            Export identifier
  TEMPLATE      Viewpoint template name
  SALE DATE MIN Earliest sale date (YYYY-MM-DD)
  SALE DATE MAX Latest sale date (YYYY-MM-DD)
  LOCATION IDS  Location IDs included in the export
  OMIT HEADER   Whether the header row is omitted

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List exports
  xbe view lehman-roberts-apex-viewpoint-ticket-exports list

  # JSON output
  xbe view lehman-roberts-apex-viewpoint-ticket-exports list --json`,
		Args: cobra.NoArgs,
		RunE: runLehmanRobertsApexViewpointTicketExportsList,
	}
	initLehmanRobertsApexViewpointTicketExportsListFlags(cmd)
	return cmd
}

func init() {
	lehmanRobertsApexViewpointTicketExportsCmd.AddCommand(newLehmanRobertsApexViewpointTicketExportsListCmd())
}

func initLehmanRobertsApexViewpointTicketExportsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runLehmanRobertsApexViewpointTicketExportsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseLehmanRobertsApexViewpointTicketExportsListOptions(cmd)
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
	query.Set("fields[lehman-roberts-apex-viewpoint-ticket-exports]", "sale-date-min,sale-date-max,location-ids,template-name,omit-header-row")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	body, _, err := client.Get(cmd.Context(), "/v1/lehman-roberts-apex-viewpoint-ticket-exports", query)
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

	rows := buildLehmanRobertsApexViewpointTicketExportRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderLehmanRobertsApexViewpointTicketExportsTable(cmd, rows)
}

func parseLehmanRobertsApexViewpointTicketExportsListOptions(cmd *cobra.Command) (lehmanRobertsApexViewpointTicketExportsListOptions, error) {
	jsonOut, err := cmd.Flags().GetBool("json")
	if err != nil {
		return lehmanRobertsApexViewpointTicketExportsListOptions{}, err
	}
	noAuth, err := cmd.Flags().GetBool("no-auth")
	if err != nil {
		return lehmanRobertsApexViewpointTicketExportsListOptions{}, err
	}
	limit, err := cmd.Flags().GetInt("limit")
	if err != nil {
		return lehmanRobertsApexViewpointTicketExportsListOptions{}, err
	}
	offset, err := cmd.Flags().GetInt("offset")
	if err != nil {
		return lehmanRobertsApexViewpointTicketExportsListOptions{}, err
	}
	sort, err := cmd.Flags().GetString("sort")
	if err != nil {
		return lehmanRobertsApexViewpointTicketExportsListOptions{}, err
	}
	baseURL, err := cmd.Flags().GetString("base-url")
	if err != nil {
		return lehmanRobertsApexViewpointTicketExportsListOptions{}, err
	}
	token, err := cmd.Flags().GetString("token")
	if err != nil {
		return lehmanRobertsApexViewpointTicketExportsListOptions{}, err
	}

	return lehmanRobertsApexViewpointTicketExportsListOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
		Limit:   limit,
		Offset:  offset,
		Sort:    sort,
	}, nil
}

func buildLehmanRobertsApexViewpointTicketExportRows(resp jsonAPIResponse) []lehmanRobertsApexViewpointTicketExportRow {
	rows := make([]lehmanRobertsApexViewpointTicketExportRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, lehmanRobertsApexViewpointTicketExportRowFromResource(resource))
	}
	return rows
}

func lehmanRobertsApexViewpointTicketExportRowFromResource(resource jsonAPIResource) lehmanRobertsApexViewpointTicketExportRow {
	attrs := resource.Attributes
	return lehmanRobertsApexViewpointTicketExportRow{
		ID:            resource.ID,
		TemplateName:  stringAttr(attrs, "template-name"),
		SaleDateMin:   formatDate(stringAttr(attrs, "sale-date-min")),
		SaleDateMax:   formatDate(stringAttr(attrs, "sale-date-max")),
		LocationIDs:   stringSliceAttr(attrs, "location-ids"),
		OmitHeaderRow: boolAttr(attrs, "omit-header-row"),
	}
}

func renderLehmanRobertsApexViewpointTicketExportsTable(cmd *cobra.Command, rows []lehmanRobertsApexViewpointTicketExportRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No Lehman Roberts Apex Viewpoint ticket exports found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, 32, 0)
	fmt.Fprintln(writer, "ID\tTEMPLATE\tSALE DATE MIN\tSALE DATE MAX\tLOCATION IDS\tOMIT HEADER")

	for _, row := range rows {
		omitHeader := "no"
		if row.OmitHeaderRow {
			omitHeader = "yes"
		}
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.TemplateName,
			row.SaleDateMin,
			row.SaleDateMax,
			strings.Join(row.LocationIDs, ","),
			omitHeader,
		)
	}

	return writer.Flush()
}
