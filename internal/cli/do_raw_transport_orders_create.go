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

type doRawTransportOrdersCreateOptions struct {
	BaseURL             string
	Token               string
	JSON                bool
	ExternalOrderNumber string
	Tables              string
	TablesRowversionMin string
	TablesRowversionMax string
	IsManaged           bool
	Importer            string
	Broker              string
}

func newDoRawTransportOrdersCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a raw transport order",
		Long: `Create a raw transport order.

Required flags:
  --external-order-number  External order number from the import source
  --broker                 Broker ID

Optional flags:
  --importer               Importer key (e.g., quantix_tmw)
  --is-managed             Mark as managed
  --tables                 Raw tables JSON array
  --tables-rowversion-min  Minimum tables rowversion
  --tables-rowversion-max  Maximum tables rowversion

Relationships:
  --broker                 Broker ID`,
		Example: `  # Create a raw transport order
  xbe do raw-transport-orders create --external-order-number ORD-1001 --broker 123

  # Create with importer and managed flag
  xbe do raw-transport-orders create --external-order-number ORD-1002 --broker 123 \
    --importer quantix_tmw --is-managed

  # Create with tables payload
  xbe do raw-transport-orders create --external-order-number ORD-1003 --broker 123 \
    --tables '[{"table_name":"orderheader","query":"select *","primary_key_column":"ord_hdrnumber","rows":[{"columns":[{"key":"ord_hdrnumber","value":"ORD-1003"}]}]}]'`,
		RunE: runDoRawTransportOrdersCreate,
	}
	initDoRawTransportOrdersCreateFlags(cmd)
	return cmd
}

func init() {
	doRawTransportOrdersCmd.AddCommand(newDoRawTransportOrdersCreateCmd())
}

func initDoRawTransportOrdersCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("external-order-number", "", "External order number (required)")
	cmd.Flags().String("tables", "", "Raw tables JSON array")
	cmd.Flags().String("tables-rowversion-min", "", "Minimum tables rowversion")
	cmd.Flags().String("tables-rowversion-max", "", "Maximum tables rowversion")
	cmd.Flags().Bool("is-managed", false, "Mark as managed")
	cmd.Flags().String("importer", "", "Importer key (e.g., quantix_tmw)")
	cmd.Flags().String("broker", "", "Broker ID (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	_ = cmd.MarkFlagRequired("external-order-number")
	_ = cmd.MarkFlagRequired("broker")
}

func runDoRawTransportOrdersCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoRawTransportOrdersCreateOptions(cmd)
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
		"external-order-number": opts.ExternalOrderNumber,
	}

	if opts.Tables != "" {
		tables, err := parseRawTransportOrderTables(opts.Tables)
		if err != nil {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		attributes["tables"] = tables
	}
	if opts.TablesRowversionMin != "" {
		attributes["tables-rowversion-min"] = opts.TablesRowversionMin
	}
	if opts.TablesRowversionMax != "" {
		attributes["tables-rowversion-max"] = opts.TablesRowversionMax
	}
	if cmd.Flags().Changed("is-managed") {
		attributes["is-managed"] = opts.IsManaged
	}
	if opts.Importer != "" {
		attributes["importer"] = opts.Importer
	}

	relationships := map[string]any{
		"broker": map[string]any{
			"data": map[string]any{
				"type": "brokers",
				"id":   opts.Broker,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "raw-transport-orders",
			"attributes":    attributes,
			"relationships": relationships,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/raw-transport-orders", jsonBody)
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

	row := buildRawTransportOrderRow(resp.Data)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	label := row.ExternalOrderNumber
	if label != "" {
		fmt.Fprintf(cmd.OutOrStdout(), "Created raw transport order %s (%s)\n", row.ID, label)
		return nil
	}
	fmt.Fprintf(cmd.OutOrStdout(), "Created raw transport order %s\n", row.ID)
	return nil
}

func parseDoRawTransportOrdersCreateOptions(cmd *cobra.Command) (doRawTransportOrdersCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	externalOrderNumber, _ := cmd.Flags().GetString("external-order-number")
	tables, _ := cmd.Flags().GetString("tables")
	rowversionMin, _ := cmd.Flags().GetString("tables-rowversion-min")
	rowversionMax, _ := cmd.Flags().GetString("tables-rowversion-max")
	isManaged, _ := cmd.Flags().GetBool("is-managed")
	importer, _ := cmd.Flags().GetString("importer")
	broker, _ := cmd.Flags().GetString("broker")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doRawTransportOrdersCreateOptions{
		BaseURL:             baseURL,
		Token:               token,
		JSON:                jsonOut,
		ExternalOrderNumber: externalOrderNumber,
		Tables:              tables,
		TablesRowversionMin: rowversionMin,
		TablesRowversionMax: rowversionMax,
		IsManaged:           isManaged,
		Importer:            importer,
		Broker:              broker,
	}, nil
}

func parseRawTransportOrderTables(raw string) ([]any, error) {
	var tables []any
	if err := json.Unmarshal([]byte(raw), &tables); err != nil {
		return nil, fmt.Errorf("invalid tables JSON (expected array): %w", err)
	}
	return tables, nil
}
