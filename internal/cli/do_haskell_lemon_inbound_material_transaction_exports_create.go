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

type doHaskellLemonInboundMaterialTransactionExportsCreateOptions struct {
	BaseURL         string
	Token           string
	JSON            bool
	TransactionDate string
	ToAddresses     []string
	CcAddresses     []string
	IsTest          bool
}

func newDoHaskellLemonInboundMaterialTransactionExportsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a Haskell Lemon inbound material transaction export",
		Long: `Create a Haskell Lemon inbound material transaction export.

Required flags:
  --transaction-date  Transaction date (YYYY-MM-DD) (required)

Optional flags:
  --to-addresses  Recipient email addresses (comma-separated or repeated)
  --cc-addresses  CC email addresses (comma-separated or repeated)
  --is-test       Create a test export (requires --to-addresses)

Notes:
  When --is-test is not set, recipient defaults are used unless overridden.`,
		Example: `  # Create an export for a transaction date
  xbe do haskell-lemon-inbound-material-transaction-exports create \
    --transaction-date 2025-01-15

  # Create a test export with explicit recipients
  xbe do haskell-lemon-inbound-material-transaction-exports create \
    --transaction-date 2025-01-15 \
    --is-test \
    --to-addresses "ops@example.com"`,
		Args: cobra.NoArgs,
		RunE: runDoHaskellLemonInboundMaterialTransactionExportsCreate,
	}
	initDoHaskellLemonInboundMaterialTransactionExportsCreateFlags(cmd)
	return cmd
}

func init() {
	doHaskellLemonInboundMaterialTransactionExportsCmd.AddCommand(newDoHaskellLemonInboundMaterialTransactionExportsCreateCmd())
}

func initDoHaskellLemonInboundMaterialTransactionExportsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("transaction-date", "", "Transaction date (YYYY-MM-DD) (required)")
	cmd.Flags().StringSlice("to-addresses", nil, "Recipient email addresses (comma-separated or repeated)")
	cmd.Flags().StringSlice("cc-addresses", nil, "CC email addresses (comma-separated or repeated)")
	cmd.Flags().Bool("is-test", false, "Create a test export (requires --to-addresses)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoHaskellLemonInboundMaterialTransactionExportsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoHaskellLemonInboundMaterialTransactionExportsCreateOptions(cmd)
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

	opts.TransactionDate = strings.TrimSpace(opts.TransactionDate)
	if opts.TransactionDate == "" {
		err := fmt.Errorf("--transaction-date is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	toAddresses := normalizeStringSlice(opts.ToAddresses)
	ccAddresses := normalizeStringSlice(opts.CcAddresses)

	if cmd.Flags().Changed("is-test") && opts.IsTest && len(toAddresses) == 0 {
		err := fmt.Errorf("--to-addresses is required when --is-test is set")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{
		"transaction-date": opts.TransactionDate,
	}

	if len(toAddresses) > 0 {
		attributes["to-addresses"] = toAddresses
	}
	if len(ccAddresses) > 0 {
		attributes["cc-addresses"] = ccAddresses
	}
	if cmd.Flags().Changed("is-test") {
		attributes["is-test"] = opts.IsTest
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "haskell-lemon-inbound-material-transaction-exports",
			"attributes": attributes,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/haskell-lemon-inbound-material-transaction-exports", jsonBody)
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

	row := haskellLemonInboundMaterialTransactionExportRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created Haskell Lemon inbound material transaction export %s (%s)\n", row.ID, row.TransactionDate)
	return nil
}

func parseDoHaskellLemonInboundMaterialTransactionExportsCreateOptions(cmd *cobra.Command) (doHaskellLemonInboundMaterialTransactionExportsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	transactionDate, _ := cmd.Flags().GetString("transaction-date")
	toAddresses, _ := cmd.Flags().GetStringSlice("to-addresses")
	ccAddresses, _ := cmd.Flags().GetStringSlice("cc-addresses")
	isTest, _ := cmd.Flags().GetBool("is-test")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doHaskellLemonInboundMaterialTransactionExportsCreateOptions{
		BaseURL:         baseURL,
		Token:           token,
		JSON:            jsonOut,
		TransactionDate: transactionDate,
		ToAddresses:     toAddresses,
		CcAddresses:     ccAddresses,
		IsTest:          isTest,
	}, nil
}

func haskellLemonInboundMaterialTransactionExportRowFromSingle(resp jsonAPISingleResponse) haskellLemonInboundMaterialTransactionExportRow {
	attrs := resp.Data.Attributes
	return haskellLemonInboundMaterialTransactionExportRow{
		ID:              resp.Data.ID,
		TransactionDate: formatDate(stringAttr(attrs, "transaction-date")),
		IsTest:          boolAttr(attrs, "is-test"),
		CreatedAt:       formatDateTime(stringAttr(attrs, "created-at")),
	}
}

func normalizeStringSlice(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	cleaned := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		cleaned = append(cleaned, trimmed)
	}
	return cleaned
}
