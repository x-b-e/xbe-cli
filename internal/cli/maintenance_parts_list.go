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

type maintenancePartsListOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
	Limit   int
	Offset  int
	Sort    string
}

func newMaintenancePartsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List maintenance requirement parts",
		Long: `List maintenance requirement parts with pagination.

Returns a list of parts from the parts catalog that can be used
in maintenance requirements.

Output Columns (table format):
  ID            Unique part identifier
  PART_NUMBER   Part number
  NAME          Part name
  MANUFACTURER  Part manufacturer
  UNIT_COST     Cost per unit

Sorting:
  Use --sort to specify sort order. Prefix with - for descending.
  Default: part-number`,
		Example: `  # List all parts
  xbe view maintenance parts list

  # Paginate results
  xbe view maintenance parts list --limit 50 --offset 100

  # Output as JSON
  xbe view maintenance parts list --json`,
		RunE: runMaintenancePartsList,
	}
	initMaintenancePartsListFlags(cmd)
	return cmd
}

func init() {
	maintenancePartsCmd.AddCommand(newMaintenancePartsListCmd())
}

func initMaintenancePartsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 0, "Page size (defaults to server default)")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort order (default: part-number)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runMaintenancePartsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseMaintenancePartsListOptions(cmd)
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

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}

	// Apply sort
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	} else {
		query.Set("sort", "part-number")
	}

	body, _, err := client.Get(cmd.Context(), "/v1/maintenance-requirement-parts", query)
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

	if opts.JSON {
		rows := buildPartRows(resp)
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderPartsList(cmd, resp)
}

func parseMaintenancePartsListOptions(cmd *cobra.Command) (maintenancePartsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return maintenancePartsListOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
		Limit:   limit,
		Offset:  offset,
		Sort:    sort,
	}, nil
}

func buildPartRows(resp jsonAPIResponse) []partRow {
	rows := make([]partRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		attrs := resource.Attributes

		row := partRow{
			ID:            resource.ID,
			PartNumber:    stringAttr(attrs, "part-number"),
			Name:          strings.TrimSpace(stringAttr(attrs, "name")),
			Description:   strings.TrimSpace(stringAttr(attrs, "description")),
			Manufacturer:  stringAttr(attrs, "manufacturer"),
			UnitCost:      float64Attr(attrs, "unit-cost"),
			UnitOfMeasure: stringAttr(attrs, "unit-of-measure"),
		}

		rows = append(rows, row)
	}
	return rows
}

func renderPartsList(cmd *cobra.Command, resp jsonAPIResponse) error {
	rows := buildPartRows(resp)
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No parts found.")
		return nil
	}

	const partNumMax = 20
	const nameMax = 30
	const mfgMax = 20

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tPART_NUMBER\tNAME\tMANUFACTURER\tUNIT_COST")
	for _, row := range rows {
		cost := "-"
		if row.UnitCost > 0 {
			cost = fmt.Sprintf("$%.2f", row.UnitCost)
		}
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(row.PartNumber, partNumMax),
			truncateString(row.Name, nameMax),
			truncateString(row.Manufacturer, mfgMax),
			cost,
		)
	}
	return writer.Flush()
}
