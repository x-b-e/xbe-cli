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

type transportOrderStopsListOptions struct {
	BaseURL                   string
	Token                     string
	JSON                      bool
	NoAuth                    bool
	Limit                     int
	Offset                    int
	Sort                      string
	TransportOrder            string
	Location                  string
	Role                      string
	AtMinMin                  string
	AtMinMax                  string
	IsAtMin                   string
	AtMaxMin                  string
	AtMaxMax                  string
	IsAtMax                   string
	ExternalTmsStopNumber     string
	ExternalIdentificationVal string
	CreatedAtMin              string
	CreatedAtMax              string
	IsCreatedAt               string
	UpdatedAtMin              string
	UpdatedAtMax              string
	IsUpdatedAt               string
}

type transportOrderStopRow struct {
	ID                    string `json:"id"`
	TransportOrderID      string `json:"transport_order_id,omitempty"`
	TransportOrderNumber  string `json:"transport_order_number,omitempty"`
	LocationID            string `json:"location_id,omitempty"`
	LocationName          string `json:"location_name,omitempty"`
	Role                  string `json:"role,omitempty"`
	Status                string `json:"status,omitempty"`
	Position              int    `json:"position,omitempty"`
	AtMin                 string `json:"at_min,omitempty"`
	AtMax                 string `json:"at_max,omitempty"`
	ExternalTmsStopNumber string `json:"external_tms_stop_number,omitempty"`
}

func newTransportOrderStopsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List transport order stops",
		Long: `List transport order stops.

Output Columns:
  ID        Transport order stop identifier
  ORDER     Transport order number or ID
  LOCATION  Transport location name or ID
  ROLE      Stop role (pickup, delivery)
  STATUS    Stop status
  POS       Stop position in the order
  AT_MIN    Earliest scheduled time
  AT_MAX    Latest scheduled time
  TMS_STOP  External TMS stop number (if set)

Filters:
  --transport-order            Filter by transport order ID
  --location                   Filter by transport location ID
  --role                       Filter by role (pickup, delivery)
  --at-min-min                 Filter by at-min on/after (ISO 8601)
  --at-min-max                 Filter by at-min on/before (ISO 8601)
  --is-at-min                  Filter by presence of at-min (true/false)
  --at-max-min                 Filter by at-max on/after (ISO 8601)
  --at-max-max                 Filter by at-max on/before (ISO 8601)
  --is-at-max                  Filter by presence of at-max (true/false)
  --external-tms-stop-number   Filter by external TMS stop number
  --external-identification-value Filter by external identification value
  --created-at-min             Filter by created-at on/after (ISO 8601)
  --created-at-max             Filter by created-at on/before (ISO 8601)
  --is-created-at              Filter by presence of created-at (true/false)
  --updated-at-min             Filter by updated-at on/after (ISO 8601)
  --updated-at-max             Filter by updated-at on/before (ISO 8601)
  --is-updated-at              Filter by presence of updated-at (true/false)

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List transport order stops
  xbe view transport-order-stops list

  # Filter by transport order
  xbe view transport-order-stops list --transport-order 123

  # Filter by location and role
  xbe view transport-order-stops list --location 456 --role pickup

  # Filter by at-min range
  xbe view transport-order-stops list \
    --at-min-min 2026-01-23T00:00:00Z \
    --at-min-max 2026-01-24T00:00:00Z

  # Output as JSON
  xbe view transport-order-stops list --json`,
		Args: cobra.NoArgs,
		RunE: runTransportOrderStopsList,
	}
	initTransportOrderStopsListFlags(cmd)
	return cmd
}

func init() {
	transportOrderStopsCmd.AddCommand(newTransportOrderStopsListCmd())
}

func initTransportOrderStopsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("transport-order", "", "Filter by transport order ID")
	cmd.Flags().String("location", "", "Filter by transport location ID")
	cmd.Flags().String("role", "", "Filter by role (pickup, delivery)")
	cmd.Flags().String("at-min-min", "", "Filter by at-min on/after (ISO 8601)")
	cmd.Flags().String("at-min-max", "", "Filter by at-min on/before (ISO 8601)")
	cmd.Flags().String("is-at-min", "", "Filter by presence of at-min (true/false)")
	cmd.Flags().String("at-max-min", "", "Filter by at-max on/after (ISO 8601)")
	cmd.Flags().String("at-max-max", "", "Filter by at-max on/before (ISO 8601)")
	cmd.Flags().String("is-at-max", "", "Filter by presence of at-max (true/false)")
	cmd.Flags().String("external-tms-stop-number", "", "Filter by external TMS stop number")
	cmd.Flags().String("external-identification-value", "", "Filter by external identification value")
	cmd.Flags().String("created-at-min", "", "Filter by created-at on/after (ISO 8601)")
	cmd.Flags().String("created-at-max", "", "Filter by created-at on/before (ISO 8601)")
	cmd.Flags().String("is-created-at", "", "Filter by presence of created-at (true/false)")
	cmd.Flags().String("updated-at-min", "", "Filter by updated-at on/after (ISO 8601)")
	cmd.Flags().String("updated-at-max", "", "Filter by updated-at on/before (ISO 8601)")
	cmd.Flags().String("is-updated-at", "", "Filter by presence of updated-at (true/false)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runTransportOrderStopsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseTransportOrderStopsListOptions(cmd)
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
	query.Set("fields[transport-order-stops]", "position,role,status,at-min,at-max,external-tms-stop-number,transport-order,location")
	query.Set("fields[transport-orders]", "external-order-number,order-number")
	query.Set("fields[project-transport-locations]", "name")
	query.Set("include", "transport-order,location")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[transport-order]", opts.TransportOrder)
	setFilterIfPresent(query, "filter[location]", opts.Location)
	setFilterIfPresent(query, "filter[role]", opts.Role)
	setFilterIfPresent(query, "filter[at-min-min]", opts.AtMinMin)
	setFilterIfPresent(query, "filter[at-min-max]", opts.AtMinMax)
	setFilterIfPresent(query, "filter[is-at-min]", opts.IsAtMin)
	setFilterIfPresent(query, "filter[at-max-min]", opts.AtMaxMin)
	setFilterIfPresent(query, "filter[at-max-max]", opts.AtMaxMax)
	setFilterIfPresent(query, "filter[is-at-max]", opts.IsAtMax)
	setFilterIfPresent(query, "filter[external-tms-stop-number]", opts.ExternalTmsStopNumber)
	setFilterIfPresent(query, "filter[external-identification-value]", opts.ExternalIdentificationVal)
	setFilterIfPresent(query, "filter[created-at-min]", opts.CreatedAtMin)
	setFilterIfPresent(query, "filter[created-at-max]", opts.CreatedAtMax)
	setFilterIfPresent(query, "filter[is-created-at]", opts.IsCreatedAt)
	setFilterIfPresent(query, "filter[updated-at-min]", opts.UpdatedAtMin)
	setFilterIfPresent(query, "filter[updated-at-max]", opts.UpdatedAtMax)
	setFilterIfPresent(query, "filter[is-updated-at]", opts.IsUpdatedAt)

	body, _, err := client.Get(cmd.Context(), "/v1/transport-order-stops", query)
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

	rows := buildTransportOrderStopRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderTransportOrderStopsTable(cmd, rows)
}

func parseTransportOrderStopsListOptions(cmd *cobra.Command) (transportOrderStopsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	transportOrder, _ := cmd.Flags().GetString("transport-order")
	location, _ := cmd.Flags().GetString("location")
	role, _ := cmd.Flags().GetString("role")
	atMinMin, _ := cmd.Flags().GetString("at-min-min")
	atMinMax, _ := cmd.Flags().GetString("at-min-max")
	isAtMin, _ := cmd.Flags().GetString("is-at-min")
	atMaxMin, _ := cmd.Flags().GetString("at-max-min")
	atMaxMax, _ := cmd.Flags().GetString("at-max-max")
	isAtMax, _ := cmd.Flags().GetString("is-at-max")
	externalTmsStopNumber, _ := cmd.Flags().GetString("external-tms-stop-number")
	externalIdentificationVal, _ := cmd.Flags().GetString("external-identification-value")
	createdAtMin, _ := cmd.Flags().GetString("created-at-min")
	createdAtMax, _ := cmd.Flags().GetString("created-at-max")
	isCreatedAt, _ := cmd.Flags().GetString("is-created-at")
	updatedAtMin, _ := cmd.Flags().GetString("updated-at-min")
	updatedAtMax, _ := cmd.Flags().GetString("updated-at-max")
	isUpdatedAt, _ := cmd.Flags().GetString("is-updated-at")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return transportOrderStopsListOptions{
		BaseURL:                   baseURL,
		Token:                     token,
		JSON:                      jsonOut,
		NoAuth:                    noAuth,
		Limit:                     limit,
		Offset:                    offset,
		Sort:                      sort,
		TransportOrder:            transportOrder,
		Location:                  location,
		Role:                      role,
		AtMinMin:                  atMinMin,
		AtMinMax:                  atMinMax,
		IsAtMin:                   isAtMin,
		AtMaxMin:                  atMaxMin,
		AtMaxMax:                  atMaxMax,
		IsAtMax:                   isAtMax,
		ExternalTmsStopNumber:     externalTmsStopNumber,
		ExternalIdentificationVal: externalIdentificationVal,
		CreatedAtMin:              createdAtMin,
		CreatedAtMax:              createdAtMax,
		IsCreatedAt:               isCreatedAt,
		UpdatedAtMin:              updatedAtMin,
		UpdatedAtMax:              updatedAtMax,
		IsUpdatedAt:               isUpdatedAt,
	}, nil
}

func buildTransportOrderStopRows(resp jsonAPIResponse) []transportOrderStopRow {
	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	rows := make([]transportOrderStopRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := transportOrderStopRow{
			ID:                    resource.ID,
			Role:                  stringAttr(resource.Attributes, "role"),
			Status:                stringAttr(resource.Attributes, "status"),
			Position:              intAttr(resource.Attributes, "position"),
			AtMin:                 formatDateTime(stringAttr(resource.Attributes, "at-min")),
			AtMax:                 formatDateTime(stringAttr(resource.Attributes, "at-max")),
			ExternalTmsStopNumber: stringAttr(resource.Attributes, "external-tms-stop-number"),
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
		if rel, ok := resource.Relationships["location"]; ok && rel.Data != nil {
			row.LocationID = rel.Data.ID
			if location, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
				row.LocationName = stringAttr(location.Attributes, "name")
			}
		}

		rows = append(rows, row)
	}

	return rows
}

func renderTransportOrderStopsTable(cmd *cobra.Command, rows []transportOrderStopRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No transport order stops found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tORDER\tLOCATION\tROLE\tSTATUS\tPOS\tAT_MIN\tAT_MAX\tTMS_STOP")

	for _, row := range rows {
		orderDisplay := firstNonEmpty(row.TransportOrderNumber, row.TransportOrderID)
		locationDisplay := firstNonEmpty(row.LocationName, row.LocationID)

		position := ""
		if row.Position != 0 {
			position = strconv.Itoa(row.Position)
		}

		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			orderDisplay,
			locationDisplay,
			row.Role,
			row.Status,
			position,
			row.AtMin,
			row.AtMax,
			row.ExternalTmsStopNumber,
		)
	}

	return writer.Flush()
}
