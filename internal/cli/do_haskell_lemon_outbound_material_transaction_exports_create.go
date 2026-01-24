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

type doHaskellLemonOutboundMaterialTransactionExportsCreateOptions struct {
	BaseURL         string
	Token           string
	JSON            bool
	TransactionDate string
	ToAddresses     []string
	CCAddresses     []string
	IsTest          bool
}

func newDoHaskellLemonOutboundMaterialTransactionExportsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a Haskell Lemon outbound material transaction export",
		Long: `Create a Haskell Lemon outbound material transaction export.

Required:
  --transaction-date   Transaction date for the export (YYYY-MM-DD)

Optional:
  --to-addresses       Email recipients (comma-separated or repeated)
  --cc-addresses       CC recipients (comma-separated or repeated)
  --is-test            Mark export as a test (requires --to-addresses)

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Create an export for a date
  xbe do haskell-lemon-outbound-material-transaction-exports create --transaction-date 2025-01-15

  # Create a test export
  xbe do haskell-lemon-outbound-material-transaction-exports create \\
    --transaction-date 2025-01-15 \\
    --is-test \\
    --to-addresses test@example.com \\
    --cc-addresses cc@example.com`,
		Args: cobra.NoArgs,
		RunE: runDoHaskellLemonOutboundMaterialTransactionExportsCreate,
	}
	initDoHaskellLemonOutboundMaterialTransactionExportsCreateFlags(cmd)
	return cmd
}

func init() {
	doHaskellLemonOutboundMaterialTransactionExportsCmd.AddCommand(newDoHaskellLemonOutboundMaterialTransactionExportsCreateCmd())
}

func initDoHaskellLemonOutboundMaterialTransactionExportsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("transaction-date", "", "Transaction date for the export (YYYY-MM-DD)")
	cmd.Flags().StringSlice("to-addresses", nil, "Email recipients (comma-separated or repeated)")
	cmd.Flags().StringSlice("cc-addresses", nil, "CC recipients (comma-separated or repeated)")
	cmd.Flags().Bool("is-test", false, "Mark export as a test (requires --to-addresses)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	_ = cmd.MarkFlagRequired("transaction-date")
}

func runDoHaskellLemonOutboundMaterialTransactionExportsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoHaskellLemonOutboundMaterialTransactionExportsCreateOptions(cmd)
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

	transactionDate := strings.TrimSpace(opts.TransactionDate)
	if transactionDate == "" {
		err := fmt.Errorf("--transaction-date is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	toAddresses := compactStringSlice(opts.ToAddresses)
	ccAddresses := compactStringSlice(opts.CCAddresses)
	if opts.IsTest && len(toAddresses) == 0 {
		err := fmt.Errorf("--to-addresses is required when --is-test is set")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{
		"transaction-date": transactionDate,
	}
	if cmd.Flags().Changed("to-addresses") {
		attributes["to-addresses"] = toAddresses
	}
	if cmd.Flags().Changed("cc-addresses") {
		attributes["cc-addresses"] = ccAddresses
	}
	if cmd.Flags().Changed("is-test") {
		attributes["is-test"] = opts.IsTest
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "haskell-lemon-outbound-material-transaction-exports",
			"attributes": attributes,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/haskell-lemon-outbound-material-transaction-exports", jsonBody)
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

	row := haskellLemonOutboundMaterialTransactionExportRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created Haskell Lemon outbound material transaction export %s\n", row.ID)
	return nil
}

func parseDoHaskellLemonOutboundMaterialTransactionExportsCreateOptions(cmd *cobra.Command) (doHaskellLemonOutboundMaterialTransactionExportsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	transactionDate, _ := cmd.Flags().GetString("transaction-date")
	toAddresses, _ := cmd.Flags().GetStringSlice("to-addresses")
	ccAddresses, _ := cmd.Flags().GetStringSlice("cc-addresses")
	isTest, _ := cmd.Flags().GetBool("is-test")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doHaskellLemonOutboundMaterialTransactionExportsCreateOptions{
		BaseURL:         baseURL,
		Token:           token,
		JSON:            jsonOut,
		TransactionDate: transactionDate,
		ToAddresses:     toAddresses,
		CCAddresses:     ccAddresses,
		IsTest:          isTest,
	}, nil
}
