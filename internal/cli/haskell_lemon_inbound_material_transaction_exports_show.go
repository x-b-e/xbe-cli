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

type haskellLemonInboundMaterialTransactionExportsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type haskellLemonInboundMaterialTransactionExportDetails struct {
	ID              string   `json:"id"`
	TransactionDate string   `json:"transaction_date,omitempty"`
	IsTest          bool     `json:"is_test,omitempty"`
	ToAddresses     []string `json:"to_addresses,omitempty"`
	CcAddresses     []string `json:"cc_addresses,omitempty"`
	CSV             string   `json:"csv,omitempty"`
	CreatedByID     string   `json:"created_by_id,omitempty"`
	CreatedByName   string   `json:"created_by_name,omitempty"`
	CreatedAt       string   `json:"created_at,omitempty"`
	UpdatedAt       string   `json:"updated_at,omitempty"`
}

func newHaskellLemonInboundMaterialTransactionExportsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show Haskell Lemon inbound material transaction export details",
		Long: `Show the full details of a Haskell Lemon inbound material transaction export.

Arguments:
  <id>  Export ID (required). Find IDs using the list command.`,
		Example: `  # Show an export
  xbe view haskell-lemon-inbound-material-transaction-exports show 123

  # Output as JSON
  xbe view haskell-lemon-inbound-material-transaction-exports show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runHaskellLemonInboundMaterialTransactionExportsShow,
	}
	initHaskellLemonInboundMaterialTransactionExportsShowFlags(cmd)
	return cmd
}

func init() {
	haskellLemonInboundMaterialTransactionExportsCmd.AddCommand(newHaskellLemonInboundMaterialTransactionExportsShowCmd())
}

func initHaskellLemonInboundMaterialTransactionExportsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runHaskellLemonInboundMaterialTransactionExportsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseHaskellLemonInboundMaterialTransactionExportsShowOptions(cmd)
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
	query.Set("fields[haskell-lemon-inbound-material-transaction-exports]", strings.Join([]string{
		"transaction-date",
		"to-addresses",
		"cc-addresses",
		"is-test",
		"csv",
		"created-by",
		"created-at",
		"updated-at",
	}, ","))
	query.Set("include", "created-by")
	query.Set("fields[users]", "name")

	body, _, err := client.Get(cmd.Context(), "/v1/haskell-lemon-inbound-material-transaction-exports/"+id, query)
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

	details := buildHaskellLemonInboundMaterialTransactionExportDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderHaskellLemonInboundMaterialTransactionExportDetails(cmd, details)
}

func parseHaskellLemonInboundMaterialTransactionExportsShowOptions(cmd *cobra.Command) (haskellLemonInboundMaterialTransactionExportsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return haskellLemonInboundMaterialTransactionExportsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildHaskellLemonInboundMaterialTransactionExportDetails(resp jsonAPISingleResponse) haskellLemonInboundMaterialTransactionExportDetails {
	resource := resp.Data
	attrs := resource.Attributes

	details := haskellLemonInboundMaterialTransactionExportDetails{
		ID:              resource.ID,
		TransactionDate: formatDate(stringAttr(attrs, "transaction-date")),
		IsTest:          boolAttr(attrs, "is-test"),
		ToAddresses:     stringSliceAttr(attrs, "to-addresses"),
		CcAddresses:     stringSliceAttr(attrs, "cc-addresses"),
		CSV:             stringAttr(attrs, "csv"),
		CreatedAt:       formatDateTime(stringAttr(attrs, "created-at")),
		UpdatedAt:       formatDateTime(stringAttr(attrs, "updated-at")),
	}

	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	if rel, ok := resource.Relationships["created-by"]; ok && rel.Data != nil {
		details.CreatedByID = rel.Data.ID
		if inc, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.CreatedByName = strings.TrimSpace(stringAttr(inc.Attributes, "name"))
		}
	}

	return details
}

func renderHaskellLemonInboundMaterialTransactionExportDetails(cmd *cobra.Command, details haskellLemonInboundMaterialTransactionExportDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	fmt.Fprintf(out, "Transaction Date: %s\n", details.TransactionDate)
	fmt.Fprintf(out, "Is Test: %t\n", details.IsTest)
	if details.CreatedByID != "" || details.CreatedByName != "" {
		fmt.Fprintf(out, "Created By: %s\n", formatRelated(details.CreatedByName, details.CreatedByID))
	}
	if details.CreatedAt != "" {
		fmt.Fprintf(out, "Created At: %s\n", details.CreatedAt)
	}
	if details.UpdatedAt != "" {
		fmt.Fprintf(out, "Updated At: %s\n", details.UpdatedAt)
	}
	if len(details.ToAddresses) > 0 {
		fmt.Fprintf(out, "To Addresses: %s\n", strings.Join(details.ToAddresses, ", "))
	}
	if len(details.CcAddresses) > 0 {
		fmt.Fprintf(out, "CC Addresses: %s\n", strings.Join(details.CcAddresses, ", "))
	}

	if strings.TrimSpace(details.CSV) != "" {
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "CSV:")
		fmt.Fprintln(out, details.CSV)
	}

	return nil
}
