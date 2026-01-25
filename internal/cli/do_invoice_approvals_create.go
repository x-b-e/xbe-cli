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

type doInvoiceApprovalsCreateOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	Invoice string
	Comment string
}

func newDoInvoiceApprovalsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Approve an invoice",
		Long: `Approve an invoice.

Invoices must be in sent or batched status.

Required flags:
  --invoice   Invoice ID

Optional flags:
  --comment   Approval comment

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Approve an invoice with a comment
  xbe do invoice-approvals create \
    --invoice 123 \
    --comment "Approved"

  # Approve an invoice without a comment
  xbe do invoice-approvals create --invoice 123`,
		Args: cobra.NoArgs,
		RunE: runDoInvoiceApprovalsCreate,
	}
	initDoInvoiceApprovalsCreateFlags(cmd)
	return cmd
}

func init() {
	doInvoiceApprovalsCmd.AddCommand(newDoInvoiceApprovalsCreateCmd())
}

func initDoInvoiceApprovalsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("invoice", "", "Invoice ID")
	cmd.Flags().String("comment", "", "Approval comment")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoInvoiceApprovalsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoInvoiceApprovalsCreateOptions(cmd)
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

	if strings.TrimSpace(opts.Invoice) == "" {
		err := fmt.Errorf("--invoice is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{}
	if opts.Comment != "" {
		attributes["comment"] = opts.Comment
	}

	relationships := map[string]any{
		"invoice": map[string]any{
			"data": map[string]any{
				"type": "invoices",
				"id":   opts.Invoice,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "invoice-approvals",
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

	body, _, err := client.Post(cmd.Context(), "/v1/invoice-approvals", jsonBody)
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

	row := buildInvoiceApprovalRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created invoice approval %s\n", row.ID)
	return nil
}

func parseDoInvoiceApprovalsCreateOptions(cmd *cobra.Command) (doInvoiceApprovalsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	invoice, _ := cmd.Flags().GetString("invoice")
	comment, _ := cmd.Flags().GetString("comment")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doInvoiceApprovalsCreateOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		Invoice: invoice,
		Comment: comment,
	}, nil
}
