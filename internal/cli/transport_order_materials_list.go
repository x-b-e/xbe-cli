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

type transportOrderMaterialsListOptions struct {
	BaseURL          string
	Token            string
	JSON             bool
	NoAuth           bool
	Limit            int
	Offset           int
	Sort             string
	TransportOrder   string
	MaterialType     string
	UnitOfMeasure    string
	QuantityExplicit string
}

type transportOrderMaterialRow struct {
	ID                   string `json:"id"`
	TransportOrderID     string `json:"transport_order_id,omitempty"`
	TransportOrderNumber string `json:"transport_order_number,omitempty"`
	MaterialTypeID       string `json:"material_type_id,omitempty"`
	MaterialType         string `json:"material_type,omitempty"`
	UnitOfMeasureID      string `json:"unit_of_measure_id,omitempty"`
	UnitOfMeasure        string `json:"unit_of_measure,omitempty"`
	Quantity             string `json:"quantity,omitempty"`
	QuantityExplicit     string `json:"quantity_explicit,omitempty"`
}

func newTransportOrderMaterialsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List transport order materials",
		Long: `List transport order materials with filtering and pagination.

Output Columns:
  ID        Transport order material identifier
  ORDER     Transport order number (or ID)
  MATERIAL  Material type
  UNIT      Unit of measure
  QUANTITY  Calculated quantity
  EXPLICIT  Explicit quantity (if set)

Filters:
  --transport-order   Filter by transport order ID
  --material-type     Filter by material type ID
  --unit-of-measure   Filter by unit of measure ID
  --quantity-explicit Filter by explicit quantity

Sorting:
  Use --sort to specify sort order. Prefix with - for descending.

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List transport order materials
  xbe view transport-order-materials list

  # Filter by transport order
  xbe view transport-order-materials list --transport-order 123

  # Filter by material type
  xbe view transport-order-materials list --material-type 456

  # Output as JSON
  xbe view transport-order-materials list --json`,
		RunE: runTransportOrderMaterialsList,
	}
	initTransportOrderMaterialsListFlags(cmd)
	return cmd
}

func init() {
	transportOrderMaterialsCmd.AddCommand(newTransportOrderMaterialsListCmd())
}

func initTransportOrderMaterialsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("transport-order", "", "Filter by transport order ID")
	cmd.Flags().String("material-type", "", "Filter by material type ID")
	cmd.Flags().String("unit-of-measure", "", "Filter by unit of measure ID")
	cmd.Flags().String("quantity-explicit", "", "Filter by explicit quantity")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runTransportOrderMaterialsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseTransportOrderMaterialsListOptions(cmd)
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
	query.Set("fields[transport-order-materials]", "quantity-explicit,quantity-implicit-cached,quantity,transport-order,material-type,unit-of-measure")
	query.Set("include", "transport-order,material-type,unit-of-measure")
	query.Set("fields[transport-orders]", "external-order-number")
	query.Set("fields[material-types]", "name,display-name")
	query.Set("fields[unit-of-measures]", "name,abbreviation")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if strings.TrimSpace(opts.Sort) != "" {
		query.Set("sort", opts.Sort)
	} else {
		query.Set("sort", "id")
	}

	setFilterIfPresent(query, "filter[transport_order]", opts.TransportOrder)
	setFilterIfPresent(query, "filter[material_type]", opts.MaterialType)
	setFilterIfPresent(query, "filter[unit_of_measure]", opts.UnitOfMeasure)
	setFilterIfPresent(query, "filter[quantity_explicit]", opts.QuantityExplicit)

	body, _, err := client.Get(cmd.Context(), "/v1/transport-order-materials", query)
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

	rows := buildTransportOrderMaterialRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderTransportOrderMaterialsTable(cmd, rows)
}

func parseTransportOrderMaterialsListOptions(cmd *cobra.Command) (transportOrderMaterialsListOptions, error) {
	jsonOut, err := cmd.Flags().GetBool("json")
	if err != nil {
		return transportOrderMaterialsListOptions{}, err
	}
	noAuth, err := cmd.Flags().GetBool("no-auth")
	if err != nil {
		return transportOrderMaterialsListOptions{}, err
	}
	limit, err := cmd.Flags().GetInt("limit")
	if err != nil {
		return transportOrderMaterialsListOptions{}, err
	}
	offset, err := cmd.Flags().GetInt("offset")
	if err != nil {
		return transportOrderMaterialsListOptions{}, err
	}
	sort, err := cmd.Flags().GetString("sort")
	if err != nil {
		return transportOrderMaterialsListOptions{}, err
	}
	transportOrder, err := cmd.Flags().GetString("transport-order")
	if err != nil {
		return transportOrderMaterialsListOptions{}, err
	}
	materialType, err := cmd.Flags().GetString("material-type")
	if err != nil {
		return transportOrderMaterialsListOptions{}, err
	}
	unitOfMeasure, err := cmd.Flags().GetString("unit-of-measure")
	if err != nil {
		return transportOrderMaterialsListOptions{}, err
	}
	quantityExplicit, err := cmd.Flags().GetString("quantity-explicit")
	if err != nil {
		return transportOrderMaterialsListOptions{}, err
	}
	baseURL, err := cmd.Flags().GetString("base-url")
	if err != nil {
		return transportOrderMaterialsListOptions{}, err
	}
	token, err := cmd.Flags().GetString("token")
	if err != nil {
		return transportOrderMaterialsListOptions{}, err
	}

	return transportOrderMaterialsListOptions{
		BaseURL:          baseURL,
		Token:            token,
		JSON:             jsonOut,
		NoAuth:           noAuth,
		Limit:            limit,
		Offset:           offset,
		Sort:             sort,
		TransportOrder:   transportOrder,
		MaterialType:     materialType,
		UnitOfMeasure:    unitOfMeasure,
		QuantityExplicit: quantityExplicit,
	}, nil
}

func buildTransportOrderMaterialRows(resp jsonAPIResponse) []transportOrderMaterialRow {
	included := make(map[string]jsonAPIResource, len(resp.Included))
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	rows := make([]transportOrderMaterialRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, buildTransportOrderMaterialRow(resource, included))
	}

	return rows
}

func buildTransportOrderMaterialRow(resource jsonAPIResource, included map[string]jsonAPIResource) transportOrderMaterialRow {
	attrs := resource.Attributes
	row := transportOrderMaterialRow{
		ID:               resource.ID,
		Quantity:         stringAttr(attrs, "quantity"),
		QuantityExplicit: stringAttr(attrs, "quantity-explicit"),
	}

	if rel, ok := resource.Relationships["transport-order"]; ok && rel.Data != nil {
		row.TransportOrderID = rel.Data.ID
		if order, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			row.TransportOrderNumber = firstNonEmpty(
				stringAttr(order.Attributes, "external-order-number"),
				stringAttr(order.Attributes, "order-number"),
			)
		}
	}

	if rel, ok := resource.Relationships["material-type"]; ok && rel.Data != nil {
		row.MaterialTypeID = rel.Data.ID
		if mt, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			row.MaterialType = firstNonEmpty(
				stringAttr(mt.Attributes, "display-name"),
				stringAttr(mt.Attributes, "name"),
			)
		}
	}

	if rel, ok := resource.Relationships["unit-of-measure"]; ok && rel.Data != nil {
		row.UnitOfMeasureID = rel.Data.ID
		if uom, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			row.UnitOfMeasure = firstNonEmpty(
				stringAttr(uom.Attributes, "abbreviation"),
				stringAttr(uom.Attributes, "name"),
			)
		}
	}

	return row
}

func renderTransportOrderMaterialsTable(cmd *cobra.Command, rows []transportOrderMaterialRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No transport order materials found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tORDER\tMATERIAL\tUNIT\tQUANTITY\tEXPLICIT")
	for _, row := range rows {
		orderLabel := firstNonEmpty(row.TransportOrderNumber, row.TransportOrderID)
		materialLabel := firstNonEmpty(row.MaterialType, row.MaterialTypeID)
		unitLabel := firstNonEmpty(row.UnitOfMeasure, row.UnitOfMeasureID)

		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(orderLabel, 18),
			truncateString(materialLabel, 24),
			truncateString(unitLabel, 12),
			row.Quantity,
			row.QuantityExplicit,
		)
	}
	return writer.Flush()
}
