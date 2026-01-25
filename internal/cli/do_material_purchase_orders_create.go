package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type doMaterialPurchaseOrdersCreateOptions struct {
	BaseURL                 string
	Token                   string
	JSON                    bool
	Status                  string
	IsManagingRedemption    bool
	TransactionAtMin        string
	TransactionAtMax        string
	Quantity                string
	ExternalPurchaseOrderID string
	ExternalSalesOrderID    string
	BrokerID                string
	MaterialSupplierID      string
	CustomerID              string
	MaterialTypeID          string
	MaterialSiteID          string
	JobSiteID               string
	UnitOfMeasureID         string
}

func newDoMaterialPurchaseOrdersCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a material purchase order",
		Long: `Create a material purchase order.

Required flags:
  --broker             Broker ID
  --material-supplier  Material supplier ID
  --material-type      Material type ID
  --unit-of-measure    Unit of measure ID (load or ton)
  --quantity           Ordered quantity

Optional flags:
  --status                     Initial status (editing, approved, closed)
  --is-managing-redemption     Whether this order manages redemption
  --transaction-at-min         Minimum transaction datetime (ISO 8601)
  --transaction-at-max         Maximum transaction datetime (ISO 8601)
  --external-purchase-order-id External purchase order ID
  --external-sales-order-id    External sales order ID

Relationships:
  --customer        Customer ID (required if --job-site is set)
  --material-site   Material site ID
  --job-site        Job site ID

Notes:
  Job site requires customer to be set. Material site must belong to the
  material supplier.`,
		Example: `  # Create a material purchase order
  xbe do material-purchase-orders create \
    --broker 123 \
    --material-supplier 456 \
    --material-type 789 \
    --unit-of-measure 10 \
    --quantity 500

  # Create with customer and job site
  xbe do material-purchase-orders create \
    --broker 123 \
    --material-supplier 456 \
    --material-type 789 \
    --unit-of-measure 10 \
    --quantity 500 \
    --customer 222 \
    --job-site 333`,
		RunE: runDoMaterialPurchaseOrdersCreate,
	}
	initDoMaterialPurchaseOrdersCreateFlags(cmd)
	return cmd
}

func init() {
	doMaterialPurchaseOrdersCmd.AddCommand(newDoMaterialPurchaseOrdersCreateCmd())
}

func initDoMaterialPurchaseOrdersCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("status", "", "Initial status (editing,approved,closed)")
	cmd.Flags().Bool("is-managing-redemption", false, "Whether this order manages redemption")
	cmd.Flags().String("transaction-at-min", "", "Minimum transaction datetime (ISO 8601)")
	cmd.Flags().String("transaction-at-max", "", "Maximum transaction datetime (ISO 8601)")
	cmd.Flags().String("quantity", "", "Ordered quantity (required)")
	cmd.Flags().String("external-purchase-order-id", "", "External purchase order ID")
	cmd.Flags().String("external-sales-order-id", "", "External sales order ID")
	cmd.Flags().String("broker", "", "Broker ID (required)")
	cmd.Flags().String("material-supplier", "", "Material supplier ID (required)")
	cmd.Flags().String("customer", "", "Customer ID")
	cmd.Flags().String("material-type", "", "Material type ID (required)")
	cmd.Flags().String("material-site", "", "Material site ID")
	cmd.Flags().String("job-site", "", "Job site ID")
	cmd.Flags().String("unit-of-measure", "", "Unit of measure ID (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	cmd.MarkFlagRequired("broker")
	cmd.MarkFlagRequired("material-supplier")
	cmd.MarkFlagRequired("material-type")
	cmd.MarkFlagRequired("unit-of-measure")
	cmd.MarkFlagRequired("quantity")
}

func runDoMaterialPurchaseOrdersCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoMaterialPurchaseOrdersCreateOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if strings.TrimSpace(opts.Token) == "" {
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

	attributes := map[string]any{
		"quantity": opts.Quantity,
	}

	if opts.Status != "" {
		attributes["status"] = opts.Status
	}
	if cmd.Flags().Changed("is-managing-redemption") {
		attributes["is-managing-redemption"] = opts.IsManagingRedemption
	}
	if opts.TransactionAtMin != "" {
		attributes["transaction-at-min"] = opts.TransactionAtMin
	}
	if opts.TransactionAtMax != "" {
		attributes["transaction-at-max"] = opts.TransactionAtMax
	}
	if opts.ExternalPurchaseOrderID != "" {
		attributes["external-purchase-order-id"] = opts.ExternalPurchaseOrderID
	}
	if opts.ExternalSalesOrderID != "" {
		attributes["external-sales-order-id"] = opts.ExternalSalesOrderID
	}

	relationships := map[string]any{
		"broker": map[string]any{
			"data": map[string]any{
				"type": "brokers",
				"id":   opts.BrokerID,
			},
		},
		"material-supplier": map[string]any{
			"data": map[string]any{
				"type": "material-suppliers",
				"id":   opts.MaterialSupplierID,
			},
		},
		"material-type": map[string]any{
			"data": map[string]any{
				"type": "material-types",
				"id":   opts.MaterialTypeID,
			},
		},
		"unit-of-measure": map[string]any{
			"data": map[string]any{
				"type": "unit-of-measures",
				"id":   opts.UnitOfMeasureID,
			},
		},
	}

	if opts.CustomerID != "" {
		relationships["customer"] = map[string]any{
			"data": map[string]any{
				"type": "customers",
				"id":   opts.CustomerID,
			},
		}
	}
	if opts.MaterialSiteID != "" {
		relationships["material-site"] = map[string]any{
			"data": map[string]any{
				"type": "material-sites",
				"id":   opts.MaterialSiteID,
			},
		}
	}
	if opts.JobSiteID != "" {
		relationships["job-site"] = map[string]any{
			"data": map[string]any{
				"type": "job-sites",
				"id":   opts.JobSiteID,
			},
		}
	}

	data := map[string]any{
		"type":          "material-purchase-orders",
		"attributes":    attributes,
		"relationships": relationships,
	}

	requestBody := map[string]any{
		"data": data,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/material-purchase-orders", jsonBody)
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

	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), map[string]string{
			"id":     resp.Data.ID,
			"status": stringAttr(resp.Data.Attributes, "status"),
		})
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created material purchase order %s\n", resp.Data.ID)
	return nil
}

func parseDoMaterialPurchaseOrdersCreateOptions(cmd *cobra.Command) (doMaterialPurchaseOrdersCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	status, _ := cmd.Flags().GetString("status")
	isManagingRedemption, _ := cmd.Flags().GetBool("is-managing-redemption")
	transactionAtMin, _ := cmd.Flags().GetString("transaction-at-min")
	transactionAtMax, _ := cmd.Flags().GetString("transaction-at-max")
	quantity, _ := cmd.Flags().GetString("quantity")
	externalPurchaseOrderID, _ := cmd.Flags().GetString("external-purchase-order-id")
	externalSalesOrderID, _ := cmd.Flags().GetString("external-sales-order-id")
	brokerID, _ := cmd.Flags().GetString("broker")
	materialSupplierID, _ := cmd.Flags().GetString("material-supplier")
	customerID, _ := cmd.Flags().GetString("customer")
	materialTypeID, _ := cmd.Flags().GetString("material-type")
	materialSiteID, _ := cmd.Flags().GetString("material-site")
	jobSiteID, _ := cmd.Flags().GetString("job-site")
	unitOfMeasureID, _ := cmd.Flags().GetString("unit-of-measure")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doMaterialPurchaseOrdersCreateOptions{
		BaseURL:                 baseURL,
		Token:                   token,
		JSON:                    jsonOut,
		Status:                  status,
		IsManagingRedemption:    isManagingRedemption,
		TransactionAtMin:        transactionAtMin,
		TransactionAtMax:        transactionAtMax,
		Quantity:                quantity,
		ExternalPurchaseOrderID: externalPurchaseOrderID,
		ExternalSalesOrderID:    externalSalesOrderID,
		BrokerID:                brokerID,
		MaterialSupplierID:      materialSupplierID,
		CustomerID:              customerID,
		MaterialTypeID:          materialTypeID,
		MaterialSiteID:          materialSiteID,
		JobSiteID:               jobSiteID,
		UnitOfMeasureID:         unitOfMeasureID,
	}, nil
}
