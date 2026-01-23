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

type doMaterialPurchaseOrdersUpdateOptions struct {
	BaseURL                 string
	Token                   string
	JSON                    bool
	ID                      string
	Status                  string
	IsManagingRedemption    bool
	TransactionAtMin        string
	TransactionAtMax        string
	Quantity                string
	ExternalPurchaseOrderID string
	ExternalSalesOrderID    string
	CustomerID              string
	MaterialTypeID          string
	MaterialSiteID          string
	JobSiteID               string
	UnitOfMeasureID         string
}

func newDoMaterialPurchaseOrdersUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a material purchase order",
		Long: `Update a material purchase order.

Optional flags:
  --status                     Set status (editing, approved, closed)
  --is-managing-redemption     Enable redemption management
  --no-is-managing-redemption  Disable redemption management
  --transaction-at-min         Minimum transaction datetime (ISO 8601)
  --transaction-at-max         Maximum transaction datetime (ISO 8601)
  --quantity                   Ordered quantity
  --external-purchase-order-id External purchase order ID
  --external-sales-order-id    External sales order ID

Relationships:
  --customer        Customer ID (empty to clear)
  --material-type   Material type ID
  --material-site   Material site ID (empty to clear)
  --job-site        Job site ID (empty to clear)
  --unit-of-measure Unit of measure ID`,
		Example: `  # Update quantity
  xbe do material-purchase-orders update 123 --quantity 750

  # Update status
  xbe do material-purchase-orders update 123 --status approved

  # Clear job site
  xbe do material-purchase-orders update 123 --job-site ""`,
		Args: cobra.ExactArgs(1),
		RunE: runDoMaterialPurchaseOrdersUpdate,
	}
	initDoMaterialPurchaseOrdersUpdateFlags(cmd)
	return cmd
}

func init() {
	doMaterialPurchaseOrdersCmd.AddCommand(newDoMaterialPurchaseOrdersUpdateCmd())
}

func initDoMaterialPurchaseOrdersUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("status", "", "Set status (editing,approved,closed)")
	cmd.Flags().Bool("is-managing-redemption", false, "Enable redemption management")
	cmd.Flags().Bool("no-is-managing-redemption", false, "Disable redemption management")
	cmd.Flags().String("transaction-at-min", "", "Minimum transaction datetime (ISO 8601)")
	cmd.Flags().String("transaction-at-max", "", "Maximum transaction datetime (ISO 8601)")
	cmd.Flags().String("quantity", "", "Ordered quantity")
	cmd.Flags().String("external-purchase-order-id", "", "External purchase order ID")
	cmd.Flags().String("external-sales-order-id", "", "External sales order ID")
	cmd.Flags().String("customer", "", "Customer ID (empty to clear)")
	cmd.Flags().String("material-type", "", "Material type ID")
	cmd.Flags().String("material-site", "", "Material site ID (empty to clear)")
	cmd.Flags().String("job-site", "", "Job site ID (empty to clear)")
	cmd.Flags().String("unit-of-measure", "", "Unit of measure ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoMaterialPurchaseOrdersUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoMaterialPurchaseOrdersUpdateOptions(cmd, args)
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

	attributes := map[string]any{}
	relationships := map[string]any{}
	hasChanges := false

	if cmd.Flags().Changed("status") {
		attributes["status"] = opts.Status
		hasChanges = true
	}
	if cmd.Flags().Changed("is-managing-redemption") {
		attributes["is-managing-redemption"] = true
		hasChanges = true
	}
	if cmd.Flags().Changed("no-is-managing-redemption") {
		attributes["is-managing-redemption"] = false
		hasChanges = true
	}
	if cmd.Flags().Changed("transaction-at-min") {
		attributes["transaction-at-min"] = opts.TransactionAtMin
		hasChanges = true
	}
	if cmd.Flags().Changed("transaction-at-max") {
		attributes["transaction-at-max"] = opts.TransactionAtMax
		hasChanges = true
	}
	if cmd.Flags().Changed("quantity") {
		attributes["quantity"] = opts.Quantity
		hasChanges = true
	}
	if cmd.Flags().Changed("external-purchase-order-id") {
		attributes["external-purchase-order-id"] = opts.ExternalPurchaseOrderID
		hasChanges = true
	}
	if cmd.Flags().Changed("external-sales-order-id") {
		attributes["external-sales-order-id"] = opts.ExternalSalesOrderID
		hasChanges = true
	}

	if cmd.Flags().Changed("customer") {
		if opts.CustomerID == "" {
			relationships["customer"] = map[string]any{"data": nil}
		} else {
			relationships["customer"] = map[string]any{
				"data": map[string]any{
					"type": "customers",
					"id":   opts.CustomerID,
				},
			}
		}
		hasChanges = true
	}
	if cmd.Flags().Changed("material-type") {
		if opts.MaterialTypeID == "" {
			relationships["material-type"] = map[string]any{"data": nil}
		} else {
			relationships["material-type"] = map[string]any{
				"data": map[string]any{
					"type": "material-types",
					"id":   opts.MaterialTypeID,
				},
			}
		}
		hasChanges = true
	}
	if cmd.Flags().Changed("material-site") {
		if opts.MaterialSiteID == "" {
			relationships["material-site"] = map[string]any{"data": nil}
		} else {
			relationships["material-site"] = map[string]any{
				"data": map[string]any{
					"type": "material-sites",
					"id":   opts.MaterialSiteID,
				},
			}
		}
		hasChanges = true
	}
	if cmd.Flags().Changed("job-site") {
		if opts.JobSiteID == "" {
			relationships["job-site"] = map[string]any{"data": nil}
		} else {
			relationships["job-site"] = map[string]any{
				"data": map[string]any{
					"type": "job-sites",
					"id":   opts.JobSiteID,
				},
			}
		}
		hasChanges = true
	}
	if cmd.Flags().Changed("unit-of-measure") {
		if opts.UnitOfMeasureID == "" {
			relationships["unit-of-measure"] = map[string]any{"data": nil}
		} else {
			relationships["unit-of-measure"] = map[string]any{
				"data": map[string]any{
					"type": "unit-of-measures",
					"id":   opts.UnitOfMeasureID,
				},
			}
		}
		hasChanges = true
	}

	if !hasChanges {
		err := fmt.Errorf("no attributes to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	data := map[string]any{
		"type": "material-purchase-orders",
		"id":   opts.ID,
	}
	if len(attributes) > 0 {
		data["attributes"] = attributes
	}
	if len(relationships) > 0 {
		data["relationships"] = relationships
	}

	requestBody := map[string]any{"data": data}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Patch(cmd.Context(), "/v1/material-purchase-orders/"+opts.ID, jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Updated material purchase order %s\n", resp.Data.ID)
	return nil
}

func parseDoMaterialPurchaseOrdersUpdateOptions(cmd *cobra.Command, args []string) (doMaterialPurchaseOrdersUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	status, _ := cmd.Flags().GetString("status")
	isManagingRedemption, _ := cmd.Flags().GetBool("is-managing-redemption")
	transactionAtMin, _ := cmd.Flags().GetString("transaction-at-min")
	transactionAtMax, _ := cmd.Flags().GetString("transaction-at-max")
	quantity, _ := cmd.Flags().GetString("quantity")
	externalPurchaseOrderID, _ := cmd.Flags().GetString("external-purchase-order-id")
	externalSalesOrderID, _ := cmd.Flags().GetString("external-sales-order-id")
	customerID, _ := cmd.Flags().GetString("customer")
	materialTypeID, _ := cmd.Flags().GetString("material-type")
	materialSiteID, _ := cmd.Flags().GetString("material-site")
	jobSiteID, _ := cmd.Flags().GetString("job-site")
	unitOfMeasureID, _ := cmd.Flags().GetString("unit-of-measure")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doMaterialPurchaseOrdersUpdateOptions{
		BaseURL:                 baseURL,
		Token:                   token,
		JSON:                    jsonOut,
		ID:                      args[0],
		Status:                  status,
		IsManagingRedemption:    isManagingRedemption,
		TransactionAtMin:        transactionAtMin,
		TransactionAtMax:        transactionAtMax,
		Quantity:                quantity,
		ExternalPurchaseOrderID: externalPurchaseOrderID,
		ExternalSalesOrderID:    externalSalesOrderID,
		CustomerID:              customerID,
		MaterialTypeID:          materialTypeID,
		MaterialSiteID:          materialSiteID,
		JobSiteID:               jobSiteID,
		UnitOfMeasureID:         unitOfMeasureID,
	}, nil
}
