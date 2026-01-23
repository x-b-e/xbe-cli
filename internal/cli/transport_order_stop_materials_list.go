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

type transportOrderStopMaterialsListOptions struct {
	BaseURL                string
	Token                  string
	JSON                   bool
	NoAuth                 bool
	Limit                  int
	Offset                 int
	Sort                   string
	TransportOrderMaterial string
	TransportOrderStop     string
	QuantityExplicit       string
}

type transportOrderStopMaterialRow struct {
	ID                     string `json:"id"`
	QuantityExplicit       string `json:"quantity_explicit,omitempty"`
	TransportOrderMaterial string `json:"transport_order_material_id,omitempty"`
	TransportOrderStop     string `json:"transport_order_stop_id,omitempty"`
}

func newTransportOrderStopMaterialsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List transport order stop materials",
		Long: `List transport order stop materials with filtering and pagination.

Output Columns:
  ID        Transport order stop material identifier
  QTY       Explicit quantity for the stop
  MATERIAL  Transport order material ID
  STOP      Transport order stop ID

Filters:
  --transport-order-material  Filter by transport order material ID (comma-separated for multiple)
  --transport-order-stop      Filter by transport order stop ID (comma-separated for multiple)
  --quantity-explicit         Filter by explicit quantity

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List transport order stop materials
  xbe view transport-order-stop-materials list

  # Filter by transport order stop
  xbe view transport-order-stop-materials list --transport-order-stop 456

  # Filter by transport order material
  xbe view transport-order-stop-materials list --transport-order-material 789

  # Output as JSON
  xbe view transport-order-stop-materials list --json`,
		Args: cobra.NoArgs,
		RunE: runTransportOrderStopMaterialsList,
	}
	initTransportOrderStopMaterialsListFlags(cmd)
	return cmd
}

func init() {
	transportOrderStopMaterialsCmd.AddCommand(newTransportOrderStopMaterialsListCmd())
}

func initTransportOrderStopMaterialsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("transport-order-material", "", "Filter by transport order material ID (comma-separated for multiple)")
	cmd.Flags().String("transport-order-stop", "", "Filter by transport order stop ID (comma-separated for multiple)")
	cmd.Flags().String("quantity-explicit", "", "Filter by explicit quantity")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runTransportOrderStopMaterialsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseTransportOrderStopMaterialsListOptions(cmd)
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
	query.Set("fields[transport-order-stop-materials]", "quantity-explicit,transport-order-material,transport-order-stop")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[transport-order-material]", opts.TransportOrderMaterial)
	setFilterIfPresent(query, "filter[transport-order-stop]", opts.TransportOrderStop)
	setFilterIfPresent(query, "filter[quantity-explicit]", opts.QuantityExplicit)

	body, _, err := client.Get(cmd.Context(), "/v1/transport-order-stop-materials", query)
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

	rows := buildTransportOrderStopMaterialRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderTransportOrderStopMaterialsTable(cmd, rows)
}

func parseTransportOrderStopMaterialsListOptions(cmd *cobra.Command) (transportOrderStopMaterialsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	transportOrderMaterial, _ := cmd.Flags().GetString("transport-order-material")
	transportOrderStop, _ := cmd.Flags().GetString("transport-order-stop")
	quantityExplicit, _ := cmd.Flags().GetString("quantity-explicit")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return transportOrderStopMaterialsListOptions{
		BaseURL:                baseURL,
		Token:                  token,
		JSON:                   jsonOut,
		NoAuth:                 noAuth,
		Limit:                  limit,
		Offset:                 offset,
		Sort:                   sort,
		TransportOrderMaterial: transportOrderMaterial,
		TransportOrderStop:     transportOrderStop,
		QuantityExplicit:       quantityExplicit,
	}, nil
}

func buildTransportOrderStopMaterialRows(resp jsonAPIResponse) []transportOrderStopMaterialRow {
	rows := make([]transportOrderStopMaterialRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, buildTransportOrderStopMaterialRow(resource))
	}
	return rows
}

func buildTransportOrderStopMaterialRow(resource jsonAPIResource) transportOrderStopMaterialRow {
	row := transportOrderStopMaterialRow{
		ID:               resource.ID,
		QuantityExplicit: stringAttr(resource.Attributes, "quantity-explicit"),
	}

	if rel, ok := resource.Relationships["transport-order-material"]; ok && rel.Data != nil {
		row.TransportOrderMaterial = rel.Data.ID
	}
	if rel, ok := resource.Relationships["transport-order-stop"]; ok && rel.Data != nil {
		row.TransportOrderStop = rel.Data.ID
	}

	return row
}

func renderTransportOrderStopMaterialsTable(cmd *cobra.Command, rows []transportOrderStopMaterialRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No transport order stop materials found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tQTY\tMATERIAL\tSTOP")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\n",
			row.ID,
			row.QuantityExplicit,
			row.TransportOrderMaterial,
			row.TransportOrderStop,
		)
	}
	return writer.Flush()
}
