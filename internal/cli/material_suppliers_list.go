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

type materialSuppliersListOptions struct {
	BaseURL        string
	Token          string
	JSON           bool
	NoAuth         bool
	Limit          int
	Offset         int
	Name           string
	IsActive       bool
	Broker         string
	IsBrokerActive string
}

func newMaterialSuppliersListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List material suppliers",
		Long: `List material suppliers with filtering and pagination.

Returns a list of material suppliers matching the specified criteria.

Pagination:
  Use --limit and --offset to paginate through large result sets.

Filtering:
  Multiple filters can be combined. All filters use AND logic.

Use Case:
  Find supplier IDs for filtering posts by creator:
    xbe view posts list --creator "MaterialSupplier|<id>"`,
		Example: `  # List material suppliers
  xbe view material-suppliers list

  # Search by name
  xbe view material-suppliers list --name "Acme"

  # Filter by active status
  xbe view material-suppliers list --active

  # Paginate results
  xbe view material-suppliers list --limit 20 --offset 40

  # Output as JSON
  xbe view material-suppliers list --json`,
		RunE: runMaterialSuppliersList,
	}
	initMaterialSuppliersListFlags(cmd)
	return cmd
}

func init() {
	materialSuppliersCmd.AddCommand(newMaterialSuppliersListCmd())
}

func initMaterialSuppliersListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 0, "Page size (defaults to server default)")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("name", "", "Filter by name (partial match)")
	cmd.Flags().Bool("active", false, "Filter to only active suppliers")
	cmd.Flags().String("broker", "", "Filter by broker ID (comma-separated for multiple)")
	cmd.Flags().String("is-broker-active", "", "Filter by broker active status (true/false)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runMaterialSuppliersList(cmd *cobra.Command, _ []string) error {
	opts, err := parseMaterialSuppliersListOptions(cmd)
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
	query.Set("fields[material-suppliers]", "name,is-active,broker")
	query.Set("fields[brokers]", "company-name")
	query.Set("include", "broker")
	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	setFilterIfPresent(query, "filter[name]", opts.Name)
	if opts.IsActive {
		query.Set("filter[is_active]", "true")
	}
	setFilterIfPresent(query, "filter[broker]", opts.Broker)
	setFilterIfPresent(query, "filter[is-broker-active]", opts.IsBrokerActive)

	body, _, err := client.Get(cmd.Context(), "/v1/material-suppliers", query)
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
		rows := buildMaterialSupplierRows(resp)
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderMaterialSuppliersList(cmd, resp)
}

func parseMaterialSuppliersListOptions(cmd *cobra.Command) (materialSuppliersListOptions, error) {
	jsonOut, err := cmd.Flags().GetBool("json")
	if err != nil {
		return materialSuppliersListOptions{}, err
	}
	noAuth, err := cmd.Flags().GetBool("no-auth")
	if err != nil {
		return materialSuppliersListOptions{}, err
	}
	limit, err := cmd.Flags().GetInt("limit")
	if err != nil {
		return materialSuppliersListOptions{}, err
	}
	offset, err := cmd.Flags().GetInt("offset")
	if err != nil {
		return materialSuppliersListOptions{}, err
	}
	name, err := cmd.Flags().GetString("name")
	if err != nil {
		return materialSuppliersListOptions{}, err
	}
	isActive, err := cmd.Flags().GetBool("active")
	if err != nil {
		return materialSuppliersListOptions{}, err
	}
	broker, err := cmd.Flags().GetString("broker")
	if err != nil {
		return materialSuppliersListOptions{}, err
	}
	isBrokerActive, err := cmd.Flags().GetString("is-broker-active")
	if err != nil {
		return materialSuppliersListOptions{}, err
	}
	baseURL, err := cmd.Flags().GetString("base-url")
	if err != nil {
		return materialSuppliersListOptions{}, err
	}
	token, err := cmd.Flags().GetString("token")
	if err != nil {
		return materialSuppliersListOptions{}, err
	}

	return materialSuppliersListOptions{
		BaseURL:        baseURL,
		Token:          token,
		JSON:           jsonOut,
		NoAuth:         noAuth,
		Limit:          limit,
		Offset:         offset,
		Name:           name,
		IsActive:       isActive,
		Broker:         broker,
		IsBrokerActive: isBrokerActive,
	}, nil
}

type materialSupplierRow struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Broker   string `json:"broker"`
	IsActive bool   `json:"is_active"`
}

func buildMaterialSupplierRows(resp jsonAPIResponse) []materialSupplierRow {
	// Build included map for broker lookup
	included := make(map[string]map[string]any)
	for _, inc := range resp.Included {
		key := resourceKey(inc.Type, inc.ID)
		included[key] = inc.Attributes
	}

	rows := make([]materialSupplierRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		brokerName := ""
		if rel, ok := resource.Relationships["broker"]; ok && rel.Data != nil {
			key := resourceKey(rel.Data.Type, rel.Data.ID)
			if attrs, ok := included[key]; ok {
				brokerName = strings.TrimSpace(stringAttr(attrs, "company-name"))
			}
		}

		rows = append(rows, materialSupplierRow{
			ID:       resource.ID,
			Name:     strings.TrimSpace(stringAttr(resource.Attributes, "name")),
			Broker:   brokerName,
			IsActive: boolAttr(resource.Attributes, "is-active"),
		})
	}
	return rows
}

func renderMaterialSuppliersList(cmd *cobra.Command, resp jsonAPIResponse) error {
	rows := buildMaterialSupplierRows(resp)
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No material suppliers found.")
		return nil
	}

	const nameMax = 50
	const brokerMax = 30

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tNAME\tBROKER")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\n",
			row.ID,
			truncateString(row.Name, nameMax),
			truncateString(row.Broker, brokerMax),
		)
	}
	return writer.Flush()
}
