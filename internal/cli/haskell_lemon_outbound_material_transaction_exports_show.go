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

type haskellLemonOutboundMaterialTransactionExportsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type haskellLemonOutboundMaterialTransactionExportDetails struct {
	ID              string   `json:"id"`
	TransactionDate string   `json:"transaction_date,omitempty"`
	ToAddresses     []string `json:"to_addresses,omitempty"`
	CCAddresses     []string `json:"cc_addresses,omitempty"`
	IsTest          bool     `json:"is_test"`
	CSV             string   `json:"csv,omitempty"`
	CreatedByID     string   `json:"created_by_id,omitempty"`
}

func newHaskellLemonOutboundMaterialTransactionExportsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show Haskell Lemon outbound material transaction export details",
		Long: `Show the full details of a Haskell Lemon outbound material transaction export.

Output Fields:
  ID
  Transaction Date
  Is Test
  To Addresses
  CC Addresses
  Created By
  CSV

Arguments:
  <id>    The export ID (required). You can find IDs using the list command.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show an export
  xbe view haskell-lemon-outbound-material-transaction-exports show 123

  # JSON output
  xbe view haskell-lemon-outbound-material-transaction-exports show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runHaskellLemonOutboundMaterialTransactionExportsShow,
	}
	initHaskellLemonOutboundMaterialTransactionExportsShowFlags(cmd)
	return cmd
}

func init() {
	haskellLemonOutboundMaterialTransactionExportsCmd.AddCommand(newHaskellLemonOutboundMaterialTransactionExportsShowCmd())
}

func initHaskellLemonOutboundMaterialTransactionExportsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runHaskellLemonOutboundMaterialTransactionExportsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseHaskellLemonOutboundMaterialTransactionExportsShowOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.NoAuth {
		opts.Token = ""
	} else if strings.TrimSpace(opts.Token) == "" {
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

	id := strings.TrimSpace(args[0])
	if id == "" {
		return fmt.Errorf("export id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[haskell-lemon-outbound-material-transaction-exports]", "transaction-date,to-addresses,cc-addresses,is-test,csv,created-by")

	body, _, err := client.Get(cmd.Context(), "/v1/haskell-lemon-outbound-material-transaction-exports/"+id, query)
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

	details := buildHaskellLemonOutboundMaterialTransactionExportDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderHaskellLemonOutboundMaterialTransactionExportDetails(cmd, details)
}

func parseHaskellLemonOutboundMaterialTransactionExportsShowOptions(cmd *cobra.Command) (haskellLemonOutboundMaterialTransactionExportsShowOptions, error) {
	jsonOut, err := cmd.Flags().GetBool("json")
	if err != nil {
		return haskellLemonOutboundMaterialTransactionExportsShowOptions{}, err
	}
	noAuth, err := cmd.Flags().GetBool("no-auth")
	if err != nil {
		return haskellLemonOutboundMaterialTransactionExportsShowOptions{}, err
	}
	baseURL, err := cmd.Flags().GetString("base-url")
	if err != nil {
		return haskellLemonOutboundMaterialTransactionExportsShowOptions{}, err
	}
	token, err := cmd.Flags().GetString("token")
	if err != nil {
		return haskellLemonOutboundMaterialTransactionExportsShowOptions{}, err
	}

	return haskellLemonOutboundMaterialTransactionExportsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildHaskellLemonOutboundMaterialTransactionExportDetails(resp jsonAPISingleResponse) haskellLemonOutboundMaterialTransactionExportDetails {
	resource := resp.Data
	attrs := resource.Attributes
	return haskellLemonOutboundMaterialTransactionExportDetails{
		ID:              resource.ID,
		TransactionDate: formatDate(stringAttr(attrs, "transaction-date")),
		ToAddresses:     stringSliceAttr(attrs, "to-addresses"),
		CCAddresses:     stringSliceAttr(attrs, "cc-addresses"),
		IsTest:          boolAttr(attrs, "is-test"),
		CSV:             stringAttr(attrs, "csv"),
		CreatedByID:     relationshipIDFromMap(resource.Relationships, "created-by"),
	}
}

func renderHaskellLemonOutboundMaterialTransactionExportDetails(cmd *cobra.Command, details haskellLemonOutboundMaterialTransactionExportDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.TransactionDate != "" {
		fmt.Fprintf(out, "Transaction Date: %s\n", details.TransactionDate)
	}
	fmt.Fprintf(out, "Is Test: %t\n", details.IsTest)
	if len(details.ToAddresses) > 0 {
		fmt.Fprintf(out, "To Addresses: %s\n", strings.Join(details.ToAddresses, ", "))
	}
	if len(details.CCAddresses) > 0 {
		fmt.Fprintf(out, "CC Addresses: %s\n", strings.Join(details.CCAddresses, ", "))
	}
	if details.CreatedByID != "" {
		fmt.Fprintf(out, "Created By: %s\n", details.CreatedByID)
	}
	if details.CSV != "" {
		fmt.Fprintln(out, "\nCSV:")
		fmt.Fprintln(out, details.CSV)
	}

	return nil
}
