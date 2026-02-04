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

type businessUnitCustomersShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type businessUnitCustomerDetails struct {
	ID               string `json:"id"`
	BusinessUnitID   string `json:"business_unit_id,omitempty"`
	BusinessUnitName string `json:"business_unit_name,omitempty"`
	CustomerID       string `json:"customer_id,omitempty"`
	CustomerName     string `json:"customer_name,omitempty"`
}

func newBusinessUnitCustomersShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show business unit customer details",
		Long: `Show the full details of a business unit customer link.

Output Fields:
  ID             Link identifier
  Business Unit  Linked business unit
  Customer       Linked customer

Arguments:
  <id>    Business unit customer ID (required).

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a business unit customer link
  xbe view business-unit-customers show 123

  # JSON output
  xbe view business-unit-customers show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runBusinessUnitCustomersShow,
	}
	initBusinessUnitCustomersShowFlags(cmd)
	return cmd
}

func init() {
	businessUnitCustomersCmd.AddCommand(newBusinessUnitCustomersShowCmd())
}

func initBusinessUnitCustomersShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runBusinessUnitCustomersShow(cmd *cobra.Command, args []string) error {
	if handled, err := maybeHandleClientURLShow(cmd, args); err != nil {
		return err
	} else if handled {
		return nil
	}

	opts, err := parseBusinessUnitCustomersShowOptions(cmd)
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
		return fmt.Errorf("business unit customer id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[business-unit-customers]", "business-unit,customer")
	query.Set("fields[business-units]", "company-name")
	query.Set("fields[customers]", "company-name")
	query.Set("include", "business-unit,customer")

	body, _, err := client.Get(cmd.Context(), "/v1/business-unit-customers/"+id, query)
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

	details := buildBusinessUnitCustomerDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderBusinessUnitCustomerDetails(cmd, details)
}

func parseBusinessUnitCustomersShowOptions(cmd *cobra.Command) (businessUnitCustomersShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return businessUnitCustomersShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildBusinessUnitCustomerDetails(resp jsonAPISingleResponse) businessUnitCustomerDetails {
	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	details := businessUnitCustomerDetails{
		ID: resp.Data.ID,
	}

	if rel, ok := resp.Data.Relationships["business-unit"]; ok && rel.Data != nil {
		details.BusinessUnitID = rel.Data.ID
		if unit, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.BusinessUnitName = strings.TrimSpace(stringAttr(unit.Attributes, "company-name"))
		}
	}

	if rel, ok := resp.Data.Relationships["customer"]; ok && rel.Data != nil {
		details.CustomerID = rel.Data.ID
		if customer, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.CustomerName = strings.TrimSpace(stringAttr(customer.Attributes, "company-name"))
		}
	}

	return details
}

func renderBusinessUnitCustomerDetails(cmd *cobra.Command, details businessUnitCustomerDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	writeLabelWithID(out, "Business Unit", details.BusinessUnitName, details.BusinessUnitID)
	writeLabelWithID(out, "Customer", details.CustomerName, details.CustomerID)

	return nil
}
