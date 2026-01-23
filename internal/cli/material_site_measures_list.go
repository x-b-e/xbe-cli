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

type materialSiteMeasuresListOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
	Limit   int
	Offset  int
	Slug    string
}

type materialSiteMeasureRow struct {
	ID                   string `json:"id"`
	Slug                 string `json:"slug"`
	Name                 string `json:"name"`
	ValidReadingValueMin string `json:"valid_reading_value_min,omitempty"`
	ValidReadingValueMax string `json:"valid_reading_value_max,omitempty"`
}

func newMaterialSiteMeasuresListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List material site measures",
		Long: `List material site measures with filtering and pagination.

Material site measures define the measurement types used for
material site readings and validation ranges.

Output Columns:
  ID        Material site measure identifier
  SLUG      URL-friendly identifier
  NAME      Measure name
  MIN       Minimum valid reading value
  MAX       Maximum valid reading value

Filters:
  --slug  Filter by slug (exact match)`,
		Example: `  # List all material site measures
  xbe view material-site-measures list

  # Filter by slug
  xbe view material-site-measures list --slug "mixing-temperature"

  # Output as JSON
  xbe view material-site-measures list --json`,
		RunE: runMaterialSiteMeasuresList,
	}
	initMaterialSiteMeasuresListFlags(cmd)
	return cmd
}

func init() {
	materialSiteMeasuresCmd.AddCommand(newMaterialSiteMeasuresListCmd())
}

func initMaterialSiteMeasuresListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("slug", "", "Filter by slug (exact match)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runMaterialSiteMeasuresList(cmd *cobra.Command, _ []string) error {
	opts, err := parseMaterialSiteMeasuresListOptions(cmd)
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
	query.Set("fields[material-site-measures]", "slug,name,valid-reading-value-min,valid-reading-value-max")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	setFilterIfPresent(query, "filter[slug]", opts.Slug)

	body, _, err := client.Get(cmd.Context(), "/v1/material-site-measures", query)
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

	rows := buildMaterialSiteMeasureRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderMaterialSiteMeasuresTable(cmd, rows)
}

func parseMaterialSiteMeasuresListOptions(cmd *cobra.Command) (materialSiteMeasuresListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	slug, _ := cmd.Flags().GetString("slug")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return materialSiteMeasuresListOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
		Limit:   limit,
		Offset:  offset,
		Slug:    slug,
	}, nil
}

func buildMaterialSiteMeasureRows(resp jsonAPIResponse) []materialSiteMeasureRow {
	rows := make([]materialSiteMeasureRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := materialSiteMeasureRow{
			ID:                   resource.ID,
			Slug:                 stringAttr(resource.Attributes, "slug"),
			Name:                 stringAttr(resource.Attributes, "name"),
			ValidReadingValueMin: stringAttr(resource.Attributes, "valid-reading-value-min"),
			ValidReadingValueMax: stringAttr(resource.Attributes, "valid-reading-value-max"),
		}
		rows = append(rows, row)
	}
	return rows
}

func renderMaterialSiteMeasuresTable(cmd *cobra.Command, rows []materialSiteMeasureRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No material site measures found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tSLUG\tNAME\tMIN\tMAX")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(row.Slug, 22),
			truncateString(row.Name, 28),
			truncateString(row.ValidReadingValueMin, 12),
			truncateString(row.ValidReadingValueMax, 12),
		)
	}
	return writer.Flush()
}
