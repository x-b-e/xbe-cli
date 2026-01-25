package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type transportOrderMaterialsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type transportOrderMaterialDetails struct {
	ID                            string   `json:"id"`
	QuantityExplicit              string   `json:"quantity_explicit,omitempty"`
	QuantityImplicitCached        string   `json:"quantity_implicit_cached,omitempty"`
	Quantity                      string   `json:"quantity,omitempty"`
	TransportOrderID              string   `json:"transport_order_id,omitempty"`
	TransportOrderNumber          string   `json:"transport_order_number,omitempty"`
	MaterialTypeID                string   `json:"material_type_id,omitempty"`
	MaterialType                  string   `json:"material_type,omitempty"`
	UnitOfMeasureID               string   `json:"unit_of_measure_id,omitempty"`
	UnitOfMeasure                 string   `json:"unit_of_measure,omitempty"`
	CustomerID                    string   `json:"customer_id,omitempty"`
	CustomerName                  string   `json:"customer_name,omitempty"`
	BrokerID                      string   `json:"broker_id,omitempty"`
	BrokerName                    string   `json:"broker_name,omitempty"`
	TransportOrderStopMaterialIDs []string `json:"transport_order_stop_material_ids,omitempty"`
	TransportReferenceIDs         []string `json:"transport_reference_ids,omitempty"`
}

func newTransportOrderMaterialsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show transport order material details",
		Long: `Show the full details of a transport order material.

Arguments:
  <id>  The transport order material ID (required).`,
		Example: `  # Show a transport order material
  xbe view transport-order-materials show 123

  # Output as JSON
  xbe view transport-order-materials show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runTransportOrderMaterialsShow,
	}
	initTransportOrderMaterialsShowFlags(cmd)
	return cmd
}

func init() {
	transportOrderMaterialsCmd.AddCommand(newTransportOrderMaterialsShowCmd())
}

func initTransportOrderMaterialsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runTransportOrderMaterialsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseTransportOrderMaterialsShowOptions(cmd)
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

	id := strings.TrimSpace(args[0])
	if id == "" {
		return fmt.Errorf("transport order material id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[transport-order-materials]", "quantity-explicit,quantity-implicit-cached,quantity,transport-order,material-type,unit-of-measure,transport-order-stop-materials,transport-references,customer,broker")
	query.Set("include", "transport-order,material-type,unit-of-measure,customer,broker")
	query.Set("fields[transport-orders]", "external-order-number")
	query.Set("fields[material-types]", "name,display-name")
	query.Set("fields[unit-of-measures]", "name,abbreviation")
	query.Set("fields[customers]", "company-name")
	query.Set("fields[brokers]", "company-name")

	body, _, err := client.Get(cmd.Context(), "/v1/transport-order-materials/"+id, query)
	if err != nil {
		if len(body) > 0 {
			fmt.Fprintln(cmd.ErrOrStderr(), string(body))
		}
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	var resp jsonAPISingleResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	handled, err := renderSparseShowIfRequested(cmd, resp)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if handled {
		return nil
	}

	details := buildTransportOrderMaterialDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderTransportOrderMaterialDetails(cmd, details)
}

func parseTransportOrderMaterialsShowOptions(cmd *cobra.Command) (transportOrderMaterialsShowOptions, error) {
	jsonOut, err := cmd.Flags().GetBool("json")
	if err != nil {
		return transportOrderMaterialsShowOptions{}, err
	}
	noAuth, err := cmd.Flags().GetBool("no-auth")
	if err != nil {
		return transportOrderMaterialsShowOptions{}, err
	}
	baseURL, err := cmd.Flags().GetString("base-url")
	if err != nil {
		return transportOrderMaterialsShowOptions{}, err
	}
	token, err := cmd.Flags().GetString("token")
	if err != nil {
		return transportOrderMaterialsShowOptions{}, err
	}

	return transportOrderMaterialsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildTransportOrderMaterialDetails(resp jsonAPISingleResponse) transportOrderMaterialDetails {
	included := make(map[string]jsonAPIResource, len(resp.Included))
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	attrs := resp.Data.Attributes
	details := transportOrderMaterialDetails{
		ID:                     resp.Data.ID,
		QuantityExplicit:       stringAttr(attrs, "quantity-explicit"),
		QuantityImplicitCached: stringAttr(attrs, "quantity-implicit-cached"),
		Quantity:               stringAttr(attrs, "quantity"),
	}

	if rel, ok := resp.Data.Relationships["transport-order"]; ok && rel.Data != nil {
		details.TransportOrderID = rel.Data.ID
		if order, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.TransportOrderNumber = firstNonEmpty(
				stringAttr(order.Attributes, "external-order-number"),
				stringAttr(order.Attributes, "order-number"),
			)
		}
	}

	if rel, ok := resp.Data.Relationships["material-type"]; ok && rel.Data != nil {
		details.MaterialTypeID = rel.Data.ID
		if mt, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.MaterialType = firstNonEmpty(
				stringAttr(mt.Attributes, "display-name"),
				stringAttr(mt.Attributes, "name"),
			)
		}
	}

	if rel, ok := resp.Data.Relationships["unit-of-measure"]; ok && rel.Data != nil {
		details.UnitOfMeasureID = rel.Data.ID
		if uom, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.UnitOfMeasure = firstNonEmpty(
				stringAttr(uom.Attributes, "abbreviation"),
				stringAttr(uom.Attributes, "name"),
			)
		}
	}

	if rel, ok := resp.Data.Relationships["customer"]; ok && rel.Data != nil {
		details.CustomerID = rel.Data.ID
		if customer, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.CustomerName = firstNonEmpty(
				stringAttr(customer.Attributes, "company-name"),
				stringAttr(customer.Attributes, "name"),
			)
		}
	}

	if rel, ok := resp.Data.Relationships["broker"]; ok && rel.Data != nil {
		details.BrokerID = rel.Data.ID
		if broker, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.BrokerName = firstNonEmpty(
				stringAttr(broker.Attributes, "company-name"),
				stringAttr(broker.Attributes, "name"),
			)
		}
	}

	if rel, ok := resp.Data.Relationships["transport-order-stop-materials"]; ok {
		details.TransportOrderStopMaterialIDs = relationshipIDList(rel)
	}
	if rel, ok := resp.Data.Relationships["transport-references"]; ok {
		details.TransportReferenceIDs = relationshipIDList(rel)
	}

	return details
}

func renderTransportOrderMaterialDetails(cmd *cobra.Command, details transportOrderMaterialDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.TransportOrderID != "" || details.TransportOrderNumber != "" {
		fmt.Fprintf(out, "Transport Order: %s\n", formatRelated(details.TransportOrderNumber, details.TransportOrderID))
	}
	if details.MaterialTypeID != "" || details.MaterialType != "" {
		fmt.Fprintf(out, "Material Type: %s\n", formatRelated(details.MaterialType, details.MaterialTypeID))
	}
	if details.UnitOfMeasureID != "" || details.UnitOfMeasure != "" {
		fmt.Fprintf(out, "Unit of Measure: %s\n", formatRelated(details.UnitOfMeasure, details.UnitOfMeasureID))
	}
	if details.Quantity != "" {
		fmt.Fprintf(out, "Quantity: %s\n", details.Quantity)
	}
	if details.QuantityExplicit != "" {
		fmt.Fprintf(out, "Quantity Explicit: %s\n", details.QuantityExplicit)
	}
	if details.QuantityImplicitCached != "" {
		fmt.Fprintf(out, "Quantity Implicit Cached: %s\n", details.QuantityImplicitCached)
	}
	if details.CustomerID != "" || details.CustomerName != "" {
		fmt.Fprintf(out, "Customer: %s\n", formatRelated(details.CustomerName, details.CustomerID))
	}
	if details.BrokerID != "" || details.BrokerName != "" {
		fmt.Fprintf(out, "Broker: %s\n", formatRelated(details.BrokerName, details.BrokerID))
	}
	if len(details.TransportOrderStopMaterialIDs) > 0 {
		fmt.Fprintf(out, "Transport Order Stop Materials: %s\n", strings.Join(details.TransportOrderStopMaterialIDs, ", "))
	}
	if len(details.TransportReferenceIDs) > 0 {
		fmt.Fprintf(out, "Transport References: %s\n", strings.Join(details.TransportReferenceIDs, ", "))
	}

	return nil
}
