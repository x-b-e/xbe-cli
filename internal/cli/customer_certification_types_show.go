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

type customerCertificationTypesShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type customerCertificationTypeDetails struct {
	ID                  string `json:"id"`
	CustomerID          string `json:"customer_id,omitempty"`
	CertificationTypeID string `json:"certification_type_id,omitempty"`
}

func newCustomerCertificationTypesShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show customer certification type details",
		Long: `Show the full details of a customer certification type.

Output Fields:
  ID             Customer certification type identifier
  CUSTOMER ID    Customer ID
  CERT TYPE ID   Certification type ID

Arguments:
  <id>  The customer certification type ID (required). Use the list command to find IDs.`,
		Example: `  # Show a customer certification type
  xbe view customer-certification-types show 123

  # Show as JSON
  xbe view customer-certification-types show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runCustomerCertificationTypesShow,
	}
	initCustomerCertificationTypesShowFlags(cmd)
	return cmd
}

func init() {
	customerCertificationTypesCmd.AddCommand(newCustomerCertificationTypesShowCmd())
}

func initCustomerCertificationTypesShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runCustomerCertificationTypesShow(cmd *cobra.Command, args []string) error {
	opts, err := parseCustomerCertificationTypesShowOptions(cmd)
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
		return fmt.Errorf("customer certification type id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[customer-certification-types]", "customer,certification-type")

	body, _, err := client.Get(cmd.Context(), "/v1/customer-certification-types/"+id, query)
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

	details := buildCustomerCertificationTypeDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderCustomerCertificationTypeDetails(cmd, details)
}

func parseCustomerCertificationTypesShowOptions(cmd *cobra.Command) (customerCertificationTypesShowOptions, error) {
	jsonOut, err := cmd.Flags().GetBool("json")
	if err != nil {
		return customerCertificationTypesShowOptions{}, err
	}
	noAuth, err := cmd.Flags().GetBool("no-auth")
	if err != nil {
		return customerCertificationTypesShowOptions{}, err
	}
	baseURL, err := cmd.Flags().GetString("base-url")
	if err != nil {
		return customerCertificationTypesShowOptions{}, err
	}
	token, err := cmd.Flags().GetString("token")
	if err != nil {
		return customerCertificationTypesShowOptions{}, err
	}

	return customerCertificationTypesShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildCustomerCertificationTypeDetails(resp jsonAPISingleResponse) customerCertificationTypeDetails {
	details := customerCertificationTypeDetails{
		ID: resp.Data.ID,
	}

	if rel, ok := resp.Data.Relationships["customer"]; ok && rel.Data != nil {
		details.CustomerID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["certification-type"]; ok && rel.Data != nil {
		details.CertificationTypeID = rel.Data.ID
	}

	return details
}

func renderCustomerCertificationTypeDetails(cmd *cobra.Command, details customerCertificationTypeDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.CustomerID != "" {
		fmt.Fprintf(out, "Customer ID: %s\n", details.CustomerID)
	}
	if details.CertificationTypeID != "" {
		fmt.Fprintf(out, "Certification Type ID: %s\n", details.CertificationTypeID)
	}

	return nil
}
