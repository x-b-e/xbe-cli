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

type materialSitesListOptions struct {
	BaseURL                    string
	Token                      string
	JSON                       bool
	NoAuth                     bool
	Limit                      int
	Offset                     int
	Name                       string
	Active                     bool
	Broker                     string
	Q                          string
	MaterialSupplier           string
	OperatingStatus            string
	Parent                     string
	MaterialTypes              string
	IsTicketMaker              string
	IsBrokerActive             string
	IsOnlyForEquipmentMovement string
	// NOTE: active-child filter removed due to server-side 500 error (column doesn't exist)
	AddressNear                    string
	ActiveAround                   string
	HasActivityAsOf                string
	MaterialTransactionIntegration string
	IsAvailableAt                  string
}

type materialSiteRow struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Active     bool   `json:"is_active"`
	Supplier   string `json:"supplier,omitempty"`
	SupplierID string `json:"supplier_id,omitempty"`
	Broker     string `json:"broker,omitempty"`
	BrokerID   string `json:"broker_id,omitempty"`
}

func newMaterialSitesListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List material sites",
		Long: `List material sites with filtering and pagination.

Returns a list of material sites (plants, quarries, stockpiles). Use this
to look up material site IDs for filtering job production plans.

Output Columns:
  ID        Material site identifier (use this for --material-site filter)
  NAME      Material site name
  ACTIVE    Whether the site is active
  SUPPLIER  Material supplier name
  BROKER    Broker name`,
		Example: `  # List material sites
  xbe view material-sites list

  # Search by name
  xbe view material-sites list --name "Plant"

  # List active material sites only
  xbe view material-sites list --active

  # Output as JSON
  xbe view material-sites list --json`,
		RunE: runMaterialSitesList,
	}
	initMaterialSitesListFlags(cmd)
	return cmd
}

func init() {
	materialSitesCmd.AddCommand(newMaterialSitesListCmd())
}

func initMaterialSitesListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Bool("active", false, "Show only active material sites")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("name", "", "Filter by name (partial match)")
	cmd.Flags().String("broker", "", "Filter by broker ID (comma-separated for multiple)")
	cmd.Flags().String("q", "", "Full-text search")
	cmd.Flags().String("material-supplier", "", "Filter by material supplier ID (comma-separated for multiple)")
	cmd.Flags().String("operating-status", "", "Filter by operating status")
	cmd.Flags().String("parent", "", "Filter by parent material site ID (comma-separated for multiple)")
	cmd.Flags().String("material-types", "", "Filter by material type IDs (comma-separated)")
	cmd.Flags().String("is-ticket-maker", "", "Filter by ticket maker status (true/false)")
	cmd.Flags().String("is-broker-active", "", "Filter by broker active status (true/false)")
	cmd.Flags().String("is-only-for-equipment-movement", "", "Filter by equipment movement only status (true/false)")
	cmd.Flags().String("address-near", "", "Filter by proximity to address (lat,lng,radius_miles)")
	cmd.Flags().String("active-around", "", "Filter by active around datetime (ISO 8601)")
	cmd.Flags().String("has-activity-as-of", "", "Filter by activity as of datetime (ISO 8601)")
	cmd.Flags().String("material-transaction-integration", "", "Filter by material transaction integration (true/false)")
	cmd.Flags().String("is-available-at", "", "Filter by availability at datetime (ISO 8601)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runMaterialSitesList(cmd *cobra.Command, _ []string) error {
	opts, err := parseMaterialSitesListOptions(cmd)
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
	query.Set("fields[material-sites]", "name,is-active-effective,material-supplier")
	query.Set("fields[material-suppliers]", "name,broker")
	query.Set("fields[brokers]", "company-name")
	query.Set("include", "material-supplier.broker")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Name != "" {
		query.Set("filter[name]", opts.Name)
	}
	if opts.Active {
		query.Set("filter[is-active-effective]", "true")
	}
	setFilterIfPresent(query, "filter[broker]", opts.Broker)
	setFilterIfPresent(query, "filter[q]", opts.Q)
	setFilterIfPresent(query, "filter[material-supplier]", opts.MaterialSupplier)
	setFilterIfPresent(query, "filter[operating-status]", opts.OperatingStatus)
	setFilterIfPresent(query, "filter[parent]", opts.Parent)
	setFilterIfPresent(query, "filter[material-types]", opts.MaterialTypes)
	setFilterIfPresent(query, "filter[is-ticket-maker]", opts.IsTicketMaker)
	setFilterIfPresent(query, "filter[is-broker-active]", opts.IsBrokerActive)
	setFilterIfPresent(query, "filter[is-only-for-equipment-movement]", opts.IsOnlyForEquipmentMovement)
	setFilterIfPresent(query, "filter[address-near]", opts.AddressNear)
	setFilterIfPresent(query, "filter[active-around]", opts.ActiveAround)
	setFilterIfPresent(query, "filter[has-activity-as-of]", opts.HasActivityAsOf)
	setFilterIfPresent(query, "filter[material-transaction-integration]", opts.MaterialTransactionIntegration)
	setFilterIfPresent(query, "filter[is-available-at]", opts.IsAvailableAt)

	body, _, err := client.Get(cmd.Context(), "/v1/material-sites", query)
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

	rows := buildMaterialSiteRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderMaterialSitesTable(cmd, rows)
}

func parseMaterialSitesListOptions(cmd *cobra.Command) (materialSitesListOptions, error) {
	jsonOut, err := cmd.Flags().GetBool("json")
	if err != nil {
		return materialSitesListOptions{}, err
	}
	noAuth, err := cmd.Flags().GetBool("no-auth")
	if err != nil {
		return materialSitesListOptions{}, err
	}
	active, err := cmd.Flags().GetBool("active")
	if err != nil {
		return materialSitesListOptions{}, err
	}
	limit, err := cmd.Flags().GetInt("limit")
	if err != nil {
		return materialSitesListOptions{}, err
	}
	offset, err := cmd.Flags().GetInt("offset")
	if err != nil {
		return materialSitesListOptions{}, err
	}
	name, err := cmd.Flags().GetString("name")
	if err != nil {
		return materialSitesListOptions{}, err
	}
	broker, err := cmd.Flags().GetString("broker")
	if err != nil {
		return materialSitesListOptions{}, err
	}
	q, err := cmd.Flags().GetString("q")
	if err != nil {
		return materialSitesListOptions{}, err
	}
	materialSupplier, err := cmd.Flags().GetString("material-supplier")
	if err != nil {
		return materialSitesListOptions{}, err
	}
	operatingStatus, err := cmd.Flags().GetString("operating-status")
	if err != nil {
		return materialSitesListOptions{}, err
	}
	parent, err := cmd.Flags().GetString("parent")
	if err != nil {
		return materialSitesListOptions{}, err
	}
	materialTypes, err := cmd.Flags().GetString("material-types")
	if err != nil {
		return materialSitesListOptions{}, err
	}
	isTicketMaker, err := cmd.Flags().GetString("is-ticket-maker")
	if err != nil {
		return materialSitesListOptions{}, err
	}
	isBrokerActive, err := cmd.Flags().GetString("is-broker-active")
	if err != nil {
		return materialSitesListOptions{}, err
	}
	isOnlyForEquipmentMovement, err := cmd.Flags().GetString("is-only-for-equipment-movement")
	if err != nil {
		return materialSitesListOptions{}, err
	}
	addressNear, err := cmd.Flags().GetString("address-near")
	if err != nil {
		return materialSitesListOptions{}, err
	}
	activeAround, err := cmd.Flags().GetString("active-around")
	if err != nil {
		return materialSitesListOptions{}, err
	}
	hasActivityAsOf, err := cmd.Flags().GetString("has-activity-as-of")
	if err != nil {
		return materialSitesListOptions{}, err
	}
	materialTransactionIntegration, err := cmd.Flags().GetString("material-transaction-integration")
	if err != nil {
		return materialSitesListOptions{}, err
	}
	isAvailableAt, err := cmd.Flags().GetString("is-available-at")
	if err != nil {
		return materialSitesListOptions{}, err
	}
	baseURL, err := cmd.Flags().GetString("base-url")
	if err != nil {
		return materialSitesListOptions{}, err
	}
	token, err := cmd.Flags().GetString("token")
	if err != nil {
		return materialSitesListOptions{}, err
	}

	return materialSitesListOptions{
		BaseURL:                        baseURL,
		Token:                          token,
		JSON:                           jsonOut,
		NoAuth:                         noAuth,
		Active:                         active,
		Limit:                          limit,
		Offset:                         offset,
		Name:                           name,
		Broker:                         broker,
		Q:                              q,
		MaterialSupplier:               materialSupplier,
		OperatingStatus:                operatingStatus,
		Parent:                         parent,
		MaterialTypes:                  materialTypes,
		IsTicketMaker:                  isTicketMaker,
		IsBrokerActive:                 isBrokerActive,
		IsOnlyForEquipmentMovement:     isOnlyForEquipmentMovement,
		AddressNear:                    addressNear,
		ActiveAround:                   activeAround,
		HasActivityAsOf:                hasActivityAsOf,
		MaterialTransactionIntegration: materialTransactionIntegration,
		IsAvailableAt:                  isAvailableAt,
	}, nil
}

func buildMaterialSiteRows(resp jsonAPIResponse) []materialSiteRow {
	// Build included lookup
	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		key := resourceKey(inc.Type, inc.ID)
		included[key] = inc
	}

	rows := make([]materialSiteRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := materialSiteRow{
			ID:     resource.ID,
			Name:   stringAttr(resource.Attributes, "name"),
			Active: boolAttr(resource.Attributes, "is-active-effective"),
		}

		// Resolve material supplier and broker
		if rel, ok := resource.Relationships["material-supplier"]; ok && rel.Data != nil {
			row.SupplierID = rel.Data.ID
			if supplier, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
				row.Supplier = stringAttr(supplier.Attributes, "name")
				// Get broker through supplier
				if brokerRel, ok := supplier.Relationships["broker"]; ok && brokerRel.Data != nil {
					row.BrokerID = brokerRel.Data.ID
					if broker, ok := included[resourceKey(brokerRel.Data.Type, brokerRel.Data.ID)]; ok {
						row.Broker = stringAttr(broker.Attributes, "company-name")
					}
				}
			}
		}

		rows = append(rows, row)
	}
	return rows
}

func renderMaterialSitesTable(cmd *cobra.Command, rows []materialSiteRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No material sites found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tNAME\tACTIVE\tSUPPLIER\tBROKER")
	for _, row := range rows {
		activeStr := ""
		if row.Active {
			activeStr = "Yes"
		}
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(row.Name, 35),
			activeStr,
			truncateString(row.Supplier, 25),
			truncateString(row.Broker, 25),
		)
	}
	return writer.Flush()
}
