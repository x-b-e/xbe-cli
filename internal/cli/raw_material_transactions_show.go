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

type rawMaterialTransactionsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type rawMaterialTransactionDetails struct {
	ID                        string `json:"id"`
	UniqueID                  string `json:"uniqueid,omitempty"`
	Version                   string `json:"version,omitempty"`
	TicketNumber              string `json:"ticket_number,omitempty"`
	TicketJobNumber           string `json:"ticket_job_number,omitempty"`
	TransactionAt             string `json:"transaction_at,omitempty"`
	TruckName                 string `json:"truck_name,omitempty"`
	TruckerName               string `json:"trucker_name,omitempty"`
	MaterialName              string `json:"material_name,omitempty"`
	SiteID                    string `json:"site_id,omitempty"`
	SalesCustomerID           string `json:"sales_customer_id,omitempty"`
	HaulerType                string `json:"hauler_type,omitempty"`
	Weighmaster               string `json:"weighmaster,omitempty"`
	RawDataUniqueID           string `json:"raw_data_uniqueid,omitempty"`
	RawData                   any    `json:"raw_data,omitempty"`
	RawAttributes             any    `json:"raw_attributes,omitempty"`
	SourceID                  string `json:"material_site_id,omitempty"`
	SourceName                string `json:"material_site_name,omitempty"`
	MaterialTransactionID     string `json:"material_transaction_id,omitempty"`
	MaterialTransactionTicket string `json:"material_transaction_ticket_number,omitempty"`
	MaterialTransactionStatus string `json:"material_transaction_status,omitempty"`
}

func newRawMaterialTransactionsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show raw material transaction details",
		Long: `Show full details of a raw material transaction.

Raw material transactions contain the raw ticket data ingested from material
sites and any linkage to a processed material transaction.

Output Fields:
  ID                  Raw material transaction ID
  Unique ID           Source unique identifier
  Ticket Number       Ticket identifier from the raw record
  Ticket Job Number   Job number from the ticket
  Transaction At      Timestamp of the transaction
  Truck/Trucker       Truck and trucker information
  Material/Site       Material identifier and raw site ID
  Sales Customer      Sales customer ID (if present)
  Hauler Type         Hauler classification
  Weighmaster         Weighmaster name (if present)
  Raw Data            Flattened raw ticket data
  Raw Attributes      Additional raw attributes (admins only)
  Source              Material site relationship (if available)
  Material Transaction Linked processed material transaction

Arguments:
  <id>    Raw material transaction ID (required)

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a raw material transaction
  xbe view raw-material-transactions show 123

  # JSON output
  xbe view raw-material-transactions show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runRawMaterialTransactionsShow,
	}
	initRawMaterialTransactionsShowFlags(cmd)
	return cmd
}

func init() {
	rawMaterialTransactionsCmd.AddCommand(newRawMaterialTransactionsShowCmd())
}

func initRawMaterialTransactionsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runRawMaterialTransactionsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseRawMaterialTransactionsShowOptions(cmd)
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
		return fmt.Errorf("raw material transaction id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[raw-material-transactions]", "ticket-job-number,uniqueid,ticket-number,transaction-at,truck-name,trucker-name,site-id,material-name,sales-customer-id,hauler-type,weighmaster,raw-data-uniqueid,raw-data,raw-attributes,version,source,material-transaction")
	query.Set("fields[material-sites]", "name")
	query.Set("fields[material-transactions]", "status,ticket-number,transaction-at")
	query.Set("include", "source,material-transaction")

	body, _, err := client.Get(cmd.Context(), "/v1/raw-material-transactions/"+id, query)
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

	details := buildRawMaterialTransactionDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderRawMaterialTransactionDetails(cmd, details)
}

func parseRawMaterialTransactionsShowOptions(cmd *cobra.Command) (rawMaterialTransactionsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return rawMaterialTransactionsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildRawMaterialTransactionDetails(resp jsonAPISingleResponse) rawMaterialTransactionDetails {
	attrs := resp.Data.Attributes

	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	details := rawMaterialTransactionDetails{
		ID:              resp.Data.ID,
		UniqueID:        stringAttr(attrs, "uniqueid"),
		Version:         stringAttr(attrs, "version"),
		TicketNumber:    stringAttr(attrs, "ticket-number"),
		TicketJobNumber: stringAttr(attrs, "ticket-job-number"),
		TransactionAt:   stringAttr(attrs, "transaction-at"),
		TruckName:       stringAttr(attrs, "truck-name"),
		TruckerName:     stringAttr(attrs, "trucker-name"),
		MaterialName:    stringAttr(attrs, "material-name"),
		SiteID:          stringAttr(attrs, "site-id"),
		SalesCustomerID: stringAttr(attrs, "sales-customer-id"),
		HaulerType:      stringAttr(attrs, "hauler-type"),
		Weighmaster:     stringAttr(attrs, "weighmaster"),
		RawDataUniqueID: stringAttr(attrs, "raw-data-uniqueid"),
		RawData:         anyAttr(attrs, "raw-data"),
		RawAttributes:   anyAttr(attrs, "raw-attributes"),
	}

	if rel, ok := resp.Data.Relationships["source"]; ok && rel.Data != nil {
		details.SourceID = rel.Data.ID
		if source, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.SourceName = stringAttr(source.Attributes, "name")
		}
	}
	if rel, ok := resp.Data.Relationships["material-transaction"]; ok && rel.Data != nil {
		details.MaterialTransactionID = rel.Data.ID
		if mt, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.MaterialTransactionTicket = stringAttr(mt.Attributes, "ticket-number")
			details.MaterialTransactionStatus = stringAttr(mt.Attributes, "status")
		}
	}

	return details
}

func renderRawMaterialTransactionDetails(cmd *cobra.Command, details rawMaterialTransactionDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	fmt.Fprintf(out, "Unique ID: %s\n", formatOptional(details.UniqueID))
	fmt.Fprintf(out, "Version: %s\n", formatOptional(details.Version))
	fmt.Fprintf(out, "Ticket Number: %s\n", formatOptional(details.TicketNumber))
	fmt.Fprintf(out, "Ticket Job Number: %s\n", formatOptional(details.TicketJobNumber))
	fmt.Fprintf(out, "Transaction At: %s\n", formatOptional(details.TransactionAt))
	fmt.Fprintf(out, "Truck: %s\n", formatOptional(details.TruckName))
	fmt.Fprintf(out, "Trucker: %s\n", formatOptional(details.TruckerName))
	fmt.Fprintf(out, "Material: %s\n", formatOptional(details.MaterialName))
	fmt.Fprintf(out, "Site ID: %s\n", formatOptional(details.SiteID))
	fmt.Fprintf(out, "Sales Customer ID: %s\n", formatOptional(details.SalesCustomerID))
	fmt.Fprintf(out, "Hauler Type: %s\n", formatOptional(details.HaulerType))
	fmt.Fprintf(out, "Weighmaster: %s\n", formatOptional(details.Weighmaster))
	fmt.Fprintf(out, "Raw Data Unique ID: %s\n", formatOptional(details.RawDataUniqueID))

	if details.SourceID != "" || details.SourceName != "" {
		label := details.SourceID
		if details.SourceName != "" {
			label = fmt.Sprintf("%s (%s)", details.SourceName, details.SourceID)
		}
		fmt.Fprintf(out, "Source (Material Site): %s\n", formatOptional(label))
	}
	if details.MaterialTransactionID != "" {
		label := details.MaterialTransactionID
		if details.MaterialTransactionTicket != "" || details.MaterialTransactionStatus != "" {
			parts := []string{}
			if details.MaterialTransactionTicket != "" {
				parts = append(parts, fmt.Sprintf("ticket %s", details.MaterialTransactionTicket))
			}
			if details.MaterialTransactionStatus != "" {
				parts = append(parts, fmt.Sprintf("status %s", details.MaterialTransactionStatus))
			}
			label = fmt.Sprintf("%s (%s)", details.MaterialTransactionID, strings.Join(parts, ", "))
		}
		fmt.Fprintf(out, "Material Transaction: %s\n", label)
	}

	if details.RawData != nil {
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "Raw Data:")
		formatted := formatAny(details.RawData)
		if formatted == "" {
			fmt.Fprintln(out, "  (none)")
		} else {
			fmt.Fprintln(out, indentLines(formatted, "  "))
		}
	}

	if details.RawAttributes != nil {
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "Raw Attributes:")
		formatted := formatAny(details.RawAttributes)
		if formatted == "" {
			fmt.Fprintln(out, "  (none)")
		} else {
			fmt.Fprintln(out, indentLines(formatted, "  "))
		}
	}

	return nil
}
