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

type rawTransportOrdersShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type rawTransportOrderDetails struct {
	ID                   string `json:"id"`
	ExternalOrderNumber  string `json:"external_order_number,omitempty"`
	IsManaged            bool   `json:"is_managed"`
	Importer             string `json:"importer,omitempty"`
	ImportStatus         string `json:"import_status,omitempty"`
	TablesRowversionMin  string `json:"tables_rowversion_min,omitempty"`
	TablesRowversionMax  string `json:"tables_rowversion_max,omitempty"`
	Tables               any    `json:"tables,omitempty"`
	ImportErrors         any    `json:"import_errors,omitempty"`
	BrokerID             string `json:"broker_id,omitempty"`
	BrokerName           string `json:"broker_name,omitempty"`
	TransportOrderID     string `json:"transport_order_id,omitempty"`
	TransportOrderStatus string `json:"transport_order_status,omitempty"`
	CreatedByID          string `json:"created_by_id,omitempty"`
	CreatedByName        string `json:"created_by_name,omitempty"`
}

func newRawTransportOrdersShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show raw transport order details",
		Long: `Show full details of a raw transport order.

Raw transport orders include the raw import payloads and import status
before they are normalized into transport orders.

Output Fields:
  ID                     Raw transport order ID
  External Order Number  External order reference
  Importer               Importer key
  Import Status          Current import status
  Managed                Managed flag
  Rowversion Min/Max     Rowversion range across tables
  Broker                 Broker relationship
  Transport Order        Linked transport order (if available)
  Created By             User who created the import
  Tables                 Raw table payloads
  Import Errors          Import error details (if any)

Arguments:
  <id>    Raw transport order ID (required)

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a raw transport order
  xbe view raw-transport-orders show 123

  # JSON output
  xbe view raw-transport-orders show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runRawTransportOrdersShow,
	}
	initRawTransportOrdersShowFlags(cmd)
	return cmd
}

func init() {
	rawTransportOrdersCmd.AddCommand(newRawTransportOrdersShowCmd())
}

func initRawTransportOrdersShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runRawTransportOrdersShow(cmd *cobra.Command, args []string) error {
	if handled, err := maybeHandleClientURLShow(cmd, args); err != nil {
		return err
	} else if handled {
		return nil
	}

	opts, err := parseRawTransportOrdersShowOptions(cmd)
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
		return fmt.Errorf("raw transport order id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[raw-transport-orders]", "external-order-number,is-managed,importer,import-status,import-errors,tables,tables-rowversion-min,tables-rowversion-max,broker,transport-order,created-by")
	query.Set("fields[brokers]", "company-name")
	query.Set("fields[transport-orders]", "status")
	query.Set("fields[users]", "name,email-address")
	query.Set("include", "broker,transport-order,created-by")

	body, _, err := client.Get(cmd.Context(), "/v1/raw-transport-orders/"+id, query)
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

	details := buildRawTransportOrderDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderRawTransportOrderDetails(cmd, details)
}

func parseRawTransportOrdersShowOptions(cmd *cobra.Command) (rawTransportOrdersShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return rawTransportOrdersShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildRawTransportOrderDetails(resp jsonAPISingleResponse) rawTransportOrderDetails {
	attrs := resp.Data.Attributes

	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	details := rawTransportOrderDetails{
		ID:                  resp.Data.ID,
		ExternalOrderNumber: stringAttr(attrs, "external-order-number"),
		IsManaged:           boolAttr(attrs, "is-managed"),
		Importer:            stringAttr(attrs, "importer"),
		ImportStatus:        stringAttr(attrs, "import-status"),
		TablesRowversionMin: stringAttr(attrs, "tables-rowversion-min"),
		TablesRowversionMax: stringAttr(attrs, "tables-rowversion-max"),
		Tables:              anyAttr(attrs, "tables"),
		ImportErrors:        anyAttr(attrs, "import-errors"),
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

	if rel, ok := resp.Data.Relationships["transport-order"]; ok && rel.Data != nil {
		details.TransportOrderID = rel.Data.ID
		if order, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.TransportOrderStatus = stringAttr(order.Attributes, "status")
		}
	}

	if rel, ok := resp.Data.Relationships["created-by"]; ok && rel.Data != nil {
		details.CreatedByID = rel.Data.ID
		if user, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.CreatedByName = firstNonEmpty(
				stringAttr(user.Attributes, "name"),
				stringAttr(user.Attributes, "email-address"),
			)
		}
	}

	return details
}

func renderRawTransportOrderDetails(cmd *cobra.Command, details rawTransportOrderDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	fmt.Fprintf(out, "External Order Number: %s\n", formatOptional(details.ExternalOrderNumber))
	fmt.Fprintf(out, "Importer: %s\n", formatOptional(details.Importer))
	fmt.Fprintf(out, "Import Status: %s\n", formatOptional(details.ImportStatus))
	fmt.Fprintf(out, "Managed: %s\n", formatBool(details.IsManaged))
	fmt.Fprintf(out, "Tables Rowversion Min: %s\n", formatOptional(details.TablesRowversionMin))
	fmt.Fprintf(out, "Tables Rowversion Max: %s\n", formatOptional(details.TablesRowversionMax))

	if details.BrokerID != "" || details.BrokerName != "" {
		label := details.BrokerID
		if details.BrokerName != "" {
			label = fmt.Sprintf("%s (%s)", details.BrokerName, details.BrokerID)
		}
		fmt.Fprintf(out, "Broker: %s\n", formatOptional(label))
	}

	if details.TransportOrderID != "" {
		label := details.TransportOrderID
		if details.TransportOrderStatus != "" {
			parts := []string{}
			if details.TransportOrderStatus != "" {
				parts = append(parts, fmt.Sprintf("status %s", details.TransportOrderStatus))
			}
			label = fmt.Sprintf("%s (%s)", details.TransportOrderID, strings.Join(parts, ", "))
		}
		fmt.Fprintf(out, "Transport Order: %s\n", label)
	}

	if details.CreatedByID != "" || details.CreatedByName != "" {
		label := details.CreatedByID
		if details.CreatedByName != "" {
			label = fmt.Sprintf("%s (%s)", details.CreatedByName, details.CreatedByID)
		}
		fmt.Fprintf(out, "Created By: %s\n", formatOptional(label))
	}

	fmt.Fprintln(out, "Import Errors:")
	if details.ImportErrors == nil {
		fmt.Fprintln(out, "  (none)")
	} else {
		formatted := formatAny(details.ImportErrors)
		if formatted == "" {
			fmt.Fprintln(out, "  (none)")
		} else {
			fmt.Fprintln(out, indentLines(formatted, "  "))
		}
	}

	fmt.Fprintln(out, "Tables:")
	if details.Tables == nil {
		fmt.Fprintln(out, "  (none)")
	} else {
		formatted := formatAny(details.Tables)
		if formatted == "" {
			fmt.Fprintln(out, "  (none)")
		} else {
			fmt.Fprintln(out, indentLines(formatted, "  "))
		}
	}

	return nil
}
