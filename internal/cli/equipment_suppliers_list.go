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

type equipmentSuppliersListOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
	Limit   int
	Offset  int
	Name    string
	Broker  string
}

type equipmentSupplierRow struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	ContractNumber string `json:"contract_number,omitempty"`
	BrokerID       string `json:"broker_id,omitempty"`
	BrokerName     string `json:"broker_name,omitempty"`
}

func newEquipmentSuppliersListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List equipment suppliers",
		Long: `List equipment suppliers with filtering and pagination.

Equipment suppliers provide rental equipment and services for projects.

Output Columns:
  ID        Equipment supplier identifier
  NAME      Supplier name
  CONTRACT  Contract number
  BROKER    Broker name

Filters:
  --name    Filter by supplier name
  --broker  Filter by broker ID`,
		Example: `  # List all equipment suppliers
  xbe view equipment-suppliers list

  # Filter by name
  xbe view equipment-suppliers list --name "Acme"

  # Filter by broker
  xbe view equipment-suppliers list --broker 123

  # Output as JSON
  xbe view equipment-suppliers list --json`,
		RunE: runEquipmentSuppliersList,
	}
	initEquipmentSuppliersListFlags(cmd)
	return cmd
}

func init() {
	equipmentSuppliersCmd.AddCommand(newEquipmentSuppliersListCmd())
}

func initEquipmentSuppliersListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("name", "", "Filter by supplier name")
	cmd.Flags().String("broker", "", "Filter by broker ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runEquipmentSuppliersList(cmd *cobra.Command, _ []string) error {
	opts, err := parseEquipmentSuppliersListOptions(cmd)
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
	query.Set("fields[equipment-suppliers]", "name,contract-number,broker")
	query.Set("fields[brokers]", "company-name")
	query.Set("include", "broker")
	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	setFilterIfPresent(query, "filter[name]", opts.Name)
	setFilterIfPresent(query, "filter[broker]", opts.Broker)

	body, _, err := client.Get(cmd.Context(), "/v1/equipment-suppliers", query)
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

	rows := buildEquipmentSupplierRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderEquipmentSuppliersList(cmd, rows)
}

func parseEquipmentSuppliersListOptions(cmd *cobra.Command) (equipmentSuppliersListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	name, _ := cmd.Flags().GetString("name")
	broker, _ := cmd.Flags().GetString("broker")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return equipmentSuppliersListOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
		Limit:   limit,
		Offset:  offset,
		Name:    name,
		Broker:  broker,
	}, nil
}

func buildEquipmentSupplierRows(resp jsonAPIResponse) []equipmentSupplierRow {
	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		key := resourceKey(inc.Type, inc.ID)
		included[key] = inc
	}

	rows := make([]equipmentSupplierRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := equipmentSupplierRow{
			ID:             resource.ID,
			Name:           strings.TrimSpace(stringAttr(resource.Attributes, "name")),
			ContractNumber: strings.TrimSpace(stringAttr(resource.Attributes, "contract-number")),
		}

		if rel, ok := resource.Relationships["broker"]; ok && rel.Data != nil {
			row.BrokerID = rel.Data.ID
			if broker, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
				row.BrokerName = strings.TrimSpace(stringAttr(broker.Attributes, "company-name"))
			}
		}

		rows = append(rows, row)
	}
	return rows
}

func renderEquipmentSuppliersList(cmd *cobra.Command, rows []equipmentSupplierRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No equipment suppliers found.")
		return nil
	}

	const nameMax = 50
	const contractMax = 20
	const brokerMax = 30

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tNAME\tCONTRACT\tBROKER")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(row.Name, nameMax),
			truncateString(row.ContractNumber, contractMax),
			truncateString(row.BrokerName, brokerMax),
		)
	}
	return writer.Flush()
}
