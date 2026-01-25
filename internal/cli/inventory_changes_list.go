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

type inventoryChangesListOptions struct {
	BaseURL          string
	Token            string
	JSON             bool
	NoAuth           bool
	Limit            int
	Offset           int
	Sort             string
	MaterialSite     string
	MaterialType     string
	MaterialSupplier string
	ForecastStartAt  string
}

type inventoryChangeRow struct {
	ID                 string `json:"id"`
	MaterialSiteID     string `json:"material_site_id,omitempty"`
	MaterialSite       string `json:"material_site,omitempty"`
	MaterialTypeID     string `json:"material_type_id,omitempty"`
	MaterialType       string `json:"material_type,omitempty"`
	EstimateAt         string `json:"estimate_at,omitempty"`
	ForecastStartAt    string `json:"forecast_start_at,omitempty"`
	CalculatedAt       string `json:"calculated_at,omitempty"`
	StartingAmountTons string `json:"starting_amount_tons,omitempty"`
	EndingAmountTons   string `json:"ending_amount_tons,omitempty"`
}

func newInventoryChangesListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List inventory changes",
		Long: `List inventory changes with filtering and pagination.

Output Columns:
  ID              Inventory change identifier
  SITE            Material site name
  MATERIAL        Material type name
  ESTIMATE AT     Estimate timestamp
  FORECAST START  Forecast start timestamp
  CALCULATED AT   Calculation timestamp
  START TONS      Starting inventory amount (tons)
  END TONS        Ending inventory amount (tons)

Filters:
  --material-site      Filter by material site ID (comma-separated for multiple)
  --material-type      Filter by material type ID (comma-separated for multiple)
  --material-supplier  Filter by material supplier ID (comma-separated for multiple)
  --forecast-start-at  Filter by forecast start timestamp (ISO 8601)

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List inventory changes
  xbe view inventory-changes list

  # Filter by material site
  xbe view inventory-changes list --material-site 123

  # Filter by material type
  xbe view inventory-changes list --material-type 456

  # Filter by material supplier
  xbe view inventory-changes list --material-supplier 789

  # Filter by forecast start timestamp
  xbe view inventory-changes list --forecast-start-at 2025-01-01T00:00:00Z

  # Output as JSON
  xbe view inventory-changes list --json`,
		Args: cobra.NoArgs,
		RunE: runInventoryChangesList,
	}
	initInventoryChangesListFlags(cmd)
	return cmd
}

func init() {
	inventoryChangesCmd.AddCommand(newInventoryChangesListCmd())
}

func initInventoryChangesListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("material-site", "", "Filter by material site ID (comma-separated for multiple)")
	cmd.Flags().String("material-type", "", "Filter by material type ID (comma-separated for multiple)")
	cmd.Flags().String("material-supplier", "", "Filter by material supplier ID (comma-separated for multiple)")
	cmd.Flags().String("forecast-start-at", "", "Filter by forecast start timestamp (ISO 8601)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runInventoryChangesList(cmd *cobra.Command, _ []string) error {
	opts, err := parseInventoryChangesListOptions(cmd)
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
	query.Set("fields[inventory-changes]", "estimate-at,forecast-start-at,calculated-at,starting-amount-tons,ending-amount-tons,material-site,material-type")
	query.Set("fields[material-sites]", "name")
	query.Set("fields[material-types]", "name")
	query.Set("include", "material-site,material-type")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[material-site]", opts.MaterialSite)
	setFilterIfPresent(query, "filter[material-type]", opts.MaterialType)
	setFilterIfPresent(query, "filter[material-supplier]", opts.MaterialSupplier)
	setFilterIfPresent(query, "filter[forecast-start-at]", opts.ForecastStartAt)

	body, _, err := client.Get(cmd.Context(), "/v1/inventory-changes", query)
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

	rows := buildInventoryChangeRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderInventoryChangesTable(cmd, rows)
}

func parseInventoryChangesListOptions(cmd *cobra.Command) (inventoryChangesListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	materialSite, _ := cmd.Flags().GetString("material-site")
	materialType, _ := cmd.Flags().GetString("material-type")
	materialSupplier, _ := cmd.Flags().GetString("material-supplier")
	forecastStartAt, _ := cmd.Flags().GetString("forecast-start-at")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return inventoryChangesListOptions{
		BaseURL:          baseURL,
		Token:            token,
		JSON:             jsonOut,
		NoAuth:           noAuth,
		Limit:            limit,
		Offset:           offset,
		Sort:             sort,
		MaterialSite:     materialSite,
		MaterialType:     materialType,
		MaterialSupplier: materialSupplier,
		ForecastStartAt:  forecastStartAt,
	}, nil
}

func buildInventoryChangeRows(resp jsonAPIResponse) []inventoryChangeRow {
	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	rows := make([]inventoryChangeRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, buildInventoryChangeRowFromResource(resource, included))
	}

	return rows
}

func buildInventoryChangeRowFromResource(resource jsonAPIResource, included map[string]jsonAPIResource) inventoryChangeRow {
	attrs := resource.Attributes
	row := inventoryChangeRow{
		ID:                 resource.ID,
		EstimateAt:         formatDateTime(stringAttr(attrs, "estimate-at")),
		ForecastStartAt:    formatDateTime(stringAttr(attrs, "forecast-start-at")),
		CalculatedAt:       formatDateTime(stringAttr(attrs, "calculated-at")),
		StartingAmountTons: stringAttr(attrs, "starting-amount-tons"),
		EndingAmountTons:   stringAttr(attrs, "ending-amount-tons"),
	}

	if rel, ok := resource.Relationships["material-site"]; ok && rel.Data != nil {
		row.MaterialSiteID = rel.Data.ID
		if ms, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			row.MaterialSite = stringAttr(ms.Attributes, "name")
		}
	}

	if rel, ok := resource.Relationships["material-type"]; ok && rel.Data != nil {
		row.MaterialTypeID = rel.Data.ID
		if mt, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			row.MaterialType = stringAttr(mt.Attributes, "name")
		}
	}

	return row
}

func renderInventoryChangesTable(cmd *cobra.Command, rows []inventoryChangeRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No inventory changes found.")
		return nil
	}

	const (
		maxSite     = 24
		maxMaterial = 24
	)

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tSITE\tMATERIAL\tESTIMATE AT\tFORECAST START\tCALCULATED AT\tSTART TONS\tEND TONS")
	for _, row := range rows {
		siteLabel := firstNonEmpty(row.MaterialSite, row.MaterialSiteID)
		materialLabel := firstNonEmpty(row.MaterialType, row.MaterialTypeID)
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(siteLabel, maxSite),
			truncateString(materialLabel, maxMaterial),
			row.EstimateAt,
			row.ForecastStartAt,
			row.CalculatedAt,
			row.StartingAmountTons,
			row.EndingAmountTons,
		)
	}
	return writer.Flush()
}
