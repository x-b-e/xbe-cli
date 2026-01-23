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

type doInvoiceSendsCreateOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	Invoice string
	Comment string
}

func newDoInvoiceSendsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Send an invoice",
		Long: `Send an invoice.

Invoices must be in editing status.

Required flags:
  --invoice   Invoice ID

Optional flags:
  --comment   Send comment

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Send an invoice with a comment
  xbe do invoice-sends create \
    --invoice 123 \
    --comment "Sent to customer"

  # Send an invoice without a comment
  xbe do invoice-sends create --invoice 123`,
		Args: cobra.NoArgs,
		RunE: runDoInvoiceSendsCreate,
	}
	initDoInvoiceSendsCreateFlags(cmd)
	return cmd
}

func init() {
	doInvoiceSendsCmd.AddCommand(newDoInvoiceSendsCreateCmd())
}

func initDoInvoiceSendsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("invoice", "", "Invoice ID")
	cmd.Flags().String("comment", "", "Send comment")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoInvoiceSendsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoInvoiceSendsCreateOptions(cmd)
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
			"type":          "invoice-sends",
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

	body, _, err := client.Post(cmd.Context(), "/v1/invoice-sends", jsonBody)
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

	row := buildInvoiceSendRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created invoice send %s\n", row.ID)
	return nil
}

func parseDoInvoiceSendsCreateOptions(cmd *cobra.Command) (doInvoiceSendsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	invoice, _ := cmd.Flags().GetString("invoice")
	comment, _ := cmd.Flags().GetString("comment")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doInvoiceSendsCreateOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		Invoice: invoice,
		Comment: comment,
	}, nil
}
