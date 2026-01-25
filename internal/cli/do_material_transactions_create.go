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

type doMaterialTransactionsCreateOptions struct {
	BaseURL           string
	Token             string
	JSON              bool
	TransactionAt     string
	TicketNumber      string
	TicketBOLNumber   string
	TicketItemNumber  string
	TareWeightLbs     string
	GrossWeightLbs    string
	NetWeightLbs      string
	MaxGVMWeightLbs   string
	Trip              string
	MaterialType      string
	MaterialSite      string
	MaterialMixDesign string
	CostCode          string
	SalesCustomer     string
}

func newDoMaterialTransactionsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new material transaction",
		Long: `Create a new material transaction.

Key attributes:
  --transaction-at     Transaction datetime (ISO 8601)
  --ticket-number      Ticket number
  --ticket-bol-number  Bill of lading number
  --ticket-item-number Ticket item number
  --tare-weight-lbs    Tare weight in pounds
  --gross-weight-lbs   Gross weight in pounds
  --net-weight-lbs     Net weight in pounds
  --max-gvm-weight-lbs Maximum GVM weight in pounds

Relationships:
  --trip               Trip ID
  --material-type      Material type ID
  --material-site      Material site ID
  --material-mix-design Material mix design ID
  --cost-code          Cost code ID
  --sales-customer     Sales customer ID`,
		Example: `  # Create a material transaction with weights
  xbe do material-transactions create --ticket-number "T12345" --net-weight-lbs 40000 --material-type 123 --material-site 456

  # Create a material transaction with a trip
  xbe do material-transactions create --trip 789 --ticket-number "T12345" --material-type 123`,
		RunE: runDoMaterialTransactionsCreate,
	}
	initDoMaterialTransactionsCreateFlags(cmd)
	return cmd
}

func init() {
	doMaterialTransactionsCmd.AddCommand(newDoMaterialTransactionsCreateCmd())
}

func initDoMaterialTransactionsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("transaction-at", "", "Transaction datetime (ISO 8601)")
	cmd.Flags().String("ticket-number", "", "Ticket number")
	cmd.Flags().String("ticket-bol-number", "", "Bill of lading number")
	cmd.Flags().String("ticket-item-number", "", "Ticket item number")
	cmd.Flags().String("tare-weight-lbs", "", "Tare weight in pounds")
	cmd.Flags().String("gross-weight-lbs", "", "Gross weight in pounds")
	cmd.Flags().String("net-weight-lbs", "", "Net weight in pounds")
	cmd.Flags().String("max-gvm-weight-lbs", "", "Maximum GVM weight in pounds")
	cmd.Flags().String("trip", "", "Trip ID")
	cmd.Flags().String("material-type", "", "Material type ID")
	cmd.Flags().String("material-site", "", "Material site ID")
	cmd.Flags().String("material-mix-design", "", "Material mix design ID")
	cmd.Flags().String("cost-code", "", "Cost code ID")
	cmd.Flags().String("sales-customer", "", "Sales customer ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoMaterialTransactionsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoMaterialTransactionsCreateOptions(cmd)
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

	if opts.TransactionAt != "" {
		attributes["transaction-at"] = opts.TransactionAt
	}
	if opts.TicketNumber != "" {
		attributes["ticket-number"] = opts.TicketNumber
	}
	if opts.TicketBOLNumber != "" {
		attributes["ticket-bol-number"] = opts.TicketBOLNumber
	}
	if opts.TicketItemNumber != "" {
		attributes["ticket-item-number"] = opts.TicketItemNumber
	}
	if opts.TareWeightLbs != "" {
		attributes["tare-weight-lbs"] = opts.TareWeightLbs
	}
	if opts.GrossWeightLbs != "" {
		attributes["gross-weight-lbs"] = opts.GrossWeightLbs
	}
	if opts.NetWeightLbs != "" {
		attributes["net-weight-lbs"] = opts.NetWeightLbs
	}
	if opts.MaxGVMWeightLbs != "" {
		attributes["max-gvm-weight-lbs"] = opts.MaxGVMWeightLbs
	}

	relationships := map[string]any{}

	if opts.Trip != "" {
		relationships["trip"] = map[string]any{
			"data": map[string]any{
				"type": "trips",
				"id":   opts.Trip,
			},
		}
	}
	if opts.MaterialType != "" {
		relationships["material-type"] = map[string]any{
			"data": map[string]any{
				"type": "material-types",
				"id":   opts.MaterialType,
			},
		}
	}
	if opts.MaterialSite != "" {
		relationships["material-site"] = map[string]any{
			"data": map[string]any{
				"type": "material-sites",
				"id":   opts.MaterialSite,
			},
		}
	}
	if opts.MaterialMixDesign != "" {
		relationships["material-mix-design"] = map[string]any{
			"data": map[string]any{
				"type": "material-mix-designs",
				"id":   opts.MaterialMixDesign,
			},
		}
	}
	if opts.CostCode != "" {
		relationships["cost-code"] = map[string]any{
			"data": map[string]any{
				"type": "cost-codes",
				"id":   opts.CostCode,
			},
		}
	}
	if opts.SalesCustomer != "" {
		relationships["sales-customer"] = map[string]any{
			"data": map[string]any{
				"type": "customers",
				"id":   opts.SalesCustomer,
			},
		}
	}

	data := map[string]any{
		"type":       "material-transactions",
		"attributes": attributes,
	}

	if len(relationships) > 0 {
		data["relationships"] = relationships
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

	body, _, err := client.Post(cmd.Context(), "/v1/material-transactions", jsonBody)
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

	row := materialTransactionRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created material transaction %s (%s)\n", row.ID, row.TicketNumber)
	return nil
}

func parseDoMaterialTransactionsCreateOptions(cmd *cobra.Command) (doMaterialTransactionsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	transactionAt, _ := cmd.Flags().GetString("transaction-at")
	ticketNumber, _ := cmd.Flags().GetString("ticket-number")
	ticketBOLNumber, _ := cmd.Flags().GetString("ticket-bol-number")
	ticketItemNumber, _ := cmd.Flags().GetString("ticket-item-number")
	tareWeightLbs, _ := cmd.Flags().GetString("tare-weight-lbs")
	grossWeightLbs, _ := cmd.Flags().GetString("gross-weight-lbs")
	netWeightLbs, _ := cmd.Flags().GetString("net-weight-lbs")
	maxGVMWeightLbs, _ := cmd.Flags().GetString("max-gvm-weight-lbs")
	trip, _ := cmd.Flags().GetString("trip")
	materialType, _ := cmd.Flags().GetString("material-type")
	materialSite, _ := cmd.Flags().GetString("material-site")
	materialMixDesign, _ := cmd.Flags().GetString("material-mix-design")
	costCode, _ := cmd.Flags().GetString("cost-code")
	salesCustomer, _ := cmd.Flags().GetString("sales-customer")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doMaterialTransactionsCreateOptions{
		BaseURL:           baseURL,
		Token:             token,
		JSON:              jsonOut,
		TransactionAt:     transactionAt,
		TicketNumber:      ticketNumber,
		TicketBOLNumber:   ticketBOLNumber,
		TicketItemNumber:  ticketItemNumber,
		TareWeightLbs:     tareWeightLbs,
		GrossWeightLbs:    grossWeightLbs,
		NetWeightLbs:      netWeightLbs,
		MaxGVMWeightLbs:   maxGVMWeightLbs,
		Trip:              trip,
		MaterialType:      materialType,
		MaterialSite:      materialSite,
		MaterialMixDesign: materialMixDesign,
		CostCode:          costCode,
		SalesCustomer:     salesCustomer,
	}, nil
}

type materialTransactionCreateRow struct {
	ID           string `json:"id"`
	Status       string `json:"status"`
	TicketNumber string `json:"ticket_number"`
}

func materialTransactionRowFromSingle(resp jsonAPISingleResponse) materialTransactionCreateRow {
	return materialTransactionCreateRow{
		ID:           resp.Data.ID,
		Status:       stringAttr(resp.Data.Attributes, "status"),
		TicketNumber: stringAttr(resp.Data.Attributes, "ticket-number"),
	}
}
