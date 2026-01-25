package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type brokersListOptions struct {
	BaseURL                                               string
	Token                                                 string
	JSON                                                  bool
	NoAuth                                                bool
	Limit                                                 int
	Offset                                                int
	CompanyName                                           string
	IsActive                                              string
	IsDefault                                             string
	SubDomain                                             string
	TrailerClassification                                 string
	QuickbooksEnabled                                     string
	CanCustomersSeeDriverContactInformation               string
	CanCustomerOperationsSeeDriverContactInformation      string
	HasHelpText                                           string
	SkipTenderJobScheduleShiftStartingSellerNotifications string
}

type brokerRow struct {
	ID          string `json:"id"`
	CompanyName string `json:"company_name"`
}

func newBrokersListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List brokers",
		Long: `List brokers with filtering and pagination.

Returns a list of brokers (branches) registered on the XBE platform.
Results are sorted alphabetically by company name.

Output Columns (table format):
  ID       Unique broker identifier (use with --broker-id in newsletter commands)
  COMPANY  Company/organization name

Pagination:
  Use --limit and --offset to paginate through large result sets.`,
		Example: `  # List all brokers
  xbe view brokers list

  # Search by company name (partial match)
  xbe view brokers list --company-name "Insurance"

  # List only active brokers
  xbe view brokers list --is-active true

  # Paginate results
  xbe view brokers list --limit 50 --offset 100

  # Get JSON output for scripting
  xbe view brokers list --json

  # Find broker ID for newsletter filtering
  xbe view brokers list --company-name "Acme" --json | jq '.[0].id'

  # List without authentication
  xbe view brokers list --no-auth`,
		RunE: runBrokersList,
	}
	initBrokersListFlags(cmd)
	return cmd
}

func init() {
	brokersCmd.AddCommand(newBrokersListCmd())
}

func initBrokersListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 0, "Page size (defaults to server default)")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("company-name", "", "Filter by company name (partial match)")
	cmd.Flags().String("is-active", "", "Filter by active status (true/false)")
	cmd.Flags().String("is-default", "", "Filter by default status (true/false)")
	cmd.Flags().String("sub-domain", "", "Filter by subdomain")
	cmd.Flags().String("trailer-classification", "", "Filter by trailer classification")
	cmd.Flags().String("quickbooks-enabled", "", "Filter by QuickBooks enabled status (true/false)")
	cmd.Flags().String("can-customers-see-driver-contact-information", "", "Filter by customer driver contact visibility (true/false)")
	cmd.Flags().String("can-customer-operations-see-driver-contact-information", "", "Filter by customer ops driver contact visibility (true/false)")
	cmd.Flags().String("has-help-text", "", "Filter by help text presence (true/false)")
	cmd.Flags().String("skip-tender-job-schedule-shift-starting-seller-notifications", "", "Filter by skip tender notifications (true/false)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runBrokersList(cmd *cobra.Command, _ []string) error {
	opts, err := parseBrokersListOptions(cmd)
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

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("sort", "company-name")
	query.Set("fields[brokers]", "company-name")
	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	setFilterIfPresent(query, "filter[company-name]", opts.CompanyName)
	setFilterIfPresent(query, "filter[is-active]", opts.IsActive)
	setFilterIfPresent(query, "filter[is-default]", opts.IsDefault)
	setFilterIfPresent(query, "filter[sub-domain]", opts.SubDomain)
	setFilterIfPresent(query, "filter[trailer-classification]", opts.TrailerClassification)
	setFilterIfPresent(query, "filter[quickbooks-enabled]", opts.QuickbooksEnabled)
	setFilterIfPresent(query, "filter[can-customers-see-driver-contact-information]", opts.CanCustomersSeeDriverContactInformation)
	setFilterIfPresent(query, "filter[can-customer-operations-see-driver-contact-information]", opts.CanCustomerOperationsSeeDriverContactInformation)
	setFilterIfPresent(query, "filter[has-help-text]", opts.HasHelpText)
	setFilterIfPresent(query, "filter[skip-tender-job-schedule-shift-starting-seller-notifications]", opts.SkipTenderJobScheduleShiftStartingSellerNotifications)

	body, _, err := client.Get(cmd.Context(), "/v1/brokers", query)
	if err != nil {
		if len(body) > 0 {
			fmt.Fprintln(cmd.ErrOrStderr(), string(body))
		}
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	var resp jsonAPIResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	handled, err := renderSparseListIfRequested(cmd, resp)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if handled {
		return nil
	}

	rows := buildBrokerRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderBrokersTable(cmd, rows)
}

func parseBrokersListOptions(cmd *cobra.Command) (brokersListOptions, error) {
	jsonOut, err := cmd.Flags().GetBool("json")
	if err != nil {
		return brokersListOptions{}, err
	}
	noAuth, err := cmd.Flags().GetBool("no-auth")
	if err != nil {
		return brokersListOptions{}, err
	}
	limit, err := cmd.Flags().GetInt("limit")
	if err != nil {
		return brokersListOptions{}, err
	}
	offset, err := cmd.Flags().GetInt("offset")
	if err != nil {
		return brokersListOptions{}, err
	}
	companyName, err := cmd.Flags().GetString("company-name")
	if err != nil {
		return brokersListOptions{}, err
	}
	isActive, err := cmd.Flags().GetString("is-active")
	if err != nil {
		return brokersListOptions{}, err
	}
	isDefault, err := cmd.Flags().GetString("is-default")
	if err != nil {
		return brokersListOptions{}, err
	}
	subDomain, err := cmd.Flags().GetString("sub-domain")
	if err != nil {
		return brokersListOptions{}, err
	}
	trailerClassification, err := cmd.Flags().GetString("trailer-classification")
	if err != nil {
		return brokersListOptions{}, err
	}
	quickbooksEnabled, err := cmd.Flags().GetString("quickbooks-enabled")
	if err != nil {
		return brokersListOptions{}, err
	}
	canCustomersSeeDriverContactInfo, err := cmd.Flags().GetString("can-customers-see-driver-contact-information")
	if err != nil {
		return brokersListOptions{}, err
	}
	canCustomerOpsSeeDriverContactInfo, err := cmd.Flags().GetString("can-customer-operations-see-driver-contact-information")
	if err != nil {
		return brokersListOptions{}, err
	}
	hasHelpText, err := cmd.Flags().GetString("has-help-text")
	if err != nil {
		return brokersListOptions{}, err
	}
	skipTenderNotifications, err := cmd.Flags().GetString("skip-tender-job-schedule-shift-starting-seller-notifications")
	if err != nil {
		return brokersListOptions{}, err
	}
	baseURL, err := cmd.Flags().GetString("base-url")
	if err != nil {
		return brokersListOptions{}, err
	}
	token, err := cmd.Flags().GetString("token")
	if err != nil {
		return brokersListOptions{}, err
	}

	return brokersListOptions{
		BaseURL:                                 baseURL,
		Token:                                   token,
		JSON:                                    jsonOut,
		NoAuth:                                  noAuth,
		Limit:                                   limit,
		Offset:                                  offset,
		CompanyName:                             companyName,
		IsActive:                                isActive,
		IsDefault:                               isDefault,
		SubDomain:                               subDomain,
		TrailerClassification:                   trailerClassification,
		QuickbooksEnabled:                       quickbooksEnabled,
		CanCustomersSeeDriverContactInformation: canCustomersSeeDriverContactInfo,
		CanCustomerOperationsSeeDriverContactInformation: canCustomerOpsSeeDriverContactInfo,
		HasHelpText: hasHelpText,
		SkipTenderJobScheduleShiftStartingSellerNotifications: skipTenderNotifications,
	}, nil
}

func buildBrokerRows(resp jsonAPIResponse) []brokerRow {
	rows := make([]brokerRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, brokerRow{
			ID:          resource.ID,
			CompanyName: strings.TrimSpace(stringAttr(resource.Attributes, "company-name")),
		})
	}

	return rows
}

func renderBrokersTable(cmd *cobra.Command, rows []brokerRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No brokers found.")
		return nil
	}

	const tableCompanyMax = 80

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, 32, 0)
	fmt.Fprintln(writer, "ID\tCOMPANY")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\n", row.ID, truncateString(row.CompanyName, tableCompanyMax))
	}
	return writer.Flush()
}
